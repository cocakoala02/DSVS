package protocols

import (
	"errors"
	"fmt"
	"sync"
	"time"

	libdrynx "github.com/ldsec/drynx/lib"
	libdrynxencoding "github.com/ldsec/drynx/lib/encoding"
	drynxproof "github.com/ldsec/drynx/lib/proof"
	libdrynxrange "github.com/ldsec/drynx/lib/range"
	dataunlynx "github.com/ldsec/unlynx/data"
	libunlynx "github.com/ldsec/unlynx/lib"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
)

// 该文件实现了 Drynx 中的数据收集协议（DataCollectionProtocol），
// 主要用于数据提供者（DP）在接收到查询后生成、编码、加密本地数据，
// 并将结果上传到根节点，从而支持分布式数据聚合。协议基于 onet 框架实现，
// 利用树形通信结构协调各个节点的操作。

// DataCollectionProtocolName is the registered name for the data provider protocol.
const DataCollectionProtocolName = "DataCollection"

var mutexGroups sync.Mutex //互斥锁

func init() {
	network.RegisterMessage(AnnouncementDCMessage{}) //用于根节点向其他节点广播查询启动信号。
	network.RegisterMessage(DataCollectionMessage{}) //用于数据提供者向根节点发送数据响应。
	//将协议名称 DataCollectionProtocolName（字符串 "DataCollection"）与构造函数 NewDataCollectionProtocol 关联起来
	if _, err := onet.GlobalProtocolRegister(DataCollectionProtocolName, NewDataCollectionProtocol); err != nil {
		log.Fatal("Error registering <DataCollectionProtocol>:", err)
	}
}

// Messages
// ______________________________________________________________________________________________________________________
// 用作触发数据收集协议的通知消息
// AnnouncementDCMessage message sent (with the query) to trigger a data collection protocol.
type AnnouncementDCMessage struct{}

// 是用来封装数据提供者（DP）生成的数据响应的消息
// DataCollectionMessage message that contains the data of each data provider
type DataCollectionMessage struct {
	DCMdata libdrynx.ResponseDPBytes //ResponseDPBytes 一般包含一个 map（键值对），其键为数据分组的标识，值为该组加密数据的字节数组，以及可能的数据长度信息。
}

// Structs
// ______________________________________________________________________________________________________________________
// 这个结构体用于将查询信息发送到数据提供者（DP），触发它们上传数据
// SurveyToDP is used to trigger the upload of data by a data provider
type SurveyToDP struct {
	SurveyID  string
	Aggregate kyber.Point // the joint aggregate key to encrypt the data

	// query statement
	Query libdrynx.Query // the query must be added to each node before the protocol can start
}

// 这个结构体是对 AnnouncementDCMessage 的包装，用于在数据收集协议中广播启动信号，同时携带 onet.TreeNode 的上下文信息。
// AnnouncementDCStruct announcement message
type AnnouncementDCStruct struct {
	//为消息提供了当前节点在协议树中的上下文，使得消息在分布式通信中可以携带额外的信息（例如节点身份、层次关系等）
	*onet.TreeNode
	AnnouncementDCMessage
}

// 这个结构体是对 DataCollectionMessage 的包装，用于在协议中传输数据提供者上传的数据，同时包含节点上下文信息。
// DataCollectionStruct is the wrapper of DataCollectionMessage to be used in a channel
type DataCollectionStruct struct {
	*onet.TreeNode
	DataCollectionMessage
}

// Protocol
//______________________________________________________________________________________________________________________

// DataCollectionProtocol hold the state of a data provider protocol instance.
type DataCollectionProtocol struct {
	*onet.TreeNodeInstance

	// Protocol feedback channel
	FeedbackChannel chan map[string]libunlynx.CipherVector //map containing the aggregation of all data providers' responses

	// Protocol communication channels
	AnnouncementChannel   chan AnnouncementDCStruct //用于传输 AnnouncementDCStruct 消息。
	DataCollectionChannel chan DataCollectionStruct

	// Protocol state data
	Survey SurveyToDP

	// Protocol proof data
	MapPIs map[string]onet.ProtocolInstance
}

// 用于构造并初始化一个 DataCollectionProtocol 实例
// NewDataCollectionProtocol constructs a DataCollection protocol instance.
func NewDataCollectionProtocol(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	dcp := &DataCollectionProtocol{
		TreeNodeInstance: n,
		FeedbackChannel:  make(chan map[string]libunlynx.CipherVector),
	}

	err := dcp.RegisterChannel(&dcp.AnnouncementChannel)
	if err != nil {
		return nil, errors.New("couldn't register data reference channel: " + err.Error())
	}

	err = dcp.RegisterChannel(&dcp.DataCollectionChannel)
	if err != nil {
		return nil, errors.New("couldn't register data reference channel: " + err.Error())
	}

	return dcp, nil
}

// Start is called at the root node and starts the execution of the protocol.
func (p *DataCollectionProtocol) Start() error {
	log.Lvl2("["+p.Name()+"]", "starts a Data Collection Protocol.")

	for _, node := range p.Tree().List() {
		// the root node sends an announcement message to all the nodes
		if !node.IsRoot() {
			if err := p.SendTo(node, &AnnouncementDCMessage{}); err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
}

// 用于协调各个节点上传数据并在根节点进行聚合。
// Dispatch is called on each tree node. It waits for incoming messages and handles them.
func (p *DataCollectionProtocol) Dispatch() error {
	defer p.Done()

	// 1. If not root(即数据提供者) -> wait for announcement message from root
	if !p.IsRoot() {
		//调用 p.GenerateData() 生成本地数据。该方法负责根据查询内容生成并编码、加密本地数据，
		//返回一个类型为 libdrynx.ResponseDPBytes 的数据结构
		response, err := p.GenerateData()
		if err != nil {
			log.Fatal(err)
		}

		dcm := DataCollectionMessage{DCMdata: response}
		//通过调用 p.SendTo(p.Root(), &dcm) 将封装好的数据消息发送给协议树的根节点。
		// 2. Send data to root
		if err := p.SendTo(p.Root(), &dcm); err != nil {
			return err
		}
	} else {
		// 3. If root wait for all other nodes to send their data
		dcmAggregate := make(map[string]libunlynx.CipherVector, 0)
		for i := 0; i < len(p.Tree().List())-1; i++ {
			dcm := <-p.DataCollectionChannel
			dcmData := dcm.DCMdata

			// 对于每个接收到的 DataCollectionMessage：
			// dcmData.Data 是一个 map，其值是字节数组（原始加密数据）。
			// 遍历该 map，将每个字节数组使用 cv.FromBytes(v, dcmData.Len) 解码为一个 CipherVector，并存储到新的 map dcmDecoded 中。
			// 这里 dcmData.Len 表示原始 CipherVector 的长度。

			// received map with bytes -> go back to map with CipherVector
			dcmDecoded := make(map[string]libunlynx.CipherVector, len(dcmData.Data))
			for i, v := range dcmData.Data {
				cv := libunlynx.NewCipherVector(dcmData.Len)
				cv.FromBytes(v, dcmData.Len)
				dcmDecoded[i] = *cv
			}
			// 对解码后的数据 dcmDecoded 进行聚合：
			// 对于每个组（key）：
			// 如果在 dcmAggregate 中已经存在该组的数据，则将当前收到的 CipherVector 与已有的 CipherVector 相加（调用 newCV.Add(cv, value)），更新该组的聚合数据。
			// 如果该组数据尚不存在，则直接将当前值存入 dcmAggregate。
			// aggregate values that belong to the same group (that are originated from different data providers)
			for key, value := range dcmDecoded {
				// if already in the map -> add to what is inside
				if cv, ok := dcmAggregate[key]; ok {
					//// 如果在 dcmAggregate 中找到 key，则 ok 为 true，此时 cv 保存对应的值
					newCV := libunlynx.NewCipherVector(len(cv))
					newCV.Add(cv, value)
					dcmAggregate[key] = *newCV
				} else { // otherwise create a new entry
					dcmAggregate[key] = value
				}
			}
		}
		p.FeedbackChannel <- dcmAggregate
	}
	return nil
}

// Support Functions
//______________________________________________________________________________________________________________________

// GenerateData is used to generate data at DPs, this is more for simulation's purposes
func (p *DataCollectionProtocol) GenerateData() (libdrynx.ResponseDPBytes, error) {

	// read the signatures needed to compute the range proofs
	signatures := make([][]libdrynx.PublishSignature, p.Survey.Query.IVSigs.InputValidationSize1)
	for i := 0; i < p.Survey.Query.IVSigs.InputValidationSize1; i++ {
		signatures[i] = make([]libdrynx.PublishSignature, p.Survey.Query.IVSigs.InputValidationSize2)
		for j := 0; j < p.Survey.Query.IVSigs.InputValidationSize2; j++ {
			signatures[i][j] = libdrynxrange.PublishSignatureBytesToPublishSignatures((*p.Survey.Query.IVSigs.InputValidationSigs[i])[j])
		}
	}

	var clm []int64

	var err error
	clm, err = libdrynxencoding.GetDataFromDataProvider(p.Name(), p.Survey.Query.DptoPath, p.Survey.Query.TableName, p.Survey.Query.SQL, p.Survey.Query.Operation.NameOp)
	if err != nil {
		return libdrynx.ResponseDPBytes{}, fmt.Errorf("when getting data from provider: %w", err)
	}

	// ------- START: ENCODING & ENCRYPTION -------
	encodeTime := libunlynx.StartTimer(p.Name() + "_DPencoding")
	cprf := make([]libdrynxrange.CreateProof, 0)

	// compute response
	queryResponse := make(map[string]libunlynx.CipherVector, 0)
	clearResponse := make([]int64, 0)
	encryptedResponse := make([]libunlynx.CipherText, 0)

	encryptedResponse, clearResponse, cprf = libdrynxencoding.Encode(clm, p.Survey.Aggregate, signatures, p.Survey.Query.Ranges, p.Survey.Query.Operation)

	log.Lvl2("Data Provider", p.Name(), "computes the query response", clearResponse, "with operation:", p.Survey.Query.Operation)

	queryResponse["0"] = libunlynx.CipherVector(encryptedResponse)

	//修改
	if p.Survey.Query.Proofs != 0 {
		go func() {
			startAllProofs := libunlynx.StartTimer(p.Name() + "_AllProofs")
			rpl := libdrynxrange.RangeProofList{}

			//rangeProofCreation := libunlynx.StartTimer(p.Name() + "_RangeProofCreation")
			rangeProofCreation := time.Now() //计时器

			// no range proofs (send only the ciphertexts)
			if len(cprf) == 0 {
				tmp := make([]libdrynxrange.RangeProof, 0)
				for _, ct := range queryResponse["0"] {
					tmp = append(tmp, libdrynxrange.RangeProof{Commit: ct, RP: nil})
				}
				rpl = libdrynxrange.RangeProofList{Data: tmp}
			} else { // if range proofs
				rpl = libdrynxrange.RangeProofList{Data: libdrynxrange.CreatePredicateRangeProofListForAllServers(cprf)}
			}
			// scaling for simulation purposes
			if p.Survey.Query.CuttingFactor != 0 {
				rplNew := libdrynxrange.RangeProofList{}
				rplNew.Data = make([]libdrynxrange.RangeProof, len(rpl.Data)*p.Survey.Query.CuttingFactor)
				counter := 0
				suitePair := bn256.NewSuite()
				for j := 0; j < p.Survey.Query.CuttingFactor; j++ {
					for _, v := range rpl.Data {

						rplNew.Data[counter].RP = &libdrynxrange.RangeProofData{}
						rplNew.Data[counter].RP.V = make([][]kyber.Point, len(v.RP.V))
						for k, w := range v.RP.V {
							rplNew.Data[counter].RP.V[k] = make([]kyber.Point, len(w))
							for l, x := range w {
								tmp := suitePair.G2().Point().Null()
								tmp.Add(tmp, x)
								rplNew.Data[counter].RP.V[k][l] = tmp
							}
						}
						//rplNew.Data[counter].RP.V = tmp.Add(tmp,v.RP.V)
						rplNew.Data[counter].RP.Zv = v.RP.Zv
						rplNew.Data[counter].RP.Zr = v.RP.Zr
						rplNew.Data[counter].RP.Challenge = v.RP.Challenge
						rplNew.Data[counter].RP.D = v.RP.D
						rplNew.Data[counter].RP.Zphi = v.RP.Zphi
						rplNew.Data[counter].RP.A = v.RP.A
						//rplNew.Data[counter].RP. = &newRpd
						rplNew.Data[counter].Commit = v.Commit
						counter = counter + 1
					}
				}

				rpl.Data = rplNew.Data
			}

			pi := p.MapPIs["range/"+p.ServerIdentity().String()]
			pi.(*ProofCollectionProtocol).Proof = drynxproof.ProofRequest{RangeProof: drynxproof.NewRangeProofRequest(&rpl, p.Survey.SurveyID, p.ServerIdentity().String(), "", p.Survey.Query.RosterVNs, p.Private(), nil)}
			//libunlynx.EndTimer(rangeProofCreation)

			go func() {
				if err := pi.Dispatch(); err != nil {
					log.Fatal(err)
				}
			}()
			go func() {
				if err := pi.Start(); err != nil {
					log.Fatal(err)
				}
			}()
			<-pi.(*ProofCollectionProtocol).FeedbackChannel
			log.Lvl1("rangeProofCreationAndVerfy:", time.Since(rangeProofCreation)) //计时器
			libunlynx.EndTimer(startAllProofs)

		}()
	}

	libunlynx.EndTimer(encodeTime)
	// ------- END -------

	//convert the response to bytes
	length := len(queryResponse)
	queryResponseBytes := make(map[string][]byte, length)
	lenQueryResponse := 0
	wg := libunlynx.StartParallelize(length)
	mutex := sync.Mutex{}
	for i, v := range queryResponse {
		go func(group string, cv libunlynx.CipherVector) {
			defer wg.Done()
			cvBytes, lenQ, _ := cv.ToBytes()

			mutex.Lock()
			lenQueryResponse = lenQ
			queryResponseBytes[group] = cvBytes
			mutex.Unlock()
		}(i, v)
	}
	libunlynx.EndParallelize(wg)

	return libdrynx.ResponseDPBytes{Data: queryResponseBytes, Len: lenQueryResponse}, nil
}

// createFakeDataForOperation creates fake data to be used
func createFakeDataForOperation(operation libdrynx.Operation, nbrRows, min, max int64) [][]int64 {
	//either use the min and max defined by the query or the default constants
	zero := int64(0)
	if min == zero && max == zero {
		log.Lvl2("Only generating 0s!")
	}

	//generate response tab
	tab := make([][]int64, operation.NbrInput)
	wg := libunlynx.StartParallelize(len(tab))
	for i := range tab {
		go func(i int) {
			defer wg.Done()
			tab[i] = dataunlynx.CreateInt64Slice(nbrRows, min, max)
		}(i)

	}
	libunlynx.EndParallelize(wg)
	return tab
}

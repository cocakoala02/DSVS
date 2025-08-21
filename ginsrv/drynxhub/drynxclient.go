package drynxhub

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ldsec/drynx/ginsrv/datastruct"
	libdrynx "github.com/ldsec/drynx/lib"
	libdrynxrange "github.com/ldsec/drynx/lib/range"
	drynx_services "github.com/ldsec/drynx/services"
	libunlynx "github.com/ldsec/unlynx/lib"
	uuid "github.com/satori/go.uuid"
	"go.dedis.ch/cothority/v3/skipchain"
	"go.dedis.ch/kyber/v3"
	kyber_util_encoding "go.dedis.ch/kyber/v3/util/encoding"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/network"
	onet_network "go.dedis.ch/onet/v3/network"
)

// global value, maybe ok
var config datastruct.Config

func init() {
	path := LoadToml()
	LoadConfig(path)
}

// to load the config which has the roster of cn,vn,dp
func LoadConfig(path string) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("载入数据错误")
		return
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&config); err != nil {
		fmt.Println("解析数据错误")
		return
	}

	// TriRoster = datastruct.TriRoster{
	// 	Total: len(config.CNs) + len(config.VNs) + len(config.DPs),
	// 	CNs:   ConfigToRoster(config.CNs),
	// 	VNs:   ConfigToRoster(config.VNs),
	// 	DPs:   ConfigToRoster(config.DPs),
	// }
}

func SurveyRun(req *datastruct.TriSurReq) (res float64, valres string, err error) {

	// 2. 生成查询语句
	client := drynx_services.NewDrynxClient(&onet_network.ServerIdentity{URL: "http://" + config.Client}, "ClientOfSurvey")
	sq, err := GenSQ(req, client)
	if err != nil {
		return 0, "cuowu", err
	}

	// 3. 开启验证过程
	var wg *sync.WaitGroup
	var block *skipchain.SkipBlock
	if sq.Query.Proofs != 0 {
		// send query to the skipchain and 'wait' for all proofs' verification to be done
		clientSkip := drynx_services.NewDrynxClient(sq.Query.RosterVNs.List[0], "simul-skip-"+sq.Query.Operation.NameOp)

		wg = libunlynx.StartParallelize(1)
		go func(elVNs *onet.Roster) {
			defer wg.Done()

			err = clientSkip.SendSurveyQueryToVNs(elVNs, &sq)
			if err != nil {
				log.Fatal("Error sending query to VNs:", err)
			}
		}(sq.Query.RosterVNs)
		libunlynx.EndParallelize(wg)

		wg = libunlynx.StartParallelize(1)
		go func(si *network.ServerIdentity) {
			defer wg.Done()

			block, err = clientSkip.SendEndVerification(si, sq.SurveyID)
			if err != nil {
				log.Fatal("Error starting the 'waiting' threads:", err)
			}
		}(sq.Query.RosterVNs.List[0])
	}

	// 4. 查询与结果聚合
	_, aggregations, err := client.SendSurveyQuery(sq)
	if err != nil {
		return res, "cuowu", err
	}
	result, ok := float64(0), false
	for _, a := range *aggregations {
		if len(a) != 1 {
			return res, "cuowu", errors.New("line in aggregation larger than one, dunno how to print")
		}
		if ok && result != a[0] {
			return res, "cuowu", errors.New("not same value found in aggregation, dunno how to print")
		}
		result = a[0]
		ok = true
	}

	var validres string
	if sq.Query.Proofs != 0 {
		// 5. 结束验证
		var flag bool
		if len(sq.Query.RosterVNs.List) > 0 {
			clientSkip := drynx_services.NewDrynxClient(sq.Query.RosterVNs.List[0], "simul-skip-"+sq.Query.Operation.NameOp)

			if sq.Query.Proofs != 0 {
				libunlynx.EndParallelize(wg)
				// close DB
				// if err := clientSkip.SendCloseDB(sq.Query.RosterVNs, &libdrynx.CloseDB{Close: 0}); err != nil {
				// 	log.Fatal("Error closing the DB:", err)
				// }
			}

			retrieveBlock := time.Now()
			sb, err := clientSkip.SendGetLatestBlock(sq.Query.RosterVNs, block)
			if err != nil || sb == nil {
				log.Fatal("Something wrong when fetching the last block")
			}
			log.Println(time.Since(retrieveBlock))

			//或者反序列化为具体数据结构
			_, msg, err := network.Unmarshal(sb.Data, libunlynx.SuiTe)
			if err != nil {
				log.Fatal("Error unmarshaling block data:", err)
			}
			fmt.Printf("msg 类型是：%T\n", msg)

			fmt.Printf("Block Data: %+v\n", msg)

			// 类型断言
			dataBlock, ok := msg.(*libdrynx.DataBlock)
			if !ok {
				log.Fatal("msg 不是 *libdrynx.DataBlock 类型")
			}
			fmt.Printf("  SurveyID:     %v\n", dataBlock.SurveyID)
			fmt.Printf("  Sample:       %v\n", dataBlock.Sample)
			fmt.Printf("  Time:         %v\n", dataBlock.Time)
			fmt.Printf("  ServerNumber: %v\n", dataBlock.ServerNumber)

			fmt.Println("  Proofs:")
			flag = true
			for k, v := range dataBlock.Proofs {
				if v == 0 {
					flag = false
				}
				fmt.Printf("    %s => %d\n", k, v)
			}
		}

		if flag {
			validres = "验证结果通过"
		} else {
			validres = "验证结果未通过"
		}
	}

	fmt.Println(result)

	return result, validres, nil
}

// set the params of GenerateSurvey
func GenSQ(req *datastruct.TriSurReq, client *drynx_services.API) (resp libdrynx.SurveyQuery, err error) {
	// 1. network
	network, err := MakeNetWork()
	if err != nil {
		return libdrynx.SurveyQuery{}, err
	}

	// 2. surveyID
	surveyID := uuid.NewV4().String()
	// 3. sql and Operation
	sql := req.Sql
	s := strings.ToUpper(strings.TrimSpace(sql))
	var Operation string
	switch {
	case strings.HasPrefix(s, "SELECT SUM"):
		Operation = "sum"
	case strings.HasPrefix(s, "SELECT AVG"):
		Operation = "mean"
	case strings.HasPrefix(s, "SELECT COUNT"):
		Operation = "count"
	case strings.HasPrefix(s, "SELECT MAX"):
		Operation = "max"
	case strings.HasPrefix(s, "SELECT MIN"):
		Operation = "min"
	}

	// 4. operation
	operation := libdrynx.ChooseOperation(Operation, 1, 9, 5, config.CuttingFactor)

	// 5. range for each output of operation
	var ranges []*[]int64
	if Operation == "sum" {
		ranges = append(ranges, &[]int64{500, 1})

	} else if Operation == "mean" {
		ranges = append(ranges, &[]int64{500, 1})
		ranges = append(ranges, &[]int64{101, 1})
	} else if Operation == "count" {
		ranges = append(ranges, &[]int64{101, 1})
	} else if Operation == "max" {
		for i := 0; i < 9; i++ {
			ranges = append(ranges, &[]int64{2, 1})
		}
	} else if Operation == "min" {
		for i := 0; i < 9; i++ {
			ranges = append(ranges, &[]int64{2, 1})
		}
	}

	// 6. signature of range validity
	signature := SignatureOfRanges(ranges, len(network.CNs.List))

	// 7. 0 == no proof, 1 == proof, 2 == optimized proof
	proofs := 1

	// 8. obfuscation, I don't know what it is, so set false
	obfuscation := false

	// 9. threshold of somethings...
	thresholdEntityProofsVerif := []float64{
		1.0, // threshold general
		1.0, // threshold aggregation
		1.0, // threshold range
		0.0, // obfuscation
		1.0} // threshold key switch

	// 10. differential privacy,maybe something like these need to call at config
	diffP := libdrynx.QueryDiffP{LapMean: 0.0, LapScale: 0.0, NoiseListSize: 0, Quanta: 0.0, Scale: 0}

	// 11. how to group by
	dataSet := libdrynx.QueryDPDataGen{GroupByValues: []int64{1}, GenerateRows: 1, GenerateDataMin: int64(100), GenerateDataMax: int64(1000)}

	// 12
	dptopath := map[string]string{"tcp://127.0.0.1:7013": "./data",
		"tcp://127.0.0.1:7015": "./data",
		"tcp://127.0.0.1:7017": "./data",
		"tcp://127.0.0.1:7019": "./data",
		"tcp://127.0.0.1:7021": "./data",
		"tcp://127.0.0.1:7023": "./data",
		"tcp://127.0.0.1:7025": "./data",
		"tcp://127.0.0.1:7027": "./data",
		"tcp://127.0.0.1:7029": "./data",
	}
	//13
	tablename := "statistics_experiment_data"

	return client.GenerateSurveyQuery(network.CNs, network.VNs, network.CnToDPs, network.IdToPub, surveyID, operation, ranges, signature, proofs, obfuscation, thresholdEntityProofsVerif, diffP, dataSet, config.CuttingFactor, sql, dptopath, tablename), nil
}

// signatures for Input Validation
func SignatureOfRanges(ranges []*[]int64, nbrserver int) []*[]libdrynx.PublishSignatureBytes {
	ps := make([]*[]libdrynx.PublishSignatureBytes, nbrserver)
	if !(ranges == nil) {
		wg := libunlynx.StartParallelize(nbrserver)
		for i := 0; i < nbrserver; i++ {
			go func(index int) {
				defer wg.Done()
				temp := make([]libdrynx.PublishSignatureBytes, len(ranges))
				for j := 0; j < len(ranges); j++ {
					temp[j] = libdrynxrange.InitRangeProofSignature((*ranges[j])[0]) // u is the first elem
				}
				ps[index] = &temp
				log.Println("Finished creating signatures for server", index)
			}(i)
		}
		libunlynx.EndParallelize(wg)
	} else {
		ps = nil
	}
	return ps
}

// make the network for every different request instead of being a global value
func MakeNetWork() (network *datastruct.TriRoster, err error) {

	network = &datastruct.TriRoster{
		CNs: ConfigToRoster(config.CNs),
	}

	network.VNs = ConfigToRoster(config.VNs)
	network.DPs = ConfigToRoster(config.DPs)
	network.CnToDPs = MakeCNToDP(*network.CNs, *network.DPs)
	network.Total = len(network.CNs.List) + len(network.DPs.List) + len(network.VNs.List)
	network.IdToPub = MakePubMap(network)

	return network, nil
}

// the map from node to its public
func MakePubMap(network *datastruct.TriRoster) map[string]kyber.Point {
	idToPub := make(map[string]kyber.Point, network.Total)
	for _, v := range network.CNs.List {
		idToPub[v.String()] = v.Public
	}
	for _, v := range network.VNs.List {
		idToPub[v.String()] = v.Public
	}
	for _, v := range network.DPs.List {
		idToPub[v.String()] = v.Public
	}

	return idToPub
}

// make the map from cns to dps according to the acount of dps
func MakeCNToDP(cn onet.Roster, dp onet.Roster) map[string]*[]network.ServerIdentity {
	cnToDPs := make(map[string]*[]onet_network.ServerIdentity, len(cn.List))

	for i, v := range dp.List {
		cnIndex := i % len(cn.List)
		if cnToDPs[cn.List[cnIndex].String()] == nil {
			cnToDPs[cn.List[cnIndex].String()] = &[]network.ServerIdentity{*v}
		} else {
			*cnToDPs[cn.List[cnIndex].String()] = append(*cnToDPs[cn.List[cnIndex].String()], *v)
		}
	}

	return cnToDPs
}

// make the nodeconfig into roster for the usage of survey building
func ConfigToRoster(conf []datastruct.NodeConfig) *onet.Roster {
	serverIdSlice := make([]*onet_network.ServerIdentity, len(conf))
	for i, node := range conf {
		public, err := kyber_util_encoding.StringHexToPoint(libdrynx.Suite, node.Pub)
		if err != nil {
			fmt.Println("转换数据错误")
			return nil
		}
		addr := onet_network.NewTCPAddress(node.Addr)
		serverIdSlice[i] = onet_network.NewServerIdentity(public, addr)
	}

	return onet.NewRoster(serverIdSlice)
}

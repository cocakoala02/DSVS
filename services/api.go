package services

import (
	"time"

	libdrynx "github.com/ldsec/drynx/lib"
	libdrynxencoding "github.com/ldsec/drynx/lib/encoding"
	libdrynxobfuscation "github.com/ldsec/drynx/lib/obfuscation"
	libdrynxrange "github.com/ldsec/drynx/lib/range"
	libunlynx "github.com/ldsec/unlynx/lib"
	libunlynxaggr "github.com/ldsec/unlynx/lib/aggregation"
	libunlynxkeyswitch "github.com/ldsec/unlynx/lib/key_switch"
	libunlynxshuffle "github.com/ldsec/unlynx/lib/shuffle"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/key"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
)

// API represents a client with the server to which he is connected and its public/private key pair.
type API struct {
	*onet.Client
	clientID   string
	entryPoint *network.ServerIdentity
	public     kyber.Point
	private    kyber.Scalar
}

// init of the network messages
func init() {
	network.RegisterMessage(libdrynx.GetLatestBlock{})
	network.RegisterMessage(libdrynxrange.RangeProofListBytes{})
	network.RegisterMessage(libunlynxshuffle.PublishedShufflingProofBytes{})
	network.RegisterMessage(libunlynxkeyswitch.PublishedKSListProofBytes{})
	network.RegisterMessage(libunlynxaggr.PublishedAggregationListProofBytes{})
	network.RegisterMessage(libdrynxobfuscation.PublishedListObfuscationProofBytes{})
}

// NewDrynxClient constructor of a client.
func NewDrynxClient(entryPoint *network.ServerIdentity, clientID string) *API {
	keys := key.NewKeyPair(libunlynx.SuiTe)
	newClient := &API{
		Client:     onet.NewClient(libdrynx.Suite, ServiceName),
		clientID:   clientID,
		entryPoint: entryPoint,
		public:     keys.Public, //新建了查询端的公钥
		private:    keys.Private,
	}

	limit := int64(99999)
	// log.LLvl1("limitlimitlimitlimitlimitlimitlimit:", limit)

	libunlynx.CreateDecryptionTable(limit, newClient.public, newClient.private)
	return newClient
}

// Send Query
//______________________________________________________________________________________________________________________

// GenerateSurveyQuery generates a query with all the information in parameters
func (c *API) GenerateSurveyQuery(rosterServers, rosterVNs *onet.Roster, dpToServer map[string]*[]network.ServerIdentity, idToPublic map[string]kyber.Point, surveyID string, operation libdrynx.Operation, ranges []*[]int64, ps []*[]libdrynx.PublishSignatureBytes, proofs int, obfuscation bool, thresholds []float64, diffP libdrynx.QueryDiffP, dpDataGen libdrynx.QueryDPDataGen, cuttingFactor int, sqlsurvey string, dptoPath map[string]string, tablename string) libdrynx.SurveyQuery {
	size1 := 0
	size2 := 0
	if ps != nil {
		size1 = len(ps)
		size2 = len(*ps[0])
	}

	iVSigs := libdrynx.QueryIVSigs{InputValidationSigs: ps, InputValidationSize1: size1, InputValidationSize2: size2}

	test := make([][]int64, 0)
	test = append(test, []int64{int64(1)})

	//create the query
	sq := libdrynx.SurveyQuery{
		SurveyID:                   surveyID,
		RosterServers:              *rosterServers,
		ClientPubKey:               c.public,
		IntraMessage:               false,
		ServerToDP:                 dpToServer,
		IDtoPublic:                 idToPublic,
		Threshold:                  thresholds[0],
		AggregationProofThreshold:  thresholds[1],
		RangeProofThreshold:        thresholds[2],
		ObfuscationProofThreshold:  thresholds[3],
		KeySwitchingProofThreshold: thresholds[4],

		// query statement
		Query: libdrynx.Query{
			Operation:   operation,
			Ranges:      ranges,
			DiffP:       diffP,
			Proofs:      proofs,
			Obfuscation: obfuscation,
			// data generation at DPs
			DPDataGen: dpDataGen,

			// identity blockchain infos
			IVSigs:        iVSigs,
			RosterVNs:     rosterVNs,
			CuttingFactor: cuttingFactor,
			//修改 新添
			SQL:       sqlsurvey,
			DptoPath:  dptoPath,
			TableName: tablename,
		},
	}
	return sq
}

// SendSurveyQuery creates a survey based on a set of entities (servers) and a survey description.
func (c *API) SendSurveyQuery(sq libdrynx.SurveyQuery) (*[]string, *[][]float64, error) {
	log.Lvl2("[API] <Drynx> Client", c.clientID, "is creating a query with SurveyID: ", sq.SurveyID)
	log.Lvl2(c.entryPoint)

	//send the query and get the answer
	sr := libdrynx.ResponseDP{}
	queryexecute := time.Now() //计时器

	err := c.SendProtobuf(c.entryPoint, &sq, &sr)
	log.Lvl1(time.Since(queryexecute)) //计时器

	if err != nil {
		return nil, nil, err
	}

	log.Lvl2("[API] <Drynx> Client", c.clientID, "successfully executed the query with SurveyID ", sq.SurveyID)

	// decrypt/decode the result
	clientDecode := libunlynx.StartTimer("Decode")
	log.Lvl2("[API] <Drynx> Client", c.clientID, "is decrypting the results")

	grp := make([]string, len(sr.Data))
	aggr := make([][]float64, len(sr.Data))
	count := 0
	for i, res := range sr.Data {
		grp[count] = i
		aggr[count] = libdrynxencoding.Decode(res, c.private, sq.Query.Operation)
		count++
	}
	libunlynx.EndTimer(clientDecode)

	log.Lvl2("[API] <Drynx> Client", c.clientID, "finished decrypting the results")
	return &grp, &aggr, nil
}

package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	libdrynxrange "github.com/ldsec/drynx/lib/range"

	drynx_lib "github.com/ldsec/drynx/lib"
	drynx_services "github.com/ldsec/drynx/services"
	libunlynx "github.com/ldsec/unlynx/lib"
	"github.com/urfave/cli"
	"go.dedis.ch/cothority/v3/skipchain"
	kyber "go.dedis.ch/kyber/v3"
	onet "go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
	onet_network "go.dedis.ch/onet/v3/network"
)

func surveyNew(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("need a name")
	}
	name := args.Get(0)

	conf := config{Survey: &configSurvey{Name: &name}}

	return conf.writeTo(os.Stdout)
}

func getRoster(conf configNetwork) ([]onet.Roster, error) {
	ids := make([]*onet_network.ServerIdentity, len(conf.Nodes))
	rosters := make([]onet.Roster, 3)
	for i, e := range conf.Nodes {
		e := e
		ids[i] = &e
	}

	rosterCN := onet.NewRoster(ids[0:3])
	if rosterCN == nil {
		return []onet.Roster{}, errors.New("unable to gen roster based on config")
	}

	rosterVN := onet.NewRoster(ids[3:6])
	if rosterVN == nil {
		return []onet.Roster{}, errors.New("unable to gen roster based on config")
	}

	rosterDP := onet.NewRoster(ids[6:15])
	if rosterDP == nil {
		return []onet.Roster{}, errors.New("unable to gen roster based on config")
	}

	rosters[0] = *rosterCN
	rosters[1] = *rosterVN
	rosters[2] = *rosterDP

	return rosters, nil
}

func surveySetOperation(c *cli.Context) error {
	args := c.Args()
	if len(args) != 1 {
		return errors.New("need an operation")
	}
	operation := args.Get(0)

	conf, err := readConfigFrom(os.Stdin)
	if err != nil {
		return err
	}

	conf.Survey.Operation = &operation

	return conf.writeTo(os.Stdout)
}

func surveyRun(c *cli.Context) error {
	log.SetDebugVisible(3)
	if args := c.Args(); len(args) != 0 {
		return errors.New("no args expected")
	}

	conf, err := readConfigFrom(os.Stdin)
	if err != nil {
		return err
	}

	if conf.Network == nil {
		return errors.New("need some network config")
	}
	roster, err := getRoster(*conf.Network)
	if err != nil {
		return err
	}

	if conf.Network.Client == nil {
		return errors.New("no client defined")
	}
	client := drynx_services.NewDrynxClient(conf.Network.Client, os.Args[0])
	if conf.Survey == nil {
		return errors.New("need some survey config")
	}
	if conf.Survey.Name == nil {
		return errors.New("need a survey name")
	}
	if conf.Survey.Operation == nil {
		return errors.New("need a survey operation")
	}

	// ranges := make([]*[]int64)
	var ranges []*[]int64
	if *conf.Survey.Operation == "sum" {
		ranges = append(ranges, &[]int64{500, 1})

	} else if *conf.Survey.Operation == "mean" {
		ranges = append(ranges, &[]int64{500, 1})
		ranges = append(ranges, &[]int64{101, 1})
	} else if *conf.Survey.Operation == "count" {
		ranges = append(ranges, &[]int64{101, 1})
	} else if *conf.Survey.Operation == "max" {
		for i := 0; i < 9; i++ {
			ranges = append(ranges, &[]int64{2, 1})
		}
	} else if *conf.Survey.Operation == "min" {
		for i := 0; i < 9; i++ {
			ranges = append(ranges, &[]int64{2, 1})
		}
	}

	// signatures for Input Validation
	nbrserver := len(roster[0].List)
	ps := make([]*[]drynx_lib.PublishSignatureBytes, nbrserver)
	if !(ranges == nil) {
		wg := libunlynx.StartParallelize(nbrserver)
		for i := 0; i < nbrserver; i++ {
			go func(index int) {
				defer wg.Done()
				temp := make([]drynx_lib.PublishSignatureBytes, len(ranges))
				for j := 0; j < len(ranges); j++ {
					temp[j] = libdrynxrange.InitRangeProofSignature((*ranges[j])[0]) // u is the first elem
				}
				ps[index] = &temp
				log.Lvl1("Finished creating signatures for server", index)
			}(i)
		}
		libunlynx.EndParallelize(wg)
	} else {
		ps = nil
	}

	dptopath := map[string]string{"tcp://127.0.0.1:7013": "../data",
		"tcp://127.0.0.1:7015": "../data",
		"tcp://127.0.0.1:7017": "../data",
		"tcp://127.0.0.1:7019": "../data",
		"tcp://127.0.0.1:7021": "../data",
		"tcp://127.0.0.1:7023": "../data",
		"tcp://127.0.0.1:7025": "../data",
		"tcp://127.0.0.1:7027": "../data",
		"tcp://127.0.0.1:7029": "../data",
	}

	tablename := "statistics_experiment_data"

	sq := client.GenerateSurveyQuery(

		/// network

		&roster[0], // CN roster
		&roster[1], // VN roster
		map[string]*[]onet_network.ServerIdentity{ // map CN to DPs
			roster[0].List[0].String(): {*roster[2].List[0], *roster[2].List[1], *roster[2].List[2]},
			roster[0].List[1].String(): {*roster[2].List[3], *roster[2].List[4], *roster[2].List[5]},
			roster[0].List[2].String(): {*roster[2].List[6], *roster[2].List[7], *roster[2].List[8]}},
		map[string]kyber.Point{ // map CN|DP|VN to pub key
			roster[0].List[0].String(): roster[0].List[0].Public,
			roster[0].List[1].String(): roster[0].List[1].Public,
			roster[0].List[2].String(): roster[0].List[2].Public,

			roster[1].List[0].String(): roster[1].List[0].Public,
			roster[1].List[1].String(): roster[1].List[1].Public,
			roster[1].List[2].String(): roster[1].List[2].Public,

			roster[2].List[0].String(): roster[2].List[0].Public,
			roster[2].List[1].String(): roster[2].List[1].Public,
			roster[2].List[2].String(): roster[2].List[2].Public,
			roster[2].List[3].String(): roster[2].List[3].Public,
			roster[2].List[4].String(): roster[2].List[4].Public,
			roster[2].List[5].String(): roster[2].List[5].Public,
			roster[2].List[6].String(): roster[2].List[6].Public,
			roster[2].List[7].String(): roster[2].List[7].Public,
			roster[2].List[8].String(): roster[2].List[8].Public},

		/// gen

		*conf.Survey.Name, // survey id
		drynx_lib.ChooseOperation(
			*conf.Survey.Operation, // operation
			1,                      // min num of DP to query
			9,                      // max num of DP to query
			5,                      // dimension for linear regression
			0),                     // "cutting factor", how much to remove of gen data[0:#/n]

		ranges, // range for each output of operation
		ps,     // signature of range validity
		int(1), // 0 == no proof, 1 == proof, 2 == optimized proof

		false, // obfuscation
		[]float64{
			1.0,  // threshold general
			1.0,  // threshold aggregation
			1.0,  // threshold range
			0.0,  // obfuscation
			1.0}, // threshold key switch
		drynx_lib.QueryDiffP{ // differential privacy
			LapMean: 0.0, LapScale: 0.0, NoiseListSize: 0, Quanta: 0.0, Scale: 0},
		drynx_lib.QueryDPDataGen{ // how to group by
			GroupByValues: []int64{1}, GenerateRows: 1, GenerateDataMin: int64(0), GenerateDataMax: int64(10)},
		0, // cutting factor
		//
		"SELECT SUM(visits) FROM statistics_experiment_data", //SQL
		dptopath,
		tablename,
	)

	var wg *sync.WaitGroup
	var block *skipchain.SkipBlock

	if sq.Query.Proofs != 0 {
		// send query to the skipchain and 'wait' for all proofs' verification to be done
		clientSkip := drynx_services.NewDrynxClient(roster[1].List[0], "simul-skip-"+*conf.Survey.Operation)

		wg = libunlynx.StartParallelize(1)
		go func(elVNs *onet.Roster) {
			defer wg.Done()

			err = clientSkip.SendSurveyQueryToVNs(elVNs, &sq)
			if err != nil {
				log.Fatal("Error sending query to VNs:", err)
			}
		}(&roster[1])
		libunlynx.EndParallelize(wg)

		wg = libunlynx.StartParallelize(1)
		go func(si *network.ServerIdentity) {
			defer wg.Done()

			block, err = clientSkip.SendEndVerification(si, *conf.Survey.Name)
			if err != nil {
				log.Fatal("Error starting the 'waiting' threads:", err)
			}
		}(roster[1].List[0])
	}
	// queryexecute := time.Now() //计时器
	_, aggregations, err := client.SendSurveyQuery(sq)
	// log.Lvl1(time.Since(queryexecute)) //计时器
	if err != nil {
		return err
	}

	result, ok := float64(0), false
	for _, a := range *aggregations {
		if len(a) != 1 {
			return errors.New("line in aggregation larger than one, dunno how to print")
		}
		if ok && result != a[0] {
			return errors.New("not same value found in aggregation, dunno how to print")
		}
		result = a[0]
		ok = true
	}
	fmt.Println(result)

	if len(roster[1].List) > 0 {
		clientSkip := drynx_services.NewDrynxClient(roster[1].List[0], "simul-skip-"+*conf.Survey.Operation)

		if sq.Query.Proofs != 0 {
			libunlynx.EndParallelize(wg)
			// log.Lvl1(time.Since(queryexecute)) //计时器
			// close DB
			if err := clientSkip.SendCloseDB(&roster[1], &drynx_lib.CloseDB{Close: 1}); err != nil {
				log.Fatal("Error closing the DB:", err)
			}
		}

		retrieveBlock := time.Now()
		sb, err := clientSkip.SendGetLatestBlock(&roster[1], block)
		if err != nil || sb == nil {
			log.Fatal("Something wrong when fetching the last block")
		}
		log.Lvl1(time.Since(retrieveBlock))

		//或者反序列化为具体数据结构
		_, msg, err := network.Unmarshal(sb.Data, libunlynx.SuiTe)
		if err != nil {
			log.Fatal("Error unmarshaling block data:", err)
		}
		fmt.Printf("msg 类型是：%T\n", msg)

		fmt.Printf("Block Data: %+v\n", msg)

		// 类型断言
		dataBlock, ok := msg.(*drynx_lib.DataBlock)
		if !ok {
			log.Fatal("msg 不是 *libdrynx.DataBlock 类型")
		}
		fmt.Printf("  SurveyID:     %v\n", dataBlock.SurveyID)
		fmt.Printf("  Sample:       %v\n", dataBlock.Sample)
		fmt.Printf("  Time:         %v\n", dataBlock.Time)
		fmt.Printf("  ServerNumber: %v\n", dataBlock.ServerNumber)

		fmt.Println("  Proofs:")
		for k, v := range dataBlock.Proofs {
			fmt.Printf("    %s => %d\n", k, v)
		}

	}

	fmt.Println(result)

	return nil
}

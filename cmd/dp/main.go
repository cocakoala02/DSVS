package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	// "github.com/pelletier/go-toml"
	"github.com/urfave/cli"

	"github.com/BurntSushi/toml"
	kyber_encoding "go.dedis.ch/kyber/v3/util/encoding"
	kyber_key "go.dedis.ch/kyber/v3/util/key"
	onet_app "go.dedis.ch/onet/v3/app"
	onet_log "go.dedis.ch/onet/v3/log"
	onet_network "go.dedis.ch/onet/v3/network"

	_ "github.com/go-sql-driver/mysql"
	drynx "github.com/ldsec/drynx/lib"
	_ "github.com/ldsec/drynx/protocols"
	_ "github.com/ldsec/drynx/services"

	libenc "github.com/ldsec/drynx/lib/encoding"
)

type WithDatabaseConfig struct {
	onet_app.CothorityConfig

	// 在原有 toml 基础上加一个 database 段
	Database struct {
		Driver string `toml:"driver"` // e.g. "mysql"
		DSN    string `toml:"dsn"`    // e.g. "user:pass@tcp(127.0.0.1:3306)/db?parseTime=true"
	} `toml:"database"`
}

func toTmpFile(reader io.Reader) (os.File, error) {
	file, err := ioutil.TempFile("", "onet-stdin")
	if err != nil {
		return os.File{}, err
	}
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return os.File{}, err
	}
	if _, err = file.Write(content); err != nil {
		return os.File{}, err
	}
	return *file, nil
}

func gen(c *cli.Context) error {
	args := c.Args()
	if len(args) != 2 {
		return errors.New("need two bind addresses")
	}
	addrNode, addrClient := args.Get(0), args.Get(1)

	onet_log.OutputToBuf() // reduce garbage to stdout

	serverBinding := onet_network.NewAddress(onet_network.PlainTCP, addrNode)
	kp := kyber_key.NewKeyPair(drynx.Suite)

	pub, err := kyber_encoding.PointToStringHex(drynx.Suite, kp.Public)
	if err != nil {
		return err
	}
	priv, _ := kyber_encoding.ScalarToStringHex(drynx.Suite, kp.Private)
	if err != nil {
		return err
	}

	serviceKeys := onet_app.GenerateServiceKeyPairs()

	conf := WithDatabaseConfig{
		CothorityConfig: onet_app.CothorityConfig{
			Suite:         drynx.Suite.String(),
			Public:        pub,
			Private:       priv,
			Address:       serverBinding,
			ListenAddress: addrNode,
			URL:           "https://" + addrClient,
			Description:   "drynx-dp",
			Services:      serviceKeys,
		},
	}
	// 给出一个示例默认值，方便手工改
	conf.Database.Driver = "mysql"
	conf.Database.DSN = "root:xx..2525@tcp(127.0.0.1:3306)/learn_sql?parseTime=true&timeout=2s&readTimeout=2s"

	return toml.NewEncoder(os.Stdout).Encode(conf)
}

func run(c *cli.Context) error {
	if len(c.Args()) > 0 {
		return errors.New("need no argument")
	}

	config := c.Args().First()
	if !c.Args().Present() {
		configFile, err := toTmpFile(os.Stdin)
		if err != nil {
			return err
		}
		defer os.Remove(configFile.Name())
		config = configFile.Name()
	}

	// 先读一遍 TOML，把 database 注入到 libenc
	var conf WithDatabaseConfig
	if _, err := toml.DecodeFile(config, &conf); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	libenc.SetDPLocalConfig(&libenc.DPLocalConfig{
		Database: conf.Database,
	})

	// 再用 onet_app 启动（它只认识原始结构；但多余字段不会影响运行）
	onet_app.RunServer(config)
	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = "configure and start a Drynx DP node (with database)"
	app.Description = fmt.Sprintf(strings.TrimSpace(strings.Replace(`
	configuration uses stdin/stdout.

	if you want to generate a DP server config, use something like
		%[1]s gen > $my_dp_config
	then, you can run it
		cat $my_dp_config | %[1]s run
	`, "\t", "   ", -1)), os.Args[0])

	app.Commands = []cli.Command{
		{
			Name:      "gen",
			ArgsUsage: "host:node-port host:client-port",
			Usage:     "generate a DP server config (with database section)",
			Action:    gen,
		}, {
			Name:   "run",
			Usage:  "sink of a DP server config, run the DP node",
			Action: run,
		},
	}

	if err := app.Run(os.Args); err != nil {
		onet_log.Error(err)
	}
}

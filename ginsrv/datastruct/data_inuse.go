package datastruct

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/onet/v3"
	onet_network "go.dedis.ch/onet/v3/network"
)

// the single config of a node
type NodeConfig struct {
	Pub string
	// it should includes ip and port
	Addr string
}

// the config information to build a seek network
type Config struct {
	CNs []NodeConfig
	VNs []NodeConfig
	// it maybe needs a map to build a more formal network
	DPs           []NodeConfig
	Client        string
	Ranges        int
	OutputNum     int
	ServerNum     int
	CuttingFactor int
	DpMap         map[string]string
	Scale         int //suo fang
	// 新增：客户端统一的“需要缩放的浮点列”
	FloatColumns []string `json:"float_columns"`
}

// a total roster for survey building
type TriRoster struct {
	Total   int
	CNs     *onet.Roster
	VNs     *onet.Roster
	DPs     *onet.Roster
	CnToDPs map[string]*[]onet_network.ServerIdentity
	IdToPub map[string]kyber.Point
}

// a struct used in loading the .toml
type SeverToml struct {
	ListenAddress string `toml:"ListenAddress"`
	Public        string `toml:"Public"`
}

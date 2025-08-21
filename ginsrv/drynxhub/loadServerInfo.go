package drynxhub

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ldsec/drynx/ginsrv/datastruct"
)

// just for auto loading the info from .toml created by .cmd, and it just can be used in simulation
func LoadToml() string {
	// get the path
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Get the cwd error:", err)
		return ""
	}
	if dir := strings.Split(cwd, "/"); dir[len(dir)-1] != "drynx-master" {
		fmt.Println("cwd is incorrect")
		return ""
	}
	cwdLoad := cwd + "/bin"

	// load the .toml in nodes
	nodes := make([]datastruct.SeverToml, 15)
	for i := 0; i < 15; i++ {
		tomlPath := cwdLoad + "/node" + strconv.Itoa(i+1) + "_config.toml"
		if _, err := toml.DecodeFile(tomlPath, &nodes[i]); err != nil {
			fmt.Println("load .toml err!:", err)
			return ""
		}
	}

	// open and load the .json
	file, err := os.Open(cwd + "/ginsrv/config.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()
	config := datastruct.Config{}
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return ""
	}

	// write in the datas about nodes with addr and pub
	newConfig := datastruct.Config{
		CNs:       make([]datastruct.NodeConfig, 3),
		VNs:       make([]datastruct.NodeConfig, 3),
		DPs:       make([]datastruct.NodeConfig, 99), //consum max = 99
		Client:    ClientSet(config.Client),
		Ranges:    18,
		OutputNum: 1,
		ServerNum: func(n int) int {
			if n != 0 {
				return n
			} else {
				return 3
			}
		}(config.ServerNum),
		CuttingFactor: 0,
	}
	for i, v := range nodes {
		switch i {
		case 0, 1, 2:
			newConfig.CNs[i] = datastruct.NodeConfig{
				Addr: v.ListenAddress,
				Pub:  v.Public,
			}
		case 3, 4, 5:
			newConfig.VNs[i-3] = datastruct.NodeConfig{
				Addr: v.ListenAddress,
				Pub:  v.Public,
			}
		case 6, 7, 8, 9, 10, 11, 12, 13, 14:
			newConfig.DPs[i-6] = datastruct.NodeConfig{
				Addr: v.ListenAddress,
				Pub:  v.Public,
			}
		}
	}

	// save
	outFile, err := os.Create(cwd + "/ginsrv/config.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return ""
	}
	defer outFile.Close()
	if err := json.NewEncoder(outFile).Encode(&newConfig); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return ""
	}

	// return the path for loadconfig
	return (cwd + "/ginsrv/config.json")
}

// set the default value of client
func ClientSet(client string) string {
	if client == "" {
		return "127.0.0.1:7002"
	} else {
		return client
	}
}

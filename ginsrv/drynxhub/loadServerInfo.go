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

	// loader
	loadServerToml := func(path string, dst *datastruct.SeverToml) bool {
		if _, err := toml.DecodeFile(path, dst); err != nil {
			fmt.Println("load .toml err!:", err, "path:", path)
			return false
		}
		return true
	}

	// 先准备一个默认的新配置
	newConfig := datastruct.Config{
		CNs:           make([]datastruct.NodeConfig, 3),
		VNs:           make([]datastruct.NodeConfig, 3),
		DPs:           make([]datastruct.NodeConfig, 99), // consum max = 99
		Client:        ClientSet(""),
		Ranges:        18,
		OutputNum:     1,
		ServerNum:     3, // 原逻辑：若未提供则默认 3
		CuttingFactor: 0,
	}

	// 读取 CN: cn1..cn3
	for i := 1; i <= 3; i++ {
		var st datastruct.SeverToml
		path := cwdLoad + "/cn" + strconv.Itoa(i) + "_config.toml"
		if !loadServerToml(path, &st) {
			return ""
		}
		newConfig.CNs[i-1] = datastruct.NodeConfig{
			Addr: st.ListenAddress,
			Pub:  st.Public,
		}
	}

	// 读取 VN: vn1..vn3
	for i := 1; i <= 3; i++ {
		var st datastruct.SeverToml
		path := cwdLoad + "/vn" + strconv.Itoa(i) + "_config.toml"
		if !loadServerToml(path, &st) {
			return ""
		}
		newConfig.VNs[i-1] = datastruct.NodeConfig{
			Addr: st.ListenAddress,
			Pub:  st.Public,
		}
	}

	// 读取 DP: dp7..dp15
	dpIdx := 0
	for i := 7; i <= 15; i++ {
		var st datastruct.SeverToml
		path := cwdLoad + "/dp" + strconv.Itoa(i) + "_config.toml"
		if !loadServerToml(path, &st) {
			return ""
		}
		newConfig.DPs[dpIdx] = datastruct.NodeConfig{
			Addr: st.ListenAddress,
			Pub:  st.Public,
		}
		dpIdx++
	}

	// 尝试加载当前 ginsrv/config.json，尽量保留其配置字段
	if file, err := os.Open(cwd + "/ginsrv/config.json"); err == nil {
		defer file.Close()
		old := datastruct.Config{}
		if err := json.NewDecoder(file).Decode(&old); err == nil {
			if old.Client != "" {
				newConfig.Client = ClientSet(old.Client)
			} else {
				newConfig.Client = ClientSet(newConfig.Client)
			}
			if old.ServerNum != 0 {
				newConfig.ServerNum = old.ServerNum
			}
			if old.Ranges != 0 {
				newConfig.Ranges = old.Ranges
			}
			if old.OutputNum != 0 {
				newConfig.OutputNum = old.OutputNum
			}
			newConfig.CuttingFactor = old.CuttingFactor
		}
	}

	// 保存到 ginsrv/config.json
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

	// 返回路径供外部 load 使用
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

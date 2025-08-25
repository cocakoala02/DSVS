package drynxhub

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/ldsec/drynx/ginsrv/datastruct"
)

// dpRegistryFile 是生产环境推荐的 DP 名册
const dpRegistryFile = "ginsrv/dp_registry.json"

// 仅用于 dp_registry.json 的载体
type dpRegistry struct {
	DPs []datastruct.NodeConfig `json:"DPs"`
}

// just for auto loading the info from .toml created by .cmd, and it just can be used in simulation
func LoadToml() string {
	// cwd 校验
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Get the cwd error:", err)
		return ""
	}
	if dir := strings.Split(cwd, "/"); dir[len(dir)-1] != "drynx-master" {
		fmt.Println("cwd is incorrect")
		return ""
	}
	cwdBin := filepath.Join(cwd, "bin")

	// toml loader
	loadServerToml := func(path string, dst *datastruct.SeverToml) bool {
		if _, err := toml.DecodeFile(path, dst); err != nil {
			fmt.Println("load .toml err!:", err, "path:", path)
			return false
		}
		return true
	}

	// 新配置骨架
	newCfg := datastruct.Config{
		CNs:           make([]datastruct.NodeConfig, 3),
		VNs:           make([]datastruct.NodeConfig, 3),
		DPs:           make([]datastruct.NodeConfig, 99), // 最大预留
		Client:        ClientSet(""),
		Ranges:        18,
		OutputNum:     1,
		ServerNum:     3,
		CuttingFactor: 0,
	}

	// === CN: cn1..cn3 ===
	for i := 1; i <= 3; i++ {
		var st datastruct.SeverToml
		path := filepath.Join(cwdBin, "cn"+strconv.Itoa(i)+"_config.toml")
		if !loadServerToml(path, &st) {
			return ""
		}
		newCfg.CNs[i-1] = datastruct.NodeConfig{Addr: st.ListenAddress, Pub: st.Public}
	}

	// === VN: vn1..vn3 ===
	for i := 1; i <= 3; i++ {
		var st datastruct.SeverToml
		path := filepath.Join(cwdBin, "vn"+strconv.Itoa(i)+"_config.toml")
		if !loadServerToml(path, &st) {
			return ""
		}
		newCfg.VNs[i-1] = datastruct.NodeConfig{Addr: st.ListenAddress, Pub: st.Public}
	}

	// === DPs: 生产优先从 dp_registry.json 读取 ===
	regPath := filepath.Join(cwd, dpRegistryFile)
	if f, err := os.Open(regPath); err == nil {
		defer f.Close()
		var reg dpRegistry
		if err := json.NewDecoder(f).Decode(&reg); err == nil && len(reg.DPs) > 0 {
			for i := range reg.DPs {
				newCfg.DPs[i] = reg.DPs[i]
			}
		} else {
			// 回退：兼容单机演示（dp7..dp15）
			// fillDPsFromLocalToml(cwdBin, &newCfg)
			fmt.Println("decode the dp_registry.json error:", err)
			return ""
		}
	} else {
		// 没有 dp_registry.json，继续回退本地 dp*.toml
		// fillDPsFromLocalToml(cwdBin, &newCfg)
		fmt.Println("decode the dp_registry.json error:", err)
		return ""
	}

	// === 合并旧 config.json 的可配置项（若存在） ===
	if f, err := os.Open(filepath.Join(cwd, "ginsrv", "config.json")); err == nil {
		defer f.Close()
		old := datastruct.Config{}
		if err := json.NewDecoder(f).Decode(&old); err == nil {
			if old.Client != "" {
				newCfg.Client = ClientSet(old.Client)
			} else {
				newCfg.Client = ClientSet(newCfg.Client)
			}
			if old.ServerNum != 0 {
				newCfg.ServerNum = old.ServerNum
			}
			if old.Ranges != 0 {
				newCfg.Ranges = old.Ranges
			}
			if old.OutputNum != 0 {
				newCfg.OutputNum = old.OutputNum
			}
			newCfg.CuttingFactor = old.CuttingFactor
		}
	}

	newCfg.DpMap = map[string]string{ //类似一个部门注册表 need fix  感觉可以加一个dp的注册信息数据库，然后直接取
		"A": "127.0.0.1:7013",
		"B": "127.0.0.1:7015",
		"C": "127.0.0.1:7017",
		"D": "127.0.0.1:7019",
		"E": "127.0.0.1:7021",
		"F": "127.0.0.1:7023",
		"G": "127.0.0.1:7025",
		"H": "127.0.0.1:7027",
		"I": "127.0.0.1:7029",
	}

	// === 写回 config.json ===
	outPath := filepath.Join(cwd, "ginsrv", "config.json")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return ""
	}
	defer outFile.Close()
	if err := json.NewEncoder(outFile).Encode(&newCfg); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return ""
	}

	return outPath
}

// 从本地 dp7..dp15_config.toml 兼容填充（演示专用）
// func fillDPsFromLocalToml(cwdBin string, cfg *datastruct.Config) {
// 	dpIdx := 0
// 	for i := 7; i <= 15; i++ {
// 		var st datastruct.SeverToml
// 		path := filepath.Join(cwdBin, "dp"+strconv.Itoa(i)+"_config.toml")
// 		if _, err := toml.DecodeFile(path, &st); err != nil {
// 			// 没这个文件就跳过（允许少量 dp）
// 			continue
// 		}
// 		cfg.DPs[dpIdx] = datastruct.NodeConfig{
// 			Addr: st.ListenAddress,
// 			Pub:  st.Public,
// 		}
// 		dpIdx++
// 	}
// }

// set the default value of client
func ClientSet(client string) string {
	if client == "" {
		return "127.0.0.1:7002" //need fix
	} else {
		return client
	}
}

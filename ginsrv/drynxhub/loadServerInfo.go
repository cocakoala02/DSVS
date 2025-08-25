package drynxhub

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql" // MySQL 驱动（如用 pg/duckdb，请换对应驱动）
	"github.com/ldsec/drynx/ginsrv/datastruct"
)

// LoadServerInfo 从本地 toml 读取 CN/VN，从数据库读取 DP 信息，合成并写回 ginsrv/config.json。
// dsn 例如： "user:pass@tcp(127.0.0.1:3306)/dbname?parseTime=true&timeout=2s&readTimeout=2s&multiStatements=false"
func LoadServerInfo(dsn string) string {
	// 1) cwd 校验与路径
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

	// 2) toml loader（用于 CN/VN）
	loadServerToml := func(path string, dst *datastruct.SeverToml) bool {
		if _, err := toml.DecodeFile(path, dst); err != nil {
			fmt.Println("load .toml err!:", err, "path:", path)
			return false
		}
		return true
	}

	// 3) 组装基础配置骨架
	newCfg := datastruct.Config{
		CNs:           make([]datastruct.NodeConfig, 3),
		VNs:           make([]datastruct.NodeConfig, 3),
		DPs:           make([]datastruct.NodeConfig, 0, 32), // 动态追加 DP
		DpMap:         make(map[string]string),              // ID -> Addr
		Client:        ClientSet(""),
		Ranges:        18,
		OutputNum:     1,
		ServerNum:     3,
		CuttingFactor: 0,
	}

	// 4) 读取 CN: cn1..cn3
	for i := 1; i <= 3; i++ {
		var st datastruct.SeverToml
		path := filepath.Join(cwdBin, "cn"+strconv.Itoa(i)+"_config.toml")
		if !loadServerToml(path, &st) {
			return ""
		}
		newCfg.CNs[i-1] = datastruct.NodeConfig{Addr: st.ListenAddress, Pub: st.Public}
	}

	// 5) 读取 VN: vn1..vn3
	for i := 1; i <= 3; i++ {
		var st datastruct.SeverToml
		path := filepath.Join(cwdBin, "vn"+strconv.Itoa(i)+"_config.toml")
		if !loadServerToml(path, &st) {
			return ""
		}
		newCfg.VNs[i-1] = datastruct.NodeConfig{Addr: st.ListenAddress, Pub: st.Public}
	}

	// 6) 连接数据库并读取 DP：dp_info_table(ID, Addr, Pub)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("open db error:", err)
		return ""
	}
	defer db.Close()

	rows, err := db.Query("SELECT ID, Addr, Pub FROM dp_info_table")
	if err != nil {
		fmt.Println("query db error:", err)
		return ""
	}
	defer rows.Close()

	for rows.Next() {
		var id, addr, pub string
		if err := rows.Scan(&id, &addr, &pub); err != nil {
			fmt.Println("scan row error:", err)
			return ""
		}
		// 追加 DP 信息
		newCfg.DPs = append(newCfg.DPs, datastruct.NodeConfig{
			Addr: addr,
			Pub:  pub,
		})
		// 写入 ID -> Addr 映射
		newCfg.DpMap[id] = addr
	}
	if err := rows.Err(); err != nil {
		fmt.Println("rows error:", err)
		return ""
	}

	// 7) 合并旧 config.json 的可配置项（若存在则尽量保留）
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

	// 8) 写回 ginsrv/config.json
	outPath := filepath.Join(cwd, "ginsrv", "config.json")
	outFile, err := os.Create(outPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return ""
	}
	defer outFile.Close()

	enc := json.NewEncoder(outFile)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ") // 便于阅读
	if err := enc.Encode(&newCfg); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return ""
	}

	return outPath
}

// set the default value of client
func ClientSet(client string) string {
	if client == "" {
		return "127.0.0.1:7002" // 与 cn1 的 client 端口默认一致 need fix:思考这个端口号
	}
	return client
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TriSurReq struct {
	Sql string `json:"sql"`
}



func main() {
	url := "http://0.0.0.0:8088/survey"

	req := TriSurReq{
	Sql : "SELECT SUM(visits) FROM statistics_experiment_data",
	}

	jsonReq, err1 := json.Marshal(req)
	if err1 != nil {
		fmt.Println("序列化 JSON 失败:", err1)
		return
	}

	resp, err2 := http.Post(url, "application/json", bytes.NewBuffer(jsonReq))
	if err2 != nil {
		fmt.Println("请求失败:", err2)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求失败，状态码：%d\n", resp.StatusCode)
		return
	}

	var res string
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return
		} else {
			fmt.Println("读取完毕")
			res = string(buf[:n])
			fmt.Println(res)
			break
		}
	}
}

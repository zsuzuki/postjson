package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := flag.String("port", "", "Port to use if not specified in server address")
	flag.Parse()

	args := flag.Args()
	if len(args) < 3 {
		fmt.Println("Usage: postjson [--port=PORT] <server1>[,<server2>,...] <path> arg1=value arg2=value ...")
		os.Exit(1)
	}

	serverArg := args[0]
	path := args[1]
	argPairs := args[2:]

	// JSONデータの組み立て
	payload := make(map[string]string)
	for _, arg := range argPairs {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			fmt.Fprintf(os.Stderr, "Invalid argument: %s\n", arg)
			os.Exit(1)
		}
		payload[kv[0]] = kv[1]
	}
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON encode error: %v\n", err)
		os.Exit(1)
	}

	// サーバーごとにリクエストを送信
	servers := strings.Split(serverArg, ",")
	for i, s := range servers {
		s = strings.TrimSpace(s)
		protocol := "http"
		host := s

		// プロトコル指定
		if strings.HasPrefix(s, "https://") {
			protocol = "https"
			host = strings.TrimPrefix(s, "https://")
		} else if strings.HasPrefix(s, "http://") {
			host = strings.TrimPrefix(s, "http://")
		}

		// ポート指定がなければ追加
		if !strings.Contains(host, ":") && *port != "" {
			host += ":" + *port
		}

		// URL生成してポスト
		url := fmt.Sprintf("%s://%s%s", protocol, host, path)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
		if err != nil {
			fmt.Printf("ret%d: error: %v\n", i, err)
			continue
		}
		defer resp.Body.Close()

		var respBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			fmt.Printf("ret%d: invalid json response\n", i)
			continue
		}

		// レスポンス表示
		respDump, _ := json.MarshalIndent(respBody, "", "  ")
		fmt.Printf("%s: %s\n", host, string(respDump))
	}
}

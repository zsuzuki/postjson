package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	port := flag.String("port", "", "Default port if not specified in server address")
	timeout := flag.Duration("timeout", 5*time.Second, "HTTP request timeout")
	output := flag.String("output", "", "Output file to save result (optional)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 3 {
		fmt.Println("Usage: postjson [--port=PORT] [--timeout=3s] [--output=res.json] <server1>[,<server2>,...] <path> arg1=val1 ...")
		os.Exit(1)
	}

	servers := strings.Split(args[0], ",")
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

	client := &http.Client{Timeout: *timeout}
	var wg sync.WaitGroup
	var mu sync.Mutex
	final := make(map[string]interface{})

	// サーバーごとにリクエストを送信
	for _, s := range servers {
		wg.Add(1)

		// 並列実行
		go func(s string) {
			defer wg.Done()

			serverName := strings.TrimSpace(s)
			protocol := "http"
			host := serverName

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
			resultKey := host
			url := fmt.Sprintf("%s://%s%s", protocol, host, path)

			resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonBytes))
			if err != nil {
				mu.Lock()
				final[resultKey] = map[string]string{"error": err.Error()}
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			var respBody any
			if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
				mu.Lock()
				final[resultKey] = map[string]string{"error": "invalid JSON response"}
				mu.Unlock()
				return
			}

			mu.Lock()
			final[resultKey] = respBody
			mu.Unlock()
		}(s)
	}

	wg.Wait()

	// レスポンス出力 or ファイル保存
	out, _ := json.MarshalIndent(final, "", "  ")
	if *output != "" {
		err := os.WriteFile(*output, out, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println(string(out))
	}
}

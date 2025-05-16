# JSON一斉ポストツール

いくつかのwebサーバーとちょっとしたJSONでやり取りしたい。
ということで作成。

## 使い方

サーバー名を並べて、PATHを指定。その後にJsonの値を並べる。

```shell
> postjson --port=10080 server1,server2 /app Name=Suzuki Age=20 Country=Japan
```

```text
{
  "server1:10080": {
    "code": 0,
    "result": "OK"
  }
  "server2:10080": {
    "code": 0,
    "result": "OK"
  }
}
```

### オプション

- port ポート指定
- timeout タイムアウト時刻("2s"で2秒)
- output レスポンスをファイルに保存

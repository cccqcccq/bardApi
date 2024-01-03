package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/net/proxy"
	"html/template"
	"net"
	"net/http"
	"time"
)

const URL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=获取的Key"

// 创建一个SOCKS5拨号器用于连接本地代理
var dialer, _ = proxy.SOCKS5("tcp", "127.0.0.1:10808", nil, proxy.Direct)

// 创建一个新的函数来适配DialContext
var dialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
	return dialer.(proxy.ContextDialer).DialContext(ctx, network, addr)
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.HandleFunc("/bard", handler)
	http.HandleFunc("/bard-more", handler2)
	http.ListenAndServe(":1212", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("static/index.html")
	t.Execute(w, nil)
}

// 创建处理器函数
func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Content-type", "application/json")
	// 获取发送来的json数据的text字段
	jsonData, text := "", ""
	buf, temp := make([]byte, 1024), make(map[string]any)
	for {
		n, _ := r.Body.Read(buf)
		if n == 0 {
			break
		}
		jsonData += string(buf[:n])
	}
	json.Unmarshal([]byte(jsonData), &temp)
	if temp["text"] == nil || temp["text"].(string) == "" {
		fmt.Fprintln(w, `{"text":"我不能为空"}`)
		return
	}
	text = temp["text"].(string)
	// 返回数据
	m := make(map[string]string)
	// 发送用户输入的内容,safetySettings设置bard的安全限制
	postData := []byte(`{
	"contents": [{"parts":[{"text": "` + text + `"}]}],
	"safetySettings":[{
		"category": "HARM_CATEGORY_HARASSMENT",
		"threshold": "BLOCK_NONE"
	},{
		"category": "HARM_CATEGORY_HATE_SPEECH",
		"threshold": "BLOCK_NONE"
	},{
		"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
		"threshold": "BLOCK_NONE"
	},{
		"category": "HARM_CATEGORY_DANGEROUS_CONTENT",
		"threshold": "BLOCK_NONE"
	}],
	"generationConfig":{
		"temperature": 0.9,
	}
}`)
	m["text"] = postBard(postData)
	data, _ := json.Marshal(m)
	fmt.Fprintln(w, string(data))
}

// 创建处理器函数
func handler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	w.Header().Set("Content-type", "application/json")
	// 获取发送来的json数据的text字段
	jsonData := ""
	buf := make([]byte, 1024)
	for {
		n, _ := r.Body.Read(buf)
		if n == 0 {
			break
		}
		jsonData += string(buf[:n])
	}
	// 返回数据
	m := make(map[string]string)
	//删掉jsonData最后一个字符},再添加safetySettings设置bard的安全限制
	jsonData = jsonData[:len(jsonData)-1] + `,
	"safetySettings":[{
		"category": "HARM_CATEGORY_HARASSMENT",
		"threshold": "BLOCK_NONE"
	},{
		"category": "HARM_CATEGORY_HATE_SPEECH",
		"threshold": "BLOCK_NONE"
	},{
		"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
		"threshold": "BLOCK_NONE"
	},{
		"category": "HARM_CATEGORY_DANGEROUS_CONTENT",
		"threshold": "BLOCK_NONE"
	}],
	"generationConfig":{
		"temperature": 0.9,
	}
}`
	m["text"] = postBard([]byte(jsonData))
	data, _ := json.Marshal(m)
	fmt.Fprintln(w, string(data))
}

// 向bard发送请求
func postBard(jsonData []byte) string {
	// 创建一个HTTP客户端并设置代理
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,             // 使用环境变量中的代理设置
			DialContext:           dialContext,                           // 使用SOCKS5拨号器进行拨号
			TLSHandshakeTimeout:   10 * time.Second,                      // 设置TLS握手超时
			ExpectContinueTimeout: 1 * time.Second,                       // 设置Expect-Continue超时
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // 跳过TLS证书验证
		},
	}
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json") // 设置请求头的内容类型

	// 发送请求
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	result := ""
	defer res.Body.Close()
	buf := make([]byte, 1024)
	for {
		n, _ := res.Body.Read(buf)
		if n == 0 {
			break
		}
		result += string(buf[:n])
	}
	return getText([]byte(result))
}

// 获取bard响应中的文本字段
func getText(resData []byte) string {
	var data map[string]any
	_ = json.Unmarshal(resData, &data)
	fmt.Println(data)
	// 可能是bard因为安全政策屏蔽了本次对话
	if data["candidates"].([]any)[0].(map[string]any)["content"] == nil {
		return "未知原因失败,请重试"
	}
	text := data["candidates"].([]any)[0].(map[string]any)["content"].(map[string]any)["parts"].([]any)[0].(map[string]any)["text"]
	return text.(string)
}

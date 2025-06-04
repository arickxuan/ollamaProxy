# ollamaProxy

## 介绍
    ollamaProxy 是一个模仿ollama的代理服务，用于将用户的请求转发给 GPT/Claude 等大模型，并将模型的响应返回给用户。
    通过本项目即可使支持ollama的客户端无缝切换到其他大模型。如 goland idea 等。
    本项目使用golang开发，使用gin框架实现。
    本项目的目的是为了方便用户切换到大模型，而不需要修改客户端代码。
    本项目的代码已经开源，欢迎大家使用。
## 配置

```
{
  "mock": false,
  "chatType":  "dify",
  "port": 8080,
  "apiURL": "https://123.com/681agentapi/api/chat-messages",
  "apiKey": "k-ant-this-is-a-test-for-cursor",
   "modelsURL": "https://freeaichatplayground.com/api/v1/models",
  "difyAppMap":{
    "claude-3-7-sonnet-latest": "123123",
    "Gemini-2.5-pro":"123123",
    "GPT-4.1":"123123"
  },
  "difyTokenUrl": "https://123123.com/681agentapi/api/passport"
}


mock 用于测试响应 一般为false
port 用于设置代理服务的端口
apiURL 用于设置目标模型服务的地址
modelsURL 用于获取代理服务的模型列表地址 openai 格式
chatType 用于选择代理的类型，目前支持 dify  claude
chatType !=  dify 时需要配置以下参数
apiKey 用于设置目标模型服务的密钥
chatType =  dify 时需要配置以下参数
difyAppMap 用于设置代理服务的模型和dify app的映射 用于获取access_token
difyTokenUrl 用于获取代理服务的token地址

```


```curl

 测试是否正常
curl -X "POST" "http://127.0.0.1:8080/api/chat" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "messages": [
    {
      "content": "你是",
      "role": "user"
    }
  ],
  "model": "claude-3-7-sonnet-latest",
  "options": {
    "num_ctx": 16384
  },
  "keep_alive": "30m",
  "stream": true
}'

```
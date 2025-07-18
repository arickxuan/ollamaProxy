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
  "model": "claude-3-7-sonnet-latest",
  "chatType":  "dify",
  "port": 8080,
  "baseUrl":"https://www.123.com/v1",
  "apiURL": "https://123.com/681agentapi/api/chat-messages",
  "apiURLProd":"http://123.com/api/chat-messages",
  "apiKey": "sk-1234",
  "difyAppMap":{
    "claude-3-7-sonnet-latest": "1234",
    "Gemini-2.5-pro":"1234",
    "GPT-4.1":"1234"
  },
  "difyAppMapProd":{
    "claude-4-sonnet-latest": "1234",
    "o3":"1234",
    "claude-4-sonnet-thinking":"1234"
  },
  "mapping":{
    "deepseek-chat": "claude-4-sonnet-latest",
    "deepseek-reasoner": "claude-4-sonnet-latest",
    "claude-sonnet-4": "claude-4-sonnet-latest",
    "deepseek-v3": "claude-4-sonnet-latest"
  },
  "difyTokenUrl": "https://123.com",
  "difyTokenUrlProd":"http://123.com",
  "modelsURL": "https://123.com/api/v1/models",
  "caFile":"cert/ca.pem",
  "caKeyFile": "cert/ca.key",
  "domain":"api.deepseek.com",
  "domainPemFile":"cert/domain.pem",
  "domainKeyFile":"cert/domain.key",
  "isTls":true,
  "debug":true,
  "proxyMapping":{
    "deepseek-chat": "claude-sonnet-4",
    "claude-4-sonnet-latest": "claude-3-7-sonnet-latest"
  },
  "oss": {
    "AccessKeyID": "1234",
    "AccessKeySecret": "1234",
    "BucketName": "re123",
    "Endpoint": "oss-cn-beijing.aliyuncs.com",
    "internal": false,
    "use_accelerate": false,
    "Region": "cn-beijing"
  }
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

### new feature
反代 deepseek 使trea 不在排队
#### 方案 一
1.1 nginx / caddy 进行反代 deepseek
根据下方命令生成证书


```
# 创建证书目录
mkdir -p cert

# 生成CA私钥
openssl genrsa -out cert/ca.key 2048

# 生成CA根证书
openssl req -new -x509 -key cert/ca.key -out cert/ca.pem -days 3650 -subj "/C=CN/ST=Beijing/L=Beijing/O=Arick/CN=OllamaProxy CA"

# 生成api.deepseek.com域名证书
openssl genrsa -out cert/api.deepseek.com.key 2048
openssl req -new -key cert/api.deepseek.com.key -out cert/api.deepseek.com.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=MyProxy/CN=api.deepseek.com"
openssl x509 -req -in cert/api.deepseek.com.csr -CA cert/ca.pem -CAkey cert/ca.key -CAcreateserial -out cert/api.deepseek.com.pem -days 3650 -extensions v3_req -extfile <(echo -e "subjectAltName=DNS:api.deepseek.com")
```

1.2 配置nginx / caddy
```
nginx
server {
    listen 443 ssl;
    server_name api.deepseek.com;

    ssl_certificate /etc/ssl/certs/api.deepseek.com.pem;
    ssl_certificate_key /etc/ssl/private/api.deepseek.com.key;

    location / {
        proxy_pass XXXXXXXXXXXXXXXXXXXXXX;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
caddy
api.deepseek.com:443 {
	tls /etc/ssl/arick.crt /go/etc/ssl/arick.key
	rewrite * /openai/v1{uri}

    reverse_proxy * http://127.0.0.1:8080 {
      header_up Host 127.0.0.1:8080
    }

    #reverse_proxy * https://www.uibers.com {
    #    header_up Host www.uibers.com
    #}

    
}
```

#### 方案 二
自动生成 ca 等文件 配置文件 设置 "isTls":true
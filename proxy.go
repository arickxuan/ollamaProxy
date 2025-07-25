package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var httpClient *http.Client

func init() {
	httpClient = &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 20 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func ProxyChatHandle(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}
	if XConfig.Debug {
		log.Println("Body", data)
	}
	if model, ok := data["model"].(string); ok {
		if originalModel, ok := XConfig.ProxyMapping[model]; ok {
			data["model"] = originalModel
			if XConfig.Debug {
				log.Printf("模型替换: %s -> %s\n", model, originalModel)
			}
		}

		// 重新编码JSON
		modifiedBody, err := json.Marshal(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode JSON"})
			return
		}

		// 构建目标URL
		targetURL := fmt.Sprintf("%s/chat/completions", XConfig.BaseUrl)

		// 创建新的请求
		req, err := http.NewRequest("POST", targetURL, bytes.NewReader(modifiedBody))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// 复制请求头，排除一些可能有问题的头部
		for key, values := range c.Request.Header {
			// 跳过这些头部，因为它们由HTTP客户端自动处理或可能引起问题
			switch key {
			case "Host", "Content-Length", "Transfer-Encoding", "Connection":
				continue
			}
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// 设置Content-Type
		req.Header.Set("Content-Type", "application/json")

		// 如果配置了API Key，替换Authorization头
		if XConfig.APIKey != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", XConfig.APIKey))
			if XConfig.Debug {
				log.Printf("使用配置的API Key替换Authorization头\n")
			}
		}
		if XConfig.Debug {
			log.Printf("转发请求到: %s\n", targetURL)
		}

		// 发送请求
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("请求失败: %v\n", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "Request failed"})
			return
		}
		defer resp.Body.Close()

		// 检查是否是流式响应
		isStream := false
		if streamValue, ok := data["stream"].(bool); ok && streamValue {
			isStream = true
		}

		// 复制响应头
		for key, values := range resp.Header {
			// 跳过这些头部，因为它们由Gin或HTTP协议自动处理
			switch key {
			case "Transfer-Encoding", "Content-Encoding", "Connection":
				continue
			}
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// 设置响应状态码
		c.Status(resp.StatusCode)

		if isStream {
			// 流式响应
			if XConfig.Debug {
				log.Println("处理流式响应")
			}
			c.Writer.Flush()

			// 使用更大的缓冲区提高性能
			buffer := make([]byte, 4096)
			for {
				n, err := resp.Body.Read(buffer)
				if n > 0 {
					if XConfig.Debug {
						log.Printf("流式响应: %d 字节\n", n)
					}
					_, writeErr := c.Writer.Write(buffer[:n])
					if writeErr != nil {
						log.Printf("写入响应错误: %v\n", writeErr)
						break
					}
					c.Writer.Flush()
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("流式读取错误: %v\n", err)
					break
				}
			}
		} else {
			// 非流式响应
			if XConfig.Debug {
				fmt.Println("处理非流式响应")
			}
			_, copyErr := io.Copy(c.Writer, resp.Body)
			if copyErr != nil {
				log.Printf("复制响应体错误: %v\n", copyErr)
			}
		}

	}
}

func TlsServer() {

	// 检查并安装CA根证书
	if err := checkAndInstallCARoot(XConfig); err != nil {
		log.Fatalf("处理CA根证书失败: %v", err)
	}

	// 检查或生成域名证书
	cert, err := checkOrGenerateCertificate(XConfig.Domain, XConfig)
	if err != nil {
		log.Fatalf("处理证书失败: %v", err)
	}
	printCertificateInfo(cert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// 创建一个 gin 的默认引擎
	r := gin.Default()

	// 定义一个简单的 GET 路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/chat/completions", ProxyChatHandle)
	r.GET("/models", GetGptModels)

	// 启动 HTTPS 服务（指定证书和密钥文件）
	// cert.pem: 你的 SSL 证书
	// key.pem: 你的 SSL 私钥
	// err := r.RunTLS(":443", XConfig.DomainPemFile, XConfig.DomainKeyFile)
	// if err != nil {
	// 	panic(err)
	// }

	// 启动 HTTPS 服务
	server := &http.Server{
		Addr:      ":443",    // 监听 HTTPS 默认端口
		Handler:   r,         // 使用 Gin 路由作为处理程序
		TLSConfig: tlsConfig, // 应用自定义的 TLS 配置
	}

	log.Println("Starting HTTPS server on https://localhost:443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("failed to start server: %s", err)
	}
}

package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/gin-gonic/gin"
)

// OSSConfig OSS配置结构体
type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	Region          string
}

func NewOSSConfig(endpoint, accessKeyID, accessKeySecret, bucketName, region string) *OSSConfig {
	return &OSSConfig{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		BucketName:      bucketName,
		Region:          region,
	}
}
func TestConfig() *OSSConfig {
	config := &XConfig.OSSConfig
	return config
}

func (o *OSSConfig) Provider() credentials.CredentialsProvider {
	return &OSSConfig{
		Endpoint:        o.Endpoint,
		AccessKeyID:     o.AccessKeyID,
		AccessKeySecret: o.AccessKeySecret,
		BucketName:      o.BucketName,
		Region:          o.Region,
	}
}

func (o *OSSConfig) GetCredentials(ctx context.Context) (credentials.Credentials, error) {
	t := time.Now().Add(time.Second * 20)
	crd := credentials.Credentials{
		AccessKeyID:     o.AccessKeyID,
		AccessKeySecret: o.AccessKeySecret,
		SecurityToken:   "",
		Expires:         &t,
	}
	return crd, nil
}

// UploadFile 上传文件到OSS
func UploadFile(config OSSConfig, filePath string, objectName string) (string, error) {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建OSS客户端
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(config.Provider()).
		WithRegion(config.Region)
	client := oss.NewClient(cfg)

	// 打开本地文件并获取文件信息
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 创建上传请求
	request := &oss.PutObjectRequest{
		Bucket:        oss.Ptr(config.BucketName),
		Key:           oss.Ptr(objectName),
		Body:          file,
		ContentLength: oss.Ptr(fileInfo.Size()),
	}

	// 执行上传
	log.Printf("开始上传文件: %s (大小: %d bytes)", filePath, fileInfo.Size())
	result, err := client.PutObject(ctx, request)
	if err != nil {
		return "", fmt.Errorf("上传文件到OSS失败: %v", err)
	}

	// 构造访问URL
	url := fmt.Sprintf("https://%s.%s/%s", config.BucketName, config.Endpoint, objectName)
	log.Printf("文件上传成功: %s, 访问URL: %s", filePath, url, result)

	return url, nil
}

func Download(config OSSConfig, objectName string) {
	bucketName := config.BucketName
	region := config.Region
	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithRegion(region).WithCredentialsProvider(config.Provider())

	// 创建OSS客户端
	client := oss.NewClient(cfg)

	// 创建获取对象的请求
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(bucketName), // 存储空间名称
		Key:    oss.Ptr(objectName), // 对象名称
	}

	// 执行获取对象的操作并处理结果
	result, err := client.GetObject(context.TODO(), request)
	if err != nil {
		log.Fatalf("failed to get object %v", err)
	}
	defer result.Body.Close() // 确保在函数结束时关闭响应体

	log.Printf("get object result:%#v\n", result)

	// 读取对象的内容
	data, _ := io.ReadAll(result.Body)
	log.Printf("body:%s\n", data)
}

type ConfigStruct struct {
	Expiration string          `json:"expiration"`
	Conditions [][]interface{} `json:"conditions"`
}

type PolicyToken struct {
	AccessKeyId string `json:"ossAccessKeyId"`
	Host        string `json:"host"`
	Signature   string `json:"signature"`
	Policy      string `json:"policy"`
	Directory   string `json:"dir"`
}

func getGMTISO8601(expireEnd int64) string {
	return time.Unix(expireEnd, 0).UTC().Format("2006-01-02T15:04:05Z")
}

func getPolicyToken(configs OSSConfig) string {
	// 指定过期时间，单位为秒。
	expireTime := int64(3600)
	uploadDir := "/upload/"
	now := time.Now().Unix()
	expireEnd := now + expireTime
	tokenExpire := getGMTISO8601(expireEnd)

	var config ConfigStruct
	config.Expiration = tokenExpire

	// 添加文件前缀限制
	config.Conditions = append(config.Conditions, []interface{}{"starts-with", "$key", uploadDir})

	// 添加文件大小限制，例如1KB到10MB
	minSize := int64(1024)
	maxSize := int64(10 * 1024 * 1024)
	config.Conditions = append(config.Conditions, []interface{}{"content-length-range", minSize, maxSize})

	result, err := json.Marshal(config)
	if err != nil {
		fmt.Println("callback json err:", err)
		return ""
	}

	encodedResult := base64.StdEncoding.EncodeToString(result)
	h := hmac.New(sha1.New, []byte(configs.AccessKeySecret))
	io.WriteString(h, encodedResult)
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))

	policyToken := PolicyToken{
		AccessKeyId: configs.AccessKeyID,
		Host:        configs.Endpoint,
		Signature:   signedStr,
		Policy:      encodedResult,
		Directory:   uploadDir,
	}

	response, err := json.Marshal(policyToken)
	if err != nil {
		fmt.Println("json err:", err)
		return ""
	}

	return string(response)
}

/**
		let file = fileInput.files[0];
        let filename = fileInput.files[0].name;
        fetch('/get_post_signature_for_oss_upload', { method: 'GET' })
          .then(response => response.json())
          .then(data => {
            const formData = new FormData();
            formData.append('name',filename);
            formData.append('policy', data.policy);
            formData.append('OSSAccessKeyId', data.ossAccessKeyId);
            formData.append('success_action_status', '200');
            formData.append('signature', data.signature);
            formData.append('key', data.dir + filename);
            // file必须为最后一个表单域，除file以外的其他表单域无顺序要求。
            formData.append('file', file);
            fetch(data.host, { method: 'POST', body: formData},).then((res) => {
              console.log(res);
              alert('文件已上传');
            });
          })

		  **/

func ListBucket(config OSSConfig) {
	// 加载默认配置并设置凭证提供者和区域
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(config.Provider()).
		WithRegion(config.Region)
	client := oss.NewClient(cfg)

	// 创建列出存储空间的请求
	request := &oss.ListBucketsRequest{}

	// 定义一个函数来处理 PaginatorOptions
	modifyOptions := func(opts *oss.PaginatorOptions) {
		// 在这里可以修改opts的值，比如设置每页返回的存储空间数量上限
		// 示例：opts.Limit = 5，即每页返回5个存储空间
		opts.Limit = 5
	}

	// 创建分页器
	p := client.NewListBucketsPaginator(request, modifyOptions)

	var i int
	log.Println("Buckets:")

	// 遍历分页器中的每一页
	for p.HasNext() {
		i++

		// 获取下一页的数据
		page, err := p.NextPage(context.TODO())
		if err != nil {
			log.Fatalf("failed to get page %v, %v", i, err)
		}

		// 打印该页中的每个存储空间的信息
		for _, b := range page.Buckets {
			log.Printf("Bucket: %v, StorageClass: %v, Location: %v\n", oss.ToString(b.Name), oss.ToString(b.StorageClass), oss.ToString(b.Location))
		}
	}
}

// curl -X POST -F "file=@./iShot_59.png" http://localhost:8080/upload/oss
// curl -X POST -F "files=@file1.txt" -F "files=@file2.txt" http://localhost:8080/upload/oss
func Upload(c *gin.Context) {

	// 从请求中获取多个文件
	form, _ := c.MultipartForm()
	files := form.File["files"]

	// 遍历并保存每个文件
	if len(files) > 0 {

		for _, file := range files {
			log.Printf("正在保存文件: %s (大小: %d 字节)\n", file.Filename, file.Size)
			os.Mkdir("uploads", os.ModePerm)
			savePath := filepath.Join("uploads", file.Filename)

			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件时出错!", "file": file.Filename})
				return
			}
		}
		c.JSON(200, gin.H{"status": "success"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "请提供有效的文件"})
		return
	}

	// 创建临时文件
	tempFile, err := os.CreateTemp("", "upload-*.tmp")
	if err != nil {
		c.JSON(500, gin.H{"error": "无法创建临时文件"})
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// 保存上传文件到临时文件
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "打开上传文件失败",
		})
		return
	}
	defer src.Close()

	if _, err = io.Copy(tempFile, src); err != nil {
		c.JSON(500, gin.H{
			"error": "保存文件失败",
		})
		return
	}

	// 获取配置
	config := TestConfig()

	// 上传到OSS
	objectName := c.PostForm("objectName")
	if objectName == "" {
		objectName = file.Filename
	}

	objectName = "uploads/temp/" + objectName

	url, err := UploadFile(*config, tempFile.Name(), objectName)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, map[string]interface{}{
		"url":  url,
		"size": file.Size,
		"name": file.Filename,
	})
}

func OssList(c *gin.Context) {
	config := TestConfig()
	ListBucket(*config)

	c.JSON(200, gin.H{"status": "success"})
}

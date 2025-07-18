package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
	"software.sslmate.com/src/go-pkcs12"
)

func checkOrGenerateCertificate(domain string, config *Config) (tls.Certificate, error) {
	certFile := filepath.Join("cert", domain+".crt")
	keyFile := filepath.Join("cert", domain+".key")

	// 检查证书文件是否存在
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			fmt.Printf("使用现有证书: %s\n", certFile)
			return tls.LoadX509KeyPair(certFile, keyFile)
		}
	} else {

		_, err := generateCAPrivateKey(config.CAKeyFile)
		log.Println(":", err)
		err = generateCACertificate(config.CAKeyFile, config.CAFile)
		log.Println(":", err)
	}
	err := validateKeyPair(config.CAFile, config.CAKeyFile)
	if err != nil {
		log.Println(":", err)
	}

	fmt.Printf("为域名 %s 在内存中生成证书\n", domain)
	return generateCertificate(domain, config, true, nil)
}

func generateCertificate(domain string, config *Config, genFile bool, domains []string) (tls.Certificate, error) {
	// 生成私钥
	var privateKey *rsa.PrivateKey
	var err error
	if genFile {
		privateKey, err = generateCAPrivateKey(config.DomainKeyFile)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("路径: %s,生成私钥失败: %v", config.DomainKeyFile, err)
		}
	} else {
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	}

	domainArr := []string{domain}
	domainArr = append(domainArr, domains...)
	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Proxy Server"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    domain,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:    domainArr,
	}

	// 从嵌入文件系统或指定路径加载CA证书和私钥
	var caCertPEM, caKeyPEM []byte

	// 优先使用命令行指定的CA文件
	if config.CAFile != "" && config.CAFile != "cert/ca.pem" {
		caCertPEM, err = os.ReadFile(config.CAFile)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("读取指定CA证书失败: %v", err)
		}

		caKeyPEM, err = os.ReadFile(config.CAKeyFile)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("读取指定CA私钥失败: %v", err)
		}
	} else {
		// 回退到嵌入文件系统
		caCertPEM, err = certFS.ReadFile("cert/ca.pem")
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("读取嵌入的CA证书失败: %v", err)
		}

		caKeyPEM, err = certFS.ReadFile("cert/ca.key")
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("读取嵌入的CA私钥失败: %v", err)
		}
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return tls.Certificate{}, fmt.Errorf("解析CA证书PEM格式失败")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("解析CA证书失败: %v", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return tls.Certificate{}, fmt.Errorf("解析CA私钥PEM格式失败")
	}
	caKeyInterface, err := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		// 尝试PKCS1格式作为备选
		caKeyInterface, err = x509.ParsePKCS1PrivateKey(caKeyBlock.Bytes)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("解析CA私钥失败: %v", err)
		}
	}

	caKey, ok := caKeyInterface.(*rsa.PrivateKey)
	if !ok {
		return tls.Certificate{}, fmt.Errorf("CA私钥不是RSA格式")
	}

	// 使用CA签名生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("生成证书失败: %v", err)
	}

	// 将证书和私钥转换为PEM格式
	certPEMBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}
	certPEM := pem.EncodeToMemory(certPEMBlock)
	if genFile {
		// 创建证书文件
		certFile, err := os.Create(config.DomainPemFile)
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("创建证书文件失败: %v", err)
		}
		defer certFile.Close()

		if err := pem.Encode(certFile, certPEMBlock); err != nil {
			return tls.Certificate{}, fmt.Errorf("写入证书文件失败: %v", err)
		}
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	// 从内存中加载证书
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("加载内存证书失败: %v", err)
	}

	fmt.Printf("成功在内存中生成域名 %s 的证书\n", domain)
	return cert, nil
}

func checkAndInstallCARoot(config *Config) error {
	var caCertPEM []byte
	var err error

	// 优先使用命令行指定的CA文件
	if config.CAFile != "" && config.CAFile != "cert/ca.pem" {
		caCertPEM, err = os.ReadFile(config.CAFile)
		if err != nil {
			return fmt.Errorf("读取指定CA证书失败: %v", err)
		}
	} else {
		// 优先从嵌入文件系统读取CA证书
		caCertPEM, err = certFS.ReadFile("cert/ca.pem")
		if err != nil {
			// 回退到本地文件系统
			caCertFile := filepath.Join("cert", "ca.pem")
			if _, err := os.Stat(caCertFile); os.IsNotExist(err) {
				return fmt.Errorf("CA证书文件不存在: %s", caCertFile)
			}

			caCertPEM, err = os.ReadFile(caCertFile)
			if err != nil {
				return fmt.Errorf("读取CA证书失败: %v", err)
			}
		}
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return fmt.Errorf("解析CA证书PEM格式失败")
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析CA证书失败: %v", err)
	}

	// 检查CA证书是否已在系统中被信任
	if isCAInstalled(caCert) {
		fmt.Println("CA证书已在系统中被信任")
		return nil
	}

	fmt.Println("CA证书未被信任，需要安装到系统钥匙串...")
	return installCAToKeychainFromMemory(caCertPEM)
}

func isCAInstalled(caCert *x509.Certificate) bool {
	// 使用security命令检查证书是否已安装
	cmd := exec.Command("security", "find-certificate", "-c", caCert.Subject.CommonName, "/System/Library/Keychains/SystemRootCertificates.keychain")
	err := cmd.Run()
	if err == nil {
		return true
	}

	// 检查登录钥匙串
	cmd = exec.Command("security", "find-certificate", "-c", caCert.Subject.CommonName)
	err = cmd.Run()
	return err == nil
}

func installCAToKeychainFromMemory(caCertPEM []byte) error {
	fmt.Print("需要管理员权限安装CA证书，请输入密码: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("读取密码失败: %v", err)
	}
	fmt.Println()

	// 创建临时文件
	//tmpFile, err := os.Create("ca-cert-*.pem")
	tmpFile, err := os.CreateTemp("", "ca-cert-*.pem")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 写入证书内容
	if _, err := tmpFile.Write(caCertPEM); err != nil {
		return fmt.Errorf("写入临时文件失败: %v", err)
	}
	tmpFile.Close()

	// 安装证书到系统钥匙串
	cmd := exec.Command("sudo", "-S", "security", "add-trusted-cert", "-d", "-r", "trustRoot", "-k", "/Library/Keychains/System.keychain", tmpFile.Name())
	cmd.Stdin = strings.NewReader(string(password) + "\n")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("安装CA证书失败: %v, 输出: %s", err, string(output))
	}

	fmt.Printf("成功安装CA证书到系统钥匙串\n")
	return nil
}

func printCertificateInfo(cert tls.Certificate) {
	// 解析证书以获取信息
	if len(cert.Certificate) > 0 {
		x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
		if err == nil {
			fmt.Printf("  证书状态: 已加载 (CN=%s)\n", x509Cert.Subject.CommonName)
		} else {
			fmt.Printf("  证书状态: 已加载\n")
		}
	}
}

// generateCAPrivateKey 生成CA私钥，等价于 openssl genrsa -out cert/ca.key 2048
func generateCAPrivateKey(outputPath string) (*rsa.PrivateKey, error) {
	// 生成2048位RSA私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成RSA私钥失败: %v", err)
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return privateKey, fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建输出文件
	keyFile, err := os.Create(outputPath)
	if err != nil {
		return privateKey, fmt.Errorf("创建私钥文件失败: %v", err)
	}
	defer keyFile.Close()

	// 将私钥编码为PEM格式并写入文件
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	if err := pem.Encode(keyFile, privateKeyPEM); err != nil {
		return privateKey, fmt.Errorf("写入私钥文件失败: %v", err)
	}

	fmt.Printf("成功生成CA私钥: %s\n", outputPath)
	return privateKey, nil
}

// generateCACertificate 生成自签名CA证书，等价于 openssl req -new -x509 -key cert/ca.key -out cert/ca.pem -days 3650 -subj "/C=CN/ST=Beijing/L=Beijing/O=Arick/CN=OllamaProxy CA"
func generateCACertificate(keyPath, certPath string) error {
	// 读取私钥文件
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("读取私钥文件失败: %v", err)
	}

	// 解析私钥
	keyBlock, _ := pem.Decode(keyData)
	if keyBlock == nil {
		return fmt.Errorf("解析私钥PEM格式失败")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析RSA私钥失败: %v", err)
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
			Organization: []string{"Arick"},
			CommonName:   "OllamaProxy CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(3650 * 24 * time.Hour), // 3650天
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// 生成自签名证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("生成证书失败: %v", err)
	}

	// 确保输出目录存在
	dir := filepath.Dir(certPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建证书文件
	certFile, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("创建证书文件失败: %v", err)
	}
	defer certFile.Close()

	// 将证书编码为PEM格式并写入文件
	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	}

	if err := pem.Encode(certFile, certPEM); err != nil {
		return fmt.Errorf("写入证书文件失败: %v", err)
	}

	fmt.Printf("成功生成CA证书: %s\n", certPath)
	return nil
}

// validateKeyPair 校验公私钥是否匹配
// 支持从文件路径或PEM字节数据校验
func validateKeyPair(certPath, keyPath string) error {
	// 读取证书文件
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("读取证书文件失败: %v", err)
	}

	// 读取私钥文件
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("读取私钥文件失败: %v", err)
	}

	return validateKeyPairFromPEM(certData, keyData)
}

// validateKeyPairFromPEM 从PEM格式的字节数据校验公私钥是否匹配
func validateKeyPairFromPEM(certPEM, keyPEM []byte) error {
	// 解析证书
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("解析证书PEM格式失败")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析证书失败: %v", err)
	}

	// 解析私钥
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("解析私钥PEM格式失败")
	}

	var privateKey interface{}

	// 尝试不同的私钥格式
	switch keyBlock.Type {
	case "ENCRYPTED PRIVATE KEY":
		// 处理加密私钥
		password, err := promptForPassword("私钥")
		if err != nil {
			return fmt.Errorf("获取密码失败: %v", err)
		}
		defer clearPassword(password) // 确保密码被清除

		privateKey, err = parseEncryptedPrivateKey(keyBlock.Bytes, password)
		if err != nil {
			return fmt.Errorf("解析加密私钥失败: %v", err)
		}
		log.Println("privateKey: ", privateKey)
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	default:
		return fmt.Errorf("不支持的私钥类型: %s", keyBlock.Type)
	}

	if err != nil {
		return fmt.Errorf("解析私钥失败: %v", err)
	}

	// 校验公私钥是否匹配
	return validatePublicPrivateKeyMatch(cert.PublicKey, privateKey)
}

// validatePublicPrivateKeyMatch 校验公钥和私钥是否匹配
func validatePublicPrivateKeyMatch(publicKey, privateKey interface{}) error {
	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		priv, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return fmt.Errorf("公钥是RSA类型，但私钥不是RSA类型")
		}

		// 比较RSA公钥的N和E值
		if pub.N.Cmp(priv.N) != 0 || pub.E != priv.E {
			return fmt.Errorf("RSA公私钥不匹配")
		}

		fmt.Println("RSA公私钥匹配验证成功")
		return nil

	default:
		return fmt.Errorf("暂不支持的公钥类型: %T", publicKey)
	}
}

// promptForPassword 安全地提示用户输入密码
func promptForPassword(keyType string) ([]byte, error) {
	fmt.Printf("检测到加密的%s，请输入密码: ", keyType)
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, fmt.Errorf("读取密码失败: %v", err)
	}
	fmt.Println() // 换行
	return password, nil
}

// clearPassword 安全地清除密码内存
func clearPassword(password []byte) {
	for i := range password {
		password[i] = 0
	}
}

// parseEncryptedPrivateKey 解析加密的私钥
func parseEncryptedPrivateKey(encryptedData []byte, password []byte) (interface{}, error) {
	// 使用 x509.DecryptPEMBlock 解密
	decryptedBlock, err := x509.DecryptPEMBlock(&pem.Block{
		Type:  "ENCRYPTED PRIVATE KEY",
		Bytes: encryptedData,
	}, password)
	if err != nil {
		return nil, fmt.Errorf("解密私钥失败，可能是密码错误: %v", err)
	}

	// 尝试解析解密后的私钥
	privateKey, err := x509.ParsePKCS8PrivateKey(decryptedBlock)
	if err != nil {
		// 如果PKCS8失败，尝试PKCS1
		privateKey, err = x509.ParsePKCS1PrivateKey(decryptedBlock)
		if err != nil {
			return nil, fmt.Errorf("解析解密后的私钥失败: %v", err)
		}
	}

	return privateKey, nil
}

// convertP12ToPEM 将P12文件转换为PEM格式的公钥和私钥
// p12Path: P12文件路径
// password: P12文件密码
// certOutputPath: 输出证书文件路径
// keyOutputPath: 输出私钥文件路径
func convertP12ToPEM(p12Path, password, certOutputPath, keyOutputPath string) error {
	// 读取P12文件
	p12Data, err := os.ReadFile(p12Path)
	if err != nil {
		return fmt.Errorf("读取P12文件失败: %v", err)
	}

	// 解析P12文件
	privateKey, certificate, caCerts, err := pkcs12.DecodeChain(p12Data, password)
	if err != nil {
		return fmt.Errorf("解析P12文件失败，可能是密码错误: %v", err)
	}

	// 确保输出目录存在
	if certOutputPath != "" {
		if err := os.MkdirAll(filepath.Dir(certOutputPath), 0755); err != nil {
			return fmt.Errorf("创建证书输出目录失败: %v", err)
		}
	}
	if keyOutputPath != "" {
		if err := os.MkdirAll(filepath.Dir(keyOutputPath), 0755); err != nil {
			return fmt.Errorf("创建私钥输出目录失败: %v", err)
		}
	}

	// 保存证书到PEM文件
	if certOutputPath != "" && certificate != nil {
		certFile, err := os.Create(certOutputPath)
		if err != nil {
			return fmt.Errorf("创建证书文件失败: %v", err)
		}
		defer certFile.Close()

		// 编码主证书
		certPEM := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certificate.Raw,
		}
		if err := pem.Encode(certFile, certPEM); err != nil {
			return fmt.Errorf("写入证书文件失败: %v", err)
		}

		// 如果有CA证书链，也一并写入
		for _, caCert := range caCerts {
			caCertPEM := &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: caCert.Raw,
			}
			if err := pem.Encode(certFile, caCertPEM); err != nil {
				return fmt.Errorf("写入CA证书失败: %v", err)
			}
		}

		fmt.Printf("成功保存证书到: %s\n", certOutputPath)
	}

	// 保存私钥到PEM文件
	if keyOutputPath != "" && privateKey != nil {
		keyFile, err := os.Create(keyOutputPath)
		if err != nil {
			return fmt.Errorf("创建私钥文件失败: %v", err)
		}
		defer keyFile.Close()

		// 根据私钥类型进行编码
		var keyPEM *pem.Block
		switch key := privateKey.(type) {
		case *rsa.PrivateKey:
			keyPEM = &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(key),
			}
		default:
			// 使用PKCS8格式作为通用格式
			keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
			if err != nil {
				return fmt.Errorf("编码私钥失败: %v", err)
			}
			keyPEM = &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: keyBytes,
			}
		}

		if err := pem.Encode(keyFile, keyPEM); err != nil {
			return fmt.Errorf("写入私钥文件失败: %v", err)
		}

		fmt.Printf("成功保存私钥到: %s\n", keyOutputPath)
	}

	return nil
}

// convertP12ToPEMBytes 将P12文件转换为PEM格式的字节数据
// 返回证书PEM数据和私钥PEM数据
func convertP12ToPEMBytes(p12Data []byte, password string) (certPEM, keyPEM []byte, err error) {
	// 解析P12文件
	privateKey, certificate, caCerts, err := pkcs12.DecodeChain(p12Data, password)
	if err != nil {
		return nil, nil, fmt.Errorf("解析P12文件失败，可能是密码错误: %v", err)
	}

	// 编码证书为PEM格式
	if certificate != nil {
		var certBuffer []byte

		// 编码主证书
		certBlock := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certificate.Raw,
		}
		certBuffer = append(certBuffer, pem.EncodeToMemory(certBlock)...)

		// 如果有CA证书链，也一并编码
		for _, caCert := range caCerts {
			caCertBlock := &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: caCert.Raw,
			}
			certBuffer = append(certBuffer, pem.EncodeToMemory(caCertBlock)...)
		}

		certPEM = certBuffer
	}

	// 编码私钥为PEM格式
	if privateKey != nil {
		var keyBlock *pem.Block

		switch key := privateKey.(type) {
		case *rsa.PrivateKey:
			keyBlock = &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(key),
			}
		default:
			// 使用PKCS8格式作为通用格式
			keyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
			if err != nil {
				return nil, nil, fmt.Errorf("编码私钥失败: %v", err)
			}
			keyBlock = &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: keyBytes,
			}
		}

		keyPEM = pem.EncodeToMemory(keyBlock)
	}

	return certPEM, keyPEM, nil
}

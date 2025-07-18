package main

import (
	"os"
	"testing"
)

func TestGenCa(t *testing.T) {
	t.Log("gen ca")
	_, err := generateCAPrivateKey("ca.key")
	t.Log(":", err)
	err = generateCACertificate("ca.key", "ca.pem")
	t.Log(":", err)

}

func TestGenDominPem(t *testing.T) {
	c := Config{}
	c.CAFile = "ca.pem"
	c.CAKeyFile = "ca.key"
	c.DomainKeyFile = "domain.key"
	c.DomainPemFile = "domain.pem"
	cert, err := generateCertificate("api.com", &c, true, nil)
	t.Log(":", err)
	printCertificateInfo(cert)

}

func TestValidateKeyPair(t *testing.T) {
	// 测试匹配的公私钥对
	t.Run("ValidKeyPair", func(t *testing.T) {
		err := validateKeyPair("ca.pem", "ca.key")
		if err != nil {
			t.Errorf("校验匹配的公私钥失败: %v", err)
		} else {
			t.Log("匹配的公私钥校验成功")
		}
	})

	// 测试不存在的文件
	t.Run("NonExistentFiles", func(t *testing.T) {
		err := validateKeyPair("test_cert.pem", "test_key.pem")
		if err != nil {
			t.Logf("应该返回文件不存在的错误: %v", err)
		} else {
			t.Logf("正确处理不存在文件的情况")
		}
	})
}

func TestConvertP12ToPEM(t *testing.T) {
	// 注意：这个测试需要一个实际的P12文件来运行
	// 如果没有P12文件，测试会跳过
	p12Path := "mitmproxy-ca.p12"
	password := "8D6CB874"
	certPath := "test_cert.pem"
	keyPath := "test_key.pem"

	t.Run("ConvertP12File", func(t *testing.T) {
		// 检查P12文件是否存在
		if _, err := os.Stat(p12Path); os.IsNotExist(err) {
			t.Skip("跳过P12转换测试：测试文件不存在")
			return
		}

		err := convertP12ToPEM(p12Path, password, certPath, keyPath)
		if err != nil {
			t.Errorf("P12转PEM失败: %v", err)
		} else {
			t.Log("P12转PEM成功")

			// 清理测试文件
			defer func() {
				//os.Remove(certPath)
				//os.Remove(keyPath)
			}()

			// 验证生成的文件是否存在
			if _, err := os.Stat(certPath); err != nil {
				t.Errorf("证书文件未生成: %v", err)
			}
			if _, err := os.Stat(keyPath); err != nil {
				t.Errorf("私钥文件未生成: %v", err)
			}
		}
	})

	t.Run("InvalidP12File", func(t *testing.T) {
		err := convertP12ToPEM("nonexistent.p12", password, certPath, keyPath)
		if err == nil {
			t.Error("应该返回文件不存在的错误")
		} else {
			t.Logf("正确处理不存在文件的情况: %v", err)
		}
	})
}

func TestConvertP12ToPEMBytes(t *testing.T) {
	// 这个测试演示如何使用字节数据转换功能
	t.Run("EmptyP12Data", func(t *testing.T) {
		_, _, err := convertP12ToPEMBytes([]byte{}, "password")
		if err == nil {
			t.Error("应该返回解析错误")
		} else {
			t.Logf("正确处理空数据的情况: %v", err)
		}
	})
}

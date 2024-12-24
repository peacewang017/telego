package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func GenerateCrtKey(directory string, name string) error {
	// 创建私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 设置证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Telego!!!"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0), // 有效期为10年

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// 生成证书
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// 确保目录存在
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return err
	}

	// Helper function to write PEM blocks to file
	writePEMFile := func(filename string, block *pem.Block, mode os.FileMode) error {
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		defer file.Close()
		return pem.Encode(file, block)
	}

	// 将私钥转换为PEM格式并保存
	privateKeyFile := filepath.Join(directory, name+".key")
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := writePEMFile(privateKeyFile, privateKeyPEM, 0600); err != nil {
		return err
	}

	// 将证书转换为PEM格式并保存
	certFile := filepath.Join(directory, name+".crt")
	certPEM := &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}
	if err := writePEMFile(certFile, certPEM, 0644); err != nil {
		return err
	}

	log.Println("certificate and private key generated successfully")
	return nil
}

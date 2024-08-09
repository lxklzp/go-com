package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"go-com/config"
	"math/big"
	"os"
	"time"
)

/***** 非对称加密 *****/

// GenerateRsaKey 生成rsa公钥和私钥
func GenerateRsaKey(bits int, rsaPriKeyFile string, rsaPubKeyFile string) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		panic(err)
	}
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	privateFile, err := os.Create(rsaPriKeyFile)
	if err != nil {
		panic(err)
	}
	defer privateFile.Close()
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	pem.Encode(privateFile, &privateBlock)

	// 生成公钥
	publicKey := privateKey.PublicKey
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		panic(err)
	}
	publicFile, err := os.Create(rsaPubKeyFile)
	if err != nil {
		panic(err)
	}
	defer publicFile.Close()
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	pem.Encode(publicFile, &publicBlock)
}

// RsaEncrypt rsa加密
func RsaEncrypt(plain []byte, publicKeyFilename string) (string, error) {
	buf, err := os.ReadFile(publicKeyFilename)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(buf)
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	cipherByte, err := rsa.EncryptPKCS1v15(rand.Reader, publicKeyInterface.(*rsa.PublicKey), plain)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cipherByte), nil
}

// RsaDecrypt rsa解密
func RsaDecrypt(ciphertext string, privateKeyFilename string) ([]byte, error) {
	cipherByte, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	buf, err := os.ReadFile(privateKeyFilename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(buf)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherByte)
}

// GenerateCert 生成https证书
func GenerateCert(dnsNames []string) error {
	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	// x509签证，终端证书
	template := x509.Certificate{
		Version:      1,
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"公司"}, // 证书持有者组织名称
		},
		DNSNames:  dnsNames,
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 50),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		return errors.New("failed to encode certificate to PEM")
	}
	if err = os.WriteFile(config.RuntimePath+"/cert.pem", pemCert, 0644); err != nil {
		return err
	}

	privateBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateBytes})
	if pemKey == nil {
		return errors.New("failed to encode private key to PEM")
	}
	if err = os.WriteFile(config.RuntimePath+"/key.pem", pemKey, 0600); err != nil {
		return err
	}
	return nil
}

/***** 对称加密 *****/

// PKCS7填充
func pkcs7Padding(plain []byte, blockSize int) []byte {
	padding := blockSize - len(plain)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	//padText := bytes.Repeat([]byte{'\n'}, padding)
	return append(plain, padText...)
}

// PKCS7清除
func pkcs7UnPadding(crypted []byte) []byte {
	length := len(crypted)
	unPadding := int(crypted[length-1])
	return crypted[:(length - unPadding)]
}

// TripleDesEncrypt 3des加密 CBC key长度32个字符，表示256 iv固定长度16个字符
func TripleDesEncrypt(plain []byte, key []byte, iv []byte) (string, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	encryptBytes := pkcs7Padding(plain, blockSize)

	// CBC加密
	crypted := make([]byte, len(encryptBytes))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(crypted, encryptBytes)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

// TripleDesDecrypt 3des解密 CBC key长度32个字符，表示256 iv固定长度16个字符
func TripleDesDecrypt(ciphertext string, key []byte, iv []byte) ([]byte, error) {
	cipherByte, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}

	// CBC解密
	blockMode := cipher.NewCBCDecrypter(block, iv)
	crypted := make([]byte, len(cipherByte))
	blockMode.CryptBlocks(crypted, cipherByte)

	return pkcs7UnPadding(crypted), nil
}

// AesEncrypt aes加密 CBC key长度32个字符，表示256 iv固定长度16个字符
func AesEncrypt(plain, key []byte, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	encryptBytes := pkcs7Padding(plain, blockSize)

	// CBC加密
	crypted := make([]byte, len(encryptBytes))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(crypted, encryptBytes)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

// AesDecrypt 3des解密 CBC key长度32个字符，表示256 iv固定长度16个字符
func AesDecrypt(ciphertext string, key []byte, iv []byte) ([]byte, error) {
	cipherByte, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// CBC解密
	blockMode := cipher.NewCBCDecrypter(block, iv)
	crypted := make([]byte, len(cipherByte))
	blockMode.CryptBlocks(crypted, cipherByte)

	return pkcs7UnPadding(crypted), nil
}

/***** 编码 *****/

// Md5Encrypt md5编码
func Md5Encrypt(plain string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(plain)))
}

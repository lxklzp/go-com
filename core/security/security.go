package security

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
)

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

func TripleDesEncrypt(plain []byte, key []byte, iv []byte) (string, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()

	// pkcs7Padding
	padding := blockSize - len(plain)%blockSize
	//padText := bytes.Repeat([]byte{byte(padding)}, padding)
	padText := bytes.Repeat([]byte{'\n'}, padding)
	encryptBytes := append(plain, padText...)

	// CBC加密
	crypted := make([]byte, len(encryptBytes))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(crypted, encryptBytes)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

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

	// pkcs7UnPadding
	length := len(crypted)
	unPadding := int(crypted[length-1])
	return crypted[:(length - unPadding)], nil
}

package utils

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"io"
	"os"
)

/**
 *	生成一对公私钥
 */
func GenKeys() (pubKey, priKey []byte) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	//通过x509标准将得到的ras私钥序列化为ASN.1的DER编码字符串
	x509_Privatekey := x509.MarshalPKCS1PrivateKey(privateKey)
	//将私钥字符串设置到pem格式块中
	pem_block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509_Privatekey,
	}
	priKey = pem.EncodeToMemory(&pem_block)

	//处理公钥,公钥包含在私钥中
	publickKey := privateKey.PublicKey
	//通过x509标准将得到的rsa公钥序列化为ASN.1 的 DER编码字符串
	x509_PublicKey, _ := x509.MarshalPKIXPublicKey(&publickKey)
	pem_PublickKey := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509_PublicKey,
	}
	pubKey = pem.EncodeToMemory(&pem_PublickKey)
	return
}

/**
 *	生成一对公私钥，保存到文件publicFile和privateFile中
 */
func GenKeyFiles(publicFile, privateFile string) error {
	pubKey, priKey := GenKeys()

	if err := os.WriteFile(publicFile, pubKey, 0640); err != nil {
		return err
	}
	if err := os.WriteFile(privateFile, priKey, 0640); err != nil {
		return err
	}
	return nil
}

/**
 *	使用公钥进行加密
 */
func RsaEncrypt(pubKey []byte, msg []byte) []byte {
	block, _ := pem.Decode(pubKey)
	//x509解码,得到一个interface类型的pub
	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	//加密操作,需要将接口类型的pub进行类型断言得到公钥类型
	cipherText, _ := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), msg)
	return cipherText
}

/**
 *	使用私钥进行解密
 */
func RsaDecrypt(priKey []byte, cipherText []byte) []byte {
	block, _ := pem.Decode(priKey)
	PrivateKey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	//二次解码完毕，调用解密函数
	decrypted, _ := rsa.DecryptPKCS1v15(rand.Reader, PrivateKey, cipherText)
	return decrypted
}

/**
 *	使用私钥签名，priKey是私钥，msg是要签名的信息
 */
func Sign(priKey []byte, msg []byte) ([]byte, error) {
	block, _ := pem.Decode(priKey)
	PrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return []byte{}, err
	}
	//加密操作,需要将接口类型的pub进行类型断言得到公钥类型
	hash := sha256.Sum256(msg)
	//调用签名函数,填入所需四个参数，得到签名
	sign, err := rsa.SignPKCS1v15(rand.Reader, PrivateKey, crypto.SHA256, hash[:])
	return sign, err
}

/**
 *	使用公钥pubKey和签名signText校验消息plainText的完整性
 */
func VerifySign(pubKey []byte, signText []byte, plainText []byte) error {
	block, _ := pem.Decode(pubKey)
	//x509解码,得到一个interface类型的pub
	pub, _ := x509.ParsePKIXPublicKey(block.Bytes)
	//签名函数中需要的数据散列值
	hash := sha256.Sum256(plainText)
	//验证签名
	return rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA256, hash[:], signText)
}

/**
 *	获取文件信息(大小及MD5)
 */
func CalcFileMd5(fpath string) (uint64, string, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()
	finfo, err := file.Stat()
	if err != nil {
		return 0, "", err
	}
	buf := make([]byte, 1024*1024)
	md5s := md5.New()
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, "", err
		}
		md5s.Write(buf[:n])
	}
	sum := md5s.Sum([]byte{})
	md5str := hex.EncodeToString(sum[:])
	return uint64(finfo.Size()), md5str, err
}

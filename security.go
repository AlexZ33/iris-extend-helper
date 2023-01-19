package iris_extend_helper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	jsoniter "github.com/json-iterator/go"
	"github.com/pelletier/go-toml"

	"io"
	"log"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

func Id() string {
	id, err := uuid.NewRandom()
	if err != nil {
		log.Println(err)
	}
	return id.String()
}

func CheckId(s string) bool {
	if _, err := uuid.Parse(s); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func Key(phrase string) string {
	return HS256(phrase, Base64Encode(phrase))
}

func NormalizeId(id string) string {
	return strings.ReplaceAll(id, "-", "")
}

func AccessId(id string, key string) string {
	signature := HS256(NormalizeId(id), key)[:md5.Size]
	return Base64Encode(signature)
}

func AccessKey() string {
	iv := make([]byte, sha1.Size)
	if _, err := rand.Read(iv); err != nil {
		log.Println(err)
	}
	return Base64Encode(string(iv))
}

func AesEncrypt(text string, key string) (string, bool) {
	block, err := aes.NewCipher([]byte(Base64Decode(key)))
	if err != nil {
		log.Println(err)
	} else {
		plaintext := []byte(text)
		blockSize := aes.BlockSize
		padding := blockSize - len(plaintext)%blockSize
		padtext := bytes.Repeat([]byte{byte(padding)}, padding)
		plaintext = append(plaintext, padtext...)
		ciphertext := make([]byte, blockSize+len(plaintext))
		iv := ciphertext[:blockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			log.Println(err)
		} else {
			mode := cipher.NewCBCEncrypter(block, iv)
			mode.CryptBlocks(ciphertext[blockSize:], plaintext)
			return Base64Encode(string(ciphertext)), true
		}
	}
	return "", false
}

func AesDecrypt(text string, key string) (string, bool) {
	ciphertext := []byte(Base64Decode(text))
	blockSize := aes.BlockSize
	length := len(ciphertext) - blockSize
	if length >= 0 && length%blockSize == 0 {
		block, err := aes.NewCipher([]byte(Base64Decode(key)))
		if err != nil {
			log.Println(err)
		} else {
			iv := ciphertext[:blockSize]
			ciphertext = ciphertext[blockSize:]
			mode := cipher.NewCBCDecrypter(block, iv)
			mode.CryptBlocks(ciphertext, ciphertext)
			unpadding := int(ciphertext[length-1])
			return string(ciphertext[:(length - unpadding)]), true
		}
	}
	return "", false
}

func PublicKeyEncodeString(key *rsa.PublicKey) string {
	pubkey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(key),
	})
	return string(pubkey)
}

func PublicKeyDecodeString(str string) *rsa.PublicKey {
	block, _ := pem.Decode([]byte(str))
	if block != nil {
		key, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			log.Println(err)
		}
		return key
	}
	return nil
}

func PrivateKeyEncodeString(key *rsa.PrivateKey) string {
	privkey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return string(privkey)
}

func PrivateKeyDecodeString(str string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(str))
	if block != nil {
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Println(err)
		}
		return key
	}
	return nil
}

func RsaEncrypt(text string, key *rsa.PublicKey) (string, bool) {
	hash := sha256.New()
	rng := rand.Reader
	plaintext := []byte(text)
	label := make([]byte, 0)
	ciphertext, err := rsa.EncryptOAEP(hash, rng, key, plaintext, label)
	if err != nil {
		log.Println(err)
	} else {
		return Base64Encode(string(ciphertext)), true
	}
	return "", false
}

func RsaDecrypt(text string, key *rsa.PrivateKey) (string, bool) {
	hash := sha256.New()
	rng := rand.Reader
	ciphertext := []byte(Base64Decode(text))
	label := make([]byte, 0)
	plaintext, err := rsa.DecryptOAEP(hash, rng, key, ciphertext, label)
	if err != nil {
		log.Println(err)
	} else {
		return string(plaintext), true
	}
	return "", false
}

func CheckIPWhitelist(whitelist []string, addr string) bool {
	ip := net.ParseIP(addr)
	for _, str := range whitelist {
		if strings.Contains(str, "/") {
			_, ipnet, err := net.ParseCIDR(str)
			if err != nil {
				log.Println(err)
			} else if ipnet.Contains(ip) {
				return true
			}
		} else if str == addr {
			return true
		}
	}
	return false
}

func Integrity(value interface{}, config *toml.Tree) string {
	integrity := strings.ToLower(GetString(config, "integrity", "SHA384"))
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	content := make([]byte, 0)
	bytes := make([]byte, 0)
	object := iris.Map{}
	isMap := false
	switch value.(type) {
	case iris.Map:
		object = value.(iris.Map)
		isMap = true
	case []byte:
		content = value.([]byte)
	case string:
		content = []byte(value.(string))
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			log.Println(err)
		} else {
			content = bytes
		}
	}
	if !isMap {
		err := jsoniter.Unmarshal(content, &object)
		if err != nil {
			log.Println(err)
		} else {
			isMap = true
		}
	}
	if isMap {
		bytes, err := json.Marshal(object)
		if err != nil {
			log.Println(err)
		} else {
			content = bytes
		}
	}
	switch integrity {
	case "sha256":
		sum := sha256.Sum256(content)
		bytes = sum[:]
	case "sha384":
		sum := sha512.Sum384(content)
		bytes = sum[:]
	case "sha512":
		sum := sha512.Sum512(content)
		bytes = sum[:]
	default:
		sum := sha512.Sum384(content)
		bytes = sum[:]
	}
	return integrity + "-" + base64.StdEncoding.EncodeToString(bytes)
}

func Signature(signature string, digest string, config *toml.Tree) string {
	integrity := strings.ToLower(GetString(config, "integrity", "SHA384"))
	str := signature + strings.TrimPrefix(digest, integrity+"-")
	return Hash(str, config, true)
}

package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"io"
	"log"
	"math/big"
	"sort"
	"strings"
)

var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type WeiXin struct {
	appid       string
	token       string
	encodingKey string
	aesKey      []byte
}

func NewWeiXin(appid, token, key string) *WeiXin {
	return &WeiXin{
		appid:       appid,
		token:       token,
		encodingKey: key,
	}
}

func (w *WeiXin) ValidateMsg(timestamp, nonce, msgEncrypt, msgSignatureIn string) bool {
	msgSignatureGen := w.MakeMsgSignature(timestamp, nonce, msgEncrypt)
	return msgSignatureGen == msgSignatureIn
}

func (w *WeiXin) MakeMsgSignature(timestamp, nonce, msgEncrypt string) string {
	sl := []string{w.token, timestamp, nonce, msgEncrypt}
	sort.Strings(sl)
	s := sha1.New()
	_, _ = io.WriteString(s, strings.Join(sl, ""))
	return fmt.Sprintf("%x", s.Sum(nil))
}

func (w *WeiXin) MakeSignature(timestamp, nonce string) string { //本地计算signature
	si := []string{w.token, timestamp, nonce}
	sort.Strings(si)            //字典序排序
	str := strings.Join(si, "") //组合字符串
	s := sha1.New()             //返回一个新的使用SHA1校验的hash.Hash接口
	_, err := io.WriteString(s, str)
	if err != nil {
		return ""
	}
	//WriteString函数将字符串数组str中的内容写入到s中
	return fmt.Sprintf("%x", s.Sum(nil))
}

func (w *WeiXin) EncodingAESKey2AESKey() []byte {
	if w.aesKey == nil || len(w.aesKey) == 0 {
		data, _ := base64.StdEncoding.DecodeString(w.encodingKey + "=")
		w.aesKey = data
	}
	b := make([]byte, len(w.aesKey))
	copy(b, w.aesKey)
	return b
}

func (w *WeiXin) aesDecrypt(cipherData []byte, aesKey []byte) ([]byte, error) {
	k := len(aesKey) //PKCS#7
	if len(cipherData)%k != 0 {
		return nil, errors.New("crypto/cipher: ciphertext size is not multiple of aes key length")
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainData := make([]byte, len(cipherData))
	blockMode.CryptBlocks(plainData, cipherData)
	return plainData, nil
}

func (w *WeiXin) PKCS7Pad(message []byte, blockSize int) (padded []byte) {
	// block size must be bigger or equal 2
	if blockSize < 1<<1 {
		panic("block size is too small (minimum is 2 bytes)")
	}
	// block size up to 255 requires 1 byte padding
	if blockSize < 1<<8 {
		// calculate padding length
		padLen := w.PadLength(len(message), blockSize)

		// define PKCS7 padding block
		padding := bytes.Repeat([]byte{byte(padLen)}, padLen)

		// apply padding
		padded = append(message, padding...)
		return padded
	}
	// block size bigger or equal 256 is not currently supported
	panic("unsupported block size")
}

func (w *WeiXin) PadLength(sliceLength, blockSize int) (padLen int) {
	padLen = blockSize - sliceLength%blockSize
	if padLen == 0 {
		padLen = blockSize
	}
	return padLen
}

func (w *WeiXin) ValidateAppId(id []byte) bool {
	return string(id) == w.appid
}

func (w *WeiXin) aesEncrypt(plainData []byte, aesKey []byte) ([]byte, error) {
	k := len(aesKey)
	if len(plainData)%k != 0 {
		plainData = w.PKCS7Pad(plainData, k)
	}
	fmt.Printf("aesEncrypt: after padding, plainData length = %d\n", len(plainData))

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cipherData := make([]byte, len(plainData))
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(cipherData, plainData)

	return cipherData, nil
}

func (w *WeiXin) ParseEncryptTextRequestBody(plainText []byte) (*EncryptRequestBody, error) {
	// xml Decoding
	textRequestBody := &EncryptRequestBody{}
	err := xml.Unmarshal(plainText, textRequestBody)
	return textRequestBody, err
}

func (w *WeiXin) ParseEncryptRequestBody(timestamp, nonce, msgSignature string, rawBody []byte) (*TextRequestBody, error) {
	encryptRequestBody, err := w.ParseEncryptTextRequestBody(rawBody)
	if err != nil {
		return nil, err
	}
	// Validate msg signature
	if !w.ValidateMsg(timestamp, nonce, encryptRequestBody.Encrypt, msgSignature) {
		return nil, errors.New("校验数据来源失败")
	}
	log.Println("Wechat Service: msg_signature validation is ok!")

	// Decode base64
	cipherData, err := base64.StdEncoding.DecodeString(encryptRequestBody.Encrypt)
	if err != nil {
		log.Println("Wechat Service: Decode base64 error:", err)
		return nil, err
	}

	// AES Decrypt
	plainText, err := aesDecrypt(cipherData, w.aesKey)
	if err != nil {
		logs.Error("解密微信加密数据失败 ->", err)
		return nil, err
	}
	// Read length
	buf := bytes.NewBuffer(plainText[16:20])
	var length int32
	binary.Read(buf, binary.BigEndian, &length)
	fmt.Println(string(plainText[20 : 20+length]))

	// appID validation
	appIDstart := 20 + length
	id := plainText[appIDstart : int(appIDstart)+len(w.appid)]
	if !w.ValidateAppId(id) {
		log.Println("Wechat Service: appid is invalid!")
		return nil, errors.New("Appid is invalid")
	}
	textRequestBody := &TextRequestBody{}
	err = xml.Unmarshal(plainText[20:20+length], textRequestBody)
	return textRequestBody, err
}

func (w *WeiXin) MakeEncryptXmlData(fromUserName, toUserName, timestamp, content string) (string, error) {
	textResponseBody := &PassiveUserReplyMessage{}
	textResponseBody.FromUserName = Value(fromUserName)
	textResponseBody.ToUserName = Value(toUserName)
	textResponseBody.MsgType = Value("text")
	textResponseBody.Content = Value(content)
	textResponseBody.CreateTime = Value(timestamp)

	body, err := xml.MarshalIndent(textResponseBody, " ", "  ")
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, int32(len(body)))
	if err != nil {
		return "", err
	}
	bodyLength := buf.Bytes()

	randomBytes := []byte(w.randomString(16))

	plainData := bytes.Join([][]byte{randomBytes, bodyLength, body, []byte(w.appid)}, nil)
	cipherData, err := w.aesEncrypt(plainData, w.aesKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(cipherData), nil
}

func (w *WeiXin) MakeEncryptResponseBody(fromUserName, toUserName, content, nonce, timestamp string) ([]byte, error) {
	encryptBody := &EncryptResponseBody{}

	encryptXmlData, _ := w.MakeEncryptXmlData(fromUserName, toUserName, timestamp, content)
	encryptBody.Encrypt = Value(encryptXmlData)
	encryptBody.MsgSignature = Value(w.MakeMsgSignature(timestamp, nonce, encryptXmlData))
	encryptBody.TimeStamp = timestamp
	encryptBody.Nonce = Value(nonce)

	return xml.MarshalIndent(encryptBody, " ", "  ")
}

func (w *WeiXin) randomString(n int, allowedChars ...[]rune) string {
	var letters []rune

	if len(allowedChars) == 0 {
		letters = defaultLetters
	} else {
		letters = allowedChars[0]
	}

	b := make([]rune, n)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[int(num.Int64())]
	}
	return string(b)
}

package chassis

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

func sign(key, ciphertext []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	if	_, err := mac.Write([]byte(ciphertext)); err != nil {
		return  []byte{}, err
	}
	return mac.Sum(nil), nil
}

// ValidMAC reports whether signatureMAC is a valid HMAC tag for a token ciphertext.
func ValidMAC(signature, ciphertext, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	if	_, err := mac.Write(ciphertext); err != nil {
		return false
	}
	return hmac.Equal(signature, mac.Sum(nil))
}

// Splits sso secret in two keys (encryption and signature)
func getKeys(ssoSecret string) ([]byte, []byte) {
	hash := sha256.New()
	hash.Write([]byte(ssoSecret))
	key := hash.Sum(nil)
	encryptionKey := key[0:16]
	signatureKey := key[16:]
	return encryptionKey, signatureKey
}

func GenerateToken(secret string, jsonBytes []byte) (*string, error) {
	// Step 1 - split secret into encryption and signature keys
	encKey, signKey := getKeys(secret)

	// Step 2 - encrypt the JSON data using AES
	ciphertext, err := encrypt(encKey, jsonBytes)
	if err != nil {
		return nil, err
	}

	// Step 3 - Sign the encrypted data in Step 2 using HMAC
	signedText, err := sign(signKey, *ciphertext)
	if err != nil {
		return nil, err
	}

	// Step 4 - Append ciphertext with the signature
	// Our SSO login token now consists of the 128 bit initialization vector,
	// a variable length ciphertext, and a 256 bit signature (EXACTLY in this order).
	token := append(*ciphertext, signedText...)

	// Step 5 - This data is encoded using base64 (URL-safe variant, RFC 4648).
	encoded := base64.URLEncoding.EncodeToString(token)
	return &encoded, nil
}

// RevertToken validate a SSO token by decrypting its data and checking its HMAC signature
func RevertToken(secret, token string) (*[]byte, error) {
	//Step 1 - split secret into encryption and signature keys
	encKey, signKey := getKeys(secret)

	//Step 2 - Revert the string url-safe to a byte slice ([]byte)
	tokenDecoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}

	//Step 3 - Split token into ciphertext and signature
	limit := len(tokenDecoded) - sha256.Size
	ciphertext := tokenDecoded[:limit]
	tokenSignature := tokenDecoded[limit:]

	// Step 4 - Sign the ciphertext on the token and validate
	if !ValidMAC(tokenSignature, ciphertext, signKey) {
		return nil, errors.New("invalid token signature")
	}

	//Step 3: Decrypt json data
	return decrypt(encKey, ciphertext)
}

func encrypt(key, input []byte) (*[]byte, error) {
	//Step 1 - Use PKCS5Padding to complete the blocks
	paddedInput := pkcs5Padding(input, aes.BlockSize)
	if len(paddedInput) % aes.BlockSize != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}

	//Step 2 - Initialize cipher using the AES algorithm
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize + len(paddedInput))
	iv := ciphertext[:aes.BlockSize]

	// Step 3  - Generate random IV (initialization vector)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Step 4 - Encrypt padded input using CBC
	// 128 bit key length, CBC mode of operation, random initialization vector
	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypter.CryptBlocks(ciphertext[aes.BlockSize:], paddedInput)

	return &ciphertext, nil
}

func decrypt(key, input []byte) (*[]byte, error) {
	//Step 1 - Initialize cipher using the AES algorithm
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(input) < aes.BlockSize {
		return nil, errors.New("ciphertext is too short")
	}
	//Step 2 - Split input between IV and ciphertext
	// The IV should be unique, but not secure.
	// It consists of the first 16 bytes on the input
	iv := input[:aes.BlockSize]
	ciphertext := input[aes.BlockSize:]

	decrypted := make([]byte, len(ciphertext))

	//Step 3 - Decrypt ciphertext using CBC
	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(decrypted, ciphertext)

	//Step 4 - removing padding from the decrypted data
	trimmed := pkcs5Trimming(decrypted)
	return &trimmed, nil
}

// pads text to be multiples of 8-byte blocks.
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

// removes padding from text.
func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

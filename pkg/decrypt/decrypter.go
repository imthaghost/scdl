package decrypt

import (
	"crypto/aes"
	"crypto/cipher"
)

// Decrypt ...
// TODO: implement tests
func Decrypt(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cbc := cipher.NewCBCDecrypter(block, iv)
	cbc.CryptBlocks(data, data)

	return zeroUnPadding(data), nil
}

func zeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

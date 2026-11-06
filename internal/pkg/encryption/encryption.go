package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

var ErrReadFile = errors.New("error reading file")

func CreateAesBlock(aesKey string) (cipher.Block, error) {
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES block: %w", err)
	}
	return block, nil
}

func EncryptFile(file io.Reader, aesBlock cipher.Block) ([]byte, error) {
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, ErrReadFile
	}
	return EncryptData(bytes, aesBlock)
}

func EncryptData(data []byte, aesBlock cipher.Block) ([]byte, error) {
	cipherText := make([]byte, aes.BlockSize+len(data))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(aesBlock, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], data)
	return cipherText, nil
}

func Decrypt(cipherText []byte, aesBlock cipher.Block) ([]byte, error) {
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(aesBlock, iv)
	stream.XORKeyStream(cipherText, cipherText)
	return cipherText, nil
}

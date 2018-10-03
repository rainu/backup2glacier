package backup

import (
	. "backup2glacier/log"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

type CryptModule interface {
	Encrypt(src io.Reader, dst io.Writer) error
	Decrypt(src io.Reader, dst io.Writer) error
}

type cryptModule struct {
	key         []byte
	cipherBlock cipher.Block
}

func NewCryptModule(password string) CryptModule {
	hash := sha256.New()
	io.WriteString(hash, password)

	key, err := hex.DecodeString(fmt.Sprintf("%x", hash.Sum(nil)))
	if err != nil {
		LogFatal("Error while init cipher key. Error: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		LogFatal("Error while init cipher. Error: %v", err)
	}

	return &cryptModule{
		key:         key,
		cipherBlock: block,
	}
}

func (c *cryptModule) Encrypt(src io.Reader, dst io.Writer) error {
	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(c.cipherBlock, iv[:])

	writer := &cipher.StreamWriter{S: stream, W: dst}
	// Copy the input file to the output file, encrypting as we go.

	_, err := io.Copy(writer, src)
	return err
}

func (c *cryptModule) Decrypt(src io.Reader, dst io.Writer) error {
	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(c.cipherBlock, iv[:])

	reader := &cipher.StreamReader{S: stream, R: src}
	// Copy the input file to the output file, encrypting as we go.

	_, err := io.Copy(dst, reader)
	return err
}

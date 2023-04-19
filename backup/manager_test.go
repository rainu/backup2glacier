package backup

import (
	"archive/zip"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

func Test_ZipEncryptDecryptUnzip(t *testing.T) {
	password := "test"

	srcZip, dstZip := io.Pipe()
	srcEncrypt, dstEncrypt := io.Pipe()

	//encryption
	go func() {
		defer dstEncrypt.Close()

		crypt := NewCryptModule(password)
		crypt.Encrypt(srcZip, dstEncrypt)
	}()

	go func() {
		defer dstZip.Close()

		Zip([]string{"./"}, []*regexp.Regexp{}, []*regexp.Regexp{}, dstZip, nil)
	}()

	encodedZipFile, err := ioutil.TempFile("", ".enc")
	assert.NoError(t, err)
	if err != nil {
		return
	}

	io.Copy(encodedZipFile, srcEncrypt)
	encodedZipFile.Close()

	srcDecrypt, dstDecrypt := io.Pipe()

	//decryption
	go func() {
		defer dstDecrypt.Close()

		encFile, err := os.Open(encodedZipFile.Name())
		assert.NoError(t, err)

		if err != nil {
			panic(err)
		}
		defer encFile.Close()

		crypt := NewCryptModule(password)
		crypt.Decrypt(encFile, dstDecrypt)
	}()

	zipFile, err := ioutil.TempFile("", ".zip")
	assert.NoError(t, err)

	if err != nil {
		panic(err)
	}
	defer encodedZipFile.Close()

	io.Copy(zipFile, srcDecrypt)

	unzipReader, err := zip.OpenReader(zipFile.Name())
	assert.NoError(t, err)

	if err != nil {
		return
	}

	containsTestFile := false
	for _, zipContent := range unzipReader.File {
		if strings.HasSuffix(zipContent.Name, "manager_test.go") {
			containsTestFile = true
		}
	}
	assert.True(t, containsTestFile)
}

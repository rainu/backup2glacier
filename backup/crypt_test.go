package backup

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"sync"
	"testing"
)

func TestCryptModule_Encrypt(t *testing.T) {
	//given
	testText := `This is a test text!`
	toTest := NewCryptModule("somePassword")

	inBuf := bytes.NewBufferString(testText)
	outBuf := new(bytes.Buffer)

	//when
	encErr := toTest.Encrypt(inBuf, outBuf)

	//then
	assert.NoError(t, encErr)
	assert.NotEqual(t, outBuf.String(), testText)
}

func TestCryptModule_Decrypt(t *testing.T) {
	//given
	testText := `This should simulate a ENCRYPTED text!`
	toTest := NewCryptModule("somePassword")

	inBuf := bytes.NewBufferString(testText)
	outBuf := new(bytes.Buffer)

	//when
	encErr := toTest.Decrypt(inBuf, outBuf)

	//then
	assert.NoError(t, encErr)
	assert.NotEqual(t, outBuf.String(), testText)
}

func TestCryptModule(t *testing.T) {
	//given
	testText := `This is a test text!`
	toTest := NewCryptModule("somePassword")

	inBuf := bytes.NewBufferString(testText)
	outBuf := new(bytes.Buffer)
	srcEnc, dstEnc := io.Pipe()

	//when
	var encErr error
	var decErr error

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer dstEnc.Close()

		encErr = toTest.Encrypt(inBuf, dstEnc)
	}()

	go func() {
		defer wg.Done()
		decErr = toTest.Decrypt(srcEnc, outBuf)
	}()

	wg.Wait()

	//then
	assert.NoError(t, encErr)
	assert.NoError(t, decErr)
	assert.Equal(t, outBuf.String(), testText)
}

package backup

import (
	"archive/zip"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
)

func Test_Zip(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", ".zip")
	assert.NoError(t, err)

	if err != nil {
		return
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	containsTestFile := false
	contentChan := make(chan *ZipContent)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		for content := range contentChan {
			if strings.HasSuffix(content.Realpath, "zip_test.go") {
				containsTestFile = true
			}
		}
	}()

	Zip("./", tmpFile, contentChan)

	wg.Wait()
	assert.True(t, containsTestFile)

	unzipReader, err := zip.OpenReader(tmpFile.Name())
	assert.NoError(t, err)
	if err != nil {
		return
	}

	containsTestFile = false
	for _, zipContent := range unzipReader.File {
		if strings.HasSuffix(zipContent.Name, "zip_test.go") {
			containsTestFile = true
		}
	}
	assert.True(t, containsTestFile)
}

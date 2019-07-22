package backup

import (
	"archive/zip"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"regexp"
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

	Zip([]string{"./"}, []*regexp.Regexp{}, tmpFile, contentChan)

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

func TestBlacklist(t *testing.T) {
	tests := []struct {
		name           string
		blacklist      string
		testCase       string
		expectedResult bool
	}{
		{"suffixes", `.*\.exe`, "test.exe", true },
		{"suffixes", `.*\.exe`, "test.exe_", true },
		{"suffixes", `.*\.exe$`, "test.exe_", false },
		{"suffixes", `.*\.exe`, "full/path/to/test.exe", true },
		{"suffixes", `.*\.exe`, ".exe", true },
		{"suffixes", `.*\.exe`, "exe", false },
		{"folders", `.*/log/.*`, "/path/to/log/test.txt", true },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, _ := isBlacklisted(test.testCase, []*regexp.Regexp{regexp.MustCompile(test.blacklist)})
			assert.Equal(t, result, test.expectedResult)
		})
	}
}

package backup

import (
	"archive/zip"
	. "backup2glacier/log"
	"compress/flate"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ZipContent struct {
	Zippath  string
	Realpath string
	Length   int64
	FileInfo os.FileInfo
}

// ZIP the given file/folder and write file information out in given channel
func Zip(filePaths []string, blacklist, whitelist []*regexp.Regexp, dst io.Writer, contentChan chan<- *ZipContent) {
	// Create a new zip archive.
	zipWriter := zip.NewWriter(dst)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	defer zipWriter.Close()

	for _, filePath := range filePaths {
		absFilePath, _ := filepath.Abs(filePath)
		fInfo, err := os.Stat(filePath)
		if err != nil {
			LogError("Could not read file information for '%s'. Error: %v", filePath, err)
			continue
		}

		if fInfo.IsDir() {
			addFiles(zipWriter, absFilePath+"/", filepath.Dir(absFilePath+"/")+"/", blacklist, whitelist, contentChan)
		} else {
			dir, name := filepath.Split(absFilePath)
			addFile(zipWriter, dir, dir+"/", name, blacklist, whitelist, contentChan)
		}
	}

	if contentChan != nil {
		close(contentChan)
	}
}

func addFiles(w *zip.Writer, basePath, baseInZip string, blacklist, whitelist []*regexp.Regexp, contentChan chan<- *ZipContent) {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		LogError("Could not list directory '%s'. Error: %v", basePath, err)
		return
	}

	for _, fileDesc := range files {
		if fileDesc.IsDir() {
			// recursion ahead!
			newBase := basePath + fileDesc.Name() + "/"
			addFiles(w, newBase, baseInZip+"/"+fileDesc.Name()+"/", blacklist, whitelist, contentChan)
		} else {
			addFile(w, basePath, baseInZip, fileDesc.Name(), blacklist, whitelist, contentChan)
		}
	}
}

func addFile(w *zip.Writer, basePath, baseInZip, fileName string, blacklist, whitelist []*regexp.Regexp, contentChan chan<- *ZipContent) int64 {
	//open for reading
	filePath := normalizeFilePath(basePath + fileName)
	zipPath := normalizeZipPath(baseInZip + fileName)

	osFile, err := os.Open(filePath)
	if err != nil {
		LogError("Could not open file '%s'. Error: %v", filePath, err)
		return 0
	}
	defer osFile.Close()

	if blacklisted, blExpr := isListed(filePath, blacklist); blacklisted {
		if whitelisted, wlExpr := isListed(filePath, whitelist); whitelisted {
			LogInfo(`Include file because it is whitelisted: %s -> "%s"`, filePath, wlExpr)
		} else {
			LogInfo(`Ignore file because it is blacklisted: %s -> "%s"`, filePath, blExpr)
			return 0
		}
	}

	LogInfo("Add to zip: %s -> %s", osFile.Name(), zipPath)

	// Add some files to the archive.
	fileInfo, err := osFile.Stat()
	if err != nil {
		LogError("Could not read file metadata for %s. Error: %v", filePath, err)
		return 0
	}

	zipFileInfo, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		LogError("Could not create fileinfo header for %s. Error: %v", zipPath, err)
		return 0
	}
	zipFileInfo.Name = zipPath

	zipFileHandle, err := w.CreateHeader(zipFileInfo)
	if err != nil {
		LogError("Could not add file '%s' to zip. Error %v", zipPath, err)
		return 0
	}

	written, err := io.Copy(zipFileHandle, osFile)
	if err != nil {
		LogError("Could not add file '%s' to zip. Error %v", zipPath, err)
		return 0
	}

	if contentChan != nil {
		contentChan <- &ZipContent{
			Zippath:  zipPath,
			Realpath: filePath,
			Length:   written,
			FileInfo: fileInfo,
		}
	}

	return written
}

func isListed(path string, list []*regexp.Regexp) (bool, *regexp.Regexp) {
	for _, curExpr := range list {
		if curExpr.MatchString(path) {
			return true, curExpr
		}
	}

	return false, nil
}

func normalizeFilePath(path string) string {
	return strings.Replace(path, "//", "/", -1)
}

func normalizeZipPath(path string) string {
	result := strings.Replace("/"+path, "//", "/", -1)
	return result[1:]
}

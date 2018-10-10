package backup

import (
	"archive/zip"
	. "backup2glacier/log"
	"compress/flate"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ZipContent struct {
	Zippath  string
	Realpath string
	Length   int64
	FileInfo os.FileInfo
}

//ZIP the given file/folder and write file information out in given channel
func Zip(filePath string, dst io.Writer, contentChan chan<- *ZipContent) {
	// Create a new zip archive.
	zipWriter := zip.NewWriter(dst)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	defer zipWriter.Close()

	fInfo, err := os.Stat(filePath)
	if err != nil {
		LogFatal("Could not read file information for '%s'. Error: %v", filePath, err)
	}

	if fInfo.IsDir() {
		addFiles(zipWriter, filePath, "", contentChan)
	} else {
		dir, name := filepath.Split(filePath)
		addFile(zipWriter, dir, "", name, contentChan)
	}

	if contentChan != nil {
		close(contentChan)
	}
}

func addFiles(w *zip.Writer, basePath, baseInZip string, contentChan chan<- *ZipContent) {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		LogFatal("Could not list directory '%s'. Error: %v", basePath, err)
	}

	for _, fileDesc := range files {
		if fileDesc.IsDir() {
			// recursion ahead!
			newBase := basePath + fileDesc.Name() + "/"
			addFiles(w, newBase, baseInZip+"/"+fileDesc.Name()+"/", contentChan)
		} else {
			addFile(w, basePath, baseInZip, fileDesc.Name(), contentChan)
		}
	}
}

func addFile(w *zip.Writer, basePath, baseInZip, fileName string, contentChan chan<- *ZipContent) int64 {
	//open for reading
	filePath := normalizeFilePath(basePath + fileName)
	zipPath := normalizeZipPath(baseInZip + fileName)

	osFile, err := os.Open(filePath)
	if err != nil {
		LogFatal("Could not open file '%s'. Error: %v", filePath, err)
	}
	defer osFile.Close()

	LogInfo("Add to zip: %s -> %s", osFile.Name(), zipPath)

	// Add some files to the archive.
	fileInfo, err := osFile.Stat()
	if err != nil {
		LogFatal("Could not read file metadata for %s. Error: %v", filePath, err)
	}

	zipFileInfo, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		LogFatal("Could not create fileinfo header for %s. Error: %v", zipPath, err)
	}
	zipFileInfo.Name = zipPath

	zipFileHandle, err := w.CreateHeader(zipFileInfo)
	if err != nil {
		LogFatal("Could not add file '%s' to zip. Error %v", zipPath, err)
	}

	written, err := io.Copy(zipFileHandle, osFile)
	if err != nil {
		LogFatal("Could not add file '%s' to zip. Error %v", zipPath, err)
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

func normalizeFilePath(path string) string {
	return strings.Replace(path, "//", "/", -1)
}

func normalizeZipPath(path string) string {
	result := strings.Replace("/"+path, "//", "/", -1)
	return result[1:]
}

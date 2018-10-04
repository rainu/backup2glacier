package backup

import (
	"archive/zip"
	. "backup2glacier/log"
	"compress/flate"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ZipContent struct {
	Zippath  string
	Realpath string
	Length   int64
}

//ZIP the given file/folder and write file information out in given channel
func Zip(filePath string, dst io.Writer, contentChan chan<- *ZipContent) {
	// Create a new zip archive.
	zipWriter := zip.NewWriter(dst)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	fInfo, err := os.Stat(filePath)
	if err != nil {
		LogFatal("Could not read file information for '%s'. Error: %v", filePath, err)
	}

	if fInfo.IsDir() {
		addFiles(zipWriter, filePath, "", contentChan)
	} else {
		dir, name := filepath.Split(filePath)
		written := addFile(zipWriter, dir, "", name)

		if contentChan != nil {
			contentChan <- &ZipContent{
				Zippath:  "/" + name,
				Realpath: filePath,
				Length:   written,
			}
		}
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
			addFiles(w, newBase, fileDesc.Name()+"/", contentChan)
		} else {
			written := addFile(w, basePath, baseInZip, fileDesc.Name())

			if contentChan != nil {
				contentChan <- &ZipContent{
					Zippath:  baseInZip + "/" + fileDesc.Name(),
					Realpath: basePath + "/" + fileDesc.Name(),
					Length:   written,
				}

			}
		}
	}
}

func addFile(w *zip.Writer, basePath, baseInZip, fileName string) int64 {
	//open for reading
	osFile, err := os.Open(basePath + fileName)
	if err != nil {
		LogFatal("Could not open file '%s'. Error: %v", basePath+fileName, err)
	}
	defer osFile.Close()

	LogInfo("Add to zip: %s", osFile.Name())

	// Add some files to the archive.
	zipFileHandle, err := w.Create(baseInZip + fileName)
	if err != nil {
		LogFatal("Could not add file '%s' to zip. Error %v", baseInZip+fileName, err)
	}

	written, err := io.Copy(zipFileHandle, osFile)
	if err != nil {
		LogFatal("Could not add file '%s' to zip. Error %v", baseInZip+fileName, err)
	}

	return written
}

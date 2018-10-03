package backup

import (
	"archive/zip"
	. "backup2glacier/log"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Zip(filePath string, dst io.Writer) {
	// Create a new zip archive.
	zipWriter := zip.NewWriter(dst)

	fInfo, err := os.Stat(filePath)
	if err != nil {
		LogFatal("Could not read file information for '%s'. Error: %v", filePath, err)
	}

	if fInfo.IsDir() {
		addFiles(zipWriter, filePath, "")
	} else {
		dir, name := filepath.Split(filePath)
		addFile(zipWriter, dir, "", name)
	}
}

func addFiles(w *zip.Writer, basePath, baseInZip string) {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		LogFatal("Could not list directory '%s'. Error: %v", basePath, err)
	}

	for _, fileDesc := range files {
		if fileDesc.IsDir() {
			// recursion ahead!
			newBase := basePath + fileDesc.Name() + "/"
			addFiles(w, newBase, fileDesc.Name()+"/")
		} else {
			addFile(w, basePath, baseInZip, fileDesc.Name())
		}
	}
}

func addFile(w *zip.Writer, basePath, baseInZip, fileName string) {
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

	_, err = io.Copy(zipFileHandle, osFile)
	if err != nil {
		LogFatal("Could not add file '%s' to zip. Error %v", baseInZip+fileName, err)
	}
}

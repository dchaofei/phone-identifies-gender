package control

import (
	"log"
	"os"
)

func initDir() {
	dirs := []string{searchDir, resultDir}
	for _, v := range dirs {
		if err := createDir(v); err != nil {
			log.Fatal("initDir:", err)
		}
	}
}

func createDir(path string) error {
	if isExist(path) {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

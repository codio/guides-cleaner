package files

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

func GetListFiles(rootPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func MergeDirectory(destPath, mergePath string) error {
	files, err := GetListFiles(destPath)
	if err != nil {
		return err
	}
	excludeFiles := make(map[string]bool)
	for _, item := range files {
		excludeFiles[strings.Replace(item, destPath, "", 1)] = true
	}
	err = copy.Copy(
		mergePath,
		destPath,
		copy.Options{
			Skip: func(src string) (bool, error) {
				relativPath := strings.Replace(src, mergePath, "", 1)
				_, exists := excludeFiles[relativPath]
				return exists, nil
			},
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func ReadFile(pathToFile string) (string, error) {
	data, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func PathIsExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

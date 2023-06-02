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

func CopyFiles(destAssignmentPath, mergeAssignmentPath string) error {
	files, _ := GetListFiles(destAssignmentPath)
	excludeFiles := make(map[string]bool)
	for _, item := range files {
		excludeFiles[strings.Replace(item, destAssignmentPath, "", 1)] = true
	}
	copy.Copy(
		mergeAssignmentPath,
		destAssignmentPath,
		copy.Options{
			Skip: func(src string) (bool, error) {
				relativPath := strings.Replace(src, mergeAssignmentPath, "", 1)
				_, exists := excludeFiles[relativPath]
				return exists, nil
			},
		},
	)
	return nil
}

func ReadFile(pathToFile string) (string, error) {
	data, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

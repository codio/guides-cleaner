package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("guides-cleaner clean-content <path>")
	os.Exit(1)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func loadSections(path string) ([]Section, error) {
	jsonFilePath := filepath.Join(path, "metadata.json")
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	if err != nil {
		return nil, err
	}
	var metadata Metadata
	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		return nil, err
	}
	return metadata.Sections, nil
}

func getFSContentFiles(path string) ([]string, error) {
	var files []string

	root := filepath.Join(path, "content")
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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

func cleanFolder(fileNames map[string]bool, path string) error {
	files, err := getFSContentFiles(path)
	if err != nil {
		return err
	}

	for _, filePath := range files {
		file := filepath.Base(filePath)
		fmt.Printf("File: %s\n", file)
		_, exists := fileNames[file]
		if !exists {
			fmt.Printf("DELETING FILE!: %s\n", filePath)
			err = os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func cleanContent(path string) (bool, error) {
	sections, err := loadSections(path)
	if err != nil {
		return false, err
	}
	sectionPaths := make(map[string]bool)
	for _, item := range sections {
		sectionPaths[filepath.Base(item.ContentFile)] = true
	}

	err = cleanFolder(sectionPaths, path)
	if err != nil {
		return false, err
	}

	return true, nil
}

func Main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
	} else {
		switch args[0] {
		case "clean-content":
			projectPath := ".guides"
			if len(args) > 1 {
				projectPath = args[1]
			}
			_, err := cleanContent(projectPath)
			check(err)
		default:
			printHelp()
		}

	}
}

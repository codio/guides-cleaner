package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileInfo struct {
	FullPath string
	NeedRemove bool
}

var fileMap = make(map[string]FileInfo)
var assessmentMap = make(map[string]bool)
var file_search_dict string

func printHelp() {
	fmt.Println(`Usage:
	clean content: guides-cleaner clean-content <path>
	clean assessments: guides-cleaner clean-assessments <path>
	delete unused files in img\: guides-cleaner clean-images <path>
	delete unused files in code\: guides-cleaner clean-code <path>
	full clean: guides-cleaner clean-full <path>
	<path> - path to the project`)
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

func cleanAssessments(path string) error {
	var root []interface{}
	jsonFilePath := filepath.Join(path, "assessments.json")
	jsonFile, err := os.OpenFile(jsonFilePath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &root); err != nil {
    	log.Fatal(err)
	}
	var dst []interface{}

	for _, value := range root {
		node, ok := value.(map[string]interface{})
		if !ok {
			log.Fatalf("Error clean assessment.json")
		}
		taskId, ok := node["taskId"].(string)
		if !ok {
			log.Fatal("Error fetch taskId")
		}
		if (assessmentMap[taskId]) {
			dst = append(dst, value)
		}
	}
	
	data, err := json.MarshalIndent(dst, "", " ")
	check(err)
	jsonFile.Truncate(0)
	jsonFile.Seek(0, 0)
	jsonFile.Write(data)

	return nil
}

func cleanFoldersByFileMap() error {
	for _, file := range fileMap {
		if (!file.NeedRemove) {
			continue
		}
		fmt.Printf("DELETING FILE!: %s\n", file.FullPath)
		err := os.Remove(file.FullPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func getListFiles(rootPath string) ([]string, error) {
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

func cleanFolder(fileNames map[string]bool, path string) error {
	files, err := getListFiles(filepath.Join(path, "content"))
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

func checkFilesContent(rootPath string, paths []string, includeAssessments bool) error {
	for _, path := range paths {
		files, err := getListFiles(path)
		if err != nil {
			return err
		}
		for _, filePath := range files {
			relativFilePath := strings.Replace(filePath, path, "", 1)
			fileMap[relativFilePath] = FileInfo{filePath, true}
			file_search_dict = file_search_dict + relativFilePath + "|"
		}
	}
	file_search_dict = strings.TrimRight(file_search_dict, "|")
	checkDirectory(rootPath, includeAssessments)
	return nil
}

func checkDirectory(pathToDirectory string, includeAssessments bool) {
  files, err := ioutil.ReadDir(pathToDirectory)
  if err != nil {
    log.Fatal(err)
  }
  for _, file := range files {
    pathToFile := pathToDirectory + "/" + file.Name()
    if file.IsDir() {
      checkDirectory(pathToFile, includeAssessments)
    } else {
      checkFile(pathToFile, includeAssessments)
    }
  }
}

func checkFile(pathToFile string, includeAssessments bool) {
	content := readFile(pathToFile)

	re := regexp.MustCompile(file_search_dict)
	matches := re.FindAllString(content, -1)
	for _, item := range matches {
		i := fileMap[item]
		i.NeedRemove = false
		fileMap[item] = i
	}

	if (includeAssessments) {
		re = regexp.MustCompile(`{.*?|assessment}\((?P<taskId>[a-zA-Z\d-]*)\)`)
		taskIdIndex := re.SubexpIndex("taskId")
		matches := re.FindAllStringSubmatch(content, -1)
		for _, v := range matches {
			assessmentMap[v[taskIdIndex]] = true
		}
	}
}

func readFile(pathToFile string) string {
	data, err := ioutil.ReadFile(pathToFile)
	if err != nil {
	  log.Println(err)
	}
	return string(data)
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

func getImgPath(projectPath string) string {
	return filepath.Join(projectPath, "img")
}

func getCodePath(projectPath string) string {
	return filepath.Join(projectPath, "../code")
}

func Main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
	} else {
		projectPath := ".guides"
		if len(args) > 1 {
			projectPath = args[1]
		}
		switch args[0] {
		case "clean-content":
			_, err := cleanContent(projectPath)
			check(err)
		case "clean-assessments":
			err := checkFilesContent(projectPath, []string{}, true)
			check(err)
			err = cleanAssessments(projectPath)
			check(err)
		case "clean-images":
			err := checkFilesContent(projectPath, []string{getImgPath(projectPath)}, false)
			check(err)
			err = cleanFoldersByFileMap()
			check(err)
		case "clean-code":
			err := checkFilesContent(projectPath, []string{getCodePath(projectPath)}, false)
			check(err)
			err = cleanFoldersByFileMap()
			check(err)
		case "clean-full":
			_, err := cleanContent(projectPath)
			check(err)
			err = checkFilesContent(projectPath, []string{getImgPath(projectPath), getCodePath(projectPath)}, true)
			check(err)
			err = cleanAssessments(projectPath)
			check(err)
			err = cleanFoldersByFileMap()
			check(err)
		default:
			printHelp()
		}
	}
}

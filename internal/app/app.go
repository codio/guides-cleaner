package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/otiai10/copy"
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
	clean content: guides-cleaner clean-content <path_to_the_project>
	clean assessments: guides-cleaner clean-assessments <path_to_the_project>
	delete unused files in img\: guides-cleaner clean-images <path_to_the_project>
	delete unused files in code\: guides-cleaner clean-code <path_to_the_project>
	full clean: guides-cleaner clean-full <path_to_the_project>
	merge assignments: guides-cleaner merge <destAssignmentPath> <mergeAssignmentPath>`)
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

	bytes, err := ioutil.ReadAll(jsonFile)
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

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &root); err != nil {
    	return err
	}
	var dst []interface{}

	for _, value := range root {
		node, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error clean assessment.json")
		}
		taskId, ok := node["taskId"].(string)
		if !ok {
			return fmt.Errorf("error fetch taskId")
		}
		if (assessmentMap[taskId]) {
			dst = append(dst, value)
		}
	}
	
	data, err := json.MarshalIndent(dst, "", " ")
	if err != nil {
		return err
	}
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

func checkDirectory(pathToDirectory string, includeAssessments bool) error {
  files, err := ioutil.ReadDir(pathToDirectory)
  if err != nil {
    return err
  }
  for _, file := range files {
    pathToFile := pathToDirectory + "/" + file.Name()
    if file.IsDir() {
      err = checkDirectory(pathToFile, includeAssessments)
	  if err != nil {
		return err
	  }
    } else {
      err = checkFile(pathToFile, includeAssessments)
	  if err != nil {
		return err
	  }
    }
  }
  return nil
}

func checkFile(pathToFile string, includeAssessments bool) error {
	content, err := readFile(pathToFile)
	if err != nil {
		return err
	}

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
	return nil
}

func readFile(pathToFile string) (string, error) {
	data, err := ioutil.ReadFile(pathToFile)
	if err != nil {
	  return "", err
	}
	return string(data), nil
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

func mergeAssignments(destAssignmentPath string, mergeAssignmentPath string) error {
	err := mergeAssessmentsJson(destAssignmentPath, mergeAssignmentPath)
	if err != nil {
		return err
	}
	err = mergeJson(destAssignmentPath, mergeAssignmentPath, ".guides/metadata.json", "sections")
	if err != nil {
		return err
	}
	err = mergeJson(destAssignmentPath, mergeAssignmentPath, ".guides/book.json", "children")
	if err != nil {
		return err
	}
	err = copyFiles(destAssignmentPath, mergeAssignmentPath)
	if err != nil {
		return err
	}
	return nil
}

func copyFiles(destAssignmentPath string, mergeAssignmentPath string) error {
	files, _ := getListFiles(destAssignmentPath)
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

func mergeAssessmentsJson(destAssignmentPath string, mergeAssignmentPath string) error {
	relativPathToBook := ".guides/assessments.json"
	var mergeJson []interface{}
	mergeFilePath := filepath.Join(mergeAssignmentPath, relativPathToBook)
	mergeFile, err := os.Open(mergeFilePath)
	if err != nil {
		return err
	}
	defer mergeFile.Close()

	bytes, err := ioutil.ReadAll(mergeFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &mergeJson); err != nil {
    	return err
	}

	var dstJson []interface{}
	dstFilePath := filepath.Join(destAssignmentPath, relativPathToBook)
	dstFile, err := os.OpenFile(dstFilePath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	bytes, err = ioutil.ReadAll(dstFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &dstJson); err != nil {
    	return err
	}

	dstIds := []string{}
	for _, val := range dstJson {
		node, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error processing file: %s", dstFilePath)
		}
		id, ok := node["taskId"].(string)
		if ok {
			dstIds = append(dstIds, id)
		}
	}

	for _, val := range mergeJson {
		node, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error processing file: %s", mergeFilePath)
		}
		id, ok := node["taskId"].(string)
		if ok && !containedInArray(dstIds, id){
			dstJson = append(dstJson, val)
		}
	}
	
	data, err := json.MarshalIndent(dstJson, "", " ")
	if err != nil {
		return err
	}
	dstFile.Truncate(0)
	dstFile.Seek(0, 0)
	dstFile.Write(data)
	return nil
}

func mergeJson(destAssignmentPath, mergeAssignmentPath, relativPathToFile, processedRecord string) error {
	pathToDest := filepath.Join(destAssignmentPath, relativPathToFile)
	pathToMerge := filepath.Join(mergeAssignmentPath, relativPathToFile)
	arr, err := getMergeArray(pathToMerge, processedRecord)
	if err != nil {
		return err
	}
	err = mergeIntoDst(pathToDest, processedRecord, arr)
	if err != nil {
		return err
	}
	return nil
}

func getMergeArray(pathToFile string, key string) ([]interface{}, error) {
	var root interface{}
	jsonFile, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &root); err != nil {
    	return nil, err
	}

	records, ok := root.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error processing file: %s", pathToFile)
	}
	out, ok := records[key].([]interface{})
	if ok {
		return out, nil
	}

	return []interface{}{}, nil
}

func mergeIntoDst(pathToFile string, key string, mergeArr []interface{}) error {
	var root interface{}
	jsonFile, err := os.OpenFile(pathToFile, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &root); err != nil {
    	return err
	}

	records, ok := root.(map[string]interface{})
	if !ok {
		return fmt.Errorf("error processing file: %s", pathToFile)
	}

	var srcRecord []interface{}
	srcRecord, ok = records[key].([]interface{})
	if !ok {
		srcRecord = []interface{}{}
	}

	srcIds := []string{}
	for _, val := range srcRecord {
		node, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error processing file: %s", pathToFile)
		}
		id, ok := node["id"].(string)
		if ok {
			srcIds = append(srcIds, id)
		}
	}

	for _, val := range mergeArr {
		node, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("error processing file: %s", pathToFile)
		}
		id, ok := node["id"].(string)
		if ok && !containedInArray(srcIds, id){
			srcRecord = append(srcRecord, val)
		}
	}

	records[key] = srcRecord
	
	data, err := json.MarshalIndent(root, "", " ")
	if err != nil {
		return err
	}
	jsonFile.Truncate(0)
	jsonFile.Seek(0, 0)
	jsonFile.Write(data)
	return nil
}

func containedInArray(array []string, value string) bool {
	for _, n := range array {
		if value == n {
			return true
		}
	}
	return false
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
		case "help":
			printHelp()
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
		case "merge":
			if len(args) > 2 {
				destAssignmentPath := args[1]
				mergeAssignmentPath := args[2]
				err := mergeAssignments(destAssignmentPath, mergeAssignmentPath)
				check(err)
			} else {
				printHelp()
			}

		default:
			printHelp()
		}
	}
}

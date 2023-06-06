package cleaner

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codio/guides-cleaner/internal/types"
	utils "github.com/codio/guides-cleaner/internal/utils"
	filesUtils "github.com/codio/guides-cleaner/internal/utils/files"
	versionChecker "github.com/codio/guides-cleaner/internal/utils/version-checker"
)

var fileMap = make(map[string]types.FileInfo)
var assessmentMap = make(map[string]bool)
var file_search_dict string

func Clean(action, projectPath string) error {
	isV3, err := versionChecker.IsV3(projectPath)
	if err != nil {
		return err
	}
	switch action {
	case "clean-content":
		if err := cleanContent(projectPath, isV3); err != nil {
			return err
		}
	case "clean-assessments":
		if err := checkFilesContent(projectPath, []string{}, true); err != nil {
			return err
		}
		if err := cleanAssessments(projectPath, isV3); err != nil {
			return err
		}
	case "clean-images":
		if err := checkFilesContent(projectPath, []string{getImgPath(projectPath)}, false); err != nil {
			return err
		}
		if err := cleanFoldersByFileMap(); err != nil {
			return err
		}
	case "clean-code":
		if err := checkFilesContent(projectPath, []string{getCodePath(projectPath)}, false); err != nil {
			return err
		}
		if err := cleanFoldersByFileMap(); err != nil {
			return err
		}
	case "clean-full":
		if err := cleanContent(projectPath, isV3); err != nil {
			return err
		}
		if err := checkFilesContent(projectPath, []string{getImgPath(projectPath), getCodePath(projectPath)}, true); err != nil {
			return err
		}
		if err := cleanAssessments(projectPath, isV3); err != nil {
			return err
		}
		if err := cleanFoldersByFileMap(); err != nil {
			return err
		}
	}
	return nil
}

func loadSections(path string) ([]types.Section, error) {
	jsonFilePath := filepath.Join(path, "metadata.json")
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	var metadata types.Metadata
	err = json.Unmarshal(bytes, &metadata)
	if err != nil {
		return nil, err
	}
	return metadata.Sections, nil
}

func cleanAssessments(path string, isV3 bool) error {
	if isV3 {
		return cleanAssessmentsV3(path)
	}
	var root []interface{}
	jsonFilePath := filepath.Join(path, "assessments.json")
	jsonFile, err := os.OpenFile(jsonFilePath, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
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
		if assessmentMap[taskId] {
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

func cleanAssessmentsV3(path string) error {
	files, err := filesUtils.GetListFiles(filepath.Join(path, "assessments"))
	if err != nil {
		return err
	}

	for _, filePath := range files {
		file := filepath.Base(filePath)
		fmt.Printf("File: %s\n", file)
		taskId := strings.TrimRight(file, ".json")
		_, exists := assessmentMap[taskId]
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

func cleanFoldersByFileMap() error {
	for _, file := range fileMap {
		if !file.NeedRemove {
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

func cleanContentFolder(fileNames map[string]bool, path string) error {
	files, err := filesUtils.GetListFiles(filepath.Join(path, "content"))
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
		files, err := filesUtils.GetListFiles(path)
		if err != nil {
			return err
		}
		for _, filePath := range files {
			relativFilePath := strings.Replace(filePath, path, "", 1)
			fileMap[relativFilePath] = types.FileInfo{FullPath: filePath, NeedRemove: true}
			file_search_dict = file_search_dict + relativFilePath + "|"
		}
	}
	file_search_dict = strings.TrimRight(file_search_dict, "|")
	checkDirectory(rootPath, includeAssessments)
	return nil
}

func checkDirectory(pathToDirectory string, includeAssessments bool) error {
	files, err := os.ReadDir(pathToDirectory)
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
	content, err := filesUtils.ReadFile(pathToFile)
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

	if includeAssessments {
		re = regexp.MustCompile(`{.*?|assessment}\((?P<taskId>[a-zA-Z\d-]*)\)`)
		taskIdIndex := re.SubexpIndex("taskId")
		matches := re.FindAllStringSubmatch(content, -1)
		for _, v := range matches {
			assessmentMap[v[taskIdIndex]] = true
		}
	}
	return nil
}

func cleanContent(path string, isV3 bool) error {
	if isV3 {
		return cleanContentV3(path)
	}
	sections, err := loadSections(path)
	if err != nil {
		return err
	}
	sectionPaths := make(map[string]bool)
	for _, item := range sections {
		sectionPaths[filepath.Base(item.ContentFile)] = true
	}

	err = cleanContentFolder(sectionPaths, path)
	if err != nil {
		return err
	}

	return nil
}

func cleanContentV3(path string) error {
	pathToContent := filepath.Join(path, "content")
	err := cleanContentFolderV3(pathToContent)
	if err != nil {
		return err
	}
	return nil
}

func cleanContentFolderV3(path string) error {
	indexJson := filepath.Join(path, "index.json")
	arr, err := utils.GetArrayFromJson[string](indexJson, "order")
	arr = append(arr, "index")
	if err != nil {
		return err
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, file := range files {
		pathToFile := filepath.Join(path, file.Name())
		itemOrder := strings.TrimRight(file.Name(), ".json")
		itemOrder = strings.TrimRight(itemOrder, ".md")
		if !utils.ContainedInArray(arr, itemOrder) {
			fmt.Printf("DELETING FILE!: %s\n", pathToFile)
			err = os.Remove(pathToFile)
			if err != nil {
				return err
			}
			continue
		}
		if file.IsDir() {
			cleanContentFolderV3(pathToFile)
		}
	}
	return nil
}

func getImgPath(projectPath string) string {
	return filepath.Join(projectPath, "img")
}

func getCodePath(projectPath string) string {
	return filepath.Join(projectPath, "../code")
}

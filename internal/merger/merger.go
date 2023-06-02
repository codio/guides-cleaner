package merger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	utils "github.com/codio/guides-cleaner/internal/utils"
	filesUtils "github.com/codio/guides-cleaner/internal/utils/files"
)

func MergeAssignments(destAssignmentPath, mergeAssignmentPath string) error {
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
	err = filesUtils.CopyFiles(destAssignmentPath, mergeAssignmentPath)
	if err != nil {
		return err
	}
	return nil
}

func mergeAssessmentsJson(destAssignmentPath, mergeAssignmentPath string) error {
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
		if ok && !utils.ContainedInArray(dstIds, id) {
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

func getMergeArray(pathToFile, key string) ([]interface{}, error) {
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

func mergeIntoDst(pathToFile, key string, mergeArr []interface{}) error {
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
		if ok && !utils.ContainedInArray(srcIds, id) {
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

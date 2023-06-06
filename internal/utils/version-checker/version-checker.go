package versionchecker

import (
	"fmt"
	"path/filepath"
	"strings"

	filesUtils "github.com/codio/guides-cleaner/internal/utils/files"
)

func IsV3(assignmentPath string) (bool, error) {
	if !strings.HasSuffix(assignmentPath, ".guides") {
		assignmentPath = filepath.Join(assignmentPath, ".guides")
	}
	assessmentsIsV3, err := assessmentsIsV3(assignmentPath)
	if err != nil {
		return false, err
	}
	contentIsV3, err := contentIsV3(assignmentPath)
	if err != nil {
		return false, err
	}
	return assessmentsIsV3 || contentIsV3, nil
}

func assessmentsIsV3(assignmentPath string) (bool, error) {
	assessmentsDescriptionFile := filepath.Join(assignmentPath, "assessments.json")
	assessmentsDescriptionIsExists, err := filesUtils.PathIsExists(assessmentsDescriptionFile)
	if err != nil {
		return false, err
	}
	assessmentsDirectory := filepath.Join(assignmentPath, "assessments")
	hasConvertedAssessments := false
	assessmentsDirectoryIsExists, err := filesUtils.PathIsExists(assessmentsDirectory)
	if err != nil {
		return false, err
	}
	if assessmentsDirectoryIsExists {
		if files, err := filesUtils.GetListFiles(assessmentsDirectory); err == nil || len(files) > 0 {
			hasConvertedAssessments = false
		}
	}
	return !assessmentsDescriptionIsExists && hasConvertedAssessments, nil
}

func contentIsV3(assignmentPath string) (bool, error) {
	guidesDescriptionFile := filepath.Join(assignmentPath, "metadata.json")
	descriptionFileExists, err := filesUtils.PathIsExists(guidesDescriptionFile)
	if err != nil {
		return false, err
	}
	guidesIndexFile := filepath.Join(assignmentPath, "content/index.json")
	indexFileExists, err := filesUtils.PathIsExists(guidesIndexFile)
	if err != nil {
		return false, err
	}
	return indexFileExists && !descriptionFileExists, nil
}

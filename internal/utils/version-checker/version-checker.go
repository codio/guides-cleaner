package versionchecker

import (
	"path/filepath"

	filesUtils "github.com/codio/guides-cleaner/internal/utils/files"
)

func IsV3(assignmentPath string) (bool, error) {
	assessmentsIsV3, err := assessmentsIsV3(assignmentPath)
	if err != nil {
		return false, err
	}
	contentIsV3, err := contentIsV3(assignmentPath)
	if err != nil {
		return false, err
	}
	// for default clean with path .guides not work now
	return assessmentsIsV3 || contentIsV3, nil
}

func assessmentsIsV3(assignmentPath string) (bool, error) {
	assessmentsDescriptionFile := filepath.Join(assignmentPath, ".guides/assessments.json")
	assessmentsDescriptionIsExists, err := filesUtils.PathIsExists(assessmentsDescriptionFile)
	if err != nil {
		return false, err
	}
	assessmentsDirectory := filepath.Join(assignmentPath, ".guides/assessments/")
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
	guidesDescriptionFile := filepath.Join(assignmentPath, ".guides/metadata.json")
	descriptionFileExists, err := filesUtils.PathIsExists(guidesDescriptionFile)
	if err != nil {
		return false, err
	}
	guidesIndexFile := filepath.Join(assignmentPath, ".guides/content/index.json")
	indexFileExists, err := filesUtils.PathIsExists(guidesIndexFile)
	if err != nil {
		return false, err
	}
	return indexFileExists && !descriptionFileExists, nil
}

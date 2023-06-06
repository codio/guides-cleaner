package app

import (
	"fmt"
	"os"

	"github.com/codio/guides-cleaner/internal/cleaner"
	"github.com/codio/guides-cleaner/internal/merger"
)

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
		case "clean-content", "clean-assessments", "clean-images", "clean-code", "clean-full":
			err := cleaner.Clean(args[0], projectPath)
			check(err)
		case "merge":
			if len(args) > 2 {
				destAssignmentPath := args[1]
				mergeAssignmentPath := args[2]
				err := merger.MergeAssignments(destAssignmentPath, mergeAssignmentPath)
				check(err)
			} else {
				printHelp()
			}
		default:
			printHelp()
		}
	}
}

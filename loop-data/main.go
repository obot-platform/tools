package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/gptscript-ai/go-gptscript"
)

var datasetIDRegex = regexp.MustCompile(`gds://[a-z0-9]+`)

func main() {
	fileName := os.Getenv("FILE_NAME")

	if fileName == "" {
		fmt.Println("Error: file name not provided")
		os.Exit(1)
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		fmt.Println("Error: failed to create gptscript client", err)
		os.Exit(1)
	}

	// Make sure there is no existing file with the same name.
	_ = gptscriptClient.DeleteFileInWorkspace(context.Background(), fileName)

	datasetID := os.Getenv("DATASET_ID")
	dataList := os.Getenv("DATA_LIST")

	if datasetID == "" && dataList == "" {
		fmt.Println("Error: no dataset ID or data list provided")
		os.Exit(1)
	}

	if datasetID != "" && dataList != "" {
		fmt.Println("Error: both dataset ID and data list were provided. Please provide only one of the two.")
		os.Exit(1)
	}

	var data []byte
	if datasetID != "" {
		if !datasetIDRegex.MatchString(datasetID) {
			fmt.Println("Error: invalid dataset ID")
			os.Exit(1)
		}

		data = []byte(datasetID)
	} else {
		var list []string
		if err := json.Unmarshal([]byte(dataList), &list); err != nil {
			fmt.Println("Error: failed to unmarshal JSON list of strings", err)
			os.Exit(1)
		}

		data = []byte(dataList)
	}

	if err := gptscriptClient.WriteFileInWorkspace(context.Background(), fileName, data); err != nil {
		fmt.Println("Error: failed to write data to file", err)
		os.Exit(1)
	}

	fmt.Println("success")
}

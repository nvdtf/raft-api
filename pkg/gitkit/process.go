package gitkit

import (
	"context"
	"fmt"
	"regexp"
)

type ProcessResult struct {
	Files []File `json:"files"`
}

type File struct {
	Path             string `json:"path"`
	Filename         string `json:"filename"`
	Type             string `json:"type"`
	OriginalContent  string `json:"originalContent"`
	ProcessedContent string `json:"processedContent"`
}

func (gk *GitKit) Process(network string) (*ProcessResult, error) {
	ctx := context.Background()

	query := fmt.Sprintf("filename:.cdc+repo:%s/%s", gk.owner, gk.repo)

	contracts, err := gk.GetContractDeployments(network)
	if err != nil {
		return nil, err
	}

	results, _, err := gk.client.Search.Code(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	var files []File

	for _, res := range results.CodeResults {
		contents, err := gk.Read(*res.Path)
		if err != nil {
			return nil, err
		}

		fileType := "Unknown"
		if IsContract(string(contents)) {
			fileType = "Contract"
		} else if IsTransaction(string(contents)) {
			fileType = "Transaction"
		} else if IsScript(string(contents)) {
			fileType = "Script"
		}

		processResult := ReplaceImports(contents, contracts)

		files = append(files, File{
			Path:             *res.Path,
			Filename:         *res.Name,
			Type:             fileType,
			OriginalContent:  string(contents),
			ProcessedContent: string(processResult),
		})
	}

	final := &ProcessResult{
		Files: files,
	}

	return final, nil
}

func IsContract(code string) bool {
	r, _ := regexp.Compile(`pub contract (.*){`)
	matches := r.FindAllStringSubmatch(string(code), -1)
	return len(matches) > 0
}

func IsTransaction(code string) bool {
	r, _ := regexp.Compile(`transaction(.*){`)
	matches := r.FindAllStringSubmatch(string(code), -1)
	return len(matches) > 0
}

func IsScript(code string) bool {
	r, _ := regexp.Compile(`pub fun main(.*){`)
	matches := r.FindAllStringSubmatch(string(code), -1)
	return len(matches) > 0
}

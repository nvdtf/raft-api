package gitkit

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/onflow/cadence/runtime/parser2"
)

type ProcessResult struct {
	Documents    []File          `json:"documents"`
	Contracts    []File          `json:"contracts"`
	Scripts      []ProcessedFile `json:"scripts"`
	Transactions []ProcessedFile `json:"transactions"`
}

type File struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Contents string `json:"contents"`
}

type ProcessedFile struct {
	File

	Arguments []Argument `json:"arguments"`
}

type Argument struct {
	Name    string `json:"name"`
	ArgType string `json:"type"`
}

func (gk *GitKit) Process(owner string, repo string, network string) (*ProcessResult, error) {
	ctx := context.Background()

	fmt.Printf("Processing %s/%s (%s)\n", owner, repo, network)

	contracts, scripts, transactions, err := gk.processCadenceFiles(ctx, owner, repo, network)
	if err != nil {
		return nil, err
	}

	documents, err := gk.processDocumentFiles(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return &ProcessResult{
		Documents:    documents,
		Contracts:    contracts,
		Scripts:      scripts,
		Transactions: transactions,
	}, nil
}

func (gk *GitKit) processCadenceFiles(
	ctx context.Context,
	owner string,
	repo string,
	network string,
) (
	contracts []File,
	scripts []ProcessedFile,
	transactions []ProcessedFile,
	err error,
) {
	query := fmt.Sprintf("filename:.cdc+repo:%s/%s", owner, repo)

	contractsMap, err := gk.GetContractDeployments(owner, repo, network)
	if err != nil {
		return
	}

	results, _, err := gk.client.Search.Code(ctx, query, nil)
	if err != nil {
		return
	}

	for _, res := range results.CodeResults {
		contents, errRead := gk.Read(owner, repo, *res.Path)
		if errRead != nil {
			err = errRead
			return
		}

		processResult := ReplaceImports(contents, contractsMap)

		file := File{
			Path:     *res.Path,
			Filename: *res.Name,
			Contents: string(processResult),
		}

		if IsContract(string(contents)) {
			file.Type = "Contract"
			contracts = append(contracts, file)
		} else if IsTransaction(string(contents)) {
			file.Type = "Transaction"

			args, errParse := ParseTransactionArguments(string(contents))
			if errParse != nil {
				err = errParse
				return
			}

			processedFile := ProcessedFile{
				File:      file,
				Arguments: args,
			}
			transactions = append(transactions, processedFile)
		} else if IsScript(string(contents)) {
			file.Type = "Script"

			args, errParse := ParseScriptArguments(string(contents))
			if errParse != nil {
				err = errParse
				return
			}

			processedFile := ProcessedFile{
				File:      file,
				Arguments: args,
			}
			scripts = append(scripts, processedFile)
		}
	}

	return
}

func (gk *GitKit) processDocumentFiles(
	ctx context.Context,
	owner string,
	repo string,
) (
	documents []File,
	err error,
) {
	query := fmt.Sprintf("filename:.md+repo:%s/%s", owner, repo)

	results, _, err := gk.client.Search.Code(ctx, query, nil)
	if err != nil {
		return
	}

	for _, res := range results.CodeResults {
		contents, errRead := gk.Read(owner, repo, *res.Path)
		if errRead != nil {
			err = errRead
			return
		}

		documents = append(documents, File{
			Type:     "Document",
			Path:     *res.Path,
			Filename: *res.Name,
			Contents: string(contents),
		})
	}

	return
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

func ReplaceImports(code []byte, contractRef map[string]string) []byte {
	r, _ := regexp.Compile(`import (?P<Contract>\w*) from "(.*).cdc"`)
	matches := r.FindAllStringSubmatch(string(code), -1)

	result := string(code)

	for i := range matches {
		result = strings.ReplaceAll(
			result,
			matches[i][0],
			fmt.Sprintf("import %s from 0x%s",
				matches[i][1],
				contractRef[matches[i][1]],
			),
		)
	}

	return []byte(result)

}

func ParseTransactionArguments(
	code string,
) (
	args []Argument,
	err error,
) {
	program, err := parser2.ParseProgram(code, nil)
	if err != nil {
		return
	}

	if program.SoleTransactionDeclaration() != nil {
		if program.SoleTransactionDeclaration().ParameterList != nil {
			for _, param := range program.SoleTransactionDeclaration().ParameterList.Parameters {
				args = append(args, Argument{
					Name:    param.Identifier.String(),
					ArgType: param.TypeAnnotation.String(),
				})
			}
		}
	}

	return
}

func ParseScriptArguments(
	code string,
) (
	args []Argument,
	err error,
) {
	program, err := parser2.ParseProgram(code, nil)
	if err != nil {
		return
	}

	if program.FunctionDeclarations() != nil && len(program.FunctionDeclarations()) == 1 {
		if program.FunctionDeclarations()[0].ParameterList != nil {
			for _, param := range program.FunctionDeclarations()[0].ParameterList.Parameters {
				args = append(args, Argument{
					Name:    param.Identifier.String(),
					ArgType: param.TypeAnnotation.String(),
				})
			}
		}
	}

	return
}

package gitkit

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/onflow/cadence/runtime/parser2"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/http"
)

type ProcessResult struct {
	Documents    []File           `json:"documents"`
	Contracts    []DeployedFile   `json:"contracts"`
	Scripts      []ExecutableFile `json:"scripts"`
	Transactions []ExecutableFile `json:"transactions"`
}

type File struct {
	Type     string   `json:"type"`
	Path     string   `json:"path"`
	Filename string   `json:"filename"`
	Contents string   `json:"contents"`
	Errors   []string `json:"errors"`
}

type DeployedFile struct {
	File

	Address string `json:"address"`
	Network string `json:"network"`
}

type ExecutableFile struct {
	File

	Arguments []Argument `json:"arguments"`
}

type Argument struct {
	Name    string `json:"name"`
	ArgType string `json:"type"`
}

func (gk *GitKit) Process(
	owner string,
	repo string,
	network string,
) (
	*ProcessResult,
	error,
) {
	ctx := context.Background()

	fmt.Printf("Processing %s/%s (%s)\n", owner, repo, network)

	documents, err := gk.processDocumentFiles(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	contractsMap := gk.getContractsMap(owner, repo, network, documents)

	fmt.Println(contractsMap)

	contracts, scripts, transactions, err := gk.processCadenceFiles(ctx, owner, repo, network, contractsMap)
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

func (gk *GitKit) getContractsMap(
	owner string,
	repo string,
	network string,
	documents []File,
) (
	contractsMap map[string]string,
) {
	contractsMap, err := gk.parseFlowJsonFile(owner, repo, network)
	if err != nil {
		fmt.Println(err)
	}

	var docContractMap map[string]string
	for _, f := range documents {
		if strings.EqualFold(f.Filename, "README.md") {
			docContractMap, err = gk.parseFileForContracts(network, f)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	fmt.Println(docContractMap)

	return
}

func (gk *GitKit) processCadenceFiles(
	ctx context.Context,
	owner string,
	repo string,
	network string,
	contractsMap map[string]string,
) (
	contracts []DeployedFile,
	scripts []ExecutableFile,
	transactions []ExecutableFile,
	err error,
) {
	query := fmt.Sprintf("filename:.cdc+repo:%s/%s", owner, repo)

	contracts = []DeployedFile{}
	scripts = []ExecutableFile{}
	transactions = []ExecutableFile{}

	// TODO: pass page number here + support pagination
	results, _, err := gk.client.Search.Code(ctx, query, nil)
	if err != nil {
		return
	}

	var networkHost string
	if strings.EqualFold("testnet", network) {
		networkHost = http.TestnetHost
	} else if strings.EqualFold("mainnet", network) {
		networkHost = http.MainnetHost
	}
	flowClient, err := http.NewClient(networkHost)
	if err != nil {
		return
	}

	for _, res := range results.CodeResults {
		contents, errRead := gk.Read(owner, repo, *res.Path)
		if errRead != nil {
			err = errRead
			return
		}

		file := File{
			Path:     *res.Path,
			Filename: *res.Name,
			Errors:   []string{},
		}

		isContract, contractName := IsContract(string(contents))
		if isContract {
			address := contractsMap[contractName]

			file.Type = "Contract"

			if len(address) > 0 {
				account, errFlow := flowClient.GetAccount(ctx, flow.HexToAddress(address))
				if errFlow != nil {
					err = errFlow
					return
				}
				file.Contents = string(account.Contracts[contractName])

			} else {
				file.Contents = string(contents)
			}

			contractFile := DeployedFile{
				File:    file,
				Network: network,
				Address: address,
			}

			contracts = append(contracts, contractFile)

		} else if IsTransaction(string(contents)) || IsScript(string(contents)) {

			processResult, errors := ReplaceImports(contents, contractsMap)
			file.Errors = append(file.Errors, errors...)

			file.Contents = string(processResult)

			var errParse error
			var args []Argument
			if IsTransaction(string(contents)) {
				file.Type = "Transaction"

				args, errParse = ParseTransactionArguments(string(contents))
				if errParse != nil {
					file.Errors = append(file.Errors, errParse.Error())
				}

				transactions = append(transactions, ExecutableFile{
					File:      file,
					Arguments: args,
				})
			} else {
				file.Type = "Script"

				args, errParse = ParseScriptArguments(string(contents))
				if errParse != nil {
					file.Errors = append(file.Errors, errParse.Error())
				}

				scripts = append(scripts, ExecutableFile{
					File:      file,
					Arguments: args,
				})
			}
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
			Errors:   []string{},
		})
	}

	return
}

func IsContract(code string) (bool, string) {
	r, _ := regexp.Compile(`pub contract (.*){`)
	contractName := r.FindStringSubmatch(code)
	if len(contractName) > 0 {
		return true, strings.Trim(contractName[1], " ")
	}
	return false, ""
}

func IsTransaction(code string) bool {
	r, _ := regexp.Compile(`transaction(.*){`)
	matches := r.FindAllStringSubmatch(code, -1)
	return len(matches) > 0
}

func IsScript(code string) bool {
	r, _ := regexp.Compile(`pub fun main(.*){`)
	matches := r.FindAllStringSubmatch(code, -1)
	return len(matches) > 0
}

func ReplaceImports(
	code []byte,
	contractRef map[string]string,
) (
	[]byte,
	[]string,
) {
	r, _ := regexp.Compile(`import (?P<Contract>\w*) from "(.*).cdc"`)
	matches := r.FindAllStringSubmatch(string(code), -1)

	result := string(code)

	errors := []string{}

	for i := range matches {
		contractName := matches[i][1]
		address, exists := contractRef[contractName]

		if exists {
			result = strings.ReplaceAll(
				result,
				matches[i][0],
				fmt.Sprintf("import %s from 0x%s",
					contractName,
					address,
				),
			)
		} else {
			errors = append(errors, fmt.Sprintf("Cannot resolve import for %s", contractName))
		}

	}

	return []byte(result), errors

}

func ParseTransactionArguments(
	code string,
) (
	args []Argument,
	err error,
) {
	args = []Argument{}

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
	args = []Argument{}

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

func (gk *GitKit) parseFileForContracts(
	network string,
	file File,
) (
	contractsMap map[string]string,
	err error,
) {
	r, _ := regexp.Compile(`0x[0-9a-fA-F]{16}`)
	docAddresses := r.FindAllString(file.Contents, -1)

	viableAddresses
	return
}

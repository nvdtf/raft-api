package main

import (
	"fmt"

	"github.com/nvdtf/raft-api/pkg/gitkit"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/contracts"
)

func main() {
	network := "mainnet"
	readerWriter := gitkit.NewGitReaderWriter("onflow", "flow-ft")
	configPath := []string{"flow.json"}
	codeFilename := "transactions/transfer_tokens.cdc"

	state, err := flowkit.Load(configPath, readerWriter)
	if err != nil {
		panic(err)
	}

	code, err := readerWriter.ReadFile(codeFilename)
	if err != nil {
		panic(err)
	}

	resolver, err := contracts.NewResolver(code)
	if err != nil {
		panic(err)
	}

	if resolver.HasFileImports() {
		contractsNetwork, err := state.DeploymentContractsByNetwork(network)
		if err != nil {
			panic(err)
		}

		code, err = resolver.ResolveImports(
			codeFilename,
			contractsNetwork,
			state.AliasesForNetwork(network),
		)
		if err != nil {
			panic(err)
		}

		fmt.Println(code)
	}

}

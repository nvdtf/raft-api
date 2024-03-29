package gitkit

import (
	"context"
	"regexp"
	"strings"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/http"
)

func getKnownContractAddresses(
	network string,
) map[string]string {
	if strings.EqualFold(network, "mainnet") {
		return map[string]string{
			"FungibleToken":     "f233dcee88fe0abe",
			"NonFungibleToken":  "1d7e57aa55817448",
			"MetadataViews":     "1d7e57aa55817448",
			"FlowToken":         "1654653399040a61",
			"FlowStorageFees":   "e467b9dd11fa00df",
			"DapperUtilityCoin": "ead892083b3e2c6c",
		}
	} else if strings.EqualFold(network, "testnet") {
		return map[string]string{
			"FungibleToken":     "9a0766d93b6608b7",
			"NonFungibleToken":  "631e88ae7f1d7c20",
			"MetadataViews":     "631e88ae7f1d7c20",
			"FlowToken":         "7e60df042a9c0868",
			"FlowStorageFees":   "8c5303eaa26202d6",
			"DapperUtilityCoin": "82e283f88a62e65",
		}
	}

	return map[string]string{}
}

func (gk *GitKit) getContractsMap(
	ctx context.Context,
	owner string,
	repo string,
	network string,
	flowClient *http.Client,
	documents []File,
) (
	contractsMap map[string]string,
) {
	log := gk.logger.With(
		"repo", owner+"/"+repo,
		"network", network,
	)

	// init with contract addresses from flow.json
	contractsMap, err := gk.parseFlowJsonFile(ctx, owner, repo, network)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			log.Info("No flow.json file")
		} else {
			log.With("error", err).Error("error parsing flow.json file")
		}
	}

	// add addresses from md files
	var docContractMap map[string]string
	for _, f := range documents {
		if strings.EqualFold(f.Path, "README.md") {
			docContractMap, err = gk.parseFileForContracts(ctx, network, flowClient, f)
			if err != nil {
				log.With("error", err).Error("error extracting contracts from document")
			}
		}
	}
	addToMapIfNotExists(&contractsMap, docContractMap)

	// add known addresses
	addToMapIfNotExists(&contractsMap, getKnownContractAddresses(network))

	return
}

func (gk *GitKit) parseFileForContracts(
	ctx context.Context,
	network string,
	flowClient *http.Client,
	file File,
) (
	contractsMap map[string]string,
	err error,
) {
	contractsMap = make(map[string]string)

	r, _ := regexp.Compile(`0x[0-9a-fA-F]{16}`)
	docAddresses := r.FindAllString(file.Contents, -1)

	viableAddresses := map[string]bool{}

	for _, docAddress := range docAddresses {
		address := flow.HexToAddress(docAddress)

		isValid := false
		if strings.EqualFold(network, "testnet") {
			isValid = address.IsValid(flow.Testnet)
		} else if strings.EqualFold(network, "mainnet") {
			isValid = address.IsValid(flow.Mainnet)
		}

		if isValid {
			viableAddresses[docAddress] = true
		}
	}

	for address := range viableAddresses {
		account, errFlow := flowClient.GetAccount(ctx, flow.HexToAddress(address))
		if errFlow != nil {
			gk.logger.With(
				"address", address,
				"file", file.Path,
				"error", errFlow,
			).Error("error calling GetAccount")
		}

		for contract := range account.Contracts {
			contractsMap[contract] = strings.TrimPrefix(address, "0x")
		}
	}

	return
}

func addToMapIfNotExists(input *map[string]string, add map[string]string) {
	for k, v := range add {
		_, exists := (*input)[k]
		if !exists {
			(*input)[k] = v
		}
	}
}

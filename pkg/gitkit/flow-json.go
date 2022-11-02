package gitkit

import (
	"context"
	"encoding/json"
	"strings"
)

const FLOW_JSON_PATH = "./flow.json"

type flowJsonConfig struct {
	Contracts jsonContracts `json:"contracts"`
}

type jsonContracts map[string]jsonContract

type jsonContract struct {
	Simple   string
	Advanced jsonContractAdvanced
}

type jsonContractAdvanced struct {
	Source  string            `json:"source"`
	Aliases map[string]string `json:"aliases"`
}

func (j *jsonContract) UnmarshalJSON(b []byte) error {
	var source string
	var advancedFormat jsonContractAdvanced

	// simple
	err := json.Unmarshal(b, &source)
	if err == nil {
		j.Simple = source
		return nil
	}

	// advanced
	err = json.Unmarshal(b, &advancedFormat)
	if err == nil {
		j.Advanced = advancedFormat
	} else {
		return err
	}

	return nil
}

func parseJson(flowJson []byte) (*flowJsonConfig, error) {
	result := &flowJsonConfig{}

	err := json.Unmarshal(flowJson, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (gk *GitKit) parseFlowJsonFile(
	ctx context.Context,
	owner string,
	repo string,
	network string,
) (
	contractMap map[string]string,
	err error,
) {
	contractMap = make(map[string]string)

	flowJson, err := gk.Read(ctx, owner, repo, FLOW_JSON_PATH)
	if err != nil {
		return
	}

	contracts, err := parseJson(flowJson)
	if err != nil {
		return
	}

	for name, j := range contracts.Contracts {
		for n, a := range j.Advanced.Aliases {
			if strings.EqualFold(strings.Trim(n, " "), network) {
				contractMap[name] = strings.TrimPrefix(a, "0x")
			}
		}
	}

	return
}

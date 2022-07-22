package gitkit

import "encoding/json"

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

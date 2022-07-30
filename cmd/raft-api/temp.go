package main

import (
	"fmt"

	"github.com/onflow/cadence/runtime/parser2"
)

type Argument struct {
	name    string
	argType string
}

func temp() {
	sampleCode := `
	import FlowContractAudits from "../contract/FlowContractAudits.cdc"

	pub fun main(code: String): String {
		return FlowContractAudits.hashContractCode(code)
	}
	`

	program, err := parser2.ParseProgram(sampleCode, nil)
	if err != nil {
		panic(err)
	}

	for _, param := range program.FunctionDeclarations()[0].ParameterList.Parameters {
		fmt.Println(param.Identifier)
	}

	var result []Argument

	if program.SoleTransactionDeclaration() != nil {
		if program.SoleTransactionDeclaration().ParameterList != nil {
			for _, param := range program.SoleTransactionDeclaration().ParameterList.Parameters {
				result = append(result, Argument{
					name:    param.Identifier.String(),
					argType: param.TypeAnnotation.String(),
				})
			}
		}
	}

	fmt.Println(result)
}

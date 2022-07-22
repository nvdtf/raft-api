package gitkit

import (
	"fmt"
	"regexp"
	"strings"
)

func ReplaceImports(code []byte, contractRef map[string]string) []byte {
	r, _ := regexp.Compile(`import (?P<Contract>\w*) from "(.*).cdc"`)
	matches := r.FindAllStringSubmatch(string(code), -1)

	result := string(code)

	for i := range matches {
		result = strings.ReplaceAll(
			result,
			matches[i][0],
			fmt.Sprintf("import %s from %s",
				matches[i][1],
				contractRef[matches[i][1]],
			),
		)
	}

	return []byte(result)

}

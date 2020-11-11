package utils

import (
	"encoding/json"

	"github.com/davecgh/go-spew/spew"
)

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")

	return string(s)
}

func PrettyDump(i interface{}) string {
	return spew.Sdump(i)
}

//go:build wasm && js
// +build wasm,js

package main

import (
	"encoding/json"
	"errors"
	"strings"
	"syscall/js"

	"github.com/gql-schema/gqls/internal/gql2neo4j"
	"github.com/gql-schema/gqls/internal/gql2neo4j/convert"
)

// version is set to the current version at build time.
var version = "dev"

type conversionRequest struct {
	GQL          string `json:"gql"`
	Neo4jVersion string `json:"neo4jVersion"`
	Neo4jEdition string `json:"neo4jEdition"`
	ApocEnabled  bool   `json:"apocEnabled"`
}

func main() {
	js.Global().Set("convertGQL", js.FuncOf(convertGQL))
	js.Global().Set("getVersionInfo", js.FuncOf(getVersionInfo))

	select {}
}

func convertGQL(_ js.Value, args []js.Value) any {
	if len(args) != 1 {
		return jsError(errors.New("expected a single JSON‑encoded argument"))
	}

	if args[0].Type() != js.TypeString {
		return jsError(errors.New("argument must be a JSON string"))
	}

	var req conversionRequest
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return jsError(err)
	}

	var edition convert.Neo4jEdition
	switch strings.ToLower(req.Neo4jEdition) {
	case "community":
		edition = convert.Neo4jCommunityEdition
	case "enterprise":
		edition = convert.Neo4jEnterpriseEdition
	default:
		edition = convert.Neo4jEnterpriseEdition
	}

	cypher, err := gql2neo4j.ConvertString(req.GQL, convert.DatabaseMetadata{
		Version:     req.Neo4jVersion,
		Edition:     edition,
		APOCEnabled: req.ApocEnabled,
	})
	if err != nil {
		return jsError(err)
	}

	return map[string]any{"neo4j": cypher}
}

func getVersionInfo(_ js.Value, _ []js.Value) any {
	return version
}

func jsError(err error) map[string]any {
	return map[string]any{"error": err.Error()}
}

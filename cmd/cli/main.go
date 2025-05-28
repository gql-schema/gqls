// Package main provides a command-line interface for converting GraphQL schemas to Neo4j format.
package main

import (
	"context"
	"os"

	"github.com/gql-schema/gqls/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	cli.Run(context.Background(), version, commit, os.Args)
}

# Tools

This directory contains tools used during development and testing.
We mostly use the tools directive [introduced in Go 1.24](https://tip.golang.org/doc/go1.24#tools).

To run a tool from inside this directory, you can use the `go tool` command:

```shell
go tool <tool>
```

To execute commands from other directories, use the `-modfile` flag with path of the `tools/go.mod` file.
For example, from the root of the project, run:

```shell
go tool -modfile=./tools/go.mod <tool>
```

## Generating the Parser

This directory also includes the `go:generate` command to generate ANTLR4 parser files.
Ensure the ANTLR4 JAR is available in the current directory as `antlr-4.13.2-complete.jar`.

To generate the parser files, run:

```shell
go generate ./...
```

Alternatively, you can run the `make generate` command from the root of the project.
This will automatically download the required ANTLR4 JAR file if it is not already present.

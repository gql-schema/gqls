//go:build generate

package tools

//go:generate java -jar antlr-4.13.2-complete.jar -Dlanguage=Go -o ../internal/gql/parser -Xexact-output-dir -package parser -no-visitor ../internal/gql/grammar/GQL.g4

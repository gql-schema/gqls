# GQLS - GQL to Neo4j Schema Converter

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/apache-2-0)
[![Test](https://github.com/gql-schema/gqls/actions/workflows/test.yml/badge.svg)](https://github.com/gql-schema/gqls/actions/workflows/test.yml)

<div align="center">
    <img src="web/assets/logo.png" alt="GQLS Logo" width="256px"/>
</div>

GQLS is a CLI tool that converts GQL schema definitions (ISO/IEC 39075:2024) into Neo4j-compatible Cypher statements. It
helps bridge the gap between the Graph Query Language (GQL) standard and Neo4j's implementation, making it easier to
work with both technologies.

## Features

- Parse GQL schema files and generate Neo4j constraints and indexes
- Support for multiple Neo4j versions (with version-specific validations)
- Command-line interface for ease of use
- Web interface for interactive conversion
- Docker image available for containerized usage
- APOC plugin support for advanced type constraints and validations

## Installation

### Prerequisites

- [Go](https://golang.org/doc/install) (>= 1.23)
- [Make](https://www.gnu.org/software/make/) (for building)
- [Java Runtime Environment (JRE)](https://adoptium.net/de/temurin/releases/) (e.g. Temurin; for parser generation)

### Clone and Build

```shell
git clone https://github.com/gql-schema/gqls.git
cd gqls
make build
```

This will create a `gqls` executable in your current directory.

### Generate the Parser

Before building, generate the ANTLR4 parser files:

```shell
make generate
```

This will populate the `internal/gql/parser/` directory with Go bindings for the GQL grammar.
You must have a valid Java Runtime (JRE) installed and available in your `$PATH` for this step.
Make will automatically download the required ANTLR4 jar file if it is not already present.

### Docker

You can also use the Docker image to run GQLS without installing Go or any dependencies:

```shell
docker run --rm -v $(pwd):/data gql-schema/gqls:latest --input /data/schema.gql --output /data/output.cypher
```

### Pre-built Binaries

Pre-built binaries for various platforms are available from
the [GitHub Releases](https://github.com/gql-schema/gqls/releases) page.

## Usage

### Command Line Interface

Run the `gqls` command with the required options:

```bash
./gqls --input schema.gql --neo4j-version 5.9
```

#### Options

| Flag              | Description                                                             |
|-------------------|-------------------------------------------------------------------------|
| `--input`, `-i`   | Path to the GQL schema file. Use - or omit to read from stdin.          |
| `--output`, `-o`  | Path to the output file. Defaults to stdout.                            |
| `--neo4j-version` | Target Neo4j version (e.g., 5.9). Defaults to latest supported version. |
| `--apoc`          | Generate APOC triggers. Defaults to false.                              |
| `--serve`         | Start web server on the specified address (e.g., `:8080`)               |

### Web Interface

GQLS includes a web interface for interactive schema conversion. There are two ways to use it:

#### Local Server

You can also run a local web server:

```bash
./gqls --serve :8080
```

Then open your browser to `http://localhost:8080` to use the interactive UI.

### Example

Given a GQL schema with precision and length constraints:

<!-- BEGIN EXAMPLE INPUT -->

```gql
CREATE PROPERTY GRAPH socialNetworkType {
    (:Person {
        givenName :: STRING(2, 50) NOT NULL, -- String with length constraints (2 to 50 chars)
        familyName :: VARCHAR(50) NOT NULL,  -- Variable length string (max 50 chars)
        age :: INT8 NOT NULL,                -- Integer with precision constraints (-128 to 127)
        bio :: CHAR(200)                     -- Fixed length string (200 chars)
    }) PRIMARY KEY (name),
    (:Post {
        title :: STRING(1, 100) NOT NULL,    -- String with length constraints (1 to 100 chars)
        content :: STRING,
        created_at :: DATE
    }) UNIQUE (title, content),
    (:Person)~[:is_friend { since :: DATE NOT NULL }]~(:Person),
    (:Person)-[:likes { since :: DATE NOT NULL }]->(:Post)
}
```

<!-- END EXAMPLE INPUT -->

GQLS will generate Neo4j constraints and triggers:

<!-- BEGIN EXAMPLE OUTPUT -->

```cypher
CREATE CONSTRAINT unique_name FOR (n:Person) REQUIRE n.name IS UNIQUE;
CREATE CONSTRAINT key_name FOR (n:Person) REQUIRE n.name IS NODE KEY;
CREATE CONSTRAINT type_person_givenname FOR (n:Person) REQUIRE n.givenName IS :: STRING;
CREATE CONSTRAINT existence_person_givenname FOR (n:Person) REQUIRE n.givenName IS NOT NULL;
CALL apoc.trigger.add("string_length_person_givenname", "UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.givenName IS NOT NULL AND (size(e.givenName) < 2 OR size(e.givenName) > 50) CALL apoc.util.validate(true, 'givenName must be between 2 and 50 characters', []) RETURN null UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change:Person AND change.givenName IS NOT NULL AND (size(change.givenName) < 2 OR size(change.givenName) > 50) CALL apoc.util.validate(true, 'givenName must be between 2 and 50 characters', []) RETURN null", {phase: "before"});
CREATE CONSTRAINT type_person_familyname FOR (n:Person) REQUIRE n.familyName IS :: STRING;
CREATE CONSTRAINT existence_person_familyname FOR (n:Person) REQUIRE n.familyName IS NOT NULL;
CALL apoc.trigger.add("string_length_person_familyname", "UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.familyName IS NOT NULL AND (size(e.familyName) < 0 OR size(e.familyName) > 50) CALL apoc.util.validate(true, 'familyName must be between 0 and 50 characters', []) RETURN null UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change:Person AND change.familyName IS NOT NULL AND (size(change.familyName) < 0 OR size(change.familyName) > 50) CALL apoc.util.validate(true, 'familyName must be between 0 and 50 characters', []) RETURN null", {phase: "before"});
CREATE CONSTRAINT type_person_age FOR (n:Person) REQUIRE n.age IS :: INTEGER;
CREATE CONSTRAINT existence_person_age FOR (n:Person) REQUIRE n.age IS NOT NULL;
CALL apoc.trigger.add("numeric_range_person_age", "UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.age IS NOT NULL AND (e.age < -128 OR e.age > 127) CALL apoc.util.validate(true, 'age must be between -128 and 127', []) RETURN null UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change:Person AND change.age IS NOT NULL AND (change.age < -128 OR change.age > 127) CALL apoc.util.validate(true, 'age must be between -128 and 127', []) RETURN null", {phase: "before"});
CREATE CONSTRAINT type_person_bio FOR (n:Person) REQUIRE n.bio IS :: STRING;
CALL apoc.trigger.add("string_length_person_bio", "UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.bio IS NOT NULL AND (size(e.bio) < 200 OR size(e.bio) > 200) CALL apoc.util.validate(true, 'bio must be between 200 and 200 characters', []) RETURN null UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change:Person AND change.bio IS NOT NULL AND (size(change.bio) < 200 OR size(change.bio) > 200) CALL apoc.util.validate(true, 'bio must be between 200 and 200 characters', []) RETURN null", {phase: "before"});
CREATE CONSTRAINT type_post_title FOR (n:Post) REQUIRE n.title IS :: STRING;
CREATE CONSTRAINT existence_post_title FOR (n:Post) REQUIRE n.title IS NOT NULL;
CALL apoc.trigger.add("string_length_post_title", "UNWIND $createdNodes AS e WITH e WHERE e:Post AND e.title IS NOT NULL AND (size(e.title) < 1 OR size(e.title) > 100) CALL apoc.util.validate(true, 'title must be between 1 and 100 characters', []) RETURN null UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change:Post AND change.title IS NOT NULL AND (size(change.title) < 1 OR size(change.title) > 100) CALL apoc.util.validate(true, 'title must be between 1 and 100 characters', []) RETURN null", {phase: "before"});
CREATE CONSTRAINT type_post_content FOR (n:Post) REQUIRE n.content IS :: STRING;
CREATE CONSTRAINT type_post_created_at FOR (n:Post) REQUIRE n.created_at IS :: DATE;
CALL apoc.trigger.add("edge_endpoint_is_friend_person_person", "UNWIND $createdRelationships AS rel WITH rel WHERE type(rel) = \"is_friend\" AND NOT (startNode(rel):Person AND endNode(rel):Person) CALL apoc.util.validate( true, \"is_friend relationship must be between Person and Person\", [] ) RETURN null", {phase: "before"});
CREATE CONSTRAINT type_is_friend_since FOR ()-[n:is_friend]->() REQUIRE n.since IS :: DATE;
CREATE CONSTRAINT existence_is_friend_since FOR ()-[n:is_friend]->() REQUIRE n.since IS NOT NULL;
CALL apoc.trigger.add("edge_endpoint_likes_person_post", "UNWIND $createdRelationships AS rel WITH rel WHERE type(rel) = \"likes\" AND NOT (startNode(rel):Person AND endNode(rel):Post) CALL apoc.util.validate( true, \"likes relationship must be between Person and Post\", [] ) RETURN null", {phase: "before"});
CREATE CONSTRAINT type_likes_since FOR ()-[n:likes]->() REQUIRE n.since IS :: DATE;
CREATE CONSTRAINT existence_likes_since FOR ()-[n:likes]->() REQUIRE n.since IS NOT NULL;
```

<!-- END EXAMPLE OUTPUT -->

## GQL to Neo4j Feature Mapping

GQLS translates GQL schema concepts to Neo4j constraints and triggers. Here's how different GQL features are mapped to
Neo4j:

### Data Types

| GQL Type                        | Neo4j Type       | Notes                      |
|---------------------------------|------------------|----------------------------|
| `STRING`, `VARCHAR`, `CHAR`     | `STRING`         | Character string types     |
| `INT8`, `INT16`, `INT32`, etc.  | `INTEGER`        | All exact numeric types    |
| `SMALLINT`, `INTEGER`, `BIGINT` | `INTEGER`        | Verbose numeric types      |
| `FLOAT16`, `FLOAT32`, etc.      | `FLOAT`          | Approximate numeric types  |
| `FLOAT`, `REAL`, `DOUBLE`       | `FLOAT`          | Other floating-point types |
| `BOOL`, `BOOLEAN`               | `BOOLEAN`        | Boolean types              |
| `DATE`                          | `DATE`           | Date type                  |
| `LOCAL DATETIME`, `TIMESTAMP`   | `LOCAL DATETIME` | Local datetime types       |
| `ZONED DATETIME`                | `ZONED DATETIME` | Datetime with timezone     |
| `LOCAL TIME`                    | `LOCAL TIME`     | Local time type            |
| `ZONED TIME`                    | `ZONED TIME`     | Time with timezone         |
| `DURATION`                      | `DURATION`       | Duration type              |

#### Precision and Length Support

GQL allows the specification of length and precision constraints for string and numeric types:

- **String types**: `STRING(min, max)`, `VARCHAR(max)`, `CHAR(fixed)`
    - String length constraints are enforced through APOC triggers when the `--apoc` flag is used
    - Example: `STRING(5, 50)` ensures strings are between 5 and 50 characters

- **Numeric types**:
    - **Integer precision**: `INT8` (-128 to 127), `INT16` (-32,768 to 32,767), etc.
        - Precision constraints for exact numeric types (up to precision 63) are enforced via APOC triggers
        - Example: `INT8` generates a trigger ensuring the value is between -128 and 127
    - **Decimal precision**: `DECIMAL(p,s)` where `p` is precision and `s` is scale
        - Neo4j doesn't support scale, so all numbers are stored as either INTEGER or FLOAT

- **Note**: Length and precision constraints are only enforced at the application level through APOC triggers, not at
  the database schema level

### Constraints

| GQL Feature               | Neo4j Implementation            | Notes                                  |
|---------------------------|---------------------------------|----------------------------------------|
| `NOT NULL`                | Property existence constraint   | Requires Enterprise Edition            |
| `PRIMARY KEY`             | Node key constraint             | Combined uniqueness and existence      |
| `UNIQUE`                  | Uniqueness constraint           | For single or composite properties     |
| Type specification (`::`) | Property type constraint        | Requires Neo4j 5.9+ Enterprise Edition |
| Edge relationships        | Relationship type and direction | Maps directed/undirected edges         |

### APOC Triggers

When using the APOC plugin (enabled with `--apoc` flag), GQLS generates triggers for additional validations:

- String length constraints
- Numeric range constraints
- Edge endpoint validations

## Development

### Linting

This project uses `golangci-lint` for linting. Run the following command to check for linting issues:

```bash
make lint
```

### Testing

Run unit tests using:

```bash
make test
```

To run all tests, including integration tests, use:

```bash
make test-integration
```

You need to have either `docker` or `podman` installed to run the integration tests.

#### Docker setup

When using Docker, no additional configuration is needed. Integration tests will work out of the box.

#### Podman setup

When using Podman, make sure to use a rootfull container and set the following environment variables:

```bash
export TESTCONTAINERS_DRIVER="podman"
export TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED="true"
```

### Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.

## License

This project is dual-licensed under the [MIT License](LICENSE-MIT) and the [Apache License 2.0](LICENSE-APACHE).
You can choose either license for your use.

## Acknowledgments

- [opengql/grammar](https://github.com/opengql/grammar) for the ANTLR4 translation of the GQL grammar.

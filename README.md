# GQLS - GQL to Neo4j Schema Converter

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/apache-2-0)
[![Test](https://github.com/gql-schema/gqls/actions/workflows/test.yml/badge.svg)](https://github.com/gql-schema/gqls/actions/workflows/test.yml)

<div align="center">
  <img src="web/assets/logo.png" alt="GQLS Logo" width="256px"/>
</div>

**GQLS** is a CLI and web-based tool that converts GQL schema definitions (ISO/IEC 39075:2024) into Neo4j-compatible
Cypher statements. It helps bridge the gap between the GQL standard and Neo4j’s implementation, making it easier to work
with both technologies.

## Features

### Standard GQL (ISO/IEC 39075:2024)

| Feature                   | Support | Description                                     |
|---------------------------|---------|-------------------------------------------------|
| **Type Constraints**      | ✅       | Enforces data types via Neo4j and APOC triggers |
| **Existence Constraints** | ✅       | Supports `NOT NULL`                             |
| **Edge Endpoints**        | ✅       | Validates correct relationship connections      |

### Extended GQL

| Feature                                   | Support | Description                                    |
|-------------------------------------------|---------|------------------------------------------------|
| **`XA01` – Cardinality Constraints**      | ⚠️      | Partial support for cardinality enforcement    |
| **`XB01` – Simple Key Constraints**       | ✅       | Unique key on single property                  |
| **`XB02` – Composite Key Constraints**    | ✅       | Multi-property key constraints                 |
| **`XB03` – Simple Unique Constraints**    | ✅       | Enforces uniqueness on one property            |
| **`XB04` – Composite Unique Constraints** | ✅       | Enforces uniqueness across multiple properties |
| **`XB05` – Check Constraints**            | ✅       | Value-based checks: length, range, precision   |

---

## GQL to Neo4j Feature Mapping

### Data Types

| GQL Type                      | Neo4j Type       | Notes                     |
|-------------------------------|------------------|---------------------------|
| `STRING`, `VARCHAR`, `CHAR`   | `STRING`         | Character string types    |
| `INT8`, `INT16`, etc.         | `INTEGER`        | Exact numeric types       |
| `FLOAT`, `REAL`, `DOUBLE`     | `FLOAT`          | Approximate numeric types |
| `BOOLEAN`                     | `BOOLEAN`        | Boolean                   |
| `DATE`                        | `DATE`           | Date type                 |
| `LOCAL DATETIME`, `TIMESTAMP` | `LOCAL DATETIME` | Local datetime            |
| `ZONED DATETIME`              | `ZONED DATETIME` | With timezone             |
| `LOCAL TIME`                  | `LOCAL TIME`     | Local time                |
| `ZONED TIME`                  | `ZONED TIME`     | With timezone             |
| `DURATION`                    | `DURATION`       | Duration type             |

#### Length & Precision Support

- **Strings**: `STRING(min, max)`, `VARCHAR(max)`, `CHAR(fixed)`
    - Enforced via APOC triggers
- **Integers**: e.g., `INT8` → triggers for (-128 to 127)
- **Decimals**: `DECIMAL(p, s)` (stored as INTEGER/FLOAT; scale not enforced)

### Constraints

| GQL Feature      | Neo4j Implementation          | Notes                           |
|------------------|-------------------------------|---------------------------------|
| `NOT NULL`       | Property existence constraint | Enterprise Edition only         |
| `PRIMARY KEY`    | Node key                      | Combines uniqueness + existence |
| `UNIQUE`         | Uniqueness constraint         | Supports composite keys         |
| `:: type`        | Property type constraint      | Requires Neo4j 5.9+ EE          |
| Edge definitions | Relationship validation       | Direction + type checking       |

### APOC Triggers (via `--apoc` flag)

- String length checks
- Numeric range enforcement
- Relationship endpoint validations

## Installation

### Prerequisites

- [Go 1.23+](https://golang.org/doc/install)
- [Make](https://www.gnu.org/software/make/)
- [Java (JRE)](https://adoptium.net/de/temurin/releases/) (for ANTLR parser generation)

### Clone and Build

```bash
git clone https://github.com/gql-schema/gqls.git
cd gqls
make build
```

### Generate Parser Files

```bash
make generate
```

> Requires Java in `$PATH`. Downloads ANTLR4 jar automatically.

### Build from Source

```bash
make build
```

### Pre-built Binaries

Available on [GitHub Releases](https://github.com/gql-schema/gqls/releases)

## Usage

### CLI

```bash
./gqls --input schema.gql --neo4j-version 5.9
```

#### Options

| Flag              | Description                                                       |
|-------------------|-------------------------------------------------------------------|
| `--input`, `-i`   | Input GQL file (or stdin)                                         |
| `--output`, `-o`  | Output file (default: stdout)                                     |
| `--neo4j-version` | Target Neo4j version (default: latest supported)                  |
| `--neo4j-edition` | Neo4j edition (`community`, `enterprise`) (default: `enterprise`) |
| `--apoc`          | Enable APOC triggers (default: false)                             |
| `--serve`         | Start web server (e.g., `:8080`)                                  |

### Web Interface

Start local web server:

```bash
./gqls --serve :8080
```

Access at [http://localhost:8080](http://localhost:8080)

## Development

### Linting

```bash
make lint
```

Uses `golangci-lint`.

### Testing

Run unit tests:

```bash
make test-unit
```

Run integration tests:

```bash
make test-integration
```

Run all tests:

```bash
make test
```

> Requires Docker or Podman

#### Podman Configuration

```bash
export TESTCONTAINERS_DRIVER="podman"
export TESTCONTAINERS_RYUK_CONTAINER_PRIVILEGED="true"
```

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.

## License

This project is dual-licensed under the [MIT License](LICENSE-MIT) and the [Apache License 2.0](LICENSE-APACHE).
You can choose either license for your use.

## Acknowledgments

Thanks to [opengql/grammar](https://github.com/opengql/grammar) for the ANTLR4 GQL grammar.

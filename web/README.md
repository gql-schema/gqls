# GQL to Neo4j Converter - Web Interface

> :info: **INFO:** The frontend code in this directory is primarily AI-generated and for demo purposes only.
> The main focus is on the CLI tool, which is the core of the project.

This directory contains the static website for the GQL to Neo4j converter, designed to be deployed on GitHub Pages.
It will not run as is, as it depends on a WASM module that is built during the `make build-static-site` command.

## Local Development

To run the static site locally, you can use the following commands:

```bash
make build-static-site
cd web && python3 -m http.server 8080
```

Alternatively, you can use the `serve` command from the `makefile`:

```bash
make serve
```

## Deployment

The static site is automatically deployed to GitHub Pages using GitHub Actions.
To deploy the site, simply push changes to the `main` branch.

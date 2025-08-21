# image-detector-action

[![Go](https://img.shields.io/badge/Go-1.XX-blue)](https://golang.org/)
[![GitHub Actions](https://github.com/KacperMalachowski/image-detector-action/actions/workflows/test.yml/badge.svg)](https://github.com/KacperMalachowski/image-detector-action/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> **Simple GitHub Action and CLI tool that gathers Docker image URLs from your repository.**

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Usage](#usage)
  - [As a GitHub Action](#as-a-github-action)
  - [As a CLI Tool](#as-a-cli-tool)
- [Inputs](#inputs)
- [Outputs](#outputs)
- [Examples](#examples)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)
- [References](#references)

---

## Overview

**image-detector-action** scans your repository for Docker image URLs, making it easy to audit, track, and automate container image usage within your codebase.

---

## Features

- **Automated scanning** for Docker image URLs in your repository.
- **GitHub Action** for seamless CI/CD integration.
- **CLI tool** written in Go for local or pipeline use.
- Supports custom configuration and output formats.

---

## Usage

### As a GitHub Action

Add the following to your workflow YAML (e.g., `.github/workflows/image-detector.yml`):

```yaml
name: Detect Docker Images

on:
  push:
    branches: [main]
  pull_request:

jobs:
  detect-images:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run image-detector-action
        id: images
        uses: KacperMalachowski/image-detector-action@v1
        with:
          directory: "images"
      - run: echo ${{ steps.images.outputs.image_urls }}
```

### As a CLI Tool

#### Install

```sh
git clone https://github.com/KacperMalachowski/image-detector-action.git
cd image-detector-action
go build -o image-detector ./cmd/image-detector
```

#### Run

```sh
./image-detector --path . --output report.txt
```

---

## Inputs

| Name     | Description                       | Required | Default |
|----------|-----------------------------------|:--------:|:-------:|
| directory| Path to scan for images           |   No     |   `.`   |

## Outputs

- `image_urlst` — List of detected Docker image URLs.

---

## Examples

See [`.github/workflows/image-detector.yml`](.github/workflows/image-detector.yml) for a full workflow example.

---

## Development

1. **Build locally:**
   ```sh
   go build -o image-detector ./cmd/image-detector
   ```
2. **Run tests:**
   ```sh
   go test ./...
   ```
3. **Lint (optional):**
   ```sh
   golangci-lint run
   ```

---

## Releases

This project uses [Semantic Versioning](https://semver.org/) and automated releases based on [Conventional Commits](https://conventionalcommits.org/).

### Using Specific Versions

When using this action, you can specify exact versions:

```yaml
# Use a specific version
- uses: KacperMalachowski/image-detector-action@v1.0.0

# Use the latest v1.x.x version (recommended)
- uses: KacperMalachowski/image-detector-action@v1

# Use the latest version (less stable)
- uses: KacperMalachowski/image-detector-action@main
```

### Triggering Releases

Releases are automatically created when commits are pushed to the `main` branch using conventional commit format:

- `feat: add new feature` → Minor version bump (v1.1.0)
- `fix: resolve bug` → Patch version bump (v1.0.1)  
- `feat!: breaking change` → Major version bump (v2.0.0)
- `BREAKING CHANGE:` in commit body → Major version bump

### Release Process

1. Commits are analyzed for conventional commit patterns
2. Version is automatically determined based on commit types
3. Release notes are generated from commit messages
4. GitHub release is created with changelog
5. Major version tags (v1, v2, etc.) are updated automatically
6. Docker images are built and tagged with new version

---

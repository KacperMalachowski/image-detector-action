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
      - run: echo ${{ steps.images.outputs.image_urls
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

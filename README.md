# tfcopen-go

## Overview
`tfcopen-go` is a command-line tool that simplifies the process of accessing Terraform Cloud workspaces and projects directly from your terminal. By reading configuration from a `.tfcopen` file, it constructs the appropriate URL and opens it in your default web browser.

## Features
- Reads configuration from a `.tfcopen` file.
- Supports searching for workspaces and projects.
- Automatically detects the operating system to open URLs in the appropriate browser.

## Installation
To install `tfcopen-go`, clone the repository and build the project:

```bash
git clone <repository-url>
cd tfcopen-go
go build -o tfcopen ./cmd/tfcopen.go
```

## Usage
To use the application, navigate to the directory containing your `.tfcopen` file and run:

```bash
./tfcopen
```

You can also use the `--print` or `-p` flag to print the constructed URL without opening it:

```bash
./tfcopen --print
```

## Configuration
The `.tfcopen` file should be structured as follows:

```yaml
workspace: <workspace-name>
search: <search-string>
project: <project-name>
org: <organization-name>
```

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.
# tfcopen

## Overview
`tfcopen` is a command-line tool that simplifies the process of accessing Terraform Cloud workspaces and projects directly from your terminal. By reading configuration from a `.tfcopen` file, it constructs the appropriate URL and opens it in your default web browser.

## Features
- Reads configuration from a `.tfcopen` file.
- Supports searching for workspaces and projects.
- Automatically detects the operating system to open URLs in the appropriate browser.

## Installation
To install `tfcopen`, clone the repository and build the project:

```bash
brew tap cuotos/tap
brew install tfcopen
```

```bash
git clone <repository-url>
cd tfcopen
go install .
```

## Usage
To use the application, navigate to the directory containing your `.tfcopen` file and run:

```bash
./tfcopen
```

`registry` or `-r` will open the private terraform registry pages. This uses the `org: ` field in the found `.tfcopen` file, or the `TFCOPEN_DEFAULT_ORG` environment variable.

```bash
./tfcopen --registry
https://app.terraform.io/app/OrgName/registry/private/modules
```

You can also use the `--print` or `-p` flag to print the constructed URL without opening it:

```bash
./tfcopen --print
https://app.terraform.io/app/OrgName/workspaces?search=workspace-names
or
https://app.terraform.io/app/OrgName/projects/prj-fAQqxxxxxxxxjxzB
etc...
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
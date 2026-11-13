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

`registry` or `-r` will open the private terraform registry pages. This uses the `org:` field in the repo's `.tfcopen`, then `~/.tfcopen`, then the `TFCOPEN_DEFAULT_ORG` environment variable.

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
A repo's `.tfcopen` file should be structured as follows:

```yaml
workspace: <workspace-name>
search: <search-string>
project: <project-name>
org: <organization-name>   # optional if set globally
site: <tfc|terrakube>      # optional if set globally
```

### Global defaults: `~/.tfcopen`

Settings shared by every repo go in `~/.tfcopen`. A repo's `.tfcopen` overrides them.

```yaml
org: <organization-name>
site: <tfc|terrakube>                                # defaults to tfc
terrakube_url: https://terrakube-ui.example.com      # only needed for terrakube
terrakube_api_url: https://terrakube-api.example.com # only needed for terrakube, /api/v1 is added if missing
```

The env vars `TFCOPEN_DEFAULT_ORG`, `TFCOPEN_DEFAULT_SITE`, `TFCOPEN_TERRAKUBE_URL` and `TFCOPEN_TERRAKUBE_API_URL` still work as a last fallback.

## Terrakube

`tfcopen` can also open [Terrakube](https://terrakube.io) workspaces. Terrakube links use IDs rather than names, so `tfcopen` looks the IDs up through the Terrakube API.

The API token comes from `TFCOPEN_TERRAKUBE_TOKEN` if set. Otherwise the token saved by `terraform login <api-host>` is used, and if that is missing or expired `tfcopen` runs `terraform login` for you.

Terrakube's workspace list page can't be opened pre-filtered, so `search:` works differently there: if exactly one workspace matches it is opened; if several match, each is printed with a clickable link and nothing is opened; if none match, the org's workspace list is opened.

## Contributing
Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.
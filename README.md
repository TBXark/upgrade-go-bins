# gbvm

A command line tool to manage Go binaries installed in your `GOPATH/bin` directory. It helps you list installed binaries, check their versions, and upgrade them to the latest available versions.

## Installation

```bash
go install github.com/TBXark/gbvm@latest
```

## Usage

The tool provides three main commands:

### List Command

Lists all Go binaries installed in your `GOPATH/bin` directory.

```bash
gbvm list [flags]

Flags:
  -version   Show version information (default: true)
  -json      Output in JSON format (default: false)
```

### Upgrade Command

Upgrades Go binaries to their latest versions.

```bash
gbvm upgrade [flags]

Flags:
  -all       Upgrade all binaries (default: false)
  -bin       Specify a binary name to upgrade
  -skip-dev  Skip binaries with 'devel' version (default: false)
```

### Install Command

Installs binaries from a backup JSON file.

```bash
gbvm install [flags]

Flags:
  -backup    Path to backup JSON file (required)
```

## Examples

1. List all installed binaries with their versions:
```bash
gbvm list
```

2. List binaries in JSON format:
```bash
gbvm list -json
```

3. Upgrade a specific binary:
```bash
gbvm upgrade -bin=golangci-lint
```

4. Upgrade all binaries except development versions:
```bash
gbvm upgrade -all -skip-dev
```

5. Install binaries from backup:
```bash
gbvm install -backup=binaries.json
```

## License

**gbvm** is released under the MIT license. [See LICENSE](LICENSE) for details.

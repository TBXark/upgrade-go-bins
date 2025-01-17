# gbvm

A command line tool to manage Go binaries installed in your `GOPATH/bin` directory. It helps you list installed binaries, check their versions, and upgrade them to the latest available versions.

## Installation

```bash
go install github.com/TBXark/gbvm@latest
```

## Usage

```bash
Usage: gbvm <command> [options]

A command line tool to manage Go binaries

list commands:
  -help
        show help
  -json
        json mode
  -versions
        show version

upgrade commands:
  -help
        show help
  -skip-dev
        skip dev version

install commands:
  -help
        show help
````

### Install Command

Installs binaries from a backup JSON file.

```bash
# Install binaries from backup
gbvm install backup.json

# List all installed binaries with their versions
gbvm list -versions

# List binaries in JSON format
gbvm list -json

# Upgrade a specific binary
gbvm upgrade bin1 bin2

# Upgrade all binaries except development versions
gbvm upgrade

# Install binaries from backup
gbvm install -backup=binaries.json
```

## License

**gbvm** is released under the MIT license. [See LICENSE](LICENSE) for details.

# gbvm

A command line tool to manage Go binaries installed in your `GOPATH/bin` directory. It helps you list installed binaries, check their versions, and upgrade them to the latest available versions.

## Installation

```bash
go install github.com/TBXark/gbvm@latest
```

## Usage

```bash
Usage: gbvm list [options]

List all installed Go binaries

  -help
        show help
  -json
        json mode
  -versions
        show version
```

```bash
Usage: gbvm install [options] <backup file>

Install Go binaries from backup file

  -help
        show help
```

```bash
Usage: gbvm upgrade [options] [bin1 bin2 ...]

Upgrade Go binaries

  -help
        show help
  -skip-dev
        skip dev version
```


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

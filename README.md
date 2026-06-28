# mise-dotfiles-uninstall

The missing uninstaller companion for mise's dotfiles. (Until native support)

## Why?

According to the [mise dotfiles documentation](https://mise.jdx.dev/dotfiles.html#conflicts):

> Removing an entry from config leaves its file, block, or line in place because mise keeps no state database. Delete unmanaged leftovers by hand.

`mise-dotfiles-uninstall` automates this process safely and statelessly by parsing `mise.toml` and unlinking only the configured symbolic targets.

## Usage

```text
$ go run main.go --help
mise-dotfiles-uninstall - The missing uninstaller companion for mise's dotfiles. (Until native support)

Usage:
  mise-dotfiles-uninstall [options]

Options:
  -c string
    	path to mise.toml configuration file (shorthand) (default "mise.toml")
  -config string
    	path to mise.toml configuration file (default "mise.toml")
  -v	display version information (shorthand)
  -version
    	display version information
```

## Installation

- **Homebrew**:
  ```bash
  brew install veerendra2/tap/mise-dotfiles-uninstall
  ```
- **Direct Download**: Download the pre-built binaries directly from the [GitHub Releases](https://github.com/veerendra2/mise-dotfiles-uninstall/releases) page.

## Development

You can manage the project tasks using `mise`:

```text
$ mise tasks
build                    Build production binary
fmt                      Format Go source code
lint                     Run golangci-lint static analysis
vet                      Run go vet static analysis
```

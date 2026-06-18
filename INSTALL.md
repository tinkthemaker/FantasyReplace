# Installation Guide

## Quick Start

### macOS (Homebrew)
```bash
brew install wizardify
```

### Windows (Scoop)
```powershell
scoop bucket add wizardify https://github.com/yourusername/wizardify-scoop
scoop install wizardify
```

### Linux (apt)
```bash
sudo add-apt-repository ppa:yourusername/wizardify
sudo apt update
sudo apt install wizardify
```

## Download Pre-built Binaries

Visit the [GitHub Releases](https://github.com/yourusername/wizardify/releases) page to download:
- **Windows** (wizardify.exe)
- **macOS** (Intel & Apple Silicon)
- **Linux** (x86_64 & ARM)

### Windows (Manual)
1. Download `wizardify.exe` from [Releases](https://github.com/yourusername/wizardify/releases)
2. Move to a folder in your PATH (or anywhere accessible)
3. Run: `wizardify --help`

### macOS/Linux (Manual)
1. Download the binary for your OS
2. Make it executable: `chmod +x wizardify`
3. Move to `/usr/local/bin`: `sudo mv wizardify /usr/local/bin/`
4. Run: `wizardify --help`

## Build from Source

Requires [Go 1.25+](https://golang.org/doc/install)

```bash
git clone https://github.com/yourusername/wizardify.git
cd wizardify
go build -o wizardify .
./wizardify --help
```

## Docker

```bash
docker pull wizardify:latest
docker run wizardify < your-file.md > output.md
```

Or build locally:
```bash
docker build -t wizardify .
docker run -v /path/to/files:/data wizardify -i 3 /data/input.md -o /data/output.md
```

## Verify Installation

```bash
wizardify --help
```

You should see the help menu with available commands.

## Troubleshooting

### Command not found
- **Windows**: Ensure `wizardify.exe` is in your PATH or run it from the containing directory
- **macOS/Linux**: Run `which wizardify` to check location; move to `/usr/local/bin` if needed

### Lexicon not found
By default, wizardify looks for `lexicon.json` in:
1. Current directory
2. Same folder as the executable
3. Explicitly specify: `wizardify -lexicon /path/to/lexicon.json`

### Permission denied (macOS/Linux)
```bash
chmod +x wizardify
```

## Next Steps

- **CLI usage**: `wizardify post.md -i 3`
- **Interactive TUI**: Just run `wizardify` with no arguments
- **Read more**: See [README.md](README.md) for full documentation

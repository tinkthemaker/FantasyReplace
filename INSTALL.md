# Installation Guide

## Build from source

Wizardify currently supports source builds and downloadable CI artifacts. Package-manager formulae are not published yet.

Requires Go 1.25 or newer:

```bash
git clone <repository-url>
cd FantasyReplace
go build -o wizardify .
```

On Windows, use `go build -o wizardify.exe .`.

The resulting executable is self-contained: the default lexicon is embedded. Put the executable somewhere on your `PATH`, or run it from the build directory.

## CI artifacts

Successful GitHub Actions builds produce platform-specific artifacts:

- `wizardify-windows-amd64.exe`
- `wizardify-windows-arm64.exe`
- `wizardify-macos-amd64`
- `wizardify-macos-arm64`
- `wizardify-linux-amd64`
- `wizardify-linux-arm64`

On macOS and Linux, make the downloaded file executable:

```bash
chmod +x wizardify-*
```

## Verify

```bash
wizardify -help
wizardify -version
wizardify sample-post.md -stdout
```

## Custom lexicons

Wizardify checks for `lexicon.json` in the current directory and next to the executable before using its embedded default. Select a different file explicitly with:

```bash
wizardify post.md -lexicon /path/to/lexicon.json
```

## Troubleshooting

- **Command not found:** invoke the executable by path or add its directory to `PATH`.
- **Permission denied on macOS/Linux:** run `chmod +x` on the downloaded binary.
- **Lexicon error:** validate the JSON or omit `-lexicon` to use the embedded default.

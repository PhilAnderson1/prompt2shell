# Installing prompt2shell

## Linux release package

Download `prompt2shell-v1.0.0-linux.tar.gz` from the GitHub release, then run:

```sh
tar -xzf prompt2shell-v1.0.0-linux.tar.gz
cd prompt2shell-v1.0.0-linux
sudo ./install-linux.sh
```

The installer detects AMD64, ARM64, or ARMv7 and installs the corresponding
binary as `/usr/local/bin/p2s`. It installs the supplied configuration as
`/etc/prompt2shell.conf` only if that file does not already exist.

Edit the configuration and replace its placeholder endpoint details before
running `p2s`. A system-wide configuration installed with mode `0644` can be
read by every local user. For a private per-user configuration, instead place
the file at `~/.config/prompt2shell.conf` and set its mode to `0600`.

To install a particular Linux binary manually:

```sh
sudo install -m 0755 p2s-linux-amd64 /usr/local/bin/p2s
sudo install -m 0644 prompt2shell.conf /etc/prompt2shell.conf
```

Replace `p2s-linux-amd64` with `p2s-linux-arm64` or `p2s-linux-armv7` when
appropriate.

## Windows release package

Download and extract `prompt2shell-v1.0.0-windows.zip`. Open PowerShell as an
administrator, change to the extracted folder, and run:

```powershell
powershell -ExecutionPolicy Bypass -File .\install-windows.ps1
```

The installer detects AMD64 or ARM64, installs the corresponding binary as
`p2s.exe` under `Program Files`, adds its directory to the system `PATH`, and
installs the supplied configuration under `ProgramData` only if no system-wide
configuration already exists.

Open a new PowerShell window after installation. Edit the configuration and
replace its placeholder endpoint details before running `p2s`.

## Build from source

Build Linux binaries:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o p2s-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o p2s-linux-arm64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o p2s-linux-armv7 .
```

Build Windows binaries from Linux:

```sh
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o p2s-windows-amd64.exe .
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o p2s-windows-arm64.exe .
```

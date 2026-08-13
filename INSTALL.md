# Installing prompt2shell

See the [tested API configurations](https://github.com/PhilAnderson1/prompt2shell#tested-api-configurations)
before choosing an endpoint and `api_type`.

For portable use without installation, copy the appropriate `p2s` executable
and `prompt2shell.conf` to the same directory on a USB drive. Run the executable
directly from the drive. A configuration in the same directory as the
executable takes precedence over per-user and system-wide configuration files
on both Linux and Windows.

## Linux release package

Download the Linux `.tar.gz` package from the GitHub release, then run the
following commands, replacing `<version>` with the downloaded version number:

```sh
tar -xzf prompt2shell-<version>-linux.tar.gz
cd prompt2shell-<version>-linux
sudo ./install-linux.sh
```

The installer detects AMD64, ARM64, or ARMv7 and installs the corresponding
binary as `/usr/local/bin/p2s`. It installs the supplied configuration as
`/etc/prompt2shell.conf` only if that file does not already exist.

Before running `p2s`, edit `/etc/prompt2shell.conf` and replace
`[your api token goes here]` with your OpenRouter API key. Alternatively, change
the endpoint, model, and API type to use another supported AI endpoint. A
system-wide configuration installed with mode `0644` can be read by every local
user. For a private per-user configuration, instead place the file at
`~/.config/prompt2shell.conf` and set its mode to `0600`.

To install a particular Linux binary manually:

```sh
sudo install -m 0755 p2s-linux-amd64 /usr/local/bin/p2s
sudo install -m 0644 prompt2shell.conf /etc/prompt2shell.conf
```

Replace `p2s-linux-amd64` with `p2s-linux-arm64` or `p2s-linux-armv7` when
appropriate.

## Windows release package

Download and extract the Windows `.zip` package from the GitHub release. Open
PowerShell as an administrator, change to the extracted folder, and run:

```powershell
powershell -ExecutionPolicy Bypass -File .\install-windows.ps1
```

The installer detects AMD64 or ARM64, installs the corresponding binary as
`p2s.exe` under `Program Files`, adds its directory to the system `PATH`, and
installs the supplied configuration under `ProgramData` only if no system-wide
configuration already exists.

Open a new PowerShell window after installation. Before running `p2s`, edit:

```text
C:\ProgramData\prompt2shell\prompt2shell.conf
```

Replace `[your api token goes here]` with your OpenRouter API key. Alternatively,
change the endpoint, model, and API type to use another supported AI endpoint.

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

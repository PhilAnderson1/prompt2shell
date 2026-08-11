# prompt2shell

`s` turns a plain-English request into a Linux shell command, displays it, and
runs it only after you press Enter. Press Esc to abort.

The program is written in Go so it builds as a single native executable with no
runtime, package manager, or virtual environment required on the target machine.

## Build and install

```sh
CGO_ENABLED=0 go build -o s .
install -m 0755 s "$HOME/.local/bin/s"
```

Ensure `$HOME/.local/bin` is on your `PATH`. To build for another Linux machine:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o s-linux-amd64 .
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o s-linux-arm64 .
```

## Use

```text
$ s find files bigger than 100mb in the current directory
find . -type f -size +100M
Press Enter to run, or Esc to abort:
```

The AI endpoint and model are compiled into the binary. The endpoint must be
reachable from the machine running `s`. Always inspect generated commands before
running them, especially commands that modify or delete files.

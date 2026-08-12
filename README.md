# prompt2shell

`s` turns a plain-English request into a Linux shell command, displays it, and
runs it only after you press Enter. Press Esc to abort.

The program is written in Go so it builds as a single native executable with no
runtime, package manager, or virtual environment required on the target machine.

## Configure the AI endpoint

Create `~/.config/prompt2shell.conf` for a per-user configuration, or
`/etc/prompt2shell.conf` for a system-wide configuration. The per-user file takes
precedence when both exist.

```ini
# OpenRouter example
endpoint=https://openrouter.ai/api/v1/chat/completions
model=qwen/qwen3.6-35b-a3b
api_token=your-openrouter-api-key
api_type=openrouter
```

`endpoint` and `model` are required. `api_token` may be left empty when the
endpoint does not require authentication. Set `api_type` to `openrouter`,
`llamacpp`, or `generic`; it defaults to `generic`. This lets `s` disable model
reasoning using only the parameter supported by that API. The endpoint must be reachable
from every machine where `s` will run. Protect a per-user file with:

```sh
chmod 600 ~/.config/prompt2shell.conf
```

Do not commit a real access token to a public repository.

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

Always inspect generated commands before running them, especially commands that
modify or delete files.

## Inspiration

This project was inspired by
[whatisit-nl2sh](https://github.com/ThorOdinson246/whatisit-nl2sh), created by
[ThorOdinson246](https://github.com/ThorOdinson246), which provides similar
functionality locally without requiring an external AI endpoint.

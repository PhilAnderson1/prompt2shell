# prompt2shell

prompt2shell turns a plain-English request into a Linux shell or Windows
PowerShell command, displays it, and runs it only after you press Enter. Press
Esc to abort.

The program is written in Go so it builds as a single native executable with no
runtime, package manager, or virtual environment required on the target machine.

## Configure the AI endpoint

Configuration paths, in lookup order:

- Linux: `~/.config/prompt2shell.conf`, then `/etc/prompt2shell.conf`
- Windows: `%APPDATA%\prompt2shell\prompt2shell.conf`, then
  `%ProgramData%\prompt2shell\prompt2shell.conf`

The per-user file takes precedence when both exist.

```ini
# OpenRouter example
endpoint=https://openrouter.ai/api/v1/chat/completions
model=qwen/qwen3.6-35b-a3b
api_token=your-openrouter-api-key
api_type=openrouter
```

`endpoint` and `model` are required. `api_token` may be left empty when the
endpoint does not require authentication. Set `api_type` to `openrouter`,
`llamacpp`, `openai`, or `generic`; it defaults to `generic`. This lets `p2s` disable model
reasoning using only the parameter supported by that API. For native OpenAI, use
`api_type=openai` with a Responses API endpoint such as
`https://api.openai.com/v1/responses`.

For native OpenAI endpoints, `reasoning_effort` may optionally be set to a value supported by the selected model. For example:

```ini
endpoint=https://api.openai.com/v1/responses
model=gpt-5-nano
api_token=YOUR_OPENAI_API_KEY
api_type=openai
reasoning_effort=minimal
```

Allowed configuration values are `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, and `max`, but each model supports only a subset. When omitted, the API uses the model default. The endpoint must be
reachable from every machine where `p2s` will run. Protect a per-user file with:

```sh
chmod 600 ~/.config/prompt2shell.conf
```

Do not commit a real access token to a public repository.

After installation, edit `prompt2shell.conf` and add your API key. You can also
replace the supplied OpenRouter endpoint and model with another supported AI
endpoint. The system-wide configuration is installed at:

- Linux: `/etc/prompt2shell.conf`
- Windows: `%ProgramData%\prompt2shell\prompt2shell.conf` (normally
  `C:\ProgramData\prompt2shell\prompt2shell.conf`)

## Build and install

Download the archive for your operating system from the GitHub Releases page.
Each archive contains binaries for all supported architectures and an installer
that selects the correct one automatically. See [INSTALL.md](INSTALL.md) for
complete installation and build instructions.

For a quick local Linux build and per-user installation:

```sh
CGO_ENABLED=0 go build -o p2s .
install -m 0755 p2s "$HOME/.local/bin/p2s"
```

## Use

```text
$ p2s find files bigger than 100mb in the current directory
find . -type f -size +100M
Press Enter to run, or Esc to abort:
```

On Windows, use `p2s` in PowerShell; generated commands
use PowerShell syntax.

Always inspect generated commands before running them, especially commands that
modify or delete files.

## Inspiration

This project was inspired by
[whatisit-nl2sh](https://github.com/ThorOdinson246/whatisit-nl2sh), created by
[ThorOdinson246](https://github.com/ThorOdinson246), which provides similar
functionality locally without requiring an external AI endpoint.

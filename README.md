# prompt2shell

prompt2shell uses AI to turn plain-English instructions into ready-to-run Linux
shell or Windows PowerShell commands. It helps you use command-line tools
without having to remember exact syntax, while always showing the generated
command for approval before it runs. Press Enter to execute it or Esc to abort.

## Why prompt2shell?

- **Minimal:** a standalone compiled binary with no runtime dependencies to
  worry about.
- **Flexible installation:** install it per user or system-wide, or run it
  directly from a USB drive.
- **Provider-independent:** works with OpenRouter, OpenAI, llama.cpp, and other
  compatible API endpoints.
- **Terminal-independent:** works from your existing terminal—there is no
  replacement shell or terminal application to install.
- **Cross-platform:** supports Linux and Windows, with binaries for x86-64 and
  Arm processors.

## Use

```text
$ p2s find files bigger than 100mb in the current directory
find . -type f -size +100M
Press Enter to run, or Esc to abort:
```

On Windows, use `p2s` in PowerShell; generated commands use PowerShell syntax.

Always inspect generated commands before running them, especially commands that
modify or delete files.

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

## Configure the AI endpoint

Configuration paths, in lookup order:

- Linux: `prompt2shell.conf` in the same directory as the executable, then
  `~/.config/prompt2shell.conf`, then `/etc/prompt2shell.conf`
- Windows: `prompt2shell.conf` in the same directory as the executable, then
  `%APPDATA%\prompt2shell\prompt2shell.conf`, then
  `%ProgramData%\prompt2shell\prompt2shell.conf`

This allows `p2s` and its configuration to run directly from portable media
such as a USB drive. The first configuration found takes precedence.

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

Allowed OpenAI configuration values are `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, and `max`, but each model supports only a subset. When omitted, the API uses the model default. The endpoint must be
reachable from every machine where `p2s` will run.


After installation, edit `prompt2shell.conf` and add your API key. You can also
replace the supplied OpenRouter endpoint and model with another supported AI
endpoint. The system-wide configuration is installed at:

- Linux: `/etc/prompt2shell.conf`
- Windows: `%ProgramData%\prompt2shell\prompt2shell.conf` (normally
  `C:\ProgramData\prompt2shell\prompt2shell.conf`)

### Tested API configurations

These are the API combinations tested with prompt2shell; this is not an
exhaustive compatibility list.

| `api_type` | Service | Tested endpoint type |
| --- | --- | --- |
| `openrouter` | OpenRouter | Chat Completions |
| `llamacpp` | llama.cpp | Chat Completions |
| `openai` | OpenAI | Responses |
| `generic` | Other OpenAI-compatible services | Chat Completions; compatibility varies |

```ini
# OpenRouter
endpoint=https://openrouter.ai/api/v1/chat/completions
api_type=openrouter

# llama.cpp
endpoint=https://example.com/v1/chat/completions
api_type=llamacpp

# OpenAI
endpoint=https://api.openai.com/v1/responses
api_type=openai
```

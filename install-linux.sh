#!/bin/sh

set -eu

if [ "$(id -u)" -ne 0 ]; then
    echo "Run this installer as root, for example: sudo ./install-linux.sh" >&2
    exit 1
fi

script_directory=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
architecture=$(uname -m)

case "$architecture" in
    x86_64|amd64)
        binary_name=p2s-linux-amd64
        ;;
    aarch64|arm64)
        binary_name=p2s-linux-arm64
        ;;
    armv7l|armv7)
        binary_name=p2s-linux-armv7
        ;;
    *)
        echo "Unsupported Linux architecture: $architecture" >&2
        echo "Supported architectures: AMD64, ARM64, and ARMv7." >&2
        exit 1
        ;;
esac

binary_source=$script_directory/$binary_name
config_source=$script_directory/prompt2shell.conf
binary_destination=/usr/local/bin/p2s
config_destination=/etc/prompt2shell.conf

if [ ! -f "$binary_source" ]; then
    echo "Missing $binary_source. Place all Linux binaries beside this installer." >&2
    exit 1
fi

install -m 0755 "$binary_source" "$binary_destination"
echo "Installed executable: $binary_destination"
echo "Detected architecture: $architecture"

if [ -e "$config_destination" ]; then
    echo "Kept existing configuration: $config_destination"
else
    if [ ! -f "$config_source" ]; then
        echo "Missing $config_source. The executable was installed, but no configuration was created." >&2
        exit 1
    fi
    install -m 0644 "$config_source" "$config_destination"
    echo "Installed configuration: $config_destination"
    echo "Edit it and replace the placeholder API key before running p2s."
fi

echo "Installation complete. Run: p2s print the current directory"

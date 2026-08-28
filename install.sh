#!/bin/sh
set -eu

REPO="nilay-banerjee/noob-cli"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
	x86_64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*)
		echo "unsupported arch: $arch" >&2
		exit 1
		;;
esac
case "$os" in
	darwin | linux) ;;
	*)
		echo "unsupported OS: $os" >&2
		exit 1
		;;
esac

url="https://github.com/$REPO/releases/latest/download/noob-cli_${os}_${arch}.tar.gz"
bindir="${BINDIR:-$HOME/.local/bin}"

echo "Downloading $url"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
curl -fsSL "$url" | tar -xz -C "$tmp"

mkdir -p "$bindir"
install -m 0755 "$tmp/noob-cli" "$bindir/noob-cli"
echo "Installed $("$bindir/noob-cli" --version) to $bindir/noob-cli"

case ":$PATH:" in
	*":$bindir:"*) ;;
	*) echo "Note: $bindir is not on your PATH yet. Add: export PATH=\"$bindir:\$PATH\"" ;;
esac

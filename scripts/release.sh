#!/bin/sh
set -eu

version="${1:?usage: scripts/release.sh vX.Y.Z}"
cd "$(dirname "$0")/.."

rm -rf dist
mkdir -p dist

for target in darwin/arm64 darwin/amd64 linux/amd64 linux/arm64; do
	os=${target%/*}
	arch=${target#*/}
	out="dist/noob-cli_${os}_${arch}"
	GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -trimpath \
		-ldflags "-s -w -X github.com/nilay-banerjee/noob-cli/cmd.version=$version" \
		-o "$out/noob-cli" .
	tar -czf "$out.tar.gz" -C "$out" noob-cli
	rm -r "$out"
done

(cd dist && shasum -a 256 ./*.tar.gz >checksums.txt)

git tag "$version"
git push origin main "$version"
gh release create "$version" dist/*.tar.gz dist/checksums.txt --title "$version" --generate-notes

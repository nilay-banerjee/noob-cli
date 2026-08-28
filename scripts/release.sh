#!/bin/sh
set -eu

version="${1:?usage: scripts/release.sh vX.Y.Z}"
cd "$(dirname "$0")/.."

go vet ./...
go test ./...

git tag "$version"
git push origin main "$version"
echo "Tagged $version — CI builds and publishes the release:"
echo "https://github.com/nilay-banerjee/noob-cli/actions"

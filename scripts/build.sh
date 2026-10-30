#!/bin/sh
# Provides bin/kanban on `herdr plugin install`, run from the checkout root. When the checkout is
# exactly the release of the manifest's version, it downloads and verifies that release's binary;
# at the first problem it builds from source instead.
set -eu

tmp=
trap 'rm -f ${tmp:+"$tmp"}' EXIT

fallback() {
	echo "build.sh: $1; building from source" >&2
	go build -o bin/kanban ./cmd/kanban
	exit
}

get() {
	curl -fsSL --proto '=https' --max-time 60 "$@"
}

version=$(sed -n 's/^version = "\(.*\)"$/\1/p' herdr-plugin.toml)
[ -n "$version" ] || fallback "no version in herdr-plugin.toml"
case $(uname -s) in
Darwin) os=darwin ;;
Linux) os=linux ;;
*) fallback "no release binary for OS $(uname -s)" ;;
esac
case $(uname -m) in
x86_64) arch=amd64 ;;
arm64 | aarch64) arch=arm64 ;;
*) fallback "no release binary for architecture $(uname -m)" ;;
esac
if command -v sha256sum >/dev/null; then
	sha256() { sha256sum "$1"; }
elif command -v shasum >/dev/null; then
	sha256() { shasum -a 256 "$1"; }
else
	fallback "neither sha256sum nor shasum found"
fi
command -v curl >/dev/null || fallback "curl not found"

url=https://github.com/zeotyn/openspec-kanban/releases/download/v$version
release=$(get "$url/commit.txt") || fallback "no release v$version"
head=$(git rev-parse HEAD) || fallback "cannot read the checkout's commit"
[ "$release" = "$head" ] || fallback "checkout $head is not release v$version ($release)"
sums=$(get "$url/checksums.txt") || fallback "cannot download checksums.txt of v$version"

bin=kanban-$os-$arch
mkdir -p bin
tmp=$(mktemp bin/.kanban.XXXXXX)
get -o "$tmp" "$url/$bin" || fallback "cannot download $bin of v$version"
want=$(printf '%s\n' "$sums" | awk -v f="$bin" '$2 == f { print $1 }')
got=$(sha256 "$tmp" | awk '{ print $1 }')
[ -n "$want" ] && [ "$want" = "$got" ] || fallback "checksum mismatch for $bin of v$version"
chmod 755 "$tmp"
[ "$("$tmp" --version 2>/dev/null)" = "$version" ] || fallback "$bin of v$version does not print version $version"
mv -f "$tmp" bin/kanban
echo "downloaded kanban $version"

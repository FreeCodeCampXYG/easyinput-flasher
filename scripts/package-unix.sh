#!/usr/bin/env bash
set -euo pipefail

version="$1"
platform="$2"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist="$root/dist"

test -f "$root/LICENSE"
mkdir -p "$dist"

if [[ "$platform" == darwin/* ]]; then
  app="$root/build/bin/easyinput-flasher.app"
  test -d "$app"
  cp "$root/README.md" "$root/SECURITY.md" "$root/THIRD_PARTY_NOTICES.md" "$root/LICENSE" "$app/Contents/Resources/"
  archive="$dist/easyinput-flasher-v${version}-${platform//\//-}.zip"
  rm -f "$archive"
  ditto -c -k --sequesterRsrc --keepParent "$app" "$archive"
else
  executable="$root/build/bin/easyinput-flasher"
  test -f "$executable"
  staging="$dist/easyinput-flasher-v${version}-${platform//\//-}"
  rm -rf "$staging"
  mkdir -p "$staging"
  cp "$executable" "$staging/easyinput-flasher"
  chmod +x "$staging/easyinput-flasher"
  cp "$root/README.md" "$root/SECURITY.md" "$root/THIRD_PARTY_NOTICES.md" "$root/LICENSE" "$staging/"
  archive="$dist/easyinput-flasher-v${version}-${platform//\//-}.tar.gz"
  tar -C "$dist" -czf "$archive" "$(basename "$staging")"
fi

(cd "$dist" && shasum -a 256 "$(basename "$archive")" > "$(basename "$archive").sha256")

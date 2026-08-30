#!/usr/bin/env bash
set -euo pipefail

version="$1"
platform="$2"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist="$root/dist"
helper="$root/tools/esptool/dist/esptool"

test -f "$helper"
test -f "$root/LICENSE"
mkdir -p "$dist"

if [[ "$platform" == darwin/* ]]; then
  app="$root/build/bin/easyinput-flasher.app"
  test -d "$app"
  target="$app/Contents/MacOS/tools/esptool"
  mkdir -p "$target"
  cp "$helper" "$target/esptool"
  chmod +x "$target/esptool"
  cp "$root/README.md" "$root/SECURITY.md" "$root/THIRD_PARTY_NOTICES.md" "$root/LICENSE" "$app/Contents/Resources/"
  mkdir -p "$app/Contents/Resources/LICENSES"
  cp "$root/LICENSES/GPL-2.0-or-later.txt" "$app/Contents/Resources/LICENSES/GPL-2.0-or-later.txt"
  archive="$dist/easyinput-flasher-v${version}-${platform//\//-}.zip"
  rm -f "$archive"
  ditto -c -k --sequesterRsrc --keepParent "$app" "$archive"
else
  executable="$root/build/bin/easyinput-flasher"
  test -f "$executable"
  staging="$dist/easyinput-flasher-v${version}-${platform//\//-}"
  rm -rf "$staging"
  mkdir -p "$staging/tools/esptool"
  mkdir -p "$staging/LICENSES"
  cp "$executable" "$staging/easyinput-flasher"
  cp "$helper" "$staging/tools/esptool/esptool"
  chmod +x "$staging/easyinput-flasher" "$staging/tools/esptool/esptool"
  cp "$root/README.md" "$root/SECURITY.md" "$root/THIRD_PARTY_NOTICES.md" "$root/LICENSE" "$staging/"
  cp "$root/LICENSES/GPL-2.0-or-later.txt" "$staging/LICENSES/GPL-2.0-or-later.txt"
  archive="$dist/easyinput-flasher-v${version}-${platform//\//-}.tar.gz"
  tar -C "$dist" -czf "$archive" "$(basename "$staging")"
fi

(cd "$dist" && shasum -a 256 "$(basename "$archive")" > "$(basename "$archive").sha256")

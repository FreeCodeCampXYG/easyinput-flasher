param(
    [Parameter(Mandatory = $true)]
    [string]$Version
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$source = Join-Path $root 'build\bin\easyinput-flasher.exe'
$helper = Join-Path $root 'tools\esptool\dist\esptool.exe'
$dist = Join-Path $root 'dist'
$staging = Join-Path $dist "easyinput-flasher-v$Version-windows-x64"

if (-not (Test-Path -LiteralPath $source)) { throw "缺少 Wails 可执行文件: $source" }
if (-not (Test-Path -LiteralPath $helper)) { throw "缺少受控 esptool helper: $helper" }

Remove-Item -LiteralPath $staging -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path (Join-Path $staging 'tools\esptool') -Force | Out-Null
New-Item -ItemType Directory -Path (Join-Path $staging 'LICENSES') -Force | Out-Null
Copy-Item -LiteralPath $source -Destination (Join-Path $staging 'easyinput-flasher.exe')
Copy-Item -LiteralPath $helper -Destination (Join-Path $staging 'tools\esptool\esptool.exe')
Copy-Item -LiteralPath (Join-Path $root 'README.md') -Destination $staging
Copy-Item -LiteralPath (Join-Path $root 'SECURITY.md') -Destination $staging
Copy-Item -LiteralPath (Join-Path $root 'THIRD_PARTY_NOTICES.md') -Destination $staging
Copy-Item -LiteralPath (Join-Path $root 'LICENSE') -Destination $staging
Copy-Item -LiteralPath (Join-Path $root 'LICENSES\GPL-2.0-or-later.txt') -Destination (Join-Path $staging 'LICENSES\GPL-2.0-or-later.txt') -Force

$archive = Join-Path $dist "easyinput-flasher-v$Version-windows-x64-portable.zip"
Remove-Item -LiteralPath $archive -Force -ErrorAction SilentlyContinue
Compress-Archive -LiteralPath $staging -DestinationPath $archive -CompressionLevel Optimal
$hash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
Set-Content -LiteralPath "$archive.sha256" -Value "$hash  $(Split-Path -Leaf $archive)" -Encoding ascii

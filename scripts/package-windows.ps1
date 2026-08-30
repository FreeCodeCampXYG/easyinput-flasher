param(
    [Parameter(Mandatory = $true)]
    [string]$Version,
    [ValidateSet('x64','arm64')]
    [string]$Architecture = 'x64'
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$source = Join-Path $root 'build\bin\easyinput-flasher.exe'
$dist = Join-Path $root 'dist'
$staging = Join-Path $dist "easyinput-flasher-v$Version-windows-$Architecture"

if (-not (Test-Path -LiteralPath $source)) { throw "缺少 Wails 可执行文件: $source" }

Remove-Item -LiteralPath $staging -Recurse -Force -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $staging -Force | Out-Null
Copy-Item -LiteralPath $source -Destination (Join-Path $staging 'easyinput-flasher.exe')
Copy-Item -LiteralPath (Join-Path $root 'README.md') -Destination $staging
Copy-Item -LiteralPath (Join-Path $root 'SECURITY.md') -Destination $staging
Copy-Item -LiteralPath (Join-Path $root 'THIRD_PARTY_NOTICES.md') -Destination $staging
Copy-Item -LiteralPath (Join-Path $root 'LICENSE') -Destination $staging

$archive = Join-Path $dist "easyinput-flasher-v$Version-windows-$Architecture-portable.zip"
Remove-Item -LiteralPath $archive -Force -ErrorAction SilentlyContinue
Compress-Archive -LiteralPath $staging -DestinationPath $archive -CompressionLevel Optimal
$hash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
Set-Content -LiteralPath "$archive.sha256" -Value "$hash  $(Split-Path -Leaf $archive)" -Encoding ascii

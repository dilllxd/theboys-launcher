param(
    [Parameter(Mandatory=$true)] [string]$NewVersion,
    [switch]$Tag
)

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root\.. 

if ($NewVersion.StartsWith('v')) { $NewVersion = $NewVersion.Substring(1) }

if (-not ($NewVersion -match '^[0-9]+\.[0-9]+\.[0-9]+$')) {
    Write-Error "Version must be in MAJOR.MINOR.PATCH format"
    exit 1
}

$parts = $NewVersion.Split('.')
$major=$parts[0]; $minor=$parts[1]; $patch=$parts[2]

Write-Host "Updating version.env to $NewVersion"
@"
# TheBoysLauncher Version Configuration
VERSION=$NewVersion
MAJOR=$major
MINOR=$minor
PATCH=$patch
BUILD_METADATA=
PRERELEASE=

# Full version string is constructed by scripts/get-version.sh
"@ | Set-Content version.env -Encoding UTF8

if ($Tag) {
    git add version.env
    git commit -m "chore: bump version to v$NewVersion"
    git tag "v$NewVersion"
    Write-Host "Committed and tagged v$NewVersion"
}

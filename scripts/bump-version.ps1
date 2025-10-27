param(
    [Parameter(Mandatory=$false)] [string]$NewVersion,
    [ValidateSet("dev", "stable")]
    [string]$Mode = "dev",
    [switch]$Tag
)

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root\..

# Read current version.env if no new version is specified
if (-not $NewVersion) {
    if (Test-Path "version.env") {
        $content = Get-Content "version.env"
        $currentVersion = ($content | Where-Object { $_ -match '^VERSION=' }).Split('=')[1]
        $currentPrerelease = ($content | Where-Object { $_ -match '^PRERELEASE=' }).Split('=')[1]
        
        $parts = $currentVersion.Split('.')
        $major=$parts[0]; $minor=$parts[1]; $patch=$parts[2]
        
        if ($Mode -eq "stable") {
            # For stable releases, increment patch version
            $patch = [int]$patch + 1
            $NewVersion = "$major.$minor.$patch"
            $prerelease = ""
        } else {
            # For dev releases, keep current version
            $NewVersion = $currentVersion
            $prerelease = $currentPrerelease
        }
    } else {
        Write-Error "version.env file not found and no version specified"
        exit 1
    }
} else {
    # If new version is specified, use it
    if ($NewVersion.StartsWith('v')) { $NewVersion = $NewVersion.Substring(1) }
    
    if (-not ($NewVersion -match '^[0-9]+\.[0-9]+\.[0-9]+$')) {
        Write-Error "Version must be in MAJOR.MINOR.PATCH format"
        exit 1
    }
    
    $parts = $NewVersion.Split('.')
    $major=$parts[0]; $minor=$parts[1]; $patch=$parts[2]
    
    # Set prerelease based on mode
    if ($Mode -eq "stable") {
        $prerelease = ""
    } else {
        $prerelease = "dev"
    }
}

Write-Host "Updating version.env to $NewVersion (mode: $Mode)"
@"
# TheBoysLauncher Version Configuration
VERSION=$NewVersion
MAJOR=$major
MINOR=$minor
PATCH=$patch
BUILD_METADATA=
PRERELEASE=$prerelease

# Full version string is constructed by scripts/get-version.sh
"@ | Set-Content version.env -Encoding UTF8

if ($Tag) {
    git add version.env
    $commitMessage = "chore: bump version to v$NewVersion"
    if ($Mode -eq "stable") {
        $commitMessage += " (stable release)"
    } else {
        $commitMessage += " (dev release)"
    }
    git commit -m $commitMessage
    
    # Create appropriate tag based on mode
    $tagVersion = "v$NewVersion"
    if ($Mode -eq "dev" -and $prerelease) {
        $tagVersion = "v$NewVersion-$prerelease"
    }
    
    git tag $tagVersion
    Write-Host "Committed and tagged $tagVersion"
}

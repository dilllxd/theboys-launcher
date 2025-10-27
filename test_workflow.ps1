# Test script to validate GitHub Actions workflow syntax

Write-Host "Testing GitHub Actions workflow syntax..." -ForegroundColor Blue

# Check if the workflow file exists
$workflowFile = ".github/workflows/dev-prerelease.yml"
if (-not (Test-Path $workflowFile)) {
    Write-Host "✗ Workflow file not found: $workflowFile" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Workflow file found: $workflowFile" -ForegroundColor Green

# Read and validate YAML structure
try {
    $content = Get-Content $workflowFile -Raw
    
    # Check for required top-level keys
    $requiredKeys = @("name", "on", "permissions", "jobs")
    foreach ($key in $requiredKeys) {
        $pattern = "^$key`:"
        if ($content -match $pattern) {
            Write-Host "✓ Found required key: $key" -ForegroundColor Green
        } else {
            Write-Host "✗ Missing required key: $key" -ForegroundColor Red
            exit 1
        }
    }
    
    # Check for required jobs
    $requiredJobs = @("bump_version", "build", "publish")
    foreach ($job in $requiredJobs) {
        $pattern = "  $job`:"
        if ($content -match $pattern) {
            Write-Host "✓ Found required job: $job" -ForegroundColor Green
        } else {
            Write-Host "✗ Missing required job: $job" -ForegroundColor Red
            exit 1
        }
    }
    
    # Check for platform matrix
    if ($content -match "matrix:.*" -and $content -match "os:.*") {
        Write-Host "✓ Found platform matrix configuration" -ForegroundColor Green
    } else {
        Write-Host "✗ Missing platform matrix configuration" -ForegroundColor Red
        exit 1
    }
    
    # Check for required steps in build job
    if ($content -match "Set up Go") {
        Write-Host "✓ Found Go setup step" -ForegroundColor Green
    } else {
        Write-Host "✗ Missing Go setup step" -ForegroundColor Red
        exit 1
    }
    
    if ($content -match "Build binary") {
        Write-Host "✓ Found build step" -ForegroundColor Green
    } else {
        Write-Host "✗ Missing build step" -ForegroundColor Red
        exit 1
    }
    
    # Check for proper artifact naming
    if ($content -match "artifact_name:.*") {
        Write-Host "✓ Found artifact naming configuration" -ForegroundColor Green
    } else {
        Write-Host "✗ Missing artifact naming configuration" -ForegroundColor Red
        exit 1
    }
    
    # Check for proper version handling
    if ($content -match "version\.env") {
        Write-Host "✓ Found version.env handling" -ForegroundColor Green
    } else {
        Write-Host "✗ Missing version.env handling" -ForegroundColor Red
        exit 1
    }
    
    # Check for proper release creation
    if ($content -match "create-release" -or $content -match "create_release") {
        Write-Host "✓ Found release creation step" -ForegroundColor Green
    } else {
        Write-Host "✗ Missing release creation step" -ForegroundColor Red
        exit 1
    }
    
} catch {
    Write-Host "✗ Error reading workflow file: $_" -ForegroundColor Red
    exit 1
}

Write-Host "✅ GitHub Actions workflow validation passed!" -ForegroundColor Green
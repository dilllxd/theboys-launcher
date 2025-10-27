#!/bin/bash

# Test script to validate GitHub Actions workflow syntax

set -e

echo "Testing GitHub Actions workflow syntax..."

# Check if yamllint is available, if not install it
if ! command -v yamllint &> /dev/null; then
    echo "yamllint not found, installing..."
    if command -v pip &> /dev/null; then
        pip install yamllint
    elif command -v pip3 &> /dev/null; then
        pip3 install yamllint
    else
        echo "Neither pip nor pip3 found. Skipping YAML syntax validation."
        echo "Please install yamllint manually to validate workflow syntax."
        exit 1
    fi
fi

# Validate workflow YAML syntax
echo "Validating .github/workflows/dev-prerelease.yml..."
yamllint .github/workflows/dev-prerelease.yml

# Check for required workflow sections
echo "Checking required workflow sections..."

WORKFLOW_FILE=".github/workflows/dev-prerelease.yml"

# Check for required top-level keys
required_keys=("name" "on" "permissions" "jobs")
for key in "${required_keys[@]}"; do
    if grep -q "^$key:" "$WORKFLOW_FILE"; then
        echo "✓ Found required key: $key"
    else
        echo "✗ Missing required key: $key"
        exit 1
    fi
done

# Check for required jobs
required_jobs=("bump_version" "build" "publish")
for job in "${required_jobs[@]}"; do
    if grep -q "  $job:" "$WORKFLOW_FILE"; then
        echo "✓ Found required job: $job"
    else
        echo "✗ Missing required job: $job"
        exit 1
    fi
done

# Check for platform matrix
if grep -q "matrix:" "$WORKFLOW_FILE" && grep -q "os:" "$WORKFLOW_FILE"; then
    echo "✓ Found platform matrix configuration"
else
    echo "✗ Missing platform matrix configuration"
    exit 1
fi

# Check for required steps in build job
if grep -q "Set up Go" "$WORKFLOW_FILE"; then
    echo "✓ Found Go setup step"
else
    echo "✗ Missing Go setup step"
    exit 1
fi

if grep -q "Build binary" "$WORKFLOW_FILE"; then
    echo "✓ Found build step"
else
    echo "✗ Missing build step"
    exit 1
fi

# Check for proper artifact naming
if grep -q "artifact_name:" "$WORKFLOW_FILE"; then
    echo "✓ Found artifact naming configuration"
else
    echo "✗ Missing artifact naming configuration"
    exit 1
fi

# Check for proper version handling
if grep -q "version.env" "$WORKFLOW_FILE"; then
    echo "✓ Found version.env handling"
else
    echo "✗ Missing version.env handling"
    exit 1
fi

# Check for proper release creation
if grep -q "create-release" "$WORKFLOW_FILE" || grep -q "create_release" "$WORKFLOW_FILE"; then
    echo "✓ Found release creation step"
else
    echo "✗ Missing release creation step"
    exit 1
fi

echo "✅ GitHub Actions workflow validation passed!"
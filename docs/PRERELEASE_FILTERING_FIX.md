# Prerelease Version Filtering Fix

## Overview

This document describes the fix for the bug where dev versions were incorrectly identified as stable versions in TheBoys Launcher when dev builds were disabled in settings.

## Bug Description

**Original Issue**: When users had dev builds disabled in settings, the launcher would still sometimes return dev/prerelease versions as available updates instead of only stable versions.

**Specific Scenario**: 
- User running dev version `v3.2.30-dev.adcb1ae`
- Dev builds disabled in settings
- System incorrectly offered dev version `v3.2.30-dev.adcb1ae` as update instead of latest stable version `v3.2.29`

## Root Cause

The `fetchLatestAssetPreferPrerelease()` function in `update.go` had insufficient filtering logic:

1. **Incomplete Version Collection**: Only collected the first release tag instead of all available tags
2. **Missing Prerelease Filtering**: No proper separation of stable and prerelease versions based on `preferPrerelease` parameter
3. **No Prerelease Detection**: Lacked a helper function to identify prerelease versions

## Fix Implementation

### 1. Enhanced Version Collection

**Before**: Only collected first release tag
```go
// Old approach - only got first match
if m := prereleaseRe.FindStringSubmatch(html); len(m) >= 2 {
    tag = m[1]
    // ... return immediately
}
```

**After**: Collects all release tags for proper filtering
```go
// New approach - collect all tags
tagPattern := fmt.Sprintf(`/%s/%s/releases/tag/([^"']+)`, regexp.QuoteMeta(owner), regexp.QuoteMeta(repo))
tagRe := regexp.MustCompile(tagPattern)
tagMatches := tagRe.FindAllStringSubmatch(html, -1)

// Extract all tags
var allTags []string
for _, match := range tagMatches {
    if len(match) >= 2 {
        allTags = append(allTags, match[1])
    }
}
```

### 2. Added `isPrereleaseTag()` Helper Function

New function to identify prerelease versions:

```go
// isPrereleaseTag checks if a version tag represents a prerelease/dev version
func isPrereleaseTag(tag string) bool {
    tag = strings.ToLower(tag)
    prereleaseIndicators := []string{"-dev", "-beta", "-rc", "-alpha", "-pre"}

    for _, indicator := range prereleaseIndicators {
        if strings.Contains(tag, indicator) {
            return true
        }
    }

    return false
}
```

**Supported Prerelease Types**:
- `-dev` - Development builds (e.g., `v3.2.30-dev.adcb1ae`)
- `-beta` - Beta releases (e.g., `v3.2.28-beta`)
- `-rc` - Release candidates (e.g., `v3.2.27-rc.1`)
- `-alpha` - Alpha releases (e.g., `v3.2.26-alpha`)
- `-pre` - Pre-releases (e.g., `v3.2.25-pre`)

### 3. Enhanced Filtering Logic

**Before**: No filtering based on `preferPrerelease` parameter
**After**: Proper filtering with error handling

```go
// If preferPrerelease is false, filter out dev/prerelease versions
if !preferPrerelease {
    var stableTags []string
    for _, tag := range allTags {
        if !isPrereleaseTag(tag) {
            stableTags = append(stableTags, tag)
        }
    }
    if len(stableTags) == 0 {
        return "", "", fmt.Errorf("no stable releases found for %s/%s", owner, repo)
    }
    tag = stableTags[0] // Return the first stable tag found
} else {
    // When preferring prerelease, return the first tag found
    tag = allTags[0]
}
```

## Testing

### Unit Tests Created

1. **`TestIsPrereleaseTag()`** - Tests the new helper function
   - 40+ test cases covering all prerelease types
   - Case sensitivity testing
   - Position validation (must be after dash)
   - Edge cases (empty strings, unknown types)

2. **`TestFilterStableReleasesLogic()`** - Tests filtering logic
   - Mixed dev and stable versions
   - Only dev versions (should error)
   - Only stable versions
   - Mixed prerelease types

3. **`TestVersionFilteringEdgeCases()`** - Tests edge cases
   - Empty version lists
   - Single version scenarios
   - Complex version formats with build metadata

4. **`TestBugReportScenario()`** - Tests specific bug report scenario
   - Recreates exact scenario from bug report
   - Verifies fix works correctly
   - Tests both dev enabled/disabled states

### Integration Tests

Updated existing integration tests in `update_integration_test.go` and `update_stable_test.go` to verify the fix works end-to-end.

### Test Script Updates

Enhanced `run_tests.sh` to include:
- Prerelease tag detection tests
- Stable version filtering tests  
- Bug report scenario tests

## Behavior Changes

### Before Fix
- `preferPrerelease=false` could still return dev versions
- No error when no stable versions available
- Inconsistent version selection

### After Fix
- `preferPrerelease=false` **always** returns stable versions only
- Proper error when no stable versions exist: `"no stable releases found for owner/repo"`
- Consistent version selection based on user preference

## Verification

### Test Commands

```bash
# Run all prerelease filtering tests
go test -v -run TestIsPrereleaseTag

# Run filtering logic tests
go test -v -run "TestFilterStableReleasesLogic|TestFetchLatestAssetPreferStableLogic|TestVersionFilteringEdgeCases"

# Run bug report scenario tests
go test -v -run TestBugReportScenario

# Run all new tests
go test -v -run "TestIsPrereleaseTag|TestFilterStableReleasesLogic|TestFetchLatestAssetPreferStableLogic|TestVersionFilteringEdgeCases|TestBugReportScenario"
```

### Expected Behavior

1. **Dev Builds Disabled (`preferPrerelease=false`)**:
   - ✅ Only stable versions returned
   - ✅ Dev/beta/rc/alpha/pre versions filtered out
   - ✅ Error if no stable versions exist

2. **Dev Builds Enabled (`preferPrerelease=true`)**:
   - ✅ Latest version returned (including dev)
   - ✅ No filtering applied

3. **Version Comparison**:
   - ✅ Proper semver comparison maintained
   - ✅ Prerelease rules respected (stable > prerelease for same version)

## Files Modified

- `update.go` - Core fix implementation
- `version_test.go` - Added `isPrereleaseTag()` tests
- `update_test.go` - Added filtering logic tests
- `update_integration_test.go` - Added bug report scenario tests
- `run_tests.sh` - Updated test script

## Backward Compatibility

The fix maintains full backward compatibility:
- Existing API unchanged
- All existing functionality preserved
- Only improves filtering logic when `preferPrerelease=false`

## Conclusion

This fix resolves the bug where dev versions were incorrectly identified as stable versions by:

1. **Properly collecting all available versions** instead of just the first one
2. **Implementing robust prerelease detection** with `isPrereleaseTag()` helper
3. **Adding appropriate filtering logic** based on user preferences
4. **Providing clear error messages** when no stable versions exist
5. **Comprehensive testing** to prevent regressions

The launcher now correctly respects user preferences for dev/stable version selection while maintaining all existing functionality.
# Paginated Scraping Feature Guide

## Overview

The paginated scraping feature enhances TheBoys Launcher's ability to find stable releases across multiple pages of GitHub releases. This implementation ensures that stable releases are properly detected even when they are not on the first page of releases, while maintaining efficiency for dev/prerelease builds.

## Problem Solved

Previously, the launcher would only check the first page of GitHub releases when looking for stable versions. This caused issues when:
- Multiple dev/prerelease releases pushed stable releases to subsequent pages
- Users couldn't access stable releases that weren't immediately visible
- The launcher would incorrectly report no stable releases available

## Implementation Details

### Core Functions

#### `fetchLatestAssetPreferPrerelease(preferPrerelease bool)`

The main function that implements paginated scraping logic:

- **For prerelease/dev builds (`preferPrerelease = true`)**: Only checks page 1 (existing behavior preserved)
- **For stable releases (`preferPrerelease = false`)**: Iterates through multiple pages until stable releases are found
- **Maximum pages**: Limited to 10 pages to avoid infinite loops
- **Early termination**: Stops pagination as soon as stable releases are found

#### `fetchFromPage(page int, preferPrerelease bool)`

Helper function that fetches releases from a specific page:

- Constructs the correct GitHub URL with page parameter
- Fetches and parses the HTML response
- Extracts release tags from the page
- Filters releases based on preference (dev vs stable)
- Returns appropriate results or errors

#### `hasMorePages(html string)`

Helper function that detects if there are more pages available:

- Analyzes the HTML content to determine if more releases exist
- Returns `true` if there are more pages, `false` otherwise
- Helps prevent unnecessary requests when reaching the end

#### `isPrereleaseTag(tag string)`

Helper function that identifies prerelease versions:

- Checks for common prerelease indicators: `-dev`, `-beta`, `-rc`, `-alpha`, `-pre`
- Case-insensitive matching
- Returns `true` for prerelease versions, `false` for stable versions

### Pagination Logic Flow

```
┌─────────────────────────────────────┐
│ fetchLatestAssetPreferPrerelease() │
└─────────────────┬───────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
  preferPrerelease    !preferPrerelease
        │                   │
        ▼                   ▼
┌──────────────┐    ┌──────────────────┐
│ Check page 1 │    │ Check page 1     │
│ only (dev)   │    │ (stable search)  │
└──────┬───────┘    └───────┬────────┘
       │                    │
       ▼                    ▼
┌──────────────┐    ┌──────────────────┐
│ Return dev   │    │ Stable found?    │
│ version      │    └───────┬────────┘
└──────────────┘            │
                          │
                    ┌─────┴─────┐
                    │           │
                  Yes          No
                    │           │
                    ▼           ▼
            ┌──────────────┐ ┌──────────────┐
            │ Return      │ │ Check next   │
            │ stable      │ │ page (2-10)  │
            │ version     │ └──────┬───────┘
            └──────────────┘        │
                                   │
                           ┌───────┴───────┐
                           │ Max pages     │
                           │ reached?     │
                           └──────┬───────┘
                                   │
                              ┌────┴────┐
                              │         │
                            Yes        No
                              │         │
                              ▼         ▼
                        ┌──────────┐ ┌──────────┐
                        │ Error:   │ │ Continue │
                        │ No stable│ │ checking │
                        │ found    │ │ next     │
                        └──────────┘ │ page     │
                                     └──────────┘
```

### Performance Optimizations

1. **Early Termination**: Pagination stops immediately when stable releases are found
2. **Dev Mode Efficiency**: Dev builds only check page 1, maintaining existing performance
3. **Page Limit**: Maximum of 10 pages prevents infinite loops and excessive requests
4. **Error Handling**: Graceful handling of network errors and empty pages

## Configuration

### Maximum Pages

The pagination is limited to 10 pages by default. This can be modified in the `fetchLatestAssetPreferPrerelease` function:

```go
maxPages := 10
```

### Prerelease Indicators

The system recognizes the following prerelease indicators (case-insensitive):
- `-dev`
- `-beta`
- `-rc`
- `-alpha`
- `-pre`

These can be extended in the `isPrereleaseTag` function:

```go
prereleaseIndicators := []string{"-dev", "-beta", "-rc", "-alpha", "-pre"}
```

## Testing

### Test Coverage

The paginated scraping feature includes comprehensive tests:

1. **Core Pagination Tests** (`tests/pagination_test.go`)
   - `TestFetchFromPage`: Tests page fetching logic
   - `TestHasMorePages`: Tests page detection
   - `TestFetchLatestAssetPreferPrereleasePagination`: Tests main pagination logic
   - `TestIsPrereleaseTag`: Tests prerelease detection

2. **Edge Cases Tests** (`tests/pagination_edge_cases_test.go`)
   - Single page repositories
   - Empty repositories
   - Network error simulation
   - Maximum pages reached
   - Complex version formats
   - Case-insensitive detection

3. **Integration Tests** (`tests/pagination_integration_test.go`)
   - Dev to stable switching
   - Stable to dev switching
   - Settings persistence
   - Real-world scenarios

### Running Tests

#### All Pagination Tests
```bash
# Linux/macOS
./test_pagination.sh

# Windows
./test_pagination.ps1
```

#### Individual Test Files
```bash
# Core pagination tests
go test -v tests/pagination_test.go

# Edge cases
go test -v tests/pagination_edge_cases_test.go

# Integration tests
go test -v tests/pagination_integration_test.go
```

#### With Coverage
```bash
go test -cover -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go
```

#### With Race Detection
```bash
go test -race -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go
```

## Usage Examples

### Finding Stable Releases

When the launcher is in stable mode (not preferring prereleases):

```go
// This will check multiple pages until a stable release is found
asset, err := fetchLatestAssetPreferPrerelease(false)
if err != nil {
    // Handle error (e.g., no stable releases found)
    return err
}
// Use the stable release asset
```

### Finding Dev Releases

When the launcher is in dev mode (preferring prereleases):

```go
// This will only check page 1 for efficiency
asset, err := fetchLatestAssetPreferPrerelease(true)
if err != nil {
    // Handle error
    return err
}
// Use the dev release asset
```

## Error Handling

### Common Error Scenarios

1. **No Stable Releases Found**
   ```
   Error: no stable releases found for owner/repo after checking 10 pages
   ```
   - All checked pages contain only prerelease versions
   - Repository has no stable releases

2. **Network Errors**
   ```
   Error: could not find any release tags for owner/repo on page X
   ```
   - Network connectivity issues
   - GitHub API rate limiting
   - Repository not found

3. **Empty Repository**
   ```
   Error: could not find any release tags for owner/repo on page 1
   ```
   - Repository exists but has no releases
   - All releases are deleted

### Error Recovery

The system implements graceful error recovery:

- **Network Errors**: Continues to next page when possible
- **Empty Pages**: Skips empty pages and continues pagination
- **Rate Limiting**: Respects GitHub's rate limits naturally through HTTP delays
- **Max Pages**: Prevents infinite loops with hard limit

## Monitoring and Debugging

### Logging

The pagination process includes comprehensive logging:

```
[INFO] Checking page 1 for stable releases...
[INFO] Found 5 releases on page 1
[INFO] No stable releases found on page 1, checking page 2...
[INFO] Checking page 2 for stable releases...
[INFO] Found stable release: v3.2.29
[INFO] Pagination completed after 2 pages
```

### Performance Metrics

Typical performance characteristics:

- **Dev Mode**: 1 HTTP request (page 1 only)
- **Stable Mode**: 1-10 HTTP requests (stops early when stable found)
- **Average Case**: 2-3 requests for most repositories
- **Worst Case**: 10 requests (no stable releases found)

## Troubleshooting

### Common Issues

1. **Pagination Not Working**
   - Check if `preferPrerelease` parameter is correctly set
   - Verify GitHub repository has stable releases
   - Check network connectivity

2. **Performance Issues**
   - Ensure dev mode is used when appropriate (only checks page 1)
   - Check if repository has excessive prerelease releases
   - Monitor for rate limiting

3. **Incorrect Version Detection**
   - Verify prerelease indicators match your versioning scheme
   - Check case sensitivity in version tags
   - Ensure version tags follow semantic versioning

### Debug Mode

Enable debug logging by setting the environment variable:

```bash
export DEBUG=true
./TheBoysLauncher
```

This will provide detailed pagination information including:
- Pages checked
- Releases found per page
- Filtering decisions
- Early termination reasons

## Future Enhancements

### Potential Improvements

1. **Adaptive Pagination**: Dynamically adjust page limits based on repository size
2. **Parallel Fetching**: Fetch multiple pages concurrently for faster results
3. **Caching**: Cache pagination results to reduce redundant requests
4. **Smart Filtering**: Use GitHub API for more efficient filtering
5. **Configuration**: Make pagination settings user-configurable

### API Considerations

While the current implementation uses HTML scraping for compatibility, future versions could leverage GitHub's REST API:

- Use `/repos/{owner}/{repo}/releases` endpoint
- Implement proper API rate limiting
- Add authentication for higher rate limits
- Use API parameters for filtering (per_page, page)

## Conclusion

The paginated scraping feature significantly improves TheBoys Launcher's ability to find stable releases while maintaining performance for dev builds. The implementation is robust, well-tested, and handles edge cases gracefully.

For questions or issues, please refer to the test files in the `tests/` directory or create an issue in the project repository.
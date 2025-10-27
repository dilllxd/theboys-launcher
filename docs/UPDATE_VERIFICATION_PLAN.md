# TheBoysLauncher - Auto Update Verification Plan
## Unified Mac Release Structure

### Executive Summary

This document provides a comprehensive verification plan to ensure the auto update mechanism works correctly with the new unified Mac release structure. The plan covers all critical update paths, test scenarios, verification checklists, automation strategies, and success criteria.

### 1. Critical Update Paths Identification

#### 1.1 Fresh Install of Universal Binary Receiving First Update
**Description**: User installs universal binary for the first time and receives the first update
**Critical Components**:
- Initial version detection and comparison
- Universal binary asset naming (`TheBoysLauncher-mac-universal`)
- Update download and replacement logic
- Quarantine attribute removal

#### 1.2 Migration from Native Binary to Universal Binary
**Description**: User with existing native binary (Intel/ARM) updates to universal binary
**Critical Components**:
- Asset name change detection (`TheBoysLauncher-mac-native` → `TheBoysLauncher-mac-universal`)
- File path preservation during migration
- Configuration and data continuity
- Architecture compatibility verification

#### 1.3 Standard Update Process After Migration
**Description**: User with universal binary receives subsequent updates
**Critical Components**:
- Universal binary update continuity
- Version comparison with unified naming
- Cross-architecture update verification

#### 1.4 Dev Build to Stable Build Updates
**Description**: Users switching between development and stable builds
**Critical Components**:
- Prerelease tag handling (`dev.<sha>` vs stable versions)
- Version comparison logic with prerelease semantics
- Asset availability for both build types

#### 1.5 Cross-Platform Update Consistency
**Description**: Ensuring update mechanism works consistently across all platforms
**Critical Components**:
- Platform-specific asset naming conventions
- Unified update flow across Windows, macOS, and Linux
- Platform-specific quarantine and permission handling

### 2. Detailed Test Scenarios

#### 2.1 Fresh Install Universal Binary Tests

**Test Case UU-01: Fresh Universal Binary First Update**
```
Preconditions:
- Clean macOS system with no previous installation
- Universal binary v1.0.0 installed
- Network connectivity available

Test Steps:
1. Install universal binary v1.0.0
2. Launch application and trigger update check
3. Verify update detection for v1.0.1
4. Confirm download of universal binary asset
5. Verify quarantine attribute removal
6. Confirm application restart with new version
7. Validate functionality post-update

Expected Results:
- Update correctly detected as available
- Universal binary asset downloaded successfully
- Quarantine attribute removed without errors
- Application restarts with v1.0.1
- All functionality preserved
```

**Test Case UU-02: Fresh Universal Binary No Update Available**
```
Preconditions:
- Clean macOS system with no previous installation
- Latest universal binary installed
- Network connectivity available

Test Steps:
1. Install latest universal binary
2. Launch application and trigger update check
3. Verify "up to date" message
4. Confirm no download attempts

Expected Results:
- Correctly identifies as up to date
- No unnecessary download attempts
- Appropriate user messaging
```

#### 2.2 Migration from Native to Universal Tests

**Test Case NU-01: Intel Native to Universal Migration**
```
Preconditions:
- macOS Intel system with native binary v1.0.0 installed
- User data and configuration present
- Universal binary v1.0.1 available

Test Steps:
1. Launch native binary v1.0.0
2. Trigger update check
3. Verify universal binary v1.0.1 detection
4. Confirm download of universal binary asset
5. Verify migration from native to universal
6. Validate configuration preservation
7. Confirm application restart with universal binary

Expected Results:
- Universal binary correctly detected as update
- Native binary replaced with universal binary
- All user data and settings preserved
- Application functions correctly on Intel hardware
```

**Test Case NU-02: ARM Native to Universal Migration**
```
Preconditions:
- macOS ARM system with native binary v1.0.0 installed
- User data and configuration present
- Universal binary v1.0.1 available

Test Steps:
1. Launch native binary v1.0.0 on ARM hardware
2. Trigger update check
3. Verify universal binary v1.0.1 detection
4. Confirm download of universal binary asset
5. Verify migration from native to universal
6. Validate configuration preservation
7. Confirm application restart with universal binary

Expected Results:
- Universal binary correctly detected as update
- Native binary replaced with universal binary
- All user data and settings preserved
- Application functions correctly on ARM hardware
```

#### 2.3 Standard Update Process Tests

**Test Case SU-01: Universal Binary to Universal Binary Update**
```
Preconditions:
- macOS system with universal binary v1.0.0 installed
- Universal binary v1.0.1 available
- Network connectivity available

Test Steps:
1. Launch universal binary v1.0.0
2. Trigger update check
3. Verify universal binary v1.0.1 detection
4. Confirm download of universal binary asset
5. Verify binary replacement
6. Confirm application restart

Expected Results:
- Update correctly detected and downloaded
- Universal binary naming maintained
- Smooth transition between versions
```

**Test Case SU-02: Universal Binary Update with Network Interruption**
```
Preconditions:
- macOS system with universal binary v1.0.0 installed
- Universal binary v1.0.1 available
- Network connectivity will be interrupted during download

Test Steps:
1. Launch universal binary v1.0.0
2. Trigger update check
3. Start download of universal binary v1.0.1
4. Interrupt network connection during download
5. Verify error handling and recovery
6. Restore network and retry update

Expected Results:
- Graceful handling of network interruption
- Appropriate error messaging
- Ability to retry update successfully
- No corruption of existing installation
```

#### 2.4 Dev Build to Stable Build Tests

**Test Case DS-01: Dev Build to Stable Build Update**
```
Preconditions:
- macOS system with dev build v1.0.0-dev.abc123 installed
- Stable build v1.0.0 available
- Dev builds disabled in settings

Test Steps:
1. Launch dev build v1.0.0-dev.abc123
2. Disable dev builds in settings
3. Trigger update check
4. Verify stable build v1.0.0 detection
5. Confirm download of stable build
6. Verify migration from dev to stable

Expected Results:
- Stable build correctly detected as update
- Dev build replaced with stable build
- Version comparison handles prerelease correctly
```

**Test Case DS-02: Stable Build to Dev Build Update**
```
Preconditions:
- macOS system with stable build v1.0.0 installed
- Dev build v1.0.1-dev.def456 available
- Dev builds enabled in settings

Test Steps:
1. Launch stable build v1.0.0
2. Enable dev builds in settings
3. Trigger update check
4. Verify dev build v1.0.1-dev.def456 detection
5. Confirm download of dev build
6. Verify migration from stable to dev

Expected Results:
- Dev build correctly detected as update
- Stable build replaced with dev build
- Version comparison handles prerelease correctly
```

#### 2.5 Edge Cases and Failure Scenarios

**Test Case EC-01: Corrupted Download Handling**
```
Preconditions:
- macOS system with universal binary v1.0.0 installed
- Universal binary v1.0.1 available
- Download will be corrupted

Test Steps:
1. Launch universal binary v1.0.0
2. Trigger update check
3. Simulate corrupted download
4. Verify error detection and handling
5. Confirm fallback to previous version
6. Retry update with corrected download

Expected Results:
- Corrupted download detected
- Application remains functional with previous version
- Clear error messaging to user
- Ability to retry update successfully
```

**Test Case EC-02: Insufficient Disk Space**
```
Preconditions:
- macOS system with universal binary v1.0.0 installed
- Universal binary v1.0.1 available
- Insufficient disk space for update

Test Steps:
1. Launch universal binary v1.0.0
2. Trigger update check
3. Verify disk space check
4. Confirm appropriate error messaging
5. Free up disk space
6. Retry update successfully

Expected Results:
- Disk space check performed before download
- Clear error message about insufficient space
- Application remains functional
- Update succeeds after space is freed
```

**Test Case EC-03: Permission Denied Scenarios**
```
Preconditions:
- macOS system with universal binary v1.0.0 installed
- Universal binary v1.0.1 available
- Insufficient permissions for update

Test Steps:
1. Launch universal binary v1.0.0
2. Trigger update check
3. Simulate permission denied during replacement
4. Verify error handling and fallback
5. Provide proper permissions
6. Retry update successfully

Expected Results:
- Permission errors handled gracefully
- Application remains functional
- Clear guidance to user about permissions
- Update succeeds after permissions fixed
```

### 3. Comprehensive Verification Checklist

#### 3.1 Code Review Checklist for Update Mechanism Changes

**Update Logic Verification**
- [ ] Version comparison logic handles universal binary naming
- [ ] Asset URL construction works for all Mac variants
- [ ] Prerelease version comparison is correct
- [ ] Fallback mechanisms are implemented
- [ ] Error handling covers all failure scenarios

**Platform-Specific Code**
- [ ] macOS quarantine attribute removal is implemented
- [ ] File permission handling is correct for macOS
- [ ] Process creation works on all macOS versions
- [ ] Architecture detection is accurate

**Asset Management**
- [ ] Universal binary asset naming is consistent
- [ ] Native to universal migration logic is correct
- [ ] Download verification is implemented
- [ ] Temporary file cleanup is handled

#### 3.2 Build Process Verification Steps

**Universal Binary Creation**
- [ ] Intel and ARM binaries build successfully
- [ ] `lipo` tool creates universal binary correctly
- [ ] Universal binary runs on both architectures
- [ ] File size is reasonable for universal binary

**Asset Naming Consistency**
- [ ] Universal binary follows naming convention: `TheBoysLauncher-mac-universal`
- [ ] Native binary follows naming convention: `TheBoysLauncher-mac-native`
- [ ] Version information is correctly embedded
- [ ] Asset names are consistent across build and update systems

**Release Artifact Validation**
- [ ] All required artifacts are generated
- [ ] Universal binary is included in releases
- [ ] Native binaries are included for migration scenarios
- [ ] Checksums are generated and verified

#### 3.3 Auto-Update Functionality Testing

**Update Detection**
- [ ] Correctly identifies when updates are available
- [ ] Properly handles "up to date" scenarios
- [ ] Version comparison works for all version formats
- [ ] Prerelease handling respects user settings

**Download Process**
- [ ] Downloads correct asset for current platform
- [ ] Handles network interruptions gracefully
- [ ] Verifies download integrity
- [ ] Manages insufficient disk space

**Installation Process**
- [ ] Replaces binary correctly
- [ ] Removes quarantine attributes
- [ ] Sets appropriate permissions
- [ ] Handles permission denied scenarios

**Post-Update Verification**
- [ ] Application restarts successfully
- [ ] New version is correctly identified
- [ ] User data is preserved
- [ ] All functionality works correctly

### 4. Test Automation Strategies

#### 4.1 Automated Test Scripts

**Update Simulation Script** (`scripts/test-update-mechanism.sh`)
```bash
#!/bin/bash
# Comprehensive update mechanism testing script

# Test scenarios:
# 1. Fresh install universal binary update
# 2. Native to universal migration
# 3. Universal to universal update
# 4. Dev to stable build transition
# 5. Error condition handling

# Implementation approach:
# - Create isolated test environments
# - Mock GitHub releases with test artifacts
# - Simulate various network conditions
# - Verify file system changes
# - Validate application behavior
```

**Asset Naming Verification Script** (`scripts/verify-asset-naming.sh`)
```bash
#!/bin/bash
# Verify asset naming conventions across releases

# Check:
# - Universal binary naming consistency
# - Native binary naming for migration
# - Version string formatting
# - Architecture-specific naming
# - Cross-platform naming patterns
```

**Version Comparison Test** (`scripts/test-version-comparison.sh`)
```bash
#!/bin/bash
# Test version comparison logic

# Test cases:
# - Standard semantic version comparison
# - Prerelease version handling
# - Dev build version comparison
# - Edge cases and malformed versions
# - Cross-platform version consistency
```

#### 4.2 Continuous Integration Tests

**GitHub Actions Workflow** (`.github/workflows/test-update-mechanism.yml`)
```yaml
name: Test Update Mechanism

on:
  pull_request:
    paths:
      - 'update*.go'
      - 'platform*.go'
      - 'scripts/**'
  push:
    branches: [main, dev]

jobs:
  test-update-mechanism:
    runs-on: macos-latest
    strategy:
      matrix:
        arch: [amd64, arm64]
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '^1.23'
      
      - name: Test update mechanism
        run: |
          ./scripts/test-update-mechanism.sh
          ./scripts/verify-asset-naming.sh
          ./scripts/test-version-comparison.sh
      
      - name: Test cross-architecture compatibility
        run: |
          ./scripts/test-cross-arch-updates.sh
```

#### 4.3 Cross-Platform Compatibility Tests

**Architecture Compatibility Test** (`scripts/test-cross-arch-updates.sh`)
```bash
#!/bin/bash
# Test cross-architecture update compatibility

# Test scenarios:
# - Intel system updating to universal binary
# - ARM system updating to universal binary
# - Universal binary running on both architectures
# - Performance impact of universal binary
```

**Platform Consistency Test** (`scripts/test-platform-consistency.sh`)
```bash
#!/bin/bash
# Test update consistency across platforms

# Verify:
# - Same update logic across Windows, macOS, Linux
# - Platform-specific asset handling
# - Consistent user experience
# - Cross-platform version synchronization
```

### 5. Success Criteria and Metrics

#### 5.1 Update Success Metrics

**Primary Success Metrics**
- **Update Success Rate**: ≥ 99.5% of update attempts complete successfully
- **Update Completion Time**: ≤ 2 minutes for typical update on standard broadband
- **Migration Success Rate**: ≥ 99% of native-to-universal migrations complete without data loss
- **Rollback Success Rate**: ≥ 99% of failed updates rollback to previous version successfully

**Secondary Success Metrics**
- **User Experience Score**: ≥ 4.5/5 for update process user experience
- **Support Ticket Reduction**: ≤ 5% of support tickets related to update issues
- **Adoption Rate**: ≥ 90% of users migrate to universal binary within 30 days

#### 5.2 Performance Benchmarks

**Update Process Performance**
- **Update Detection**: ≤ 5 seconds to check for updates
- **Download Speed**: Utilize ≥ 80% of available bandwidth
- **Installation Time**: ≤ 30 seconds for binary replacement
- **Restart Time**: ≤ 10 seconds for application restart

**Resource Usage**
- **Memory Usage During Update**: ≤ 100MB additional memory usage
- **Disk Space Requirements**: ≤ 2x current binary size during update
- **CPU Usage**: ≤ 50% CPU utilization during update process

#### 5.3 Quality Measures

**Code Quality**
- **Test Coverage**: ≥ 90% code coverage for update mechanism
- **Static Analysis**: Zero critical issues in static code analysis
- **Security Review**: No high-severity security vulnerabilities

**Reliability Measures**
- **Mean Time Between Failures**: ≥ 1000 update operations
- **Error Recovery**: 100% recovery from transient errors
- **Data Integrity**: Zero instances of user data corruption during updates

#### 5.4 User Experience Quality Measures

**Usability Metrics**
- **Update Process Clarity**: Clear progress indicators and status messages
- **Error Messaging**: Understandable error messages with actionable guidance
- **Interruption Handling**: Graceful handling of user interruptions

**Accessibility Compliance**
- **Screen Reader Compatibility**: Update process accessible via screen readers
- **Keyboard Navigation**: Full keyboard navigation support
- **Visual Indicators**: Clear visual feedback for all update states

### 6. Implementation Timeline

#### Phase 1: Test Infrastructure Setup (Week 1)
- Create automated test scripts
- Set up CI/CD test pipelines
- Establish test environments
- Create mock release infrastructure

#### Phase 2: Core Update Path Testing (Week 2)
- Test fresh install scenarios
- Verify native-to-universal migration
- Validate universal-to-universal updates
- Test dev-to-stable transitions

#### Phase 3: Edge Case and Failure Testing (Week 3)
- Test network interruption scenarios
- Verify error handling and recovery
- Test permission and disk space issues
- Validate rollback mechanisms

#### Phase 4: Cross-Platform and Performance Testing (Week 4)
- Test cross-architecture compatibility
- Verify platform consistency
- Performance benchmarking
- Load testing with concurrent updates

#### Phase 5: User Acceptance and Final Validation (Week 5)
- User acceptance testing
- Final validation against success criteria
- Documentation updates
- Release preparation

### 7. Risk Mitigation Strategies

#### 7.1 Technical Risks

**Universal Binary Compatibility Issues**
- **Risk**: Universal binary fails on specific hardware configurations
- **Mitigation**: Comprehensive testing on diverse hardware, fallback to native binaries

**Migration Data Loss**
- **Risk**: User data corruption during native-to-universal migration
- **Mitigation**: Backup mechanisms, data validation, rollback capabilities

**Update Process Failures**
- **Risk**: Update process fails leaving application in unusable state
- **Mitigation**: Atomic updates, rollback mechanisms, recovery procedures

#### 7.2 Operational Risks

**Release Coordination**
- **Risk**: Inconsistent release timing across platforms
- **Mitigation**: Coordinated release process, automated verification

**User Communication**
- **Risk**: Insufficient communication about universal binary transition
- **Mitigation**: Clear documentation, in-app notifications, support preparation

**Support Load**
- **Risk**: Increased support tickets during transition period
- **Mitigation**: Proactive support preparation, common issue documentation

### 8. Documentation and Training

#### 8.1 Technical Documentation

**Update Mechanism Architecture**
- Detailed explanation of update flow
- Platform-specific considerations
- Error handling procedures
- Troubleshooting guides

**Release Process Documentation**
- Step-by-step release procedures
- Asset naming conventions
- Verification checklists
- Rollback procedures

#### 8.2 User-Facing Documentation

**Update Process Guide**
- What to expect during updates
- Troubleshooting common issues
- Contact information for support
- FAQ for update-related questions

**Migration Guide**
- Explanation of universal binary benefits
- Migration process overview
- Data preservation assurances
- Performance expectations

### 9. Conclusion

This comprehensive verification plan ensures that the auto update mechanism works correctly with the new unified Mac release structure. By systematically addressing all critical update paths, implementing thorough test scenarios, and establishing clear success criteria, we can confidently deliver a robust update experience that maintains data integrity and provides a smooth transition to universal binaries.

The plan emphasizes automation, cross-platform consistency, and user experience quality while providing clear risk mitigation strategies and success metrics. Implementation of this plan will ensure that users receive reliable, efficient updates regardless of their starting point or hardware configuration.
# AI Implementation Prompt for TheBoys Launcher Migration

## Instructions for AI Assistant

You are helping migrate TheBoys Launcher from Go to Rust + Tauri. Follow these rules strictly:

1. **NO PLACEHOLDERS**: Every feature must be fully implemented
2. **NO TODOs**: Complete all functionality before moving on
3. **NO MOCK CODE**: Real implementations only
4. **REFERENCE LEGACY CODE**: Use `legacy/` folder to understand existing logic
5. **POLISH FIRST**: User experience and visual quality are paramount
6. **CROSS-PLATFORM**: Must work on Windows, macOS, and Linux

## Current Slice Context

## SLICE 12: Packaging & Distribution

### Package Requirements
1. **Windows**
   - MSI installer with proper shortcuts
   - Auto-start on windows boot option
   - File association handling
   - Windows 10/11 compatibility

2. **macOS**
   - DMG package with drag-and-drop
   - Notarization for Gatekeeper
   - App Store submission ready
   - macOS 11+ compatibility

3. **Linux**
   - AppImage for universal distribution
   - DEB package for Debian/Ubuntu
   - RPM package for Fedora/RHEL
   - AUR package for Arch Linux

### Distribution Setup
1. **Auto-Updater Integration**
   - Background update checks
   - Silent update option
   - Update notifications
   - Rollback capability

### Success Criteria
- [ ] Installers work on all platforms
- [ ] Auto-updater functions correctly
- [ ] All package formats install properly
- [ ] Release process documented

---

## Legacy Reference

The original Go code is in the `legacy/` folder. Key files to reference:
- `main.go` - Main application logic and flow
- `go.mod` - Dependencies and version info
- Any other relevant files for the current slice

## Implementation Requirements

### Backend (Rust) Requirements

1. **Code Quality**:
   - Use proper error handling with `Result<T, E>`
   - Include comprehensive documentation
   - Follow Rust best practices and idioms
   - Use `#[derive(Debug, Clone, Serialize, Deserialize)]` for data structures
   - Implement proper logging with `tracing` or similar

2. **Tauri Commands**:
   - All commands must return `Result<T, String>` for proper error handling
   - Include detailed error messages
   - Handle all edge cases
   - Validate inputs properly

3. **File Operations**:
   - Use `tokio` for async operations
   - Handle file permissions properly
   - Create directories as needed
   - Clean up temporary files

4. **HTTP Operations**:
   - Use `reqwest` for HTTP requests
   - Include proper timeout handling
   - Handle network errors gracefully
   - Include user agent headers

### Frontend (TypeScript/React) Requirements

1. **Code Quality**:
   - Use strict TypeScript
   - Include proper type definitions
   - Use functional components with hooks
   - Include proper error boundaries

2. **UI Components**:
   - Use a consistent design system
   - Include loading states for all async operations
   - Provide proper visual feedback
   - Handle empty states gracefully

3. **State Management**:
   - Use proper state management (Zustand recommended)
   - Include optimistic updates where appropriate
   - Handle race conditions properly
   - Persist important state to backend

4. **User Experience**:
   - Include smooth transitions and animations
   - Provide helpful tooltips and guidance
   - Handle errors gracefully with recovery options
   - Include keyboard navigation support

## Testing Requirements

1. **Backend Tests**:
   - Unit tests for all functions
   - Integration tests for Tauri commands
   - Mock external dependencies in tests
   - Include edge case testing

2. **Frontend Tests**:
   - Component tests for all UI components
   - Integration tests for user flows
   - Mock API responses in tests
   - Include accessibility testing

## Implementation Steps

1. **Set up the project structure** for the current slice
2. **Implement backend functionality** first, with tests
3. **Create frontend components** with proper styling
4. **Integrate frontend and backend** through Tauri commands
5. **Add comprehensive error handling** throughout
6. **Implement loading states** and user feedback
7. **Add tests** for all functionality
8. **Test on all target platforms** (Windows, macOS, Linux)
9. **Polish the UI** with animations and transitions
10. **Verify cross-platform compatibility**

## Quality Checklist

Before completing each slice, verify:

- [ ] All functionality is implemented without placeholders
- [ ] Error handling covers all edge cases
- [ ] UI provides clear feedback for all operations
- [ ] Loading states are present for async operations
- [ ] Code is properly documented
- [ ] Tests are comprehensive and passing
- [ ] Application works on all target platforms
- [ ] UI is responsive and visually polished
- [ ] Keyboard navigation works
- [ ] Accessibility features are implemented

## Deliverables

For each slice, provide:

1. **Complete Rust backend code** with all functions implemented
2. **Complete TypeScript frontend code** with all components
3. **Comprehensive tests** for both backend and frontend
4. **Updated Tauri configuration** if needed
5. **Documentation** for any new APIs or components
6. **Cross-platform testing results**

## Legacy Integration Guidelines

When referencing the legacy Go code:

1. **Understand the logic** but implement it idiomatically in Rust/TypeScript
2. **Improve upon the original** where possible (better error handling, UI, etc.)
3. **Maintain compatibility** with existing modpack formats and configurations
4. **Preserve all functionality** while enhancing the user experience

## Example Implementation Pattern

### Backend Command Example:
```rust
#[tauri::command]
async fn download_file(url: String, destination: String) -> Result<(), String> {
    // Validate inputs
    if url.is_empty() || destination.is_empty() {
        return Err("URL and destination cannot be empty".to_string());
    }
    
    // Create parent directories
    if let Some(parent) = Path::new(&destination).parent() {
        tokio::fs::create_dir_all(parent)
            .await
            .map_err(|e| format!("Failed to create directory: {}", e))?;
    }
    
    // Download file with progress tracking
    // ... complete implementation
    
    Ok(())
}
```

### Frontend Component Example:
```typescript
interface DownloadProps {
  url: string;
  destination: string;
  onComplete: () => void;
  onError: (error: string) => void;
}

const DownloadComponent: React.FC<DownloadProps> = ({ 
  url, 
  destination, 
  onComplete, 
  onError 
}) => {
  const [progress, setProgress] = useState(0);
  const [isDownloading, setIsDownloading] = useState(false);
  
  const handleDownload = async () => {
    setIsDownloading(true);
    setProgress(0);
    
    try {
      await invoke('download_file', { url, destination });
      onComplete();
    } catch (error) {
      onError(error as string);
    } finally {
      setIsDownloading(false);
    }
  };
  
  return (
    <div className="download-container">
      {/* Complete UI implementation */}
    </div>
  );
};
```

## Success Criteria

Each slice is complete only when:

1. **All functionality works** without any placeholders
2. **Error handling is comprehensive** and user-friendly
3. **UI is polished and responsive**
4. **Cross-platform compatibility** is verified
5. **Tests are comprehensive** and passing
6. **Code is documented** and maintainable
7. **User experience is excellent** with proper feedback

## Next Steps

After completing a slice:

1. **Review the implementation** against requirements
2. **Test thoroughly** on all platforms
3. **Refactor if needed** for code quality
4. **Update documentation**
5. **Proceed to next slice** only when current slice is 100% complete

Remember: Quality is paramount. Do not rush through slices or leave incomplete functionality. Every feature must be production-ready before moving forward.
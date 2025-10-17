# TheBoys Launcher Rebuild Plan

## Project Overview
Rebuild the TheBoys Launcher from a single-file Windows-only TUI application to a cross-platform GUI application with proper architecture and maintainability.

## Core Requirements
- Cross-platform support (Windows, macOS, Linux)
- Modern GUI instead of TUI
- Proper project architecture (no single monolithic files)
- Platform-specific code separation
- No placeholder/bandaid code
- Maintainable and testable structure

## Phase 1: Project Setup & Architecture (Foundation)

### 1.1 Legacy Code Management
- [ ] Create `legacy/` directory
- [ ] Move existing files to legacy folder
- [ ] Document legacy code structure

### 1.2 New Project Structure
```
theboys-launcher/
├── cmd/
│   └── launcher/
│       └── main.go              # Application entry point only
├── internal/
│   ├── app/                     # Application orchestration
│   │   ├── app.go               # Main application state
│   │   └── lifecycle.go         # Startup/shutdown logic
│   ├── gui/                     # Wails frontend code
│   │   ├── events.go            # Event definitions
│   │   ├── handlers.go          # Backend API handlers
│   │   ├── models.go            # Data models for frontend
│   │   └── components/          # UI components (optional)
│   ├── launcher/                # Core launcher logic
│   │   ├── modpack.go           # Modpack management
│   │   ├── java.go              # Java runtime management
│   │   ├── prism.go             # Prism launcher integration
│   │   ├── updater.go           # Self-update logic
│   │   ├── downloader.go        # Download utilities
│   │   └── instance.go          # Instance management
│   ├── config/                  # Configuration management
│   │   ├── settings.go          # User settings
│   │   └── modpacks.go          # Modpack configuration
│   ├── platform/                # Platform-specific code
│   │   ├── interface.go         # Platform interface definitions
│   │   ├── windows.go           # Windows-specific implementations
│   │   ├── darwin.go            # macOS-specific implementations
│   │   ├── linux.go             # Linux-specific implementations
│   │   └── common.go            # Cross-platform utilities
│   └── logging/                 # Logging infrastructure
│       ├── logger.go            # Logger setup
│       └── rotation.go          # Log rotation logic
├── pkg/
│   └── types/                   # Public type definitions
│       └── types.go
├── assets/                      # Static assets
│   ├── icons/                   # Application icons
│   └── images/                  # UI images
├── configs/                     # Configuration files
│   └── modpacks.json
├── docs/                        # Documentation
│   └── ARCHITECTURE.md
├── scripts/                     # Build scripts
├── wails.json                   # Wails configuration
├── go.mod                       # Go module file
├── go.sum                       # Go dependencies
└── README.md                    # Project documentation
```

### 1.3 Wails Project Initialization
- [ ] Install Wails v2
- [ ] Initialize Wails project structure
- [ ] Configure wails.json for cross-platform builds
- [ ] Set up basic frontend template

## Phase 2: Core Infrastructure (Backend Foundation)

### 2.1 Platform Interface Design
```go
type Platform interface {
    GetExecutablePath() (string, error)
    GetAppDataDir() (string, error)
    DetectJavaInstallations() ([]JavaInstallation, error)
    LaunchProcess(cmd string, args []string) error
    GetDefaultRAM() int
    SupportsAutoUpdate() bool
}
```

### 2.2 Configuration System
- [ ] Implement settings management with JSON persistence
- [ ] Create modpack configuration loader
- [ ] Add settings validation
- [ ] Implement default values

### 2.3 Logging Infrastructure
- [ ] Set up structured logging with proper levels
- [ ] Implement log rotation
- [ ] Add GUI log viewer interface
- [ ] Create different log destinations (file, GUI)

### 2.4 Download & Update System
- [ ] Create generic download manager
- [ ] Implement progress tracking
- [ ] Add retry logic and error handling
- [ ] Create update checker with GitHub integration

## Phase 3: Launcher Core Logic (Business Logic)

### 3.1 Modpack Management
- [ ] Port modpack loading logic
- [ ] Implement modpack selection and validation
- [ ] Add modpack metadata handling
- [ ] Create modpack update checking

### 3.2 Java Runtime Management
- [ ] Implement cross-platform Java detection
- [ ] Create Java download manager (Adoptium API)
- [ ] Add Java version validation
- [ ] Implement Java installation verification

### 3.3 Prism Launcher Integration
- [ ] Create Prism download manager
- [ ] Implement Prism instance creation
- [ ] Add Prism configuration management
- [ ] Create Prism launcher interface

### 3.4 Instance Management
- [ ] Design instance data structure
- [ ] Implement instance creation and deletion
- [ ] Add instance configuration management
- [ ] Create instance backup/restore functionality

## Phase 4: GUI Development (Frontend)

### 4.1 Basic UI Layout
- [ ] Create main window layout
- [ ] Implement navigation structure
- [ ] Add status bar and progress indicators
- [ ] Create responsive design

### 4.2 Main Components
- [ ] Modpack selection interface
- [ ] Settings/configuration panel
- [ ] Instance management view
- [ ] Log viewer component
- [ ] Update progress dialog

### 4.3 User Experience Features
- [ ] Add keyboard shortcuts
- [ ] Implement drag-and-drop functionality
- [ ] Create notification system
- [ ] Add tooltips and help text

### 4.4 Visual Design
- [ ] Design modern UI theme
- [ ] Create consistent color scheme
- [ ] Add loading animations
- [ ] Implement dark/light mode toggle

## Phase 5: Integration & Testing (Polish)

### 5.1 Backend-Frontend Integration
- [ ] Connect all GUI components to backend logic
- [ ] Implement event-driven communication
- [ ] Add error handling and user feedback
- [ ] Create data synchronization

### 5.2 Cross-Platform Testing
- [ ] Test on Windows
- [ ] Test on macOS
- [ ] Test on Linux
- [ ] Fix platform-specific issues

### 5.3 Performance Optimization
- [ ] Optimize startup time
- [ ] Reduce memory usage
- [ ] Improve download speeds
- [ ] Add caching mechanisms

### 5.4 Final Polish
- [ ] Add comprehensive error handling
- [ ] Implement graceful shutdown
- [ ] Create user documentation
- [ ] Add application icons and metadata

## Development Rules

### No Placeholder Code
- Every function must have a complete implementation
- Use interfaces and dependency injection for testability
- Implement proper error handling from the start
- Create realistic test data

### Code Quality Standards
- Follow Go best practices and idioms
- Use meaningful variable and function names
- Add comprehensive comments for complex logic
- Implement proper package organization

### Testing Strategy
- Write unit tests for all business logic
- Create integration tests for major workflows
- Add end-to-end tests for critical paths
- Test error conditions and edge cases

### Documentation Requirements
- Document all exported functions
- Create architecture documentation
- Add inline comments for complex algorithms
- Maintain up-to-date README

## Success Criteria

1. **Functional**: All features from legacy version working
2. **Cross-Platform**: Runs on Windows, macOS, Linux
3. **Maintainable**: Clean architecture, well-documented
4. **User-Friendly**: Modern GUI with good UX
5. **Robust**: Comprehensive error handling and testing

## Timeline Estimation

- **Phase 1**: 1-2 days
- **Phase 2**: 2-3 days
- **Phase 3**: 3-4 days
- **Phase 4**: 4-5 days
- **Phase 5**: 2-3 days

**Total**: 12-17 days

---

## Progress Tracking

### Phase 1: Project Setup & Architecture ✅ COMPLETED
- [x] Create `legacy/` directory
- [x] Move existing files to legacy folder
- [x] Document legacy code structure
- [x] New project structure
- [x] Install Wails v2
- [x] Initialize Wails project structure
- [x] Configure wails.json for cross-platform builds
- [x] Set up basic frontend template

### Phase 2: Core Infrastructure ✅ COMPLETED
- [x] Platform interface design (implemented in internal/platform/)
- [x] Implement settings management with JSON persistence
- [x] Create modpack configuration loader
- [x] Add settings validation
- [x] Implement default values
- [x] Set up structured logging with proper levels
- [x] Implement log rotation
- [x] Create different log destinations (file, console)

### Phase 3: Launcher Core Logic (Business Logic) ✅ COMPLETED
- [x] Port modpack loading logic
- [x] Implement modpack selection and validation
- [x] Add modpack metadata handling
- [x] Create modpack update checking
- [x] Implement cross-platform Java detection
- [x] Create Java download manager (Adoptium API)
- [x] Add Java version validation
- [x] Implement Java installation verification
- [x] Create Prism download manager
- [x] Implement Prism instance creation
- [x] Add Prism configuration management
- [x] Create Prism launcher interface
- [x] Design instance data structure
- [x] Implement instance creation and deletion
- [x] Add instance configuration management
- [x] Create instance backup/restore functionality

### Phase 4: GUI Development (Frontend) 🚧 IN PROGRESS
- [x] Create main window layout
- [x] Implement navigation structure
- [x] Add status bar and progress indicators
- [x] Create responsive design
- [x] Modpack selection interface
- [ ] Settings/configuration panel
- [x] Instance management view
- [x] Log viewer component
- [ ] Update progress dialog

### Phase 5: Integration & Testing (Polish) 📋 PENDING
- [ ] Connect all GUI components to backend logic
- [ ] Implement event-driven communication
- [ ] Add error handling and user feedback
- [ ] Create data synchronization
- [ ] Test on Windows, macOS, Linux
- [ ] Fix platform-specific issues
- [ ] Optimize startup time and memory usage
- [ ] Add comprehensive error handling and documentation

## Current Status

✅ **Phase 1, 2 & 3 Complete**: Successfully created cross-platform project structure with:
- ✅ Modular architecture separated by platform (Windows/macOS/Linux)
- ✅ Configuration management with JSON persistence and validation
- ✅ Cross-platform logging with rotation and multiple outputs
- ✅ Complete modpack management system with remote fetching and validation
- ✅ Cross-platform Java detection, download, and installation via Adoptium API
- ✅ Complete Prism Launcher integration with instance management
- ✅ Comprehensive test coverage for all core components
- ✅ Successfully compiled executable (8.7MB binary)

🚧 **Current Progress**: Phase 4 is well underway with modern GUI implementation:
- ✅ Modern dark theme with glassmorphism effects and Inter font
- ✅ Responsive main window layout with header, sidebar, and content area
- ✅ Navigation sidebar with main and management sections
- ✅ Modpack browser with card-based layout and search functionality
- ✅ Instance management interface with launch controls and status indicators
- ✅ Java installations viewer with detection information
- ✅ Log viewer container with terminal-style output
- ✅ Real-time status bar with system information and live time updates
- ✅ Full backend integration with all GUI components connected to Go functions
- ✅ Successfully compiled application (8.6MB binary) with embedded frontend

📋 **Next Steps**: Complete remaining Phase 4 items (settings panel, progress dialogs) and move to Phase 5 for final integration and testing.
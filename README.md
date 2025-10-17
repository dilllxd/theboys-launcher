# TheBoys Launcher

A cross-platform, modern GUI Minecraft modpack launcher built with Go and Wails.

## Project Status

🚧 **Currently in Development - Phase 3: Core Logic Implementation**

This is a complete rebuild of the original TheBoys Launcher, transforming it from a Windows-only TUI application to a cross-platform GUI with proper architecture.

## ✅ Completed Features

### Phase 1: Project Architecture ✅
- Cross-platform project structure
- Wails v2 GUI framework integration
- Modular architecture with proper separation of concerns
- Legacy code preservation in `/legacy` folder

### Phase 2: Core Infrastructure ✅
- **Platform Abstraction**: Windows, macOS, and Linux support
- **Configuration Management**: JSON-based settings with validation
- **Logging System**: Structured logging with rotation and multiple outputs
- **Type Safety**: Comprehensive type definitions
- **Basic GUI**: Functional application with modern interface

## 🚧 In Progress

### Phase 3: Core Launcher Logic
- [ ] Modpack management system
- [ ] Java runtime detection and management
- [ ] Prism Launcher integration
- [ ] Instance management

## 📋 Planned Features

### Phase 4: GUI Development
- Modern, responsive user interface
- Modpack selection and management UI
- Settings and configuration panels
- Progress indicators and status updates
- Dark/light theme support

### Phase 5: Integration & Polish
- Cross-platform testing
- Performance optimization
- Error handling and user feedback
- Documentation and help system

## Architecture

```
theboys-launcher/
├── cmd/launcher/          # Application entry point
├── internal/
│   ├── app/              # Main application orchestration
│   ├── config/           # Configuration management
│   ├── gui/              # GUI components and handlers
│   ├── launcher/         # Core launcher logic (modpacks, java, prism)
│   ├── logging/          # Logging infrastructure
│   └── platform/         # Platform-specific implementations
├── pkg/types/            # Public type definitions
├── configs/              # Configuration files
├── assets/               # Static assets
└── legacy/               # Original codebase
```

## Technology Stack

- **Backend**: Go 1.23
- **GUI Framework**: Wails v2 (web-based UI)
- **Frontend**: HTML, CSS, JavaScript (Vite)
- **Build System**: Go modules + Wails build system
- **Target Platforms**: Windows, macOS, Linux

## Building

### Prerequisites

- Go 1.23+
- Node.js 18+
- Wails v2

### Build Steps

1. Install dependencies:
   ```bash
   go mod tidy
   cd frontend && npm install && cd ..
   ```

2. Build frontend:
   ```bash
   cd frontend && npm run build && cd ..
   ```

3. Build application:
   ```bash
   go build -o theboys-launcher ./cmd/launcher
   ```

### Development

For development with hot reload:

```bash
# Start development server
wails dev
```

## Progress

This project follows a structured development approach outlined in [REBUILD_PLAN.md](./REBUILD_PLAN.md). Progress is tracked through phases to ensure quality and maintainability.

### Current Phase: 3/5
- ✅ Phase 1: Project Setup & Architecture
- ✅ Phase 2: Core Infrastructure
- 🚧 Phase 3: Core Launcher Logic
- 📋 Phase 4: GUI Development
- 📋 Phase 5: Integration & Testing

## Contributing

This is currently in active development. Please refer to the development plan and progress tracking in the project documentation.

## License

[Add your license here]
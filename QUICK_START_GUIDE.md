# TheBoys Launcher - Rust + Tauri Quick Start Guide

## Prerequisites

1. **Install Rust**: https://rustup.rs/
2. **Install Node.js**: https://nodejs.org/ (v18 or later)
3. **Install Tauri CLI**: `cargo install tauri-cli`
4. **Code Editor**: VS Code with Rust and TypeScript extensions recommended

## Initial Setup

### 1. Create the Tauri Project

```bash
# Create new Tauri project with TypeScript template
npm create tauri-app@latest theboys-launcher -- --template react-ts

# Navigate to project directory
cd theboys-launcher

# Install dependencies
npm install
```

### 2. Configure Project Structure

```bash
# Create additional directories
mkdir -p src-tauri/src/{commands,models,utils,downloader,launcher}
mkdir -p src/{components,pages,types,utils,styles}
mkdir -p legacy

# Copy original Go code to legacy folder
cp /path/to/original/go/code/* legacy/
```

### 3. Add Required Dependencies

#### Backend Dependencies (add to src-tauri/Cargo.toml):
```toml
[dependencies]
tauri = { version = "1.0", features = ["api-all"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tokio = { version = "1.0", features = ["full"] }
reqwest = { version = "0.11", features = ["json"] }
anyhow = "1.0"
thiserror = "1.0"
tracing = "0.1"
tracing-subscriber = "0.3"
dirs = "5.0"
zip = "0.6"
semver = "1.0"
```

#### Frontend Dependencies (add to package.json):
```json
{
  "dependencies": {
    "@tauri-apps/api": "^1.0",
    "react": "^18.0",
    "react-dom": "^18.0",
    "react-router-dom": "^6.0",
    "zustand": "^4.0",
    "styled-components": "^5.0",
    "@types/styled-components": "^5.0"
  },
  "devDependencies": {
    "@types/react": "^18.0",
    "@types/react-dom": "^18.0",
    "@typescript-eslint/eslint-plugin": "^5.0",
    "@typescript-eslint/parser": "^5.0",
    "eslint": "^8.0",
    "eslint-plugin-react": "^7.0",
    "eslint-plugin-react-hooks": "^4.0",
    "prettier": "^2.0",
    "typescript": "^4.0"
  }
}
```

### 4. Configure Tauri Settings

Update `src-tauri/tauri.conf.json`:

```json
{
  "build": {
    "beforeBuildCommand": "npm run build",
    "beforeDevCommand": "npm run dev",
    "devPath": "http://localhost:3000",
    "distDir": "../build"
  },
  "package": {
    "productName": "TheBoys Launcher",
    "version": "1.0.0"
  },
  "tauri": {
    "allowlist": {
      "all": false,
      "fs": {
        "all": true
      },
      "shell": {
        "all": true
      },
      "dialog": {
        "all": true
      },
      "notification": {
        "all": true
      }
    },
    "bundle": {
      "active": true,
      "targets": "all",
      "identifier": "com.theboys.launcher",
      "icon": [
        "icons/32x32.png",
        "icons/128x128.png",
        "icons/128x128@2x.png",
        "icons/icon.icns",
        "icons/icon.ico"
      ]
    },
    "security": {
      "csp": null
    },
    "windows": [
      {
        "fullscreen": false,
        "resizable": true,
        "title": "TheBoys Launcher",
        "width": 1200,
        "height": 800,
        "minWidth": 800,
        "minHeight": 600
      }
    ]
  }
}
```

### 5. Set Up Development Environment

#### VS Code Extensions:
- Rust Analyzer
- TypeScript and JavaScript Language Features
- ESLint
- Prettier
- Tauri

#### VS Code Settings (.vscode/settings.json):
```json
{
  "rust-analyzer.checkOnSave.command": "clippy",
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "typescript.preferences.importModuleSpecifier": "relative"
}
```

## Development Workflow

### 1. Start Development Server

```bash
# Start Tauri development mode
npm run tauri dev
```

This will:
- Start the React development server
- Launch the Tauri application
- Enable hot reload for both frontend and backend

### 2. Implement Slices Sequentially

Follow the `TAURI_MIGRATION_PLAN.md` document, implementing one slice at a time:

1. **Slice 1**: Project Foundation & Basic UI Shell
2. **Slice 2**: Settings Management System
3. **Slice 3**: Modpack Management Core
4. And so on...

### 3. Testing

```bash
# Run Rust tests
cargo test

# Run frontend tests
npm test

# Build for production
npm run tauri build
```

## Key Implementation Tips

### 1. Error Handling Pattern

Always use Result types in Rust:
```rust
use anyhow::{Result, Context};

pub fn some_function() -> Result<String> {
    let value = some_operation()
        .context("Failed to perform operation")?;
    Ok(value)
}
```

### 2. Tauri Command Pattern

```rust
#[tauri::command]
async fn command_name(param: String) -> Result<String, String> {
    // Validate input
    if param.is_empty() {
        return Err("Parameter cannot be empty".to_string());
    }
    
    // Perform operation
    let result = some_operation()
        .map_err(|e| format!("Operation failed: {}", e))?;
    
    Ok(result)
}
```

### 3. React Component Pattern

```typescript
import React, { useState, useEffect } from 'react';
import { invoke } from '@tauri-apps/api/tauri';

interface ComponentProps {
  // Define props
}

const Component: React.FC<ComponentProps> = ({ ...props }) => {
  const [state, setState] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  useEffect(() => {
    // Component initialization
  }, []);
  
  const handleAction = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await invoke('command_name', { param: 'value' });
      setState(result);
    } catch (err) {
      setError(err as string);
    } finally {
      setLoading(false);
    }
  };
  
  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;
  
  return (
    <div className="component">
      {/* Component JSX */}
    </div>
  );
};

export default Component;
```

### 4. State Management with Zustand

```typescript
import { create } from 'zustand';

interface AppState {
  // State
  modpacks: Modpack[];
  settings: Settings;
  loading: boolean;
  
  // Actions
  setModpacks: (modpacks: Modpack[]) => void;
  updateSettings: (settings: Partial<Settings>) => void;
  setLoading: (loading: boolean) => void;
}

export const useAppStore = create<AppState>((set, get) => ({
  // Initial state
  modpacks: [],
  settings: getDefaultSettings(),
  loading: false,
  
  // Actions
  setModpacks: (modpacks) => set({ modpacks }),
  updateSettings: (newSettings) => set((state) => ({ 
    settings: { ...state.settings, ...newSettings } 
  })),
  setLoading: (loading) => set({ loading }),
}));
```

## Common Pitfalls to Avoid

1. **Async/Await in Tauri Commands**: Always mark commands as `async` if they perform async operations
2. **Error Propagation**: Always handle errors properly and provide user-friendly messages
3. **File Paths**: Use proper path joining for cross-platform compatibility
4. **Permissions**: Ensure all required permissions are in `tauri.conf.json`
5. **State Management**: Avoid prop drilling by using proper state management

## Debugging Tips

1. **Rust Backend**: Use `println!` or `tracing` for debugging
2. **Frontend**: Use browser developer tools
3. **Tauri Logs**: Check `tauri://logs` in the app data directory
4. **Network Issues**: Use browser network tab to inspect HTTP requests

## Building for Distribution

```bash
# Build for current platform
npm run tauri build

# Build for specific platforms
npm run tauri build -- --target x86_64-pc-windows-msvc
npm run tauri build -- --target x86_64-apple-darwin
npm run tauri build -- --target x86_64-unknown-linux-gnu
```

## Next Steps

1. **Set up the project** using this guide
2. **Review the migration plan** in `TAURI_MIGRATION_PLAN.md`
3. **Start with Slice 1** - Project Foundation & Basic UI Shell
4. **Use the AI prompt** in `AI_IMPLEMENTATION_PROMPT.md` for each slice
5. **Test thoroughly** on all target platforms
6. **Iterate and polish** until each slice is complete

Remember: Quality is paramount. Don't rush through slices or leave incomplete functionality. Every feature must be production-ready before moving forward.

Good luck with your migration! The result will be a beautiful, cross-platform launcher that your friends will love to use.
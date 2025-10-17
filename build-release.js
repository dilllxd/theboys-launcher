#!/usr/bin/env node

/**
 * TheBoys Launcher - Build and Release Script
 *
 * This script automates the build and packaging process for all platforms
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const platforms = {
  'windows': ['x86_64'],
  'darwin': ['x86_64', 'aarch64'],
  'linux': ['x86_64']
};

function runCommand(command, description) {
  console.log(`\n🔨 ${description}...`);
  console.log(`⚡ Running: ${command}`);

  try {
    execSync(command, { stdio: 'inherit', cwd: process.cwd() });
    console.log(`✅ ${description} completed successfully`);
  } catch (error) {
    console.error(`❌ ${description} failed:`, error.message);
    process.exit(1);
  }
}

function ensureDirectoryExists(dirPath) {
  if (!fs.existsSync(dirPath)) {
    fs.mkdirSync(dirPath, { recursive: true });
    console.log(`📁 Created directory: ${dirPath}`);
  }
}

function createReleaseNotes(version) {
  const releaseNotes = `# TheBoys Launcher v${version}

## 🚀 Features
- Complete rewrite from Go to Rust + Tauri for better performance
- Modern UI with React and TypeScript
- Cross-platform support for Windows, macOS, and Linux
- Auto-updater functionality with background checks
- Improved modpack management system
- Enhanced Java and Prism Launcher integration

## 🐛 Bug Fixes
- Resolved installation issues on Windows
- Fixed memory allocation problems
- Improved error handling and logging

## 📦 Distribution
- Windows: MSI installer with proper shortcuts
- macOS: DMG package with drag-and-drop installation
- Linux: AppImage, DEB, and RPM packages

## 🔧 System Requirements
- Windows 10/11 (x64)
- macOS 11+ (Intel and Apple Silicon)
- Linux (x64) with libwebkit2gtk-4.1-0

## ⬇️ Installation
1. Download the appropriate package for your platform
2. Run the installer and follow the prompts
3. Launch TheBoys Launcher from your applications menu

---

**Important:** The launcher will automatically check for updates on startup.
`;

  const releaseNotesPath = path.join(process.cwd(), 'RELEASE_NOTES.md');
  fs.writeFileSync(releaseNotesPath, releaseNotes);
  console.log(`📝 Created release notes: ${releaseNotesPath}`);
}

function main() {
  console.log('🚀 TheBoys Launcher - Build and Release Script');
  console.log('==================================================');

  // Get version from package.json
  const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
  const version = packageJson.version;
  console.log(`📦 Building version: ${version}`);

  // Create release directories
  ensureDirectoryExists('dist');
  ensureDirectoryExists('releases');

  // Clean previous builds
  console.log('\n🧹 Cleaning previous builds...');
  if (fs.existsSync('src-tauri/target/release')) {
    runCommand('rm -rf src-tauri/target/release', 'Clean release directory');
  }

  // Install dependencies
  runCommand('npm ci', 'Install frontend dependencies');

  // Build frontend
  runCommand('npm run build', 'Build frontend application');

  // Create release notes
  createReleaseNotes(version);

  // Build for each platform
  for (const [platform, archs] of Object.entries(platforms)) {
    for (const arch of archs) {
      console.log(`\n🏗️  Building for ${platform} (${arch})...`);

      let target = '';
      switch (platform) {
        case 'windows':
          target = arch === 'x86_64' ? 'x86_64-pc-windows-msvc' : null;
          break;
        case 'darwin':
          target = arch === 'x86_64' ? 'x86_64-apple-darwin' : 'aarch64-apple-darwin';
          break;
        case 'linux':
          target = 'x86_64-unknown-linux-gnu';
          break;
      }

      if (target) {
        runCommand(
          `npm run tauri build -- --target ${target}`,
          `Build ${platform} binary (${arch})`
        );
      }
    }
  }

  // Create distribution packages
  console.log('\n📦 Creating distribution packages...');

  // Copy all built packages to releases directory
  const releaseDir = path.join(process.cwd(), 'src-tauri/target/release/bundle');
  if (fs.existsSync(releaseDir)) {
    runCommand(`cp -r ${releaseDir}/* releases/`, 'Copy packages to releases directory');
  }

  // Generate checksums
  console.log('\n🔐 Generating checksums...');
  const releasesPath = path.join(process.cwd(), 'releases');
  if (fs.existsSync(releasesPath)) {
    const checksumFile = path.join(releasesPath, 'checksums.txt');
    const checksums = [];

    // Walk through releases directory and generate checksums
    function walkDir(dir, fileList = []) {
      const files = fs.readdirSync(dir);
      files.forEach(file => {
        const filePath = path.join(dir, file);
        const stat = fs.statSync(filePath);
        if (stat.isDirectory()) {
          walkDir(filePath, fileList);
        } else if (file.endsWith('.exe') || file.endsWith('.dmg') || file.endsWith('.deb') || file.endsWith('.rpm') || file.endsWith('.AppImage')) {
          fileList.push(filePath);
        }
      });
      return fileList;
    }

    const packageFiles = walkDir(releasesPath);
    packageFiles.forEach(file => {
      const relativePath = path.relative(releasesPath, file);
      const checksum = execSync(`sha256sum "${file}"`, { encoding: 'utf8' }).trim();
      checksums.push(checksum);
    });

    fs.writeFileSync(checksumFile, checksums.join('\n'));
    console.log(`📝 Created checksums file: ${checksumFile}`);
  }

  console.log('\n🎉 Build process completed successfully!');
  console.log('📁 Check the "releases" directory for all built packages');
  console.log('📋 Release notes and checksums have been generated');
}

// Run the build process
if (require.main === module) {
  main();
}

module.exports = { runCommand, createReleaseNotes };
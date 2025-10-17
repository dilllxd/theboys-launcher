#!/usr/bin/env python3

"""
TheBoys Launcher Linux GUI Installer
A professional setup wizard for Linux installations
"""

import os
import sys
import shutil
import subprocess
import platform
from pathlib import Path

try:
    from PySide6.QtWidgets import (
        QApplication, QMainWindow, QWizard, QWizardPage, QVBoxLayout, QHBoxLayout,
        QLabel, QLineEdit, QPushButton, QCheckBox, QRadioButton, QButtonGroup,
        QProgressBar, QTextEdit, QFileDialog, QMessageBox, QFrame, QGroupBox,
        QGridLayout, QSpinBox, QComboBox
    )
    from PySide6.QtCore import Qt, QThread, Signal, QTimer
    from PySide6.QtGui import QFont, QPixmap, QIcon
except ImportError:
    print("Error: PySide6 is not installed.")
    print("Please install it with: pip install PySide6")
    sys.exit(1)

class InstallerThread(QThread):
    """Background thread for installation operations"""
    progress = Signal(int)
    log = Signal(str)
    finished = Signal(bool, str)

    def __init__(self, config):
        super().__init__()
        self.config = config
        self.should_stop = False

    def run(self):
        try:
            self.install()
        except Exception as e:
            self.finished.emit(False, str(e))

    def install(self):
        """Perform the installation"""
        config = self.config

        # Create directories
        self.log.emit("Creating directories...")
        install_dir = Path(config['install_dir'])
        install_dir.mkdir(parents=True, exist_ok=True)
        self.progress.emit(10)

        # Copy executable
        self.log.emit("Installing application files...")
        exe_source = Path(config['exe_source'])
        exe_dest = install_dir / 'theboys-launcher'

        if exe_source.exists():
            shutil.copy2(exe_source, exe_dest)
            exe_dest.chmod(0o755)
        else:
            raise FileNotFoundError(f"Executable not found: {exe_source}")

        self.progress.emit(30)

        # Create desktop entry
        self.log.emit("Creating desktop entry...")
        desktop_dir = Path.home() / '.local' / 'share' / 'applications'
        desktop_dir.mkdir(parents=True, exist_ok=True)

        desktop_entry = desktop_dir / 'theboys-launcher.desktop'
        desktop_content = f"""[Desktop Entry]
Version=1.0
Type=Application
Name=TheBoys Launcher
Comment=Modern Minecraft Modpack Launcher
Exec={exe_dest}
Icon=theboys-launcher
Terminal=false
Categories=Game;Utility;
StartupWMClass=TheBoys Launcher
"""

        with open(desktop_entry, 'w') as f:
            f.write(desktop_content)

        self.progress.emit(50)

        # Create user data directory
        self.log.emit("Setting up user data directory...")
        user_data_dir = Path.home() / '.theboys-launcher'
        user_data_dir.mkdir(parents=True, exist_ok=True)

        for subdir in ['instances', 'config', 'logs', 'prism', 'util']:
            (user_data_dir / subdir).mkdir(exist_ok=True)

        self.progress.emit(70)

        # Create icon directories and copy icon
        self.log.emit("Installing icons...")
        icon_sizes = [16, 22, 24, 32, 48, 64, 128, 256, 512]

        for size in icon_sizes:
            icon_dir = Path.home() / '.local' / 'share' / 'icons' / f'hicolor' / f'{size}x{size}' / 'apps'
            icon_dir.mkdir(parents=True, exist_ok=True)

            # For now, create a simple script that would copy icons if they exist
            # In a real installer, you would have actual icon files

        # Create symlink in user bin directory
        if config['create_symlink']:
            self.log.emit("Creating command-line symlink...")
            bin_dir = Path.home() / '.local' / 'bin'
            bin_dir.mkdir(parents=True, exist_ok=True)

            symlink_path = bin_dir / 'theboys-launcher'
            if symlink_path.exists() or symlink_path.is_symlink():
                symlink_path.unlink()
            symlink_path.symlink_to(exe_dest)

        self.progress.emit(90)

        # Update desktop database
        self.log.emit("Updating desktop database...")
        try:
            subprocess.run(['update-desktop-database', str(Path.home() / '.local' / 'share' / 'applications')],
                         check=False, capture_output=True)
        except:
            pass  # Not critical if this fails

        # Update icon cache
        try:
            subprocess.run(['gtk-update-icon-cache', '-q', '-t', '-f',
                           str(Path.home() / '.local' / 'share' / 'icons' / 'hicolor')],
                         check=False, capture_output=True)
        except:
            pass  # Not critical if this fails

        self.progress.emit(100)
        self.log.emit("Installation completed successfully!")
        self.finished.emit(True, "Installation completed successfully!")

class WelcomePage(QWizardPage):
    """Welcome page with introduction"""

    def __init__(self):
        super().__init__()
        self.setTitle("Welcome to TheBoys Launcher")
        self.setSubTitle("Modern Minecraft Modpack Launcher")

        layout = QVBoxLayout()

        # Welcome text
        welcome_label = QLabel("""
<h2>Welcome to TheBoys Launcher!</h2>

<p>This installer will guide you through the installation of TheBoys Launcher, a modern, cross-platform Minecraft modpack launcher.</p>

<p><b>Features:</b></p>
<ul>
<li>Modern graphical user interface</li>
<li>Automatic Java runtime management</li>
<li>Support for multiple modpack sources</li>
<li>Automatic updates and backups</li>
<li>Integration with Prism Launcher</li>
</ul>

<p>Your saved games and settings will be stored in your user home directory (<code>~/.theboys-launcher/</code>).</p>

<p>Click <b>Next</b> to continue with the installation.</p>
        """)
        welcome_label.setWordWrap(True)
        welcome_label.setTextFormat(Qt.RichText)
        layout.addWidget(welcome_label)

        self.setLayout(layout)

class LicensePage(QWizardPage):
    """License agreement page"""

    def __init__(self):
        super().__init__()
        self.setTitle("License Agreement")
        self.setSubTitle("Please read the license agreement")

        layout = QVBoxLayout()

        # License text
        license_text = QTextEdit()
        license_text.setPlainText("""
TheBoys Launcher - License Agreement

Copyright (c) 2024 TheBoys

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
        """)
        license_text.setReadOnly(True)
        layout.addWidget(license_text)

        # Accept checkbox
        self.accept_checkbox = QCheckBox("I accept the terms of the license agreement")
        self.accept_checkbox.stateChanged.connect(self.completeChanged)
        layout.addWidget(self.accept_checkbox)

        self.setLayout(layout)

    def isComplete(self):
        return self.accept_checkbox.isChecked()

class InstallationPage(QWizardPage):
    """Installation configuration page"""

    def __init__(self):
        super().__init__()
        self.setTitle("Installation Settings")
        self.setSubTitle("Configure your installation preferences")

        layout = QVBoxLayout()

        # Installation directory
        dir_group = QGroupBox("Installation Directory")
        dir_layout = QVBoxLayout()

        dir_label = QLabel("Choose where to install TheBoys Launcher:")
        dir_layout.addWidget(dir_label)

        dir_input_layout = QHBoxLayout()
        self.install_dir_edit = QLineEdit("/opt/theboys-launcher")
        self.install_dir_edit.textChanged.connect(self.completeChanged)

        browse_button = QPushButton("Browse...")
        browse_button.clicked.connect(self.browse_directory)

        dir_input_layout.addWidget(self.install_dir_edit)
        dir_input_layout.addWidget(browse_button)
        dir_layout.addLayout(dir_input_layout)

        dir_group.setLayout(dir_layout)
        layout.addWidget(dir_group)

        # Additional options
        options_group = QGroupBox("Additional Options")
        options_layout = QVBoxLayout()

        self.create_symlink_checkbox = QCheckBox("Create command-line symlink")
        self.create_symlink_checkbox.setChecked(True)
        self.create_symlink_checkbox.setToolTip("Creates a symlink in ~/.local/bin/ for command-line access")
        options_layout.addWidget(self.create_symlink_checkbox)

        self.create_desktop_checkbox = QCheckBox("Create desktop shortcut")
        self.create_desktop_checkbox.setChecked(True)
        self.create_desktop_checkbox.setToolTip("Adds a shortcut to your desktop")
        options_layout.addWidget(self.create_desktop_checkbox)

        options_group.setLayout(options_layout)
        layout.addWidget(options_group)

        layout.addStretch()
        self.setLayout(layout)

    def browse_directory(self):
        directory = QFileDialog.getExistingDirectory(self, "Select Installation Directory")
        if directory:
            self.install_dir_edit.setText(directory)

    def isComplete(self):
        install_dir = self.install_dir_edit.text().strip()
        if not install_dir:
            return False

        # Check if directory is writable
        try:
            path = Path(install_dir)
            if path.exists():
                return os.access(install_dir, os.W_OK)
            else:
                # Check if parent directory is writable
                return os.access(path.parent, os.W_OK)
        except:
            return False

    def get_config(self):
        return {
            'install_dir': self.install_dir_edit.text().strip(),
            'create_symlink': self.create_symlink_checkbox.isChecked(),
            'create_desktop': self.create_desktop_checkbox.isChecked()
        }

class InstallPage(QWizardPage):
    """Installation progress page"""

    def __init__(self):
        super().__init__()
        self.setTitle("Installing")
        self.setSubTitle("Please wait while TheBoys Launcher is being installed...")

        layout = QVBoxLayout()

        # Progress bar
        self.progress_bar = QProgressBar()
        self.progress_bar.setRange(0, 100)
        layout.addWidget(self.progress_bar)

        # Log output
        self.log_output = QTextEdit()
        self.log_output.setReadOnly(True)
        self.log_output.setMaximumHeight(200)
        layout.addWidget(self.log_output)

        self.setLayout(layout)

        self.install_thread = None

    def initializePage(self):
        """Start installation when page is shown"""
        # Get configuration
        config = self.field('installation_config')

        # Add executable source to config
        exe_source = self.field('exe_source')
        config['exe_source'] = exe_source

        # Start installation thread
        self.install_thread = InstallerThread(config)
        self.install_thread.progress.connect(self.progress_bar.setValue)
        self.install_thread.log.connect(self.add_log)
        self.install_thread.finished.connect(self.installation_finished)
        self.install_thread.start()

    def add_log(self, message):
        """Add message to log output"""
        self.log_output.append(message)

    def installation_finished(self, success, message):
        """Handle installation completion"""
        if success:
            self.add_log("Installation completed successfully!")
            self.wizard().next()
        else:
            self.add_log(f"Installation failed: {message}")
            QMessageBox.critical(self, "Installation Failed",
                                f"The installation failed: {message}\n\nPlease check the log output above for details.")

    def isComplete(self):
        return self.install_thread and not self.install_thread.isRunning()

class FinishPage(QWizardPage):
    """Installation completion page"""

    def __init__(self):
        super().__init__()
        self.setTitle("Installation Complete")
        self.setSubTitle("TheBoys Launcher has been successfully installed")

        layout = QVBoxLayout()

        # Success message
        success_label = QLabel("""
<h2>Installation Complete!</h2>

<p>TheBoys Launcher has been successfully installed on your system.</p>

<p><b>Next Steps:</b></p>
<ul>
<li>Launch TheBoys Launcher from your applications menu</li>
<li>Or run <code>theboys-launcher</code> from the terminal</li>
</ul>

<p><b>Important Information:</b></p>
<ul>
<li>Your user data is stored in: <code>~/.theboys-launcher/</code></li>
<li>This includes instances, settings, and downloaded files</li>
</ul>

<p>Thank you for choosing TheBoys Launcher!</p>
        """)
        success_label.setWordWrap(True)
        success_label.setTextFormat(Qt.RichText)
        layout.addWidget(success_label)

        # Launch checkbox
        self.launch_checkbox = QCheckBox("Launch TheBoys Launcher now")
        self.launch_checkbox.setChecked(True)
        layout.addWidget(self.launch_checkbox)

        layout.addStretch()
        self.setLayout(layout)

    def get_launch_app(self):
        return self.launch_checkbox.isChecked()

class LinuxInstallerWizard(QWizard):
    """Main installer wizard"""

    def __init__(self):
        super().__init__()
        self.setWindowTitle("TheBoys Launcher Setup")
        self.setWizardStyle(QWizard.ModernStyle)
        self.setOption(QWizard.HaveHelpButton, False)
        self.setMinimumSize(600, 500)

        # Set up pages
        self.addPage(WelcomePage())
        self.addPage(LicensePage())
        self.addPage(InstallationPage())
        self.addPage(InstallPage())
        self.addPage(FinishPage())

        # Register fields
        self.installation_page = self.page(2)
        self.install_page = self.installation_page
        self.registerField('installation_config', self.installation_page, 'config')
        self.registerField('exe_source', self, 'exe_source')

        self.exe_source = ""

    def setExecutableSource(self, exe_path):
        """Set the path to the executable to install"""
        self.exe_source = exe_path

def main():
    app = QApplication(sys.argv)
    app.setApplicationName("TheBoys Launcher Setup")
    app.setApplicationVersion("1.0.0")
    app.setOrganizationName("TheBoys")

    # Check if running as root for system-wide installation
    if os.geteuid() == 0:
        QMessageBox.warning(None, "Warning",
                          "You are running this installer as root. "
                          "It's recommended to run it as a normal user for user installation.")

    # Get executable path from command line
    exe_source = ""
    if len(sys.argv) > 1:
        exe_source = sys.argv[1]
    else:
        # Try to find the executable in the current directory
        current_dir = Path.cwd()
        for exe_name in ['theboys-launcher-linux-amd64', 'theboys-launcher-linux', 'theboys-launcher']:
            exe_path = current_dir / exe_name
            if exe_path.exists():
                exe_source = str(exe_path)
                break

    if not exe_source:
        QMessageBox.critical(None, "Error",
                           "Could not find TheBoys Launcher executable.\n\n"
                           "Please run this installer with the executable path as an argument:\n"
                           "python3 linux-setup.py /path/to/theboys-launcher")
        sys.exit(1)

    # Create and show wizard
    wizard = LinuxInstallerWizard()
    wizard.setExecutableSource(exe_source)

    if wizard.exec() == QWizard.Accepted:
        # Installation completed
        finish_page = wizard.page(4)
        if finish_page.get_launch_app():
            # Launch the application
            install_config = wizard.field('installation_config')
            exe_path = Path(install_config['install_dir']) / 'theboys-launcher'
            try:
                subprocess.Popen([str(exe_path)])
            except Exception as e:
                QMessageBox.warning(None, "Launch Failed",
                                  f"Failed to launch TheBoys Launcher: {e}")

if __name__ == '__main__':
    main()
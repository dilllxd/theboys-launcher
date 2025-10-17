import './style.css';
import './app.css';

// Import backend functions
import {
  GetModpacks,
  GetInstances,
  GetSettings,
  GetJavaInstallations,
  SelectModpack,
  CreateInstance,
  LaunchInstance,
  DeleteInstance,
  RefreshModpacks,
  GetBestJavaInstallation,
  GetJavaVersionForMinecraft,
  DownloadJava,
  IsPrismInstalled,
  InstallModpackWithPackwiz,
  CheckModpackUpdate,
  UpdateModpack,
  GetLWJGLVersionForMinecraft,
  CheckForUpdates,
  DownloadUpdate,
  InstallUpdate,
  GetVersion
} from '../wailsjs/go/main/App';

// State management
let currentView = 'modpacks';
let modpacks = [];
let instances = [];
let settings = {};
let javaInstallations = [];
let currentTheme = 'dark';
let downloadProgress = {};
let activeOperations = new Map();
let currentVersion = 'dev';

// Initialize the application
async function initApp() {
  console.log('Initializing TheBoys Launcher...');

  // Load version
  try {
    currentVersion = await GetVersion();
    console.log(`TheBoys Launcher v${currentVersion}`);
  } catch (error) {
    console.warn('Failed to get version:', error);
  }

  // Load theme and settings
  loadTheme();

  // Load initial data
  await loadAllData();

  // Render the main layout
  renderLayout();

  // Show default view
  showView('modpacks');

  // Set up periodic refresh
  setInterval(updateStatusBar, 5000);

  // Check for updates on startup
  checkForUpdates();
}

// Load all data from backend
async function loadAllData() {
  try {
    const [modpacksData, instancesData, settingsData, javaData] = await Promise.all([
      GetModpacks(),
      GetInstances(),
      GetSettings(),
      GetJavaInstallations()
    ]);

    modpacks = modpacksData || [];
    instances = instancesData || [];
    settings = settingsData || {};
    javaInstallations = javaData || [];

    console.log('Data loaded:', {
      modpacks: modpacks.length,
      instances: instances.length,
      javaInstallations: javaInstallations.length
    });
  } catch (error) {
    console.error('Failed to load data:', error);
    showError('Failed to load application data');
  }
}

// Render the main layout
function renderLayout() {
  const app = document.getElementById('app');
  app.innerHTML = `
    <div class="app-container">
      <!-- Progress Container -->
      <div id="progress-container" class="progress-container"></div>

      <!-- Header -->
      <header class="app-header">
        <div class="header-content">
          <div class="logo-section">
            <div class="logo"></div>
            <div class="title-section">
              <h1 class="app-title">TheBoys Launcher</h1>
              <span class="app-version">v${currentVersion}</span>
            </div>
          </div>
          <div class="header-actions">
            <button class="btn btn-secondary" onclick="refreshModpacks()" title="Refresh Modpacks">
              <span class="icon">↻</span>
            </button>
            <button class="btn btn-secondary" onclick="toggleTheme()" title="Toggle Theme">
              <span class="icon">${currentTheme === 'dark' ? '☀' : '🌙'}</span>
            </button>
            <button class="btn btn-secondary" onclick="showSettings()" title="Settings">
              <span class="icon">⚙</span>
            </button>
          </div>
        </div>
      </header>

      <!-- Main Content Area -->
      <div class="app-body">
        <!-- Navigation Sidebar -->
        <nav class="sidebar">
          <div class="nav-section">
            <div class="nav-section-title">Main</div>
            <button class="nav-item ${currentView === 'modpacks' ? 'active' : ''}"
                    onclick="showView('modpacks')" data-view="modpacks">
              <span class="nav-icon">📦</span>
              <span class="nav-text">Modpacks</span>
            </button>
            <button class="nav-item ${currentView === 'instances' ? 'active' : ''}"
                    onclick="showView('instances')" data-view="instances">
              <span class="nav-icon">🎮</span>
              <span class="nav-text">Instances</span>
            </button>
          </div>

          <div class="nav-section">
            <div class="nav-section-title">Management</div>
            <button class="nav-item ${currentView === 'downloads' ? 'active' : ''}"
                    onclick="showView('downloads')" data-view="downloads">
              <span class="nav-icon">⬇</span>
              <span class="nav-text">Downloads</span>
            </button>
            <button class="nav-item ${currentView === 'java' ? 'active' : ''}"
                    onclick="showView('java')" data-view="java">
              <span class="nav-icon">☕</span>
              <span class="nav-text">Java</span>
            </button>
            <button class="nav-item ${currentView === 'logs' ? 'active' : ''}"
                    onclick="showView('logs')" data-view="logs">
              <span class="nav-icon">📋</span>
              <span class="nav-text">Logs</span>
            </button>
          </div>
        </nav>

        <!-- Main Content -->
        <main class="main-content">
          <div id="view-container" class="view-container">
            <!-- Dynamic content will be loaded here -->
          </div>
        </main>
      </div>

      <!-- Status Bar -->
      <footer class="status-bar">
        <div class="status-left">
          <span class="status-item">
            <span class="status-label">Modpacks:</span>
            <span class="status-value" id="modpack-count">${modpacks.length}</span>
          </span>
          <span class="status-item">
            <span class="status-label">Instances:</span>
            <span class="status-value" id="instance-count">${instances.length}</span>
          </span>
          <span class="status-item">
            <span class="status-label">Java:</span>
            <span class="status-value" id="java-count">${javaInstallations.length}</span>
          </span>
        </div>
        <div class="status-right">
          <span class="status-item" id="status-text">Ready</span>
          <span class="status-item" id="current-time"></span>
        </div>
      </footer>
    </div>
  `;
}

// Show different views
function showView(viewName) {
  currentView = viewName;

  // Update navigation active state
  document.querySelectorAll('.nav-item').forEach(item => {
    item.classList.toggle('active', item.dataset.view === viewName);
  });

  // Clear current view
  const container = document.getElementById('view-container');

  // Render appropriate view
  switch (viewName) {
    case 'modpacks':
      renderModpacksView(container);
      break;
    case 'instances':
      renderInstancesView(container);
      break;
    case 'downloads':
      renderDownloadsView(container);
      break;
    case 'java':
      renderJavaView(container);
      break;
    case 'logs':
      renderLogsView(container);
      break;
    case 'settings':
      renderSettingsView(container);
      break;
    default:
      renderModpacksView(container);
  }

  updateStatusBar();
}

// Render modpacks view
function renderModpacksView(container) {
  container.innerHTML = `
    <div class="view-header">
      <h2>Available Modpacks</h2>
      <div class="view-actions">
        <div class="search-box">
          <input type="text" placeholder="Search modpacks..." id="modpack-search"
                 onkeyup="filterModpacks()" />
        </div>
      </div>
    </div>

    <div class="modpack-grid" id="modpack-grid">
      ${modpacks.length === 0 ?
        '<div class="empty-state"><p>No modpacks available. Click refresh to update.</p></div>' :
        modpacks.map(modpack => `
          <div class="modpack-card" data-modpack-id="${modpack.id}">
            <div class="modpack-header">
              <h3 class="modpack-name">${modpack.displayName}</h3>
              <span class="modpack-version">${modpack.version || 'Latest'}</span>
            </div>
            <div class="modpack-description">
              <p>${modpack.description || 'No description available'}</p>
            </div>
            <div class="modpack-info">
              <span class="info-item">🎮 ${modpack.minecraftVersion || 'Unknown'}</span>
              <span class="info-item">🔧 ${modpack.modLoader || 'Unknown'}</span>
            </div>
            <div class="modpack-actions">
              <button class="btn btn-primary" onclick="selectModpack('${modpack.id}')">
                Select & Install
              </button>
            </div>
          </div>
        `).join('')
      }
    </div>
  `;
}

// Render instances view
function renderInstancesView(container) {
  container.innerHTML = `
    <div class="view-header">
      <h2>Minecraft Instances</h2>
      <div class="view-actions">
        <button class="btn btn-secondary" onclick="refreshInstances()">Refresh</button>
      </div>
    </div>

    <div class="instance-list" id="instance-list">
      ${instances.length === 0 ?
        '<div class="empty-state"><p>No instances created yet. Select a modpack to get started.</p></div>' :
        instances.map(instance => `
          <div class="instance-card" data-instance-id="${instance.id}">
            <div class="instance-header">
              <h3 class="instance-name">${instance.name}</h3>
              <div class="instance-status ${instance.lastPlayed ? 'played' : 'new'}">
                ${instance.lastPlayed ? 'Played' : 'New'}
              </div>
            </div>
            <div class="instance-info">
              <span class="info-item">🎮 ${instance.minecraft}</span>
              <span class="info-item">🔧 ${instance.modLoader}</span>
              <span class="info-item">☕ ${instance.javaVersion}</span>
            </div>
            <div class="instance-actions">
              <button class="btn btn-primary" onclick="launchInstance('${instance.id}')">
                Launch
              </button>
              <button class="btn btn-secondary" onclick="manageInstance('${instance.id}')">
                Manage
              </button>
              <button class="btn btn-danger" onclick="deleteInstance('${instance.id}')">
                Delete
              </button>
            </div>
          </div>
        `).join('')
      }
    </div>
  `;
}

// Render downloads view
function renderDownloadsView(container) {
  container.innerHTML = `
    <div class="view-header">
      <h2>Downloads</h2>
    </div>
    <div class="empty-state">
      <p>No active downloads.</p>
    </div>
  `;
}

// Render Java view
function renderJavaView(container) {
  container.innerHTML = `
    <div class="view-header">
      <h2>Java Installations</h2>
      <div class="view-actions">
        <button class="btn btn-secondary" onclick="detectJava()">Detect Java</button>
      </div>
    </div>

    <div class="java-list">
      ${javaInstallations.length === 0 ?
        '<div class="empty-state"><p>No Java installations detected. The launcher will download Java as needed.</p></div>' :
        javaInstallations.map(java => `
          <div class="java-card">
            <div class="java-info">
              <h4>Java ${java.version}</h4>
              <p class="java-path">${java.path}</p>
            </div>
            <div class="java-meta">
              <span class="java-type">${java.isJDK ? 'JDK' : 'JRE'}</span>
              <span class="java-arch">${java.architecture || 'Unknown'}</span>
            </div>
          </div>
        `).join('')
      }
    </div>
  `;
}

// Render logs view
function renderLogsView(container) {
  container.innerHTML = `
    <div class="view-header">
      <h2>Application Logs</h2>
      <div class="view-actions">
        <button class="btn btn-secondary" onclick="refreshLogs()">Refresh</button>
        <button class="btn btn-secondary" onclick="clearLogs()">Clear</button>
      </div>
    </div>
    <div class="logs-container">
      <div class="empty-state">
        <p>Log viewer will be implemented here.</p>
      </div>
    </div>
  `;
}

// Render settings view
function renderSettingsView(container) {
  container.innerHTML = `
    <div class="view-header">
      <h2>Settings</h2>
      <div class="view-actions">
        <button class="btn btn-primary" onclick="saveSettings()">Save Settings</button>
        <button class="btn btn-secondary" onclick="resetSettings()">Reset to Defaults</button>
      </div>
    </div>

    <div class="settings-container">
      <!-- General Settings -->
      <div class="settings-section">
        <h3 class="section-title">General</h3>
        <div class="settings-grid">
          <div class="setting-item">
            <label class="setting-label">Theme</label>
            <select class="setting-input" id="theme-select" onchange="updateThemeFromSettings()">
              <option value="dark" ${currentTheme === 'dark' ? 'selected' : ''}>Dark</option>
              <option value="light" ${currentTheme === 'light' ? 'selected' : ''}>Light</option>
            </select>
            <span class="setting-description">Choose the application theme</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Auto-check Updates</label>
            <div class="toggle-switch">
              <input type="checkbox" id="auto-updates" ${settings.autoUpdates !== false ? 'checked' : ''}>
              <label for="auto-updates" class="toggle-label"></label>
            </div>
            <span class="setting-description">Automatically check for launcher updates</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Keep Launcher Open</label>
            <div class="toggle-switch">
              <input type="checkbox" id="keep-open" ${settings.keepLauncherOpen !== false ? 'checked' : ''}>
              <label for="keep-open" class="toggle-label"></label>
            </div>
            <span class="setting-description">Keep launcher open after launching Minecraft</span>
          </div>
        </div>
      </div>

      <!-- Java Settings -->
      <div class="settings-section">
        <h3 class="section-title">Java</h3>
        <div class="settings-grid">
          <div class="setting-item">
            <label class="setting-label">Memory Allocation</label>
            <div class="memory-presets">
              <button class="preset-btn ${settings.memoryMB <= 2048 ? 'active' : ''}"
                      onclick="applyMemoryPreset(1024, 2048)" data-preset="low">
                Low (1-2GB)
              </button>
              <button class="preset-btn ${settings.memoryMB > 2048 && settings.memoryMB <= 4096 ? 'active' : ''}"
                      onclick="applyMemoryPreset(2048, 4096)" data-preset="medium">
                Medium (2-4GB)
              </button>
              <button class="preset-btn ${settings.memoryMB > 4096 && settings.memoryMB <= 8192 ? 'active' : ''}"
                      onclick="applyMemoryPreset(4096, 8192)" data-preset="high">
                High (4-8GB)
              </button>
              <button class="preset-btn ${settings.memoryMB > 8192 ? 'active' : ''}"
                      onclick="applyMemoryPreset(6144, 12288)" data-preset="ultra">
                Ultra (6-12GB)
              </button>
              <button class="preset-btn ${settings.autoDetectMemory ? 'active' : ''}"
                      onclick="autoDetectMemory()" id="auto-detect-btn">
                Auto-Detect
              </button>
            </div>
            <div class="memory-inputs">
              <input type="number" id="min-memory" min="512" max="32768" step="512"
                     value="${settings.minMemory || 2048}" class="memory-input">
              <span class="memory-separator">-</span>
              <input type="number" id="max-memory" min="512" max="32768" step="512"
                     value="${settings.maxMemory || 4096}" class="memory-input">
              <span class="memory-unit">MB</span>
            </div>
            <span class="setting-description">Memory allocation for Minecraft. Click presets or enter custom values.</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Auto-detect Java</label>
            <div class="toggle-switch">
              <input type="checkbox" id="auto-detect-java" ${settings.autoDetectJava !== false ? 'checked' : ''}>
              <label for="auto-detect-java" class="toggle-label"></label>
            </div>
            <span class="setting-description">Automatically detect and install Java versions</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Additional JVM Arguments</label>
            <textarea id="jvm-args" class="setting-textarea" rows="3"
                      placeholder="e.g., -XX:+UseG1GC -XX:+UnlockExperimentalVMOptions">${settings.jvmArgs || ''}</textarea>
            <span class="setting-description">Additional JVM arguments for Minecraft</span>
          </div>
        </div>
      </div>

      <!-- Instance Settings -->
      <div class="settings-section">
        <h3 class="section-title">Instances</h3>
        <div class="settings-grid">
          <div class="setting-item">
            <label class="setting-label">Default Instance Directory</label>
            <div class="path-input">
              <input type="text" id="instance-dir" class="setting-input"
                     value="${settings.instanceDirectory || ''}"
                     placeholder="Default instance directory">
              <button class="btn btn-secondary" onclick="browseInstanceDirectory()">Browse</button>
            </div>
            <span class="setting-description">Default location for Minecraft instances</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Create Backups Before Updates</label>
            <div class="toggle-switch">
              <input type="checkbox" id="auto-backup" ${settings.autoBackup !== false ? 'checked' : ''}>
              <label for="auto-backup" class="toggle-label"></label>
            </div>
            <span class="setting-description">Automatically create backups before modpack updates</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Max Backup Retention</label>
            <select class="setting-input" id="backup-retention">
              <option value="3" ${settings.maxBackups === 3 ? 'selected' : ''}>3 backups</option>
              <option value="5" ${settings.maxBackups === 5 ? 'selected' : ''}>5 backups</option>
              <option value="10" ${settings.maxBackups === 10 ? 'selected' : ''}>10 backups</option>
              <option value="unlimited" ${settings.maxBackups === 'unlimited' ? 'selected' : ''}>Unlimited</option>
            </select>
            <span class="setting-description">Maximum number of backups to keep</span>
          </div>
        </div>
      </div>

      <!-- Network Settings -->
      <div class="settings-section">
        <h3 class="section-title">Network</h3>
        <div class="settings-grid">
          <div class="setting-item">
            <label class="setting-label">Download Timeout (seconds)</label>
            <input type="number" id="download-timeout" min="30" max="300" step="10"
                   value="${settings.downloadTimeout || 60}" class="setting-input">
            <span class="setting-description">Timeout for download operations</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Max Concurrent Downloads</label>
            <input type="number" id="max-downloads" min="1" max="10" step="1"
                   value="${settings.maxConcurrentDownloads || 3}" class="setting-input">
            <span class="setting-description">Maximum number of concurrent downloads</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Use Proxy</label>
            <div class="toggle-switch">
              <input type="checkbox" id="use-proxy" ${settings.useProxy ? 'checked' : ''}>
              <label for="use-proxy" class="toggle-label"></label>
            </div>
            <span class="setting-description">Use proxy for network operations</span>
          </div>

          <div class="setting-item" id="proxy-settings" style="${settings.useProxy ? '' : 'display: none;'}">
            <label class="setting-label">Proxy URL</label>
            <input type="text" id="proxy-url" class="setting-input"
                   value="${settings.proxyUrl || ''}"
                   placeholder="http://proxy.example.com:8080">
            <span class="setting-description">Proxy server URL (if enabled)</span>
          </div>
        </div>
      </div>

      <!-- Advanced Settings -->
      <div class="settings-section">
        <h3 class="section-title">Advanced</h3>
        <div class="settings-grid">
          <div class="setting-item">
            <label class="setting-label">Debug Mode</label>
            <div class="toggle-switch">
              <input type="checkbox" id="debug-mode" ${settings.debugMode ? 'checked' : ''}>
              <label for="debug-mode" class="toggle-label"></label>
            </div>
            <span class="setting-description">Enable debug logging and additional information</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Pause on Error</label>
            <div class="toggle-switch">
              <input type="checkbox" id="pause-on-error" ${settings.pauseOnError !== false ? 'checked' : ''}>
              <label for="pause-on-error" class="toggle-label"></label>
            </div>
            <span class="setting-description">Pause execution when errors occur</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Clear Cache</label>
            <button class="btn btn-secondary" onclick="clearCache()">Clear Cache</button>
            <span class="setting-description">Clear downloaded files and temporary data</span>
          </div>

          <div class="setting-item">
            <label class="setting-label">Reset Configuration</label>
            <button class="btn btn-danger" onclick="confirmResetConfiguration()">Reset All Settings</button>
            <span class="setting-description">Reset all settings to default values</span>
          </div>
        </div>
      </div>
    </div>
  `;

  // Setup event listeners
  setupSettingsListeners();
}

// Action functions
async function refreshModpacks() {
  updateStatus('Refreshing modpacks...');
  try {
    await RefreshModpacks();
    await loadAllData();
    showView('modpacks');
    updateStatus('Modpacks refreshed successfully');
  } catch (error) {
    console.error('Failed to refresh modpacks:', error);
    showError('Failed to refresh modpacks');
  }
}

async function selectModpack(modpackId) {
  updateStatus(`Selecting modpack ${modpackId}...`);
  try {
    const modpack = await SelectModpack(modpackId);
    if (modpack) {
      await createInstanceForModpack(modpack);
    }
  } catch (error) {
    console.error('Failed to select modpack:', error);
    showError('Failed to select modpack');
  }
}

async function createInstanceForModpack(modpack) {
  updateStatus(`Creating instance for ${modpack.displayName}...`);
  try {
    const instance = await CreateInstance(modpack);
    if (instance) {
      await loadAllData();
      showView('instances');
      updateStatus(`Instance "${instance.name}" created successfully`);
    }
  } catch (error) {
    console.error('Failed to create instance:', error);
    showError('Failed to create instance');
  }
}

async function launchInstance(instanceId) {
  updateStatus(`Launching instance ${instanceId}...`);
  try {
    await LaunchInstance(instanceId);
    updateStatus('Instance launched successfully');
  } catch (error) {
    console.error('Failed to launch instance:', error);
    showError('Failed to launch instance');
  }
}

async function deleteInstance(instanceId) {
  if (!confirm('Are you sure you want to delete this instance? This action cannot be undone.')) {
    return;
  }

  updateStatus(`Deleting instance ${instanceId}...`);
  try {
    await DeleteInstance(instanceId);
    await loadAllData();
    showView('instances');
    updateStatus('Instance deleted successfully');
  } catch (error) {
    console.error('Failed to delete instance:', error);
    showError('Failed to delete instance');
  }
}

function showSettings() {
  showView('settings');
}

function filterModpacks() {
  const searchTerm = document.getElementById('modpack-search').value.toLowerCase();
  const cards = document.querySelectorAll('.modpack-card');

  cards.forEach(card => {
    const name = card.querySelector('.modpack-name').textContent.toLowerCase();
    const description = card.querySelector('.modpack-description p').textContent.toLowerCase();

    if (name.includes(searchTerm) || description.includes(searchTerm)) {
      card.style.display = 'block';
    } else {
      card.style.display = 'none';
    }
  });
}

// Theme management
function loadTheme() {
  // Load theme from settings or localStorage
  const savedTheme = settings.theme || localStorage.getItem('launcher-theme') || 'dark';
  currentTheme = savedTheme;
  applyTheme(currentTheme);
}

function applyTheme(theme) {
  document.body.classList.remove('theme-dark', 'theme-light');
  document.body.classList.add(`theme-${theme}`);
  currentTheme = theme;
}

function toggleTheme() {
  const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
  applyTheme(newTheme);
  localStorage.setItem('launcher-theme', newTheme);

  // Update settings
  if (settings) {
    settings.theme = newTheme;
    saveSettings();
  }
}

// Progress tracking
function startOperation(operationId, operationName) {
  activeOperations.set(operationId, {
    name: operationName,
    progress: 0,
    startTime: Date.now()
  });
}

function updateProgress(operationId, progress) {
  if (activeOperations.has(operationId)) {
    const operation = activeOperations.get(operationId);
    operation.progress = progress;
    downloadProgress[operationId] = progress;
    renderProgressIndicators();
  }
}

function completeOperation(operationId) {
  if (activeOperations.has(operationId)) {
    activeOperations.delete(operationId);
    delete downloadProgress[operationId];
    renderProgressIndicators();
  }
}

function renderProgressIndicators() {
  const container = document.getElementById('progress-container');
  if (!container) return;

  if (activeOperations.size === 0) {
    container.innerHTML = '';
    return;
  }

  let html = '<div class="progress-indicators">';
  for (const [id, operation] of activeOperations) {
    html += `
      <div class="progress-item">
        <div class="progress-info">
          <span class="progress-name">${operation.name}</span>
          <span class="progress-percentage">${Math.round(operation.progress)}%</span>
        </div>
        <div class="progress-bar">
          <div class="progress-fill" style="width: ${operation.progress}%"></div>
        </div>
      </div>
    `;
  }
  html += '</div>';
  container.innerHTML = html;
}


// Update checking
async function checkForUpdates() {
  try {
    const updateInfo = await CheckForUpdates();
    if (updateInfo && updateInfo.updateAvailable) {
      showUpdateNotification(updateInfo);
    }
  } catch (error) {
    console.warn('Failed to check for updates:', error);
  }
}

function showUpdateNotification(updateInfo) {
  const notification = document.createElement('div');
  notification.className = 'update-notification';
  notification.innerHTML = `
    <div class="update-content">
      <h4>Update Available</h4>
      <p>A new version (${updateInfo.latestVersion}) is available. You are currently on v${currentVersion}.</p>
      <div class="update-actions">
        <button class="btn btn-primary" onclick="downloadUpdate('${updateInfo.downloadUrl}')">Download Update</button>
        <button class="btn btn-secondary" onclick="dismissUpdate()">Skip</button>
      </div>
    </div>
  `;

  document.body.appendChild(notification);

  // Auto-dismiss after 10 seconds
  setTimeout(() => {
    if (notification.parentNode) {
      notification.remove();
    }
  }, 10000);
}

function dismissUpdate() {
  const notification = document.querySelector('.update-notification');
  if (notification) {
    notification.remove();
  }
}

async function downloadUpdate(downloadUrl) {
  try {
    updateStatus('Downloading update...');
    startOperation('update-download', 'Downloading Update');

    const updatePath = await DownloadUpdate(downloadUrl, (progress) => {
      updateProgress('update-download', progress * 100);
    });

    completeOperation('update-download');
    updateStatus('Update downloaded. Installing...');

    // Install update
    await InstallUpdate(updatePath);

    updateStatus('Update installed successfully. Restart may be required.');
  } catch (error) {
    completeOperation('update-download');
    console.error('Failed to download/update:', error);
    showError('Failed to download update');
  }
}

// Utility functions
function updateStatus(message) {
  document.getElementById('status-text').textContent = message;
  console.log('Status:', message);
}

function showError(message) {
  updateStatus(`Error: ${message}`);
  showNotification(message, 'error');
}

function showNotification(message, type = 'info') {
  const notification = document.createElement('div');
  notification.className = `notification notification-${type}`;
  notification.innerHTML = `
    <div class="notification-content">
      <span class="notification-message">${message}</span>
      <button class="notification-close" onclick="this.parentElement.parentElement.remove()">×</button>
    </div>
  `;

  document.body.appendChild(notification);

  // Auto-dismiss after 5 seconds for info, 10 seconds for errors
  const timeout = type === 'error' ? 10000 : 5000;
  setTimeout(() => {
    if (notification.parentNode) {
      notification.remove();
    }
  }, timeout);
}

function updateStatusBar() {
  // Update counts
  document.getElementById('modpack-count').textContent = modpacks.length;
  document.getElementById('instance-count').textContent = instances.length;
  document.getElementById('java-count').textContent = javaInstallations.length;

  // Update time
  const now = new Date();
  document.getElementById('current-time').textContent = now.toLocaleTimeString();
}

// Placeholder functions
async function refreshInstances() {
  await loadAllData();
  showView('instances');
}

async function detectJava() {
  updateStatus('Detecting Java installations...');
  await loadAllData();
  showView('java');
  updateStatus('Java detection complete');
}

function refreshLogs() {
  updateStatus('Refreshing logs...');
}

function clearLogs() {
  updateStatus('Logs cleared');
}

function manageInstance(instanceId) {
  updateStatus(`Managing instance ${instanceId}`);
  // Could open instance management dialog
}

// Settings management functions
function setupSettingsListeners() {
  // Proxy toggle visibility
  const useProxyCheckbox = document.getElementById('use-proxy');
  const proxySettings = document.getElementById('proxy-settings');

  if (useProxyCheckbox && proxySettings) {
    useProxyCheckbox.addEventListener('change', (e) => {
      proxySettings.style.display = e.target.checked ? 'block' : 'none';
    });
  }

  // Memory validation
  const minMemoryInput = document.getElementById('min-memory');
  const maxMemoryInput = document.getElementById('max-memory');

  if (minMemoryInput && maxMemoryInput) {
    minMemoryInput.addEventListener('change', validateMemorySettings);
    maxMemoryInput.addEventListener('change', validateMemorySettings);
  }
}

function updateThemeFromSettings() {
  const themeSelect = document.getElementById('theme-select');
  if (themeSelect) {
    applyTheme(themeSelect.value);
  }
}

function validateMemorySettings() {
  const minMemory = parseInt(document.getElementById('min-memory').value);
  const maxMemory = parseInt(document.getElementById('max-memory').value);

  if (minMemory > maxMemory) {
    document.getElementById('min-memory').value = maxMemory;
    showNotification('Minimum memory cannot exceed maximum memory', 'error');
  }

  // Validate ranges
  if (minMemory < 512) {
    document.getElementById('min-memory').value = 512;
    showNotification('Minimum memory must be at least 512MB', 'error');
  }

  if (maxMemory > 32768) {
    document.getElementById('max-memory').value = 32768;
    showNotification('Maximum memory cannot exceed 32GB', 'error');
  }
}

async function saveSettings() {
  updateStatus('Saving settings...');

  // Collect settings from form
  const newSettings = {
    theme: document.getElementById('theme-select').value,
    autoUpdates: document.getElementById('auto-updates').checked,
    keepLauncherOpen: document.getElementById('keep-open').checked,
    minMemory: parseInt(document.getElementById('min-memory').value),
    maxMemory: parseInt(document.getElementById('max-memory').value),
    autoDetectJava: document.getElementById('auto-detect-java').checked,
    jvmArgs: document.getElementById('jvm-args').value.trim(),
    instanceDirectory: document.getElementById('instance-dir').value.trim(),
    autoBackup: document.getElementById('auto-backup').checked,
    maxBackups: document.getElementById('backup-retention').value,
    downloadTimeout: parseInt(document.getElementById('download-timeout').value),
    maxConcurrentDownloads: parseInt(document.getElementById('max-downloads').value),
    useProxy: document.getElementById('use-proxy').checked,
    proxyUrl: document.getElementById('proxy-url').value.trim(),
    debugMode: document.getElementById('debug-mode').checked,
    pauseOnError: document.getElementById('pause-on-error').checked,
  };

  // Validate settings
  if (newSettings.minMemory > newSettings.maxMemory) {
    showError('Minimum memory cannot exceed maximum memory');
    return;
  }

  if (newSettings.useProxy && !newSettings.proxyUrl) {
    showError('Proxy URL is required when proxy is enabled');
    return;
  }

  try {
    // Save to backend (would need backend function)
    settings = { ...settings, ...newSettings };
    console.log('Settings saved:', settings);

    showNotification('Settings saved successfully', 'success');
    updateStatus('Settings saved');
  } catch (error) {
    console.error('Failed to save settings:', error);
    showError('Failed to save settings');
  }
}

async function resetSettings() {
  if (confirm('Are you sure you want to reset all settings to their default values? This action cannot be undone.')) {
    try {
      // Reset to defaults
      settings = {
        theme: 'dark',
        autoUpdates: true,
        keepLauncherOpen: true,
        minMemory: 2048,
        maxMemory: 4096,
        autoDetectJava: true,
        jvmArgs: '',
        instanceDirectory: '',
        autoBackup: true,
        maxBackups: 5,
        downloadTimeout: 60,
        maxConcurrentDownloads: 3,
        useProxy: false,
        proxyUrl: '',
        debugMode: false,
        pauseOnError: true,
      };

      // Apply theme
      applyTheme(settings.theme);

      // Re-render settings view
      showView('settings');

      showNotification('Settings reset to defaults', 'success');
    } catch (error) {
      console.error('Failed to reset settings:', error);
      showError('Failed to reset settings');
    }
  }
}

function browseInstanceDirectory() {
  // In a real implementation, this would open a directory picker
  // For now, just show a placeholder
  showNotification('Directory browser not yet implemented', 'info');
}

async function clearCache() {
  if (confirm('Are you sure you want to clear all cached files? This will remove downloaded files and temporary data.')) {
    try {
      // Would call backend function to clear cache
      updateStatus('Clearing cache...');

      // Simulate cache clearing
      await new Promise(resolve => setTimeout(resolve, 2000));

      showNotification('Cache cleared successfully', 'success');
      updateStatus('Cache cleared');
    } catch (error) {
      console.error('Failed to clear cache:', error);
      showError('Failed to clear cache');
    }
  }
}

function confirmResetConfiguration() {
  if (confirm('WARNING: This will reset ALL settings to their default values and cannot be undone. Are you absolutely sure?')) {
    if (confirm('This is your final warning. All configuration will be lost. Continue?')) {
      resetSettings();
    }
  }
}

// Memory management functions
function applyMemoryPreset(minMB, maxMB) {
  document.getElementById('min-memory').value = minMB;
  document.getElementById('max-memory').value = maxMB;

  // Update active preset button
  document.querySelectorAll('.preset-btn').forEach(btn => {
    btn.classList.remove('active');
  });

  const activeBtn = document.querySelector(`[onclick="applyMemoryPreset(${minMB}, ${maxMB})"]`);
  if (activeBtn) {
    activeBtn.classList.add('active');
  }

  // Clear auto-detect button
  document.getElementById('auto-detect-btn').classList.remove('active');

  showNotification(`Memory preset applied: ${minMB}MB - ${maxMB}MB`, 'success');
}

async function autoDetectMemory() {
  updateStatus('Detecting optimal memory settings...');
  startOperation('memory-detect', 'Detecting Memory');

  try {
    // This would call a backend function to auto-detect memory
    // For now, simulate the detection process
    await new Promise(resolve => setTimeout(resolve, 1500));

    // Simulate detection based on available memory
    const totalMemory = navigator.deviceMemory ? navigator.deviceMemory * 1024 : 8192; // Convert GB to MB

    let minMB, maxMB;
    if (totalMemory >= 32768) { // 32GB+
      minMB = 4096;
      maxMB = 12288;
    } else if (totalMemory >= 16384) { // 16GB+
      minMB = 3072;
      maxMB = 8192;
    } else if (totalMemory >= 8192) { // 8GB+
      minMB = 2048;
      maxMB = 6144;
    } else if (totalMemory >= 4096) { // 4GB+
      minMB = 1024;
      maxMB = 3072;
    } else {
      minMB = 512;
      maxMB = 2048;
    }

    document.getElementById('min-memory').value = minMB;
    document.getElementById('max-memory').value = maxMB;

    // Update active preset button
    document.querySelectorAll('.preset-btn').forEach(btn => {
      btn.classList.remove('active');
    });
    document.getElementById('auto-detect-btn').classList.add('active');

    completeOperation('memory-detect');
    showNotification(`Auto-detected optimal memory: ${minMB}MB - ${maxMB}MB`, 'success');
    updateStatus('Memory detection completed');

  } catch (error) {
    completeOperation('memory-detect');
    console.error('Memory detection failed:', error);
    showError('Failed to auto-detect memory settings');
  }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', initApp);

// Handle window resize for responsive layout
window.addEventListener('resize', () => {
  // Could implement responsive adjustments here
});
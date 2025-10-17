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
  RefreshModpacks
} from '../wailsjs/go/main/App';

// State management
let currentView = 'modpacks';
let modpacks = [];
let instances = [];
let settings = {};
let javaInstallations = [];

// Initialize the application
async function initApp() {
  console.log('Initializing TheBoys Launcher...');

  // Load initial data
  await loadAllData();

  // Render the main layout
  renderLayout();

  // Show default view
  showView('modpacks');

  // Set up periodic refresh
  setInterval(updateStatusBar, 5000);
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
      <!-- Header -->
      <header class="app-header">
        <div class="header-content">
          <div class="logo-section">
            <div class="logo"></div>
            <h1 class="app-title">TheBoys Launcher</h1>
          </div>
          <div class="header-actions">
            <button class="btn btn-secondary" onclick="refreshModpacks()" title="Refresh Modpacks">
              <span class="icon">↻</span>
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
    </div>
    <div class="settings-container">
      <div class="empty-state">
        <p>Settings panel will be implemented here.</p>
      </div>
    </div>
  `;
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

// Utility functions
function updateStatus(message) {
  document.getElementById('status-text').textContent = message;
  console.log('Status:', message);
}

function showError(message) {
  updateStatus(`Error: ${message}`);
  // Could implement a toast notification here
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

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', initApp);

// Handle window resize for responsive layout
window.addEventListener('resize', () => {
  // Could implement responsive adjustments here
});
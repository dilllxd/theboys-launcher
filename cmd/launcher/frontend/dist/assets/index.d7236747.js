(function(){const a=document.createElement("link").relList;if(a&&a.supports&&a.supports("modulepreload"))return;for(const t of document.querySelectorAll('link[rel="modulepreload"]'))l(t);new MutationObserver(t=>{for(const n of t)if(n.type==="childList")for(const r of n.addedNodes)r.tagName==="LINK"&&r.rel==="modulepreload"&&l(r)}).observe(document,{childList:!0,subtree:!0});function e(t){const n={};return t.integrity&&(n.integrity=t.integrity),t.referrerpolicy&&(n.referrerPolicy=t.referrerpolicy),t.crossorigin==="use-credentials"?n.credentials="include":t.crossorigin==="anonymous"?n.credentials="omit":n.credentials="same-origin",n}function l(t){if(t.ep)return;t.ep=!0;const n=e(t);fetch(t.href,n)}})();function u(){return window.go.main.App.GetModpacks()}function m(){return window.go.main.App.GetInstances()}function h(){return window.go.main.App.GetSettings()}function b(){return window.go.main.App.GetJavaInstallations()}let i="modpacks",c=[],o=[],w={},d=[];async function g(){console.log("Initializing TheBoys Launcher..."),await f(),k(),y("modpacks"),setInterval(p,5e3)}async function f(){try{const[s,a,e,l]=await Promise.all([u(),m(),h(),b()]);c=s||[],o=a||[],w=e||{},d=l||[],console.log("Data loaded:",{modpacks:c.length,instances:o.length,javaInstallations:d.length})}catch(s){console.error("Failed to load data:",s),D("Failed to load application data")}}function k(){const s=document.getElementById("app");s.innerHTML=`
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
              <span class="icon">\u21BB</span>
            </button>
            <button class="btn btn-secondary" onclick="showSettings()" title="Settings">
              <span class="icon">\u2699</span>
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
            <button class="nav-item ${i==="modpacks"?"active":""}"
                    onclick="showView('modpacks')" data-view="modpacks">
              <span class="nav-icon">\u{1F4E6}</span>
              <span class="nav-text">Modpacks</span>
            </button>
            <button class="nav-item ${i==="instances"?"active":""}"
                    onclick="showView('instances')" data-view="instances">
              <span class="nav-icon">\u{1F3AE}</span>
              <span class="nav-text">Instances</span>
            </button>
          </div>

          <div class="nav-section">
            <div class="nav-section-title">Management</div>
            <button class="nav-item ${i==="downloads"?"active":""}"
                    onclick="showView('downloads')" data-view="downloads">
              <span class="nav-icon">\u2B07</span>
              <span class="nav-text">Downloads</span>
            </button>
            <button class="nav-item ${i==="java"?"active":""}"
                    onclick="showView('java')" data-view="java">
              <span class="nav-icon">\u2615</span>
              <span class="nav-text">Java</span>
            </button>
            <button class="nav-item ${i==="logs"?"active":""}"
                    onclick="showView('logs')" data-view="logs">
              <span class="nav-icon">\u{1F4CB}</span>
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
            <span class="status-value" id="modpack-count">${c.length}</span>
          </span>
          <span class="status-item">
            <span class="status-label">Instances:</span>
            <span class="status-value" id="instance-count">${o.length}</span>
          </span>
          <span class="status-item">
            <span class="status-label">Java:</span>
            <span class="status-value" id="java-count">${d.length}</span>
          </span>
        </div>
        <div class="status-right">
          <span class="status-item" id="status-text">Ready</span>
          <span class="status-item" id="current-time"></span>
        </div>
      </footer>
    </div>
  `}function y(s){i=s,document.querySelectorAll(".nav-item").forEach(e=>{e.classList.toggle("active",e.dataset.view===s)});const a=document.getElementById("view-container");switch(s){case"modpacks":v(a);break;case"instances":$(a);break;case"downloads":L(a);break;case"java":I(a);break;case"logs":M(a);break;case"settings":S(a);break;default:v(a)}p()}function v(s){s.innerHTML=`
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
      ${c.length===0?'<div class="empty-state"><p>No modpacks available. Click refresh to update.</p></div>':c.map(a=>`
          <div class="modpack-card" data-modpack-id="${a.id}">
            <div class="modpack-header">
              <h3 class="modpack-name">${a.displayName}</h3>
              <span class="modpack-version">${a.version||"Latest"}</span>
            </div>
            <div class="modpack-description">
              <p>${a.description||"No description available"}</p>
            </div>
            <div class="modpack-info">
              <span class="info-item">\u{1F3AE} ${a.minecraftVersion||"Unknown"}</span>
              <span class="info-item">\u{1F527} ${a.modLoader||"Unknown"}</span>
            </div>
            <div class="modpack-actions">
              <button class="btn btn-primary" onclick="selectModpack('${a.id}')">
                Select & Install
              </button>
            </div>
          </div>
        `).join("")}
    </div>
  `}function $(s){s.innerHTML=`
    <div class="view-header">
      <h2>Minecraft Instances</h2>
      <div class="view-actions">
        <button class="btn btn-secondary" onclick="refreshInstances()">Refresh</button>
      </div>
    </div>

    <div class="instance-list" id="instance-list">
      ${o.length===0?'<div class="empty-state"><p>No instances created yet. Select a modpack to get started.</p></div>':o.map(a=>`
          <div class="instance-card" data-instance-id="${a.id}">
            <div class="instance-header">
              <h3 class="instance-name">${a.name}</h3>
              <div class="instance-status ${a.lastPlayed?"played":"new"}">
                ${a.lastPlayed?"Played":"New"}
              </div>
            </div>
            <div class="instance-info">
              <span class="info-item">\u{1F3AE} ${a.minecraft}</span>
              <span class="info-item">\u{1F527} ${a.modLoader}</span>
              <span class="info-item">\u2615 ${a.javaVersion}</span>
            </div>
            <div class="instance-actions">
              <button class="btn btn-primary" onclick="launchInstance('${a.id}')">
                Launch
              </button>
              <button class="btn btn-secondary" onclick="manageInstance('${a.id}')">
                Manage
              </button>
              <button class="btn btn-danger" onclick="deleteInstance('${a.id}')">
                Delete
              </button>
            </div>
          </div>
        `).join("")}
    </div>
  `}function L(s){s.innerHTML=`
    <div class="view-header">
      <h2>Downloads</h2>
    </div>
    <div class="empty-state">
      <p>No active downloads.</p>
    </div>
  `}function I(s){s.innerHTML=`
    <div class="view-header">
      <h2>Java Installations</h2>
      <div class="view-actions">
        <button class="btn btn-secondary" onclick="detectJava()">Detect Java</button>
      </div>
    </div>

    <div class="java-list">
      ${d.length===0?'<div class="empty-state"><p>No Java installations detected. The launcher will download Java as needed.</p></div>':d.map(a=>`
          <div class="java-card">
            <div class="java-info">
              <h4>Java ${a.version}</h4>
              <p class="java-path">${a.path}</p>
            </div>
            <div class="java-meta">
              <span class="java-type">${a.isJDK?"JDK":"JRE"}</span>
              <span class="java-arch">${a.architecture||"Unknown"}</span>
            </div>
          </div>
        `).join("")}
    </div>
  `}function M(s){s.innerHTML=`
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
  `}function S(s){s.innerHTML=`
    <div class="view-header">
      <h2>Settings</h2>
    </div>
    <div class="settings-container">
      <div class="empty-state">
        <p>Settings panel will be implemented here.</p>
      </div>
    </div>
  `}function j(s){document.getElementById("status-text").textContent=s,console.log("Status:",s)}function D(s){j(`Error: ${s}`)}function p(){document.getElementById("modpack-count").textContent=c.length,document.getElementById("instance-count").textContent=o.length,document.getElementById("java-count").textContent=d.length;const s=new Date;document.getElementById("current-time").textContent=s.toLocaleTimeString()}document.addEventListener("DOMContentLoaded",g);window.addEventListener("resize",()=>{});

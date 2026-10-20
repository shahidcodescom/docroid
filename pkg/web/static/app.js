// Docroid Web Dashboard Client Logic — Zoho Desk, CRM & Mail Inspired Interface

// Reusable Feather / Lucide Style SVG Line Icons
const Icons = {
  refresh: `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67"/></svg>`,
  folder: `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/></svg>`,
  zap: `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>`,
  rocket: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4.5 16.5c-1.5 1.26-2 5-2 5s3.74-.5 5-2c.71-.84.7-2.13-.09-2.91a2.18 2.18 0 0 0-2.91-.09z"/><path d="M12 15l-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z"/><path d="M9 12H4s.55-3.03 2-4c1.62-1.08 5 0 5 0"/><path d="M12 9V4s3.03.55 4 2c1.08 1.62 0 5 0 5"/></svg>`,
  package: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>`,
  trash: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>`,
  stop: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="5" width="14" height="14" rx="2"/></svg>`,
  play: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"/></svg>`,
  globe: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>`,
  cluster: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/></svg>`,
  phone: `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>`,
  monitor: `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>`,
  laptop: `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="12" rx="2"/><path d="M2 20h20"/></svg>`,
  smartphone: `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"/><line x1="12" y1="18" x2="12.01" y2="18"/></svg>`,
  file: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>`,
  fileCode: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>`,
  fileImage: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>`,
};

function getFileIconSvg(name, isDir) {
  if (isDir) return Icons.folder;
  const lower = (name || '').toLowerCase();
  if (lower.endsWith('.html') || lower.endsWith('.htm')) return Icons.globe;
  if (lower.endsWith('.css') || lower.endsWith('.js') || lower.endsWith('.json') || lower.endsWith('.sh') || lower.endsWith('.py') || lower.endsWith('.go')) return Icons.fileCode;
  if (lower.endsWith('.png') || lower.endsWith('.jpg') || lower.endsWith('.jpeg') || lower.endsWith('.svg') || lower.endsWith('.webp') || lower.endsWith('.gif') || lower.endsWith('.ico')) return Icons.fileImage;
  return Icons.file;
}

let apps = [];
let sites = [];
let networks = [];
let currentFilter = 'all';
let currentSiteFilter = 'all';
let currentNetFilter = 'all';
let currentViewMode = 'cards'; // 'cards' | 'table'
let currentWorkspace = 'apps'; // 'apps' | 'sites'
let searchKeyword = '';
let activeDrawerApp = null;
let logPollTimer = null;
let autoScrollLogs = true;

// Immediate theme execution to prevent flash
(function() {
  const savedTheme = localStorage.getItem('docroid-theme') || 
    (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark');
  document.documentElement.setAttribute('data-theme', savedTheme);
})();

function initTheme() {
  const current = document.documentElement.getAttribute('data-theme') || 'dark';
  applyTheme(current, false);

  const btnToggle = document.getElementById('btnThemeToggle');
  if (btnToggle) {
    btnToggle.addEventListener('click', toggleTheme);
  }
}

function applyTheme(theme, notify = true) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('docroid-theme', theme);

  const sunIcon = document.getElementById('themeIconSun');
  const moonIcon = document.getElementById('themeIconMoon');
  const btnToggle = document.getElementById('btnThemeToggle');

  if (theme === 'light') {
    if (sunIcon) sunIcon.style.display = 'none';
    if (moonIcon) moonIcon.style.display = 'block';
    if (btnToggle) btnToggle.setAttribute('title', 'Switch to Dark Theme');
  } else {
    if (sunIcon) sunIcon.style.display = 'block';
    if (moonIcon) moonIcon.style.display = 'none';
    if (btnToggle) btnToggle.setAttribute('title', 'Switch to Light Theme');
  }

  if (notify && typeof showToast === 'function') {
    showToast(`Switched to ${theme} theme`, 'info');
  }
}

function toggleTheme() {
  const current = document.documentElement.getAttribute('data-theme') || 'dark';
  const target = current === 'dark' ? 'light' : 'dark';
  applyTheme(target, true);
}

document.addEventListener('DOMContentLoaded', () => {
  initTheme();
  initEventListeners();
  fetchApps();
  fetchSites();
  fetchNetworks();
  initDocs();
  handleUrlHash();
  window.addEventListener('hashchange', handleUrlHash);
  // Poll every 1.5 seconds for live status & telemetry
  setInterval(() => {
    fetchApps();
    fetchSites();
    if (currentWorkspace === 'networks') {
      fetchNetworks();
    }
  }, 1500);
});

function initEventListeners() {
  // Mobile Sidebar & Backdrop Toggle
  const sidebar = document.getElementById('appSidebar');
  const sidebarToggle = document.getElementById('sidebarToggle');
  const sidebarCloseBtn = document.getElementById('sidebarCloseBtn');
  const sidebarBackdrop = document.getElementById('sidebarBackdrop');

  function openMobileSidebar() {
    if (sidebar) sidebar.classList.add('mobile-open');
    if (sidebarBackdrop) sidebarBackdrop.classList.add('active');
  }

  function closeMobileSidebar() {
    if (sidebar) sidebar.classList.remove('mobile-open');
    if (sidebarBackdrop) sidebarBackdrop.classList.remove('active');
  }

  if (sidebarToggle) {
    sidebarToggle.addEventListener('click', openMobileSidebar);
  }
  if (sidebarCloseBtn) {
    sidebarCloseBtn.addEventListener('click', closeMobileSidebar);
  }
  if (sidebarBackdrop) {
    sidebarBackdrop.addEventListener('click', closeMobileSidebar);
  }

  // Auto-close sidebar on mobile navigation
  if (sidebar) {
    const navLinks = sidebar.querySelectorAll('.nav-item');
    navLinks.forEach(link => {
      link.addEventListener('click', () => {
        if (window.innerWidth <= 768) {
          closeMobileSidebar();
        }
      });
    });
  }

  // Mobile Bottom Navigation Bar Links
  const bnavApps = document.getElementById('bnavApps');
  if (bnavApps) {
    bnavApps.addEventListener('click', () => {
      const navApps = document.getElementById('navApps');
      if (navApps) setActiveNavItem(navApps);
      switchWorkspaceView('apps');
      currentFilter = 'all';
      updateFilterButtons();
      renderApps();
    });
  }

  const bnavSites = document.getElementById('bnavSites');
  if (bnavSites) {
    bnavSites.addEventListener('click', () => {
      const navSites = document.getElementById('navSites');
      if (navSites) setActiveNavItem(navSites);
      switchWorkspaceView('sites');
      currentSiteFilter = 'all';
      updateSiteFilterButtons();
      fetchSites();
    });
  }

  const bnavCluster = document.getElementById('bnavCluster');
  if (bnavCluster) {
    bnavCluster.addEventListener('click', () => {
      const navCluster = document.getElementById('navCluster');
      if (navCluster) setActiveNavItem(navCluster);
      switchWorkspaceView('cluster');
      currentPodFilter = 'all';
      if (typeof updatePodFilterButtons === 'function') updatePodFilterButtons();
      if (typeof fetchClusterData === 'function') fetchClusterData();
    });
  }

  const bnavNetworks = document.getElementById('bnavNetworks');
  if (bnavNetworks) {
    bnavNetworks.addEventListener('click', () => {
      const navNetworks = document.getElementById('navNetworks');
      if (navNetworks) setActiveNavItem(navNetworks);
      switchWorkspaceView('networks');
      if (typeof fetchNetworks === 'function') fetchNetworks();
    });
  }

  const bnavMore = document.getElementById('bnavMore');
  if (bnavMore) {
    bnavMore.addEventListener('click', openMobileSidebar);
  }

  // Mobile Search Input synchronization
  const mobileSearchInput = document.getElementById('mobileSearchInput');
  const searchInput = document.getElementById('searchInput');
  if (mobileSearchInput && searchInput) {
    mobileSearchInput.addEventListener('input', (e) => {
      searchInput.value = e.target.value;
      searchKeyword = e.target.value.toLowerCase().trim();
      if (currentWorkspace === 'sites') {
        renderSites();
      } else {
        renderApps();
      }
    });
    searchInput.addEventListener('input', (e) => {
      if (mobileSearchInput.value !== e.target.value) {
        mobileSearchInput.value = e.target.value;
      }
    });
  }

  // Sidebar Nav Links
  const navApps = document.getElementById('navApps');
  if (navApps) {
    navApps.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navApps);
      switchWorkspaceView('apps');
      currentFilter = 'all';
      updateFilterButtons();
      renderApps();
    });
  }

  const navSites = document.getElementById('navSites');
  if (navSites) {
    navSites.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navSites);
      switchWorkspaceView('sites');
      currentSiteFilter = 'all';
      updateSiteFilterButtons();
      fetchSites();
    });
  }

  const navCluster = document.getElementById('navCluster');
  if (navCluster) {
    navCluster.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navCluster);
      switchWorkspaceView('cluster');
      currentPodFilter = 'all';
      updatePodFilterButtons();
      fetchClusterData();
    });
  }

  const navNetworks = document.getElementById('navNetworks');
  if (navNetworks) {
    navNetworks.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navNetworks);
      switchWorkspaceView('networks');
      currentNetFilter = 'all';
      updateNetFilterButtons();
      fetchNetworks();
    });
  }

  const navTunnels = document.getElementById('navTunnels');
  if (navTunnels) {
    navTunnels.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navTunnels);
      switchWorkspaceView('tunnels');
      currentTunnelFilter = 'all';
      updateTunnelFilterButtons();
      fetchTunnels();
    });
  }

  const navDocs = document.getElementById('navDocs');
  if (navDocs) {
    navDocs.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navDocs);
      switchWorkspaceView('docs');
      window.location.hash = 'docs';
    });
  }

  const navCompose = document.getElementById('navCompose');
  if (navCompose) {
    navCompose.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navCompose);
      const deployModal = document.getElementById('deployModal');
      deployModal.classList.add('active');
    });
  }

  const navSecrets = document.getElementById('navSecrets');
  if (navSecrets) {
    navSecrets.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navSecrets);
      document.getElementById('secretsModal').classList.add('active');
      loadSecrets();
    });
  }

  const navPackages = document.getElementById('navPackages');
  if (navPackages) {
    navPackages.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navPackages);
      document.getElementById('importModal').classList.add('active');
    });
  }

  const navDoctor = document.getElementById('navDoctor');
  if (navDoctor) {
    navDoctor.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navDoctor);
      document.getElementById('doctorModal').classList.add('active');
      loadDoctorDiagnostics();
    });
  }

  const navConsole = document.getElementById('navConsole');
  if (navConsole) {
    navConsole.addEventListener('click', (e) => {
      e.preventDefault();
      setActiveNavItem(navConsole);
      if (apps.length > 0) {
        openDetailDrawer(apps[0].app.id, apps[0].app.name, 'logs');
      } else {
        showToast('No active applications to stream logs', 'info');
      }
    });
  }

  // View Switcher (Cards vs Table)
  const viewCardsBtn = document.getElementById('viewModeCards');
  const viewTableBtn = document.getElementById('viewModeTable');
  const cardsContainer = document.getElementById('appsContainer');
  const tableContainer = document.getElementById('appsTableContainer');

  if (viewCardsBtn && viewTableBtn) {
    viewCardsBtn.addEventListener('click', () => {
      currentViewMode = 'cards';
      viewCardsBtn.classList.add('active');
      viewTableBtn.classList.remove('active');
      cardsContainer.style.display = 'grid';
      tableContainer.style.display = 'none';
      const sitesGrid = document.getElementById('sitesContainer');
      const sitesTable = document.getElementById('sitesTableContainer');
      if (sitesGrid) sitesGrid.style.display = 'grid';
      if (sitesTable) sitesTable.style.display = 'none';
      renderApps();
      renderSites();
    });

    viewTableBtn.addEventListener('click', () => {
      currentViewMode = 'table';
      viewTableBtn.classList.add('active');
      viewCardsBtn.classList.remove('active');
      cardsContainer.style.display = 'none';
      tableContainer.style.display = 'block';
      const sitesGrid = document.getElementById('sitesContainer');
      const sitesTable = document.getElementById('sitesTableContainer');
      if (sitesGrid) sitesGrid.style.display = 'none';
      if (sitesTable) sitesTable.style.display = 'block';
      renderApps();
      renderSites();
    });
  }

  // Filter Buttons
  document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      currentFilter = btn.dataset.filter;
      renderApps();
    });
  });

  // Search Input & Global Shortcut
  if (searchInput) {
    searchInput.addEventListener('input', (e) => {
      searchKeyword = e.target.value.toLowerCase().trim();
      if (currentWorkspace === 'sites') {
        renderSites();
      } else {
        renderApps();
      }
    });

    window.addEventListener('keydown', (e) => {
      if (e.key === '/' && document.activeElement !== searchInput && !['INPUT', 'TEXTAREA'].includes(document.activeElement.tagName)) {
        e.preventDefault();
        searchInput.focus();
      }
    });
  }

  // Refresh Button
  document.getElementById('btnRefresh').addEventListener('click', fetchApps);

  // Deploy Modal
  const deployModal = document.getElementById('deployModal');
  document.getElementById('btnDeployModal').addEventListener('click', () => {
    deployModal.classList.add('active');
  });
  document.getElementById('closeDeployModal').addEventListener('click', () => {
    deployModal.classList.remove('active');
  });
  document.getElementById('cancelDeployModal').addEventListener('click', () => {
    deployModal.classList.remove('active');
  });

  // Engine selection changes
  document.getElementById('appEngine').addEventListener('change', (e) => {
    const rootfsGroup = document.getElementById('rootfsSelectGroup');
    if (e.target.value === 'rootfs') {
      rootfsGroup.style.display = 'block';
    } else {
      rootfsGroup.style.display = 'none';
    }
  });

  // Deploy Form Submission
  document.getElementById('deployForm').addEventListener('submit', handleDeploySubmit);

  // Compose Upload & Auto-Parse
  const composeFileInput = document.getElementById('composeFileInput');
  if (composeFileInput) {
    composeFileInput.addEventListener('change', (e) => {
      if (e.target.files && e.target.files[0]) {
        handleComposeFile(e.target.files[0]);
      }
    });
  }

  const dropZone = document.getElementById('composeDropZone');
  if (dropZone) {
    dropZone.addEventListener('dragover', (e) => {
      e.preventDefault();
      dropZone.classList.add('drag-over');
    });
    dropZone.addEventListener('dragleave', () => {
      dropZone.classList.remove('drag-over');
    });
    dropZone.addEventListener('drop', (e) => {
      e.preventDefault();
      dropZone.classList.remove('drag-over');
      if (e.dataTransfer.files && e.dataTransfer.files[0]) {
        handleComposeFile(e.dataTransfer.files[0]);
      }
    });
  }

  // Compose Service Select
  const composeServiceSelect = document.getElementById('composeServiceSelect');
  if (composeServiceSelect) {
    composeServiceSelect.addEventListener('change', (e) => {
      const idx = parseInt(e.target.value, 10);
      if (uploadedComposeConfig && uploadedComposeConfig.services && uploadedComposeConfig.services[idx]) {
        populateFormFromService(uploadedComposeConfig.services[idx]);
      }
    });
  }

  // Download Compose YAML
  const btnDownloadCompose = document.getElementById('btnDownloadCompose');
  if (btnDownloadCompose) {
    btnDownloadCompose.addEventListener('click', handleDownloadCompose);
  }

  // Deploy All Services (Compose Up)
  const btnDeployAllCompose = document.getElementById('btnDeployAllCompose');
  if (btnDeployAllCompose) {
    btnDeployAllCompose.addEventListener('click', handleDeployAllCompose);
  }

  // Images & Registry Modal
  const importModal = document.getElementById('importModal');
  const openImagesModal = () => {
    if (importModal) {
      importModal.classList.add('active');
      loadImagesList();
    }
  };
  const btnImport = document.getElementById('btnImportModal');
  if (btnImport) btnImport.addEventListener('click', openImagesModal);

  const closeImport = document.getElementById('closeImportModal');
  if (closeImport) closeImport.addEventListener('click', () => importModal.classList.remove('active'));

  const cancelImport = document.getElementById('cancelImportModal');
  if (cancelImport) cancelImport.addEventListener('click', () => importModal.classList.remove('active'));

  const importForm = document.getElementById('importForm');
  if (importForm) importForm.addEventListener('submit', handleImportSubmit);

  const btnRefreshImages = document.getElementById('btnRefreshImages');
  if (btnRefreshImages) btnRefreshImages.addEventListener('click', loadImagesList);

  // Tab switching for Images Modal
  document.querySelectorAll('.img-modal-tab').forEach(tab => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('.img-modal-tab').forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      const targetId = tab.dataset.target;
      document.querySelectorAll('.img-tab-panel').forEach(p => p.style.display = 'none');
      const targetPanel = document.getElementById(targetId);
      if (targetPanel) targetPanel.style.display = 'block';
      if (targetId === 'tabImagesList') loadImagesList();
    });
  });

  // Pull presets
  document.querySelectorAll('.preset-pull-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const pullInput = document.getElementById('pullImageRef');
      if (pullInput) pullInput.value = btn.dataset.ref;
    });
  });

  // Pull, Build, Push forms
  const pullForm = document.getElementById('pullImageForm');
  if (pullForm) pullForm.addEventListener('submit', handlePullImageSubmit);

  const buildForm = document.getElementById('buildImageForm');
  if (buildForm) buildForm.addEventListener('submit', handleBuildImageSubmit);

  const pushForm = document.getElementById('pushImageForm');
  if (pushForm) pushForm.addEventListener('submit', handlePushImageSubmit);

  // Zoho Desk Slide-Over Detail Drawer & Tabs
  document.getElementById('closeLogsDrawer').addEventListener('click', closeDetailDrawer);
  document.getElementById('btnRefreshLogs').addEventListener('click', refreshLogs);
  document.getElementById('autoScrollCheck').addEventListener('change', (e) => {
    autoScrollLogs = e.target.checked;
  });

  document.querySelectorAll('.drawer-tab').forEach(tabBtn => {
    tabBtn.addEventListener('click', () => {
      const tabName = tabBtn.dataset.tab;
      switchDrawerTab(tabName);
    });
  });

  // Doctor Modal
  const doctorModal = document.getElementById('doctorModal');
  document.getElementById('btnDoctorModal').addEventListener('click', () => {
    doctorModal.classList.add('active');
    loadDoctorDiagnostics();
  });
  document.getElementById('closeDoctorModal').addEventListener('click', () => {
    doctorModal.classList.remove('active');
  });
  document.getElementById('closeDoctorBtn').addEventListener('click', () => {
    doctorModal.classList.remove('active');
  });
  document.getElementById('btnRerunDoctor').addEventListener('click', loadDoctorDiagnostics);

  // Secrets Modal
  const secretsModal = document.getElementById('secretsModal');
  document.getElementById('btnSecretsModal').addEventListener('click', () => {
    secretsModal.classList.add('active');
    loadSecrets();
  });
  document.getElementById('closeSecretsModal').addEventListener('click', () => {
    secretsModal.classList.remove('active');
  });
  document.getElementById('closeSecretsBtn').addEventListener('click', () => {
    secretsModal.classList.remove('active');
  });
  document.getElementById('addSecretForm').addEventListener('submit', handleAddSecret);

  // Update Modal
  const updateModal = document.getElementById('updateModal');
  document.getElementById('closeUpdateModal').addEventListener('click', () => {
    updateModal.classList.remove('active');
  });
  document.getElementById('cancelUpdateModal').addEventListener('click', () => {
    updateModal.classList.remove('active');
  });
  document.getElementById('updateForm').addEventListener('submit', handleUpdateSubmit);
}

function setActiveNavItem(element) {
  document.querySelectorAll('.nav-item').forEach(item => item.classList.remove('active'));
  element.classList.add('active');
}

function updateFilterButtons() {
  document.querySelectorAll('.filter-btn').forEach(btn => {
    if (btn.dataset.filter === currentFilter) {
      btn.classList.add('active');
    } else {
      btn.classList.remove('active');
    }
  });
}

// Fetch apps from REST API
async function fetchApps() {
  try {
    const res = await fetch('/api/apps');
    if (!res.ok) throw new Error('API error: ' + res.status);
    const data = await res.json();
    apps = data || [];
    updateTopStats();
    renderApps();

    // If detail drawer is open, keep its overview/specs fresh
    if (activeDrawerApp) {
      const liveTarget = apps.find(item => item.app.id === activeDrawerApp.id);
      if (liveTarget) {
        updateDrawerContent(liveTarget);
      }
    }
  } catch (err) {
    const statusText = document.getElementById('liveStatusText');
    if (statusText) {
      statusText.textContent = 'Engine Offline';
      statusText.parentElement.style.borderColor = '#f43f5e';
    }
    console.error('Failed to fetch apps:', err);
  }
}

// Update Top Metric Cards
function updateTopStats() {
  let runningCount = 0;
  let totalMem = 0;
  let totalCpu = 0;

  apps.forEach(item => {
    const a = item.app;
    const st = item.stats;
    if (a.status === 'running') {
      runningCount++;
      if (st) {
        totalMem += (st.MemoryRSS || 0);
        totalCpu += (st.CPUPercent || 0);
      }
    }
  });

  document.getElementById('statRunning').textContent = runningCount;
  document.getElementById('statTotal').textContent = `${apps.length} total registered`;
  document.getElementById('statCpu').textContent = `${totalCpu.toFixed(1)}%`;
  document.getElementById('statMem').textContent = formatBytes(totalMem);

  document.getElementById('countAll').textContent = apps.length;
  document.getElementById('countRunning').textContent = runningCount;
  document.getElementById('countStopped').textContent = apps.length - runningCount;
  document.getElementById('appsCountBadge').textContent = `${apps.length} apps`;

  const navBadge = document.getElementById('navAppsBadge');
  if (navBadge) navBadge.textContent = apps.length;
}

// Get filtered app list
function getFilteredApps() {
  return apps.filter(item => {
    const a = item.app;
    if (currentFilter === 'running' && a.status !== 'running') return false;
    if (currentFilter === 'stopped' && a.status === 'running') return false;
    if (searchKeyword) {
      const matchName = (a.name || '').toLowerCase().includes(searchKeyword);
      const matchId = (a.id || '').toLowerCase().includes(searchKeyword);
      const matchCmd = (a.command || '').toLowerCase().includes(searchKeyword);
      const matchPort = a.port ? a.port.toString().includes(searchKeyword) : false;
      return matchName || matchId || matchCmd || matchPort;
    }
    return true;
  });
}

// Render applications (Cards Grid or Table View)
function renderApps() {
  const filtered = getFilteredApps();

  if (currentViewMode === 'table') {
    renderAppsTable(filtered);
  } else {
    renderAppsCards(filtered);
  }
}

// Render Cards Grid
function renderAppsCards(filtered) {
  const container = document.getElementById('appsContainer');
  
  if (filtered.length === 0) {
    container.innerHTML = `
      <div class="loading-placeholder">
        <p>No applications match current filters.</p>
        <button class="btn btn-primary btn-sm" style="margin-top:12px;" onclick="document.getElementById('deployModal').classList.add('active')">
          Deploy First App
        </button>
      </div>
    `;
    return;
  }

  container.innerHTML = filtered.map(item => {
    const a = item.app;
    const st = item.stats || {};
    const isRunning = a.status === 'running';

    let memStr = '-';
    let cpuStr = '-';

    if (isRunning && st.Alive) {
      memStr = formatBytes(st.MemoryRSS || 0);
      cpuStr = (st.CPUPercent || 0).toFixed(1) + '%';
    }

    let portDisplay = '-';
    if (a.port_mapping) {
      let hostPort = a.port;
      const parts = a.port_mapping.split(':');
      if (parts.length === 2) hostPort = parseInt(parts[0]) || a.port;
      if (parts.length === 3) hostPort = parseInt(parts[1]) || a.port;
      portDisplay = `<a href="http://${window.location.hostname}:${hostPort}" target="_blank" class="port-link">${escapeHtml(a.port_mapping)} ↗</a>`;
    } else if (a.port) {
      portDisplay = `<a href="http://${window.location.hostname}:${a.port}" target="_blank" class="port-link">:${a.port} ↗</a>`;
    }

    const engineTagClass = a.engine === 'rootfs' ? 'tag-rootfs' : 'tag-native';
    const statusBadgeClass = isRunning ? 'running' : (a.status === 'failed' ? 'failed' : 'stopped');

    return `
      <div class="app-card ${statusBadgeClass}">
        <div>
          <div class="card-top">
            <div class="app-identity" style="cursor:pointer;" onclick="openDetailDrawer('${a.id}', '${escapeHtml(a.name)}')">
              <span class="app-name">${escapeHtml(a.name)}</span>
              <span class="app-id-tag">ID: ${a.id}</span>
            </div>
            <div class="card-badges">
              <span class="tag ${engineTagClass}">${a.engine}</span>
              ${a.restart_policy && a.restart_policy !== 'no' ? `<span class="tag" style="background:rgba(99,102,241,0.2); color:#818cf8;" title="Restart Policy: ${escapeHtml(a.restart_policy)}">${Icons.refresh} <span>${escapeHtml(a.restart_policy)}${a.restart_count ? ' (' + a.restart_count + ')' : ''}</span></span>` : ''}
              ${a.volumes && a.volumes.length > 0 ? `<span class="tag" style="background:rgba(16,185,129,0.2); color:#34d399;" title="${a.volumes.length} Volume Bindings">${Icons.folder} <span>${a.volumes.length} vol</span></span>` : ''}
              ${a.network ? `<span class="tag" style="background:rgba(6,182,212,0.2); color:#22d3ee;" title="Network: ${escapeHtml(a.network)}${a.network_aliases && a.network_aliases.length ? ' (Aliases: ' + a.network_aliases.join(', ') + ')' : ''}"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:3px; vertical-align:-1px;"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg><span>${escapeHtml(a.network)}</span></span>` : ''}
              <span class="status-badge ${statusBadgeClass}">
                ${isRunning ? '<span class="pulse-dot"></span>' : ''}
                ${a.status}
              </span>
            </div>
          </div>

          <div class="app-command-box" title="${escapeHtml(a.command + ' ' + (a.args || []).join(' '))}">
            ${escapeHtml(a.command)} ${(a.args || []).map(escapeHtml).join(' ')}
          </div>

          <div class="app-telemetry">
            <div class="telemetry-item">
              <span class="telemetry-label">CPU</span>
              <span class="telemetry-val">${cpuStr}</span>
            </div>
            <div class="telemetry-item">
              <span class="telemetry-label">RAM</span>
              <span class="telemetry-val">${memStr}</span>
            </div>
            <div class="telemetry-item">
              <span class="telemetry-label">Port</span>
              <span class="telemetry-val">${portDisplay}</span>
            </div>
          </div>
        </div>

        <div class="card-actions">
          ${isRunning ? `
            <button class="btn btn-secondary btn-sm" onclick="stopApp('${a.id}', '${escapeHtml(a.name)}')">Stop</button>
            <button class="btn btn-secondary btn-sm" title="Zero-Downtime Hot Restart" onclick="rollingRestartApp('${a.id}', '${escapeHtml(a.name)}')" style="display:inline-flex; align-items:center; gap:4px;">${Icons.zap} <span>Hot Restart</span></button>
            <button class="btn btn-secondary btn-sm" onclick="openUpdateModal('${a.id}', '${escapeHtml(a.name)}')">Update</button>
            <button class="btn btn-secondary btn-sm" title="Expose via Cloudflare Tunnel" onclick="openTunnelModalForApp('${a.id}', '${escapeHtml(a.name)}', ${a.port || 0})" style="display:inline-flex; align-items:center; gap:4px; color:#f38020; border-color:rgba(243,128,32,0.35);"><svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/></svg> <span>Tunnel</span></button>
          ` : `
            <button class="btn btn-primary btn-sm" onclick="startApp('${a.id}', '${escapeHtml(a.name)}')">Start</button>
          `}
          <button class="btn btn-secondary btn-sm" onclick="openDetailDrawer('${a.id}', '${escapeHtml(a.name)}', 'logs')">Logs</button>
          <button class="btn btn-secondary btn-sm" onclick="exportApp('${a.id}', '${escapeHtml(a.name)}')">Export</button>
          <button class="btn btn-danger btn-sm" onclick="deleteApp('${a.id}', '${escapeHtml(a.name)}')">Delete</button>
        </div>
      </div>
    `;
  }).join('');
}

// Render Zoho CRM / Desk High-Density Table View
function renderAppsTable(filtered) {
  const tbody = document.getElementById('appsTableBody');
  if (!tbody) return;

  if (filtered.length === 0) {
    tbody.innerHTML = `
      <tr>
        <td colspan="8" style="text-align: center; padding: 40px; color: var(--text-muted);">
          No applications match current filters.
        </td>
      </tr>
    `;
    return;
  }

  tbody.innerHTML = filtered.map(item => {
    const a = item.app;
    const st = item.stats || {};
    const isRunning = a.status === 'running';

    let memStr = '-';
    let cpuStr = '-';
    let uptimeStr = a.status;

    if (isRunning && st.Alive) {
      memStr = formatBytes(st.MemoryRSS || 0);
      cpuStr = (st.CPUPercent || 0).toFixed(1) + '%';
      uptimeStr = `Up ${formatTimeAgo(a.started_at)}`;
    }

    let portDisplay = '-';
    if (a.port_mapping) {
      let hostPort = a.port;
      const parts = a.port_mapping.split(':');
      if (parts.length === 2) hostPort = parseInt(parts[0]) || a.port;
      if (parts.length === 3) hostPort = parseInt(parts[1]) || a.port;
      portDisplay = `<a href="http://${window.location.hostname}:${hostPort}" target="_blank" class="port-link" onclick="event.stopPropagation();">${escapeHtml(a.port_mapping)} ↗</a>`;
    } else if (a.port) {
      portDisplay = `<a href="http://${window.location.hostname}:${a.port}" target="_blank" class="port-link" onclick="event.stopPropagation();">:${a.port} ↗</a>`;
    }

    const engineTagClass = a.engine === 'rootfs' ? 'tag-rootfs' : 'tag-native';
    const statusBadgeClass = isRunning ? 'running' : (a.status === 'failed' ? 'failed' : 'stopped');

    return `
      <tr onclick="openDetailDrawer('${a.id}', '${escapeHtml(a.name)}')">
        <td>
          <div class="table-app-cell">
            <span class="table-app-name">${escapeHtml(a.name)}</span>
            <span class="table-app-id">ID: ${a.id}</span>
          </div>
        </td>
        <td>
          <span class="status-badge ${statusBadgeClass}">
            ${isRunning ? '<span class="pulse-dot"></span>' : ''}
            ${a.status}
          </span>
        </td>
        <td>
          <span class="tag ${engineTagClass}">${a.engine}</span>
        </td>
        <td>${portDisplay}</td>
        <td style="font-family: var(--font-mono); font-weight:600;">${cpuStr}</td>
        <td style="font-family: var(--font-mono); font-weight:600;">${memStr}</td>
        <td style="color: var(--text-secondary); font-size:12px;">${uptimeStr}</td>
        <td>
          <div class="table-actions-cell" onclick="event.stopPropagation();">
            ${isRunning ? `
              <button class="btn btn-secondary btn-sm" onclick="stopApp('${a.id}', '${escapeHtml(a.name)}')">Stop</button>
              <button class="btn btn-secondary btn-sm" title="Zero-Downtime Hot Restart" onclick="rollingRestartApp('${a.id}', '${escapeHtml(a.name)}')" style="display:inline-flex; align-items:center; gap:4px;">${Icons.zap} <span>Hot</span></button>
              <button class="btn btn-secondary btn-sm" onclick="openUpdateModal('${a.id}', '${escapeHtml(a.name)}')">Update</button>
              <button class="btn btn-secondary btn-sm" title="Expose via Cloudflare Tunnel" onclick="openTunnelModalForApp('${a.id}', '${escapeHtml(a.name)}', ${a.port || 0})" style="color:#f38020; border-color:rgba(243,128,32,0.35);">Tunnel</button>
            ` : `
              <button class="btn btn-primary btn-sm" onclick="startApp('${a.id}', '${escapeHtml(a.name)}')">Start</button>
            `}
            <button class="btn btn-secondary btn-sm" onclick="openDetailDrawer('${a.id}', '${escapeHtml(a.name)}', 'logs')">Logs</button>
            <button class="btn btn-secondary btn-sm" onclick="exportApp('${a.id}', '${escapeHtml(a.name)}')">Export</button>
            <button class="btn btn-danger btn-sm" onclick="deleteApp('${a.id}', '${escapeHtml(a.name)}')">Delete</button>
          </div>
        </td>
      </tr>
    `;
  }).join('');
}

// ==========================================
// Zoho Desk Slide-Over Detail Drawer
// ==========================================

function openDetailDrawer(id, name, defaultTab = 'overview') {
  activeDrawerApp = { id, name };
  document.getElementById('logsAppName').textContent = name;
  document.getElementById('logsAppId').textContent = `ID: ${id}`;
  document.getElementById('logsDrawer').classList.add('active');

  const targetApp = apps.find(item => item.app.id === id);
  if (targetApp) {
    updateDrawerContent(targetApp);
  }

  switchDrawerTab(defaultTab);

  if (logPollTimer) clearInterval(logPollTimer);
  logPollTimer = setInterval(refreshLogs, 1200);
}

// Backward compatibility alias for openLogsDrawer
function openLogsDrawer(id, name) {
  openDetailDrawer(id, name, 'logs');
}

function closeDetailDrawer() {
  document.getElementById('logsDrawer').classList.remove('active');
  activeDrawerApp = null;
  if (logPollTimer) clearInterval(logPollTimer);
}

// Backward compatibility alias for closeLogsDrawer
function closeLogsDrawer() {
  closeDetailDrawer();
}

function switchDrawerTab(tabName) {
  document.querySelectorAll('.drawer-tab').forEach(b => {
    if (b.dataset.tab === tabName) {
      b.classList.add('active');
    } else {
      b.classList.remove('active');
    }
  });

  const panels = {
    logs: document.getElementById('panelLogs'),
    overview: document.getElementById('panelOverview'),
    config: document.getElementById('panelConfig'),
    actions: document.getElementById('panelActions')
  };

  Object.entries(panels).forEach(([k, panel]) => {
    if (panel) {
      if (k === tabName) {
        panel.classList.add('active');
      } else {
        panel.classList.remove('active');
      }
    }
  });

  if (tabName === 'logs') {
    refreshLogs();
  }
}

function updateDrawerContent(target) {
  const a = target.app;
  const st = target.stats || {};
  const isRunning = a.status === 'running';

  // Update Status Badge in Drawer Header
  const statusBadge = document.getElementById('drawerStatusBadge');
  if (statusBadge) {
    statusBadge.className = `status-badge ${isRunning ? 'running' : 'stopped'}`;
    statusBadge.innerHTML = `${isRunning ? '<span class="pulse-dot"></span>' : ''}${a.status}`;
  }

  // Update Overview / Specs Tab
  const overviewContent = document.getElementById('drawerOverviewContent');
  if (overviewContent) {
    overviewContent.innerHTML = `
      <div class="spec-card">
        <div class="spec-title">Process Telemetry & Resources</div>
        <div class="spec-row">
          <span class="spec-label">Process ID (PID)</span>
          <span class="spec-value">${a.pid || '-'}</span>
        </div>
        <div class="spec-row">
          <span class="spec-label">CPU Utilization</span>
          <span class="spec-value">${isRunning ? (st.CPUPercent || 0).toFixed(1) + '%' : '-'}</span>
        </div>
        <div class="spec-row">
          <span class="spec-label">Resident Memory (RSS)</span>
          <span class="spec-value">${isRunning ? formatBytes(st.MemoryRSS || 0) : '-'}</span>
        </div>
        <div class="spec-row">
          <span class="spec-label">Engine Mode</span>
          <span class="spec-value">${escapeHtml(a.engine)}</span>
        </div>
        <div class="spec-row">
          <span class="spec-label">Restart Policy</span>
          <span class="spec-value">${escapeHtml(a.restart_policy || 'no')} (restarts: ${a.restart_count || 0})</span>
        </div>
        <div class="spec-row">
          <span class="spec-label">Started At</span>
          <span class="spec-value">${a.started_at ? new Date(a.started_at).toLocaleString() : '-'}</span>
        </div>
      </div>

      <div class="spec-card">
        <div class="spec-title">Network & Entrypoint</div>
        <div class="spec-row">
          <span class="spec-label">Port Mapping</span>
          <span class="spec-value">${escapeHtml(a.port_mapping || (a.port ? String(a.port) : '-'))}</span>
        </div>
        <div class="spec-row">
          <span class="spec-label">Command</span>
          <span class="spec-value" style="word-break: break-all;">${escapeHtml(a.command)}</span>
        </div>
        <div class="spec-row">
          <span class="spec-label">Arguments</span>
          <span class="spec-value">${(a.args || []).map(escapeHtml).join(' ') || '-'}</span>
        </div>
      </div>
    `;
  }

  // Update Config & Mounts Tab
  const configContent = document.getElementById('drawerConfigContent');
  if (configContent) {
    const envEntries = Object.entries(a.env || {});
    const volEntries = a.volumes || [];

    configContent.innerHTML = `
      <div class="spec-card" style="margin-bottom:16px;">
        <div class="spec-title">Environment Variables (${envEntries.length})</div>
        ${envEntries.length === 0 ? '<p style="color:var(--text-muted); font-size:12px;">No custom environment variables.</p>' : `
          <div style="display:flex; flex-direction:column; gap:6px;">
            ${envEntries.map(([k, v]) => `
              <div style="display:flex; justify-content:space-between; font-size:12px; font-family:var(--font-mono); background:rgba(0,0,0,0.25); padding:6px 10px; border-radius:4px;">
                <span style="color:var(--text-secondary);">${escapeHtml(k)}</span>
                <span style="color:#60a5fa;">${escapeHtml(v)}</span>
              </div>
            `).join('')}
          </div>
        `}
      </div>

      <div class="spec-card">
        <div class="spec-title">Volume Mounts (${volEntries.length})</div>
        ${volEntries.length === 0 ? '<p style="color:var(--text-muted); font-size:12px;">No volume bindings configured.</p>' : `
          <div style="display:flex; flex-direction:column; gap:6px;">
            ${volEntries.map(v => `
              <div style="display:flex; justify-content:space-between; align-items:center; font-size:12px; font-family:var(--font-mono); background:rgba(0,0,0,0.25); padding:6px 10px; border-radius:4px;">
                <span style="color:#34d399;">${escapeHtml(v.host_path)}</span>
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color:var(--text-muted);"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
                <span style="color:#a78bfa;">${escapeHtml(v.container_path)}</span>
              </div>
            `).join('')}
          </div>
        `}
      </div>
    `;
  }

  // Update Actions Tab
  const actionsContent = document.getElementById('drawerActionsContent');
  if (actionsContent) {
    actionsContent.innerHTML = `
      <div class="spec-card">
        <div class="spec-title">Lifecycle Operations</div>
        <p style="font-size:12px; color:var(--text-secondary); margin-bottom:14px;">
          Manage process execution, live updates, and package export.
        </p>
        <div style="display:flex; flex-direction:column; gap:10px;">
          ${isRunning ? `
            <button class="btn btn-secondary" onclick="stopApp('${a.id}', '${escapeHtml(a.name)}')">
              <span style="display:inline-flex; align-items:center; gap:6px;">${Icons.stop} Stop Service</span>
            </button>
            <button class="btn btn-secondary" onclick="rollingRestartApp('${a.id}', '${escapeHtml(a.name)}')">
              <span style="display:inline-flex; align-items:center; gap:6px;">${Icons.zap} Zero-Downtime Hot Restart</span>
            </button>
            <button class="btn btn-secondary" onclick="openUpdateModal('${a.id}', '${escapeHtml(a.name)}')">
              <span style="display:inline-flex; align-items:center; gap:6px;">${Icons.rocket} Live Cutover Update</span>
            </button>
          ` : `
            <button class="btn btn-primary" onclick="startApp('${a.id}', '${escapeHtml(a.name)}')">
              <span style="display:inline-flex; align-items:center; gap:6px;">${Icons.play} Start Service</span>
            </button>
          `}
          <button class="btn btn-secondary" onclick="exportApp('${a.id}', '${escapeHtml(a.name)}')">
            <span style="display:inline-flex; align-items:center; gap:6px;">${Icons.package} Export as .dpack Bundle</span>
          </button>
          <button class="btn btn-danger" onclick="deleteApp('${a.id}', '${escapeHtml(a.name)}')">
            <span style="display:inline-flex; align-items:center; gap:6px;">${Icons.trash} Delete Application</span>
          </button>
        </div>
      </div>
    `;
  }
}

async function refreshLogs() {
  if (!activeDrawerApp) return;
  const terminal = document.getElementById('logsTerminal');
  if (!terminal) return;

  try {
    const res = await fetch(`/api/apps/${activeDrawerApp.id}/logs?lines=100`);
    if (!res.ok) throw new Error('Failed to fetch logs');
    const text = await res.text();
    
    if (!text.trim()) {
      terminal.innerHTML = '<div class="terminal-line text-muted">(no log output generated yet)</div>';
      return;
    }

    const lines = text.split('\n');
    terminal.innerHTML = lines.map(l => `<div class="terminal-line">${escapeHtml(l)}</div>`).join('');

    if (autoScrollLogs) {
      terminal.scrollTop = terminal.scrollHeight;
    }
  } catch (err) {
    terminal.innerHTML = `<div class="terminal-line text-muted">Error reading logs: ${err.message}</div>`;
  }
}

// Deploy form handler
async function handleDeploySubmit(e) {
  e.preventDefault();
  const btn = document.getElementById('btnSubmitDeploy');
  btn.disabled = true;
  btn.textContent = 'Deploying...';

  const name = document.getElementById('appName').value.trim();
  const engine = document.getElementById('appEngine').value;
  const rootfs = document.getElementById('appRootfs').value;
  const command = document.getElementById('appCommand').value.trim();
  const rawArgs = document.getElementById('appArgs').value.trim();
  const portMapping = document.getElementById('appPort').value.trim();
  const restartPolicy = document.getElementById('appRestart').value;
  const rawVolumes = document.getElementById('appVolumes').value.trim();
  const workDir = document.getElementById('appWorkDir').value.trim();
  const rawEnv = document.getElementById('appEnv').value.trim();
  const detached = document.getElementById('appDetached').checked;

  const args = rawArgs ? rawArgs.split(' ').filter(Boolean) : [];
  const envMap = {};
  if (rawEnv) {
    rawEnv.split(',').forEach(pair => {
      const parts = pair.split('=');
      if (parts.length === 2) {
        envMap[parts[0].trim()] = parts[1].trim();
      }
    });
  }

  const volumes = [];
  if (rawVolumes) {
    rawVolumes.split(',').forEach(v => {
      const parts = v.trim().split(':');
      if (parts.length >= 2) {
        volumes.push({ host_path: parts[0].trim(), container_path: parts[1].trim() });
      }
    });
  }

  const netInput = document.getElementById('appNetwork');
  const network = netInput ? netInput.value.trim() : 'bridge';
  const rawAliases = document.getElementById('appNetworkAliases') ? document.getElementById('appNetworkAliases').value.trim() : '';
  const networkAliases = rawAliases ? rawAliases.split(',').map(s => s.trim()).filter(Boolean) : [];

  const payload = {
    name,
    engine,
    command,
    args,
    port_mapping: portMapping,
    restart_policy: restartPolicy,
    volumes: volumes,
    work_dir: workDir,
    env: envMap,
    rootfs: engine === 'rootfs' ? rootfs : '',
    detached,
    network,
    network_aliases: networkAliases
  };

  try {
    const res = await fetch('/api/apps', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    const result = await res.json();
    if (!res.ok) throw new Error(result.error || 'Failed to deploy app');

    showToast(`Successfully deployed ${result.name}!`, 'success');
    document.getElementById('deployModal').classList.remove('active');
    document.getElementById('deployForm').reset();
    fetchApps();
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Launch App';
  }
}

// App control actions
async function stopApp(id, name) {
  try {
    const res = await fetch(`/api/apps/${id}/stop`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to stop');
    showToast(`Stopped ${name}`, 'success');
    fetchApps();
  } catch (err) {
    showToast(`Error stopping ${name}`, 'error');
  }
}

async function startApp(id, name) {
  try {
    const res = await fetch(`/api/apps/${id}/start`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to start');
    showToast(`Started ${name}`, 'success');
    fetchApps();
  } catch (err) {
    showToast(`Error starting ${name}`, 'error');
  }
}

async function restartApp(id, name) {
  try {
    const res = await fetch(`/api/apps/${id}/restart`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to restart');
    showToast(`Restarted ${name}`, 'success');
    fetchApps();
  } catch (err) {
    showToast(`Error restarting ${name}`, 'error');
  }
}

async function deleteApp(id, name) {
  if (!confirm(`Are you sure you want to delete '${name}' and its workspace?`)) return;
  try {
    const res = await fetch(`/api/apps/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to delete');
    showToast(`Removed ${name}`, 'success');
    if (activeDrawerApp && activeDrawerApp.id === id) {
      closeDetailDrawer();
    }
    fetchApps();
  } catch (err) {
    showToast(`Error removing ${name}`, 'error');
  }
}

// Export app as a portable .dpack bundle
function exportApp(id, name) {
  showToast(`Exporting ${name} bundle...`, 'info');
  const link = document.createElement('a');
  link.href = `/api/apps/${id}/export`;
  link.download = `${name}.dpack`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

// Image Registry and Lifecycle Handlers (Pull, Build, Push, Images)
async function loadImagesList() {
  const tbody = document.getElementById('imagesTableBody');
  if (!tbody) return;
  tbody.innerHTML = '<tr><td colspan="5" style="text-align:center; padding:16px; color:var(--text-secondary);">Loading images...</td></tr>';

  try {
    const res = await fetch('/api/images');
    if (!res.ok) throw new Error('Failed to fetch images');
    const images = await res.json() || [];

    if (images.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" style="text-align:center; padding:24px; color:var(--text-secondary);">No images found. Pull an image from Docker Hub / OCI or build one using the Build tab.</td></tr>';
      return;
    }

    tbody.innerHTML = images.map(img => {
      const tagDisplay = img.version || 'latest';
      const engineDisplay = img.engine || 'native';
      const archDisplay = img.architecture || 'all';
      const fullRef = `${img.name}:${tagDisplay}`;

      return `
        <tr style="border-bottom: 1px solid var(--border-card);">
          <td style="padding: 10px 12px; font-weight:600; color:var(--text-primary); font-family:var(--font-mono);">${escapeHtml(img.name)}</td>
          <td style="padding: 10px 12px; font-family:var(--font-mono);"><span class="badge" style="background:rgba(16,185,129,0.15); color:#10b981; padding:2px 8px; border-radius:4px; font-size:11px;">${escapeHtml(tagDisplay)}</span></td>
          <td style="padding: 10px 12px; text-transform:uppercase; font-size:11px; color:var(--text-secondary);">${escapeHtml(engineDisplay)}</td>
          <td style="padding: 10px 12px; font-size:12px; color:var(--text-secondary);">${escapeHtml(archDisplay)}</td>
          <td style="padding: 10px 12px; text-align:right;">
            <div style="display:flex; gap:6px; justify-content:flex-end;">
              <button type="button" class="btn btn-secondary btn-xs" onclick="deployFromImagePrompt('${escapeHtml(fullRef)}')">Deploy</button>
              <button type="button" class="btn btn-secondary btn-xs" onclick="switchPushTab('${escapeHtml(fullRef)}')">Push</button>
              <button type="button" class="btn btn-secondary btn-xs" style="color:#ef4444;" onclick="deleteImage('${escapeHtml(fullRef)}')">Delete</button>
            </div>
          </td>
        </tr>
      `;
    }).join('');
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="5" style="text-align:center; padding:16px; color:#ef4444;">Error: ${escapeHtml(err.message)}</td></tr>`;
  }
}

function switchPushTab(ref) {
  const pushTabBtn = document.querySelector('.img-modal-tab[data-target="tabImagePush"]');
  if (pushTabBtn) pushTabBtn.click();
  const pushInput = document.getElementById('pushImageRef');
  if (pushInput) pushInput.value = ref;
}

function deployFromImagePrompt(ref) {
  document.getElementById('importModal').classList.remove('active');
  const deployModal = document.getElementById('deployModal');
  if (deployModal) {
    deployModal.classList.add('active');
    const cmdInput = document.getElementById('appCommand');
    if (cmdInput) cmdInput.value = ref;
    const nameInput = document.getElementById('appName');
    if (nameInput && !nameInput.value) {
      const cleanName = ref.split(':')[0].split('/').pop();
      nameInput.value = cleanName;
    }
  }
}

async function deleteImage(tag) {
  if (!confirm(`Are you sure you want to untag and delete image '${tag}'?`)) return;
  try {
    const res = await fetch(`/api/images/${encodeURIComponent(tag)}`, { method: 'DELETE' });
    const result = await res.json();
    if (!res.ok) throw new Error(result.error || 'Failed to delete');
    showToast(`Image ${tag} deleted`, 'success');
    loadImagesList();
  } catch (err) {
    showToast(`Delete failed: ${err.message}`, 'error');
  }
}

async function handlePullImageSubmit(e) {
  e.preventDefault();
  const input = document.getElementById('pullImageRef');
  const btn = document.getElementById('btnSubmitPull');
  const logsContainer = document.getElementById('pullLogsContainer');
  const logsPre = document.getElementById('pullLogs');

  const imageRef = input.value.trim();
  if (!imageRef) return;

  btn.disabled = true;
  btn.innerHTML = '<span class="spinner" style="width:14px;height:14px;margin-right:6px;display:inline-block;vertical-align:middle;"></span> Pulling...';
  logsContainer.style.display = 'block';
  logsPre.textContent = `Connecting to registry and downloading ${imageRef}...\n`;

  try {
    const res = await fetch('/api/images/pull', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ image: imageRef })
    });
    const result = await res.json();
    if (!res.ok) {
      logsPre.textContent += (result.logs || '') + `\nError: ${result.error || 'Pull failed'}`;
      throw new Error(result.error || 'Pull failed');
    }

    logsPre.textContent = result.logs || `Successfully pulled and unpacked ${imageRef}`;
    showToast(`Successfully pulled ${imageRef}!`, 'success');
    input.value = '';
    loadImagesList();
  } catch (err) {
    showToast(`Pull error: ${err.message}`, 'error');
  } finally {
    btn.disabled = false;
    btn.innerHTML = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg><span>Pull Image</span>';
  }
}

async function handleBuildImageSubmit(e) {
  e.preventDefault();
  const tagInput = document.getElementById('buildImageTag');
  const dockerfileInput = document.getElementById('buildDockerfile');
  const btn = document.getElementById('btnSubmitBuild');
  const logsContainer = document.getElementById('buildLogsContainer');
  const logsPre = document.getElementById('buildLogs');

  const tag = tagInput.value.trim();
  const dockerfile = dockerfileInput.value;
  if (!tag) return;

  btn.disabled = true;
  btn.innerHTML = '<span class="spinner" style="width:14px;height:14px;margin-right:6px;display:inline-block;vertical-align:middle;"></span> Building...';
  logsContainer.style.display = 'block';
  logsPre.textContent = `Building image '${tag}'...\n`;

  try {
    const res = await fetch('/api/images/build', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tag: tag, dockerfile: dockerfile })
    });
    const result = await res.json();
    if (!res.ok) {
      logsPre.textContent += (result.logs || '') + `\nBuild Error: ${result.error || 'Build failed'}`;
      throw new Error(result.error || 'Build failed');
    }

    logsPre.textContent = result.logs || `Successfully built and tagged ${tag}`;
    showToast(`Successfully built ${tag}!`, 'success');
    loadImagesList();
  } catch (err) {
    showToast(`Build error: ${err.message}`, 'error');
  } finally {
    btn.disabled = false;
    btn.innerHTML = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg><span>Build Image</span>';
  }
}

async function handlePushImageSubmit(e) {
  e.preventDefault();
  const input = document.getElementById('pushImageRef');
  const btn = document.getElementById('btnSubmitPush');
  const logsContainer = document.getElementById('pushLogsContainer');
  const logsPre = document.getElementById('pushLogs');

  const imageRef = input.value.trim();
  if (!imageRef) return;

  btn.disabled = true;
  btn.innerHTML = '<span class="spinner" style="width:14px;height:14px;margin-right:6px;display:inline-block;vertical-align:middle;"></span> Pushing...';
  logsContainer.style.display = 'block';
  logsPre.textContent = `Pushing ${imageRef} to registry...\n`;

  try {
    const res = await fetch('/api/images/push', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ image: imageRef })
    });
    const result = await res.json();
    if (!res.ok) {
      logsPre.textContent += (result.logs || '') + `\nPush Error: ${result.error || 'Push failed'}`;
      throw new Error(result.error || 'Push failed');
    }

    logsPre.textContent = (result.logs || '') + `\nDigest: ${result.digest}`;
    showToast(`Successfully pushed ${imageRef}!`, 'success');
  } catch (err) {
    showToast(`Push error: ${err.message}`, 'error');
  } finally {
    btn.disabled = false;
    btn.innerHTML = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/></svg><span>Push Image</span>';
  }
}

// Handle .dpack bundle import
async function handleImportSubmit(e) {
  e.preventDefault();
  const fileInput = document.getElementById('bundleFile');
  if (!fileInput.files || fileInput.files.length === 0) {
    showToast('Please select a .dpack bundle file', 'error');
    return;
  }

  const btn = document.getElementById('btnSubmitImport');
  btn.disabled = true;
  btn.textContent = 'Importing...';

  const formData = new FormData();
  formData.append('bundle', fileInput.files[0]);

  try {
    const res = await fetch('/api/bundles/import', {
      method: 'POST',
      body: formData
    });
    const result = await res.json();
    if (!res.ok) throw new Error(result.error || 'Import failed');

    showToast(`Successfully imported ${result.name}:${result.version}!`, 'success');
    document.getElementById('importModal').classList.remove('active');
    document.getElementById('importForm').reset();
    fetchApps();
  } catch (err) {
    showToast(`Import error: ${err.message}`, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Import Package';
  }
}

// System Doctor Diagnostics
async function loadDoctorDiagnostics() {
  const container = document.getElementById('doctorContent');
  container.innerHTML = `
    <div class="loading-placeholder">
      <div class="spinner"></div>
      <p>Inspecting Android kernel, Termux environment & PRoot status...</p>
    </div>
  `;

  try {
    const res = await fetch('/api/system');
    if (!res.ok) throw new Error('Failed to load system diagnostics');
    const data = await res.json();

    container.innerHTML = `
      <table style="width:100%; border-collapse: collapse; margin-bottom: 20px;">
        <thead>
          <tr style="border-bottom: 1px solid var(--border-subtle); text-align: left; font-size: 11px; color: var(--text-muted); text-transform: uppercase;">
            <th style="padding: 8px 12px;">CHECK</th>
            <th style="padding: 8px 12px;">STATUS</th>
            <th style="padding: 8px 12px;">DETAILS</th>
          </tr>
        </thead>
        <tbody>
          ${(data.checks || []).map(c => `
            <tr style="border-bottom: 1px solid rgba(255,255,255,0.04); font-size: 13px;">
              <td style="padding: 10px 12px; font-weight: 600; color: #ffffff;">${escapeHtml(c.Name)}</td>
              <td style="padding: 10px 12px;">
                <span class="badge" style="color:${c.Status === 'OK' ? '#34d399' : '#fb7185'}; background:${c.Status === 'OK' ? 'rgba(16,185,129,0.1)' : 'rgba(244,63,94,0.1)'}; display:inline-flex; align-items:center; gap:4px;">
                  ${c.Status === 'OK' ? '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg> OK' : '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg> ' + c.Status}
                </span>
              </td>
              <td style="padding: 10px 12px; color: var(--text-secondary);">${escapeHtml(c.Message)}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>

      <div style="background: rgba(255,255,255,0.02); padding: 14px; border-radius: 8px; border: 1px solid var(--border-subtle);">
        <div style="font-size: 12px; font-weight: 700; color: var(--accent-amber); margin-bottom: 6px;">Android Termux Performance Advisory</div>
        <p style="font-size: 12px; color: var(--text-muted); line-height: 1.5;">
          • To prevent process eviction when Android turns the display off, execute <code>termux-wake-lock</code>.<br>
          • On Android 12+, disable the Phantom Process Killer in Termux or via ADB.
        </p>
      </div>
    `;
  } catch (err) {
    container.innerHTML = `<p style="color:var(--accent-rose);">Failed to run diagnostics: ${err.message}</p>`;
  }
}

// Toast helper
function showToast(msg, type = 'info') {
  const container = document.getElementById('toastContainer');
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.textContent = msg;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transition = 'opacity 0.3s ease';
    setTimeout(() => toast.remove(), 300);
  }, 3000);
}

// Utility formatting
function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  let i = 0;
  let val = bytes;
  while (val >= 1024 && i < units.length - 1) {
    val /= 1024;
    i++;
  }
  return `${val.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

function formatTimeAgo(isoString) {
  if (!isoString) return 'recently';
  const diffSec = Math.floor((new Date() - new Date(isoString)) / 1000);
  if (diffSec < 60) return `${diffSec}s`;
  if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m`;
  if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h`;
  return `${Math.floor(diffSec / 86400)}d`;
}

function escapeHtml(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

// Secrets Store UI logic
async function loadSecrets() {
  const tbody = document.getElementById('secretsTableBody');
  tbody.innerHTML = '<tr><td colspan="3" style="text-align:center; padding:16px; color:var(--text-secondary);">Loading secrets...</td></tr>';

  try {
    const res = await fetch('/api/secrets');
    if (!res.ok) throw new Error('Failed to fetch secrets');
    const list = await res.json();

    if (!list || list.length === 0) {
      tbody.innerHTML = '<tr><td colspan="3" style="text-align:center; padding:16px; color:var(--text-secondary);">No secrets stored yet. Add one above!</td></tr>';
      return;
    }

    tbody.innerHTML = list.map(s => {
      const masked = '••••••••';
      return `
        <tr style="border-bottom: 1px solid rgba(255,255,255,0.05);">
          <td style="padding: 8px 12px; font-family: var(--font-mono); font-weight: 600; color: #38bdf8;">${escapeHtml(s.key)}</td>
          <td style="padding: 8px 12px; font-family: var(--font-mono);" id="sec-val-${escapeHtml(s.key)}">
            <span class="masked-val">${masked}</span>
            <span class="plain-val" style="display:none;">${escapeHtml(s.value)}</span>
          </td>
          <td style="padding: 8px 12px; text-align: right;">
            <button class="btn btn-secondary btn-sm" style="padding: 2px 8px; font-size: 11px; margin-right: 6px;" onclick="toggleSecretVisibility(this, '${escapeHtml(s.key)}')">Reveal</button>
            <button class="btn btn-danger btn-sm" style="padding: 2px 8px; font-size: 11px;" onclick="deleteSecret('${escapeHtml(s.key)}')">Delete</button>
          </td>
        </tr>
      `;
    }).join('');
  } catch (err) {
    tbody.innerHTML = `<tr><td colspan="3" style="text-align:center; padding:16px; color:var(--accent-rose);">${err.message}</td></tr>`;
  }
}

async function handleAddSecret(e) {
  e.preventDefault();
  const keyInput = document.getElementById('newSecretKey');
  const valInput = document.getElementById('newSecretVal');

  const key = keyInput.value.trim();
  const val = valInput.value;
  if (!key) return;

  try {
    const res = await fetch('/api/secrets', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ key, value: val })
    });
    if (!res.ok) throw new Error('Failed to save secret');

    showToast(`Secret '${key}' saved!`, 'success');
    keyInput.value = '';
    valInput.value = '';
    loadSecrets();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function deleteSecret(key) {
  if (!confirm(`Delete secret '${key}'?`)) return;

  try {
    const res = await fetch(`/api/secrets/${encodeURIComponent(key)}`, {
      method: 'DELETE'
    });
    if (!res.ok) throw new Error('Failed to delete secret');

    showToast(`Secret '${key}' deleted`, 'info');
    loadSecrets();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

function toggleSecretVisibility(btn, key) {
  const container = document.getElementById(`sec-val-${key}`);
  if (!container) return;
  const masked = container.querySelector('.masked-val');
  const plain = container.querySelector('.plain-val');

  if (plain.style.display === 'none') {
    plain.style.display = 'inline';
    masked.style.display = 'none';
    btn.textContent = 'Hide';
  } else {
    plain.style.display = 'none';
    masked.style.display = 'inline';
    btn.textContent = 'Reveal';
  }
}

// Zero-Downtime Hot Restart
async function rollingRestartApp(id, name) {
  showToast(`Initiating zero-downtime restart for ${name}...`, 'info');
  try {
    const res = await fetch(`/api/apps/${id}/rolling-restart`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Rolling restart failed');
    showToast(`Zero-downtime restart complete for ${name}! (PID: ${data.old_pid} -> ${data.new_pid})`, 'success');
    fetchApps();
  } catch (err) {
    showToast(`Restart error: ${err.message}`, 'error');
  }
}

// Open Update Modal
function openUpdateModal(id, name) {
  const modal = document.getElementById('updateModal');
  document.getElementById('updateAppId').value = id;
  document.getElementById('updateAppNameDisplay').textContent = name;
  const targetApp = apps.find(a => a.app.id === id);
  if (targetApp && targetApp.app) {
    const a = targetApp.app;
    document.getElementById('updateCommand').value = a.command || '';
    document.getElementById('updateArgs').value = (a.args || []).join(' ');
    if (a.env) {
      const envPairs = Object.entries(a.env).map(([k, v]) => `${k}=${v}`).join(',');
      document.getElementById('updateEnv').value = envPairs;
    } else {
      document.getElementById('updateEnv').value = '';
    }
  }
  modal.classList.add('active');
}

// Handle Update Form Submission
async function handleUpdateSubmit(e) {
  e.preventDefault();
  const id = document.getElementById('updateAppId').value;
  const name = document.getElementById('updateAppNameDisplay').textContent;
  const btn = document.getElementById('btnSubmitUpdate');
  btn.disabled = true;
  btn.textContent = 'Executing Live Cutover...';

  const command = document.getElementById('updateCommand').value.trim();
  const rawArgs = document.getElementById('updateArgs').value.trim();
  const rawEnv = document.getElementById('updateEnv').value.trim();
  const drainTimeout = parseFloat(document.getElementById('updateDrain').value) || 1.5;
  const probeTimeout = parseFloat(document.getElementById('updateTimeout').value) || 10;

  const args = rawArgs ? rawArgs.split(' ').filter(Boolean) : [];
  const env = {};
  if (rawEnv) {
    rawEnv.split(',').forEach(pair => {
      const parts = pair.split('=');
      if (parts.length === 2) {
        env[parts[0].trim()] = parts[1].trim();
      }
    });
  }

  try {
    const res = await fetch(`/api/apps/${id}/update`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        command: command,
        args: args,
        env: env,
        drain_timeout: drainTimeout,
        probe_timeout: probeTimeout
      })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Update failed');

    showToast(`Successfully updated ${name} with zero downtime! (PID: ${data.old_pid} -> ${data.new_pid})`, 'success');
    document.getElementById('updateModal').classList.remove('active');
    fetchApps();
  } catch (err) {
    showToast(`Live update error: ${err.message}`, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Execute Live Cutover';
  }
}

// ==========================================
// Docroid Compose (Upload, Parse, Generate)
// ==========================================

let uploadedComposeConfig = null;
let uploadedRawYAML = null;

async function handleComposeFile(file) {
  if (!file) return;
  showToast(`Parsing ${file.name}...`, 'info');

  const formData = new FormData();
  formData.append('compose', file);

  try {
    const res = await fetch('/api/compose/parse', {
      method: 'POST',
      body: formData
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to parse compose file');

    uploadedComposeConfig = data;
    uploadedRawYAML = await file.text();

    const services = data.services || [];
    if (services.length === 0) {
      showToast('No valid services found in compose file', 'error');
      return;
    }

    // Populate first service
    populateFormFromService(services[0]);

    // Handle service selector if multiple services
    const serviceGroup = document.getElementById('composeServiceGroup');
    const serviceSelect = document.getElementById('composeServiceSelect');
    const btnDeployAll = document.getElementById('btnDeployAllCompose');

    if (services.length > 1) {
      serviceSelect.innerHTML = services.map((s, idx) => 
        `<option value="${idx}">${escapeHtml(s.name || s.service_name)}</option>`
      ).join('');
      serviceGroup.style.display = 'flex';
      btnDeployAll.style.display = 'inline-flex';
      showToast(`Loaded ${file.name} with ${services.length} services!`, 'success');
    } else {
      serviceGroup.style.display = 'none';
      btnDeployAll.style.display = 'none';
      showToast(`Loaded and populated configuration for '${services[0].name || services[0].service_name}'!`, 'success');
    }
  } catch (err) {
    showToast(`Compose error: ${err.message}`, 'error');
  }
}

function populateFormFromService(svc) {
  if (!svc) return;

  document.getElementById('appName').value = svc.name || svc.service_name || '';
  
  const engineSelect = document.getElementById('appEngine');
  engineSelect.value = svc.engine || 'native';
  engineSelect.dispatchEvent(new Event('change'));

  if (svc.rootfs) {
    document.getElementById('appRootfs').value = svc.rootfs;
  }
  
  document.getElementById('appCommand').value = svc.command || '';
  document.getElementById('appArgs').value = (svc.args || []).join(' ');
  document.getElementById('appPort').value = svc.port_mapping || (svc.port ? String(svc.port) : '');
  document.getElementById('appRestart').value = svc.restart_policy || 'no';
  document.getElementById('appVolumes').value = (svc.volumes || []).join(', ');
  document.getElementById('appWorkDir').value = svc.working_dir || '';
  
  if (svc.environment && typeof svc.environment === 'object') {
    const pairs = Object.entries(svc.environment).map(([k, v]) => `${k}=${v}`).join(', ');
    document.getElementById('appEnv').value = pairs;
  } else {
    document.getElementById('appEnv').value = '';
  }

  document.getElementById('appDetached').checked = svc.detached !== false;
}

async function handleDownloadCompose() {
  const name = document.getElementById('appName').value.trim() || 'my-app';
  const engine = document.getElementById('appEngine').value;
  const rootfs = document.getElementById('appRootfs').value;
  const command = document.getElementById('appCommand').value.trim() || './run';
  const rawArgs = document.getElementById('appArgs').value.trim();
  const portMapping = document.getElementById('appPort').value.trim();
  const restartPolicy = document.getElementById('appRestart').value;
  const rawVolumes = document.getElementById('appVolumes').value.trim();
  const workDir = document.getElementById('appWorkDir').value.trim();
  const rawEnv = document.getElementById('appEnv').value.trim();
  const detached = document.getElementById('appDetached').checked;

  const args = rawArgs ? rawArgs.split(' ').filter(Boolean) : [];
  const envMap = {};
  if (rawEnv) {
    rawEnv.split(',').forEach(pair => {
      const parts = pair.split('=');
      if (parts.length === 2) {
        envMap[parts[0].trim()] = parts[1].trim();
      }
    });
  }

  const volumes = rawVolumes ? rawVolumes.split(',').map(v => v.trim()).filter(Boolean) : [];

  const payload = {
    name: name,
    engine: engine,
    rootfs: rootfs,
    command: command,
    args: args,
    port_mapping: portMapping,
    restart_policy: restartPolicy,
    work_dir: workDir,
    env: envMap,
    volumes: volumes,
    detached: detached
  };

  try {
    showToast('Generating docroid-compose.yml...', 'info');
    const res = await fetch('/api/compose/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    if (!res.ok) throw new Error('Failed to generate compose file');

    const blob = await res.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `docroid-compose-${name}.yml`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);

    showToast(`Downloaded docroid-compose-${name}.yml!`, 'success');
  } catch (err) {
    showToast(`Error: ${err.message}`, 'error');
  }
}

async function handleDeployAllCompose() {
  if (!uploadedRawYAML) {
    showToast('No compose YAML loaded', 'error');
    return;
  }

  const btn = document.getElementById('btnDeployAllCompose');
  btn.disabled = true;
  btn.textContent = 'Deploying all services...';

  try {
    const res = await fetch('/api/compose/up', {
      method: 'POST',
      headers: { 'Content-Type': 'text/yaml' },
      body: uploadedRawYAML
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to deploy all services');

    showToast(`Successfully deployed ${data.count} services!`, 'success');
    document.getElementById('deployModal').classList.remove('active');
    fetchApps();
  } catch (err) {
    showToast(`Deployment error: ${err.message}`, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Deploy All Services (Compose Up)';
  }
}

/* ====================================================
   Static Websites Controller & Logic
==================================================== */

let selectedSiteZip = null;
let selectedSiteFiles = [];
let currentSiteTab = 'zip';
let activeExplorerSite = null;

function switchWorkspaceView(view) {
  currentWorkspace = view;
  const sectionApps = document.getElementById('sectionApps');
  const sectionSites = document.getElementById('sectionSites');
  const sectionCluster = document.getElementById('sectionCluster');
  const sectionNetworks = document.getElementById('sectionNetworks');
  const sectionTunnels = document.getElementById('sectionTunnels');
  const sectionDocs = document.getElementById('sectionDocs');
  const breadcrumb = document.getElementById('breadcrumbActive');
  const btnDeployApp = document.getElementById('btnDeployModal');
  const btnDeploySite = document.getElementById('btnDeploySiteTop');
  const btnDeployPod = document.getElementById('btnDeployPodTop');
  const btnClusterSetup = document.getElementById('btnClusterSetupTop');
  const btnCreateNetwork = document.getElementById('btnCreateNetworkTop');
  const btnConfigureTunnel = document.getElementById('btnConfigureTunnelTop');

  if (view === 'docs') {
    if (sectionApps) sectionApps.style.display = 'none';
    if (sectionSites) sectionSites.style.display = 'none';
    if (sectionCluster) sectionCluster.style.display = 'none';
    if (sectionNetworks) sectionNetworks.style.display = 'none';
    if (sectionTunnels) sectionTunnels.style.display = 'none';
    if (sectionDocs) sectionDocs.style.display = 'block';
    if (breadcrumb) breadcrumb.textContent = 'Documentation & Developer Reference';
    if (btnDeployApp) btnDeployApp.style.display = 'none';
    if (btnDeploySite) btnDeploySite.style.display = 'none';
    if (btnDeployPod) btnDeployPod.style.display = 'none';
    if (btnClusterSetup) btnClusterSetup.style.display = 'none';
    if (btnCreateNetwork) btnCreateNetwork.style.display = 'none';
    if (btnConfigureTunnel) btnConfigureTunnel.style.display = 'none';
  } else if (view === 'sites') {
    if (sectionApps) sectionApps.style.display = 'none';
    if (sectionSites) sectionSites.style.display = 'block';
    if (sectionCluster) sectionCluster.style.display = 'none';
    if (sectionNetworks) sectionNetworks.style.display = 'none';
    if (sectionTunnels) sectionTunnels.style.display = 'none';
    if (sectionDocs) sectionDocs.style.display = 'none';
    if (breadcrumb) breadcrumb.textContent = 'Static Websites';
    if (btnDeployApp) btnDeployApp.style.display = 'none';
    if (btnDeploySite) btnDeploySite.style.display = 'inline-flex';
    if (btnDeployPod) btnDeployPod.style.display = 'none';
    if (btnClusterSetup) btnClusterSetup.style.display = 'none';
    if (btnCreateNetwork) btnCreateNetwork.style.display = 'none';
    if (btnConfigureTunnel) btnConfigureTunnel.style.display = 'none';
  } else if (view === 'cluster') {
    if (sectionApps) sectionApps.style.display = 'none';
    if (sectionSites) sectionSites.style.display = 'none';
    if (sectionCluster) sectionCluster.style.display = 'block';
    if (sectionNetworks) sectionNetworks.style.display = 'none';
    if (sectionTunnels) sectionTunnels.style.display = 'none';
    if (sectionDocs) sectionDocs.style.display = 'none';
    if (breadcrumb) breadcrumb.textContent = 'Cluster & Pods Scaling';
    if (btnDeployApp) btnDeployApp.style.display = 'none';
    if (btnDeploySite) btnDeploySite.style.display = 'none';
    if (btnDeployPod) btnDeployPod.style.display = 'inline-flex';
    if (btnClusterSetup) btnClusterSetup.style.display = 'inline-flex';
    if (btnCreateNetwork) btnCreateNetwork.style.display = 'none';
    if (btnConfigureTunnel) btnConfigureTunnel.style.display = 'none';
  } else if (view === 'networks') {
    if (sectionApps) sectionApps.style.display = 'none';
    if (sectionSites) sectionSites.style.display = 'none';
    if (sectionCluster) sectionCluster.style.display = 'none';
    if (sectionNetworks) sectionNetworks.style.display = 'block';
    if (sectionTunnels) sectionTunnels.style.display = 'none';
    if (sectionDocs) sectionDocs.style.display = 'none';
    if (breadcrumb) breadcrumb.textContent = 'Virtual Networks & Inter-Container Intercom';
    if (btnDeployApp) btnDeployApp.style.display = 'none';
    if (btnDeploySite) btnDeploySite.style.display = 'none';
    if (btnDeployPod) btnDeployPod.style.display = 'none';
    if (btnClusterSetup) btnClusterSetup.style.display = 'none';
    if (btnCreateNetwork) btnCreateNetwork.style.display = 'inline-flex';
    if (btnConfigureTunnel) btnConfigureTunnel.style.display = 'none';
  } else if (view === 'tunnels') {
    if (sectionApps) sectionApps.style.display = 'none';
    if (sectionSites) sectionSites.style.display = 'none';
    if (sectionCluster) sectionCluster.style.display = 'none';
    if (sectionNetworks) sectionNetworks.style.display = 'none';
    if (sectionTunnels) sectionTunnels.style.display = 'block';
    if (sectionDocs) sectionDocs.style.display = 'none';
    if (breadcrumb) breadcrumb.textContent = 'Cloudflare Tunnels & Direct Port Routing';
    if (btnDeployApp) btnDeployApp.style.display = 'none';
    if (btnDeploySite) btnDeploySite.style.display = 'none';
    if (btnDeployPod) btnDeployPod.style.display = 'none';
    if (btnClusterSetup) btnClusterSetup.style.display = 'none';
    if (btnCreateNetwork) btnCreateNetwork.style.display = 'none';
    if (btnConfigureTunnel) btnConfigureTunnel.style.display = 'inline-flex';
  } else {
    if (sectionApps) sectionApps.style.display = 'block';
    if (sectionSites) sectionSites.style.display = 'none';
    if (sectionCluster) sectionCluster.style.display = 'none';
    if (sectionNetworks) sectionNetworks.style.display = 'none';
    if (sectionTunnels) sectionTunnels.style.display = 'none';
    if (sectionDocs) sectionDocs.style.display = 'none';
    if (breadcrumb) breadcrumb.textContent = 'Applications Workspace';
    if (btnDeployApp) btnDeployApp.style.display = 'inline-flex';
    if (btnDeploySite) btnDeploySite.style.display = 'none';
    if (btnDeployPod) btnDeployPod.style.display = 'none';
    if (btnClusterSetup) btnClusterSetup.style.display = 'none';
    if (btnCreateNetwork) btnCreateNetwork.style.display = 'none';
    if (btnConfigureTunnel) btnConfigureTunnel.style.display = 'none';
  }

  if (breadcrumb) breadcrumb.title = breadcrumb.textContent;

  // Synchronize Mobile Bottom Navigation Active State
  const bottomNavItems = document.querySelectorAll('.bottom-nav-item[data-view]');
  bottomNavItems.forEach(item => {
    if (item.getAttribute('data-view') === view) {
      item.classList.add('active');
    } else {
      item.classList.remove('active');
    }
  });
}

function initStaticWebsites() {
  // Topbar and Section Deploy Buttons
  const btnTopDeploy = document.getElementById('btnDeploySiteTop');
  const btnSectionDeploy = document.getElementById('btnOpenDeploySiteModal');
  const modalDeploy = document.getElementById('deploySiteModal');
  const closeDeploy = document.getElementById('closeDeploySiteModal');
  const cancelDeploy = document.getElementById('cancelDeploySiteModal');

  const openDeploy = () => {
    modalDeploy.classList.add('active');
    suggestFreePort();
  };

  if (btnTopDeploy) btnTopDeploy.addEventListener('click', openDeploy);
  if (btnSectionDeploy) btnSectionDeploy.addEventListener('click', openDeploy);
  if (closeDeploy) closeDeploy.addEventListener('click', () => modalDeploy.classList.remove('active'));
  if (cancelDeploy) cancelDeploy.addEventListener('click', () => modalDeploy.classList.remove('active'));

  // Upload Mode Tabs (ZIP, Files, HTML)
  const tabZip = document.getElementById('tabSiteZip');
  const tabFiles = document.getElementById('tabSiteFiles');
  const tabHtml = document.getElementById('tabSiteHtml');
  const paneZip = document.getElementById('paneSiteZip');
  const paneFiles = document.getElementById('paneSiteFiles');
  const paneHtml = document.getElementById('paneSiteHtml');

  const switchTab = (mode) => {
    currentSiteTab = mode;
    [tabZip, tabFiles, tabHtml].forEach(t => t && t.classList.remove('active'));
    [paneZip, paneFiles, paneHtml].forEach(p => p && p.classList.remove('active'));

    if (mode === 'zip') {
      tabZip.classList.add('active');
      paneZip.classList.add('active');
    } else if (mode === 'files') {
      tabFiles.classList.add('active');
      paneFiles.classList.add('active');
    } else {
      tabHtml.classList.add('active');
      paneHtml.classList.add('active');
    }
  };

  if (tabZip) tabZip.addEventListener('click', () => switchTab('zip'));
  if (tabFiles) tabFiles.addEventListener('click', () => switchTab('files'));
  if (tabHtml) tabHtml.addEventListener('click', () => switchTab('html'));

  // ZIP Drag & Drop
  const zipDropzone = document.getElementById('siteZipDropzone');
  const zipInput = document.getElementById('inputSiteZip');
  const zipChip = document.getElementById('zipFileChip');
  const zipName = document.getElementById('zipFileName');
  const btnRemoveZip = document.getElementById('btnRemoveZip');

  if (zipDropzone && zipInput) {
    zipDropzone.addEventListener('click', (e) => {
      if (e.target !== btnRemoveZip) zipInput.click();
    });
    zipDropzone.addEventListener('dragover', (e) => {
      e.preventDefault();
      zipDropzone.classList.add('dragover');
    });
    zipDropzone.addEventListener('dragleave', () => zipDropzone.classList.remove('dragover'));
    zipDropzone.addEventListener('drop', (e) => {
      e.preventDefault();
      zipDropzone.classList.remove('dragover');
      if (e.dataTransfer.files.length > 0) {
        handleZipSelected(e.dataTransfer.files[0]);
      }
    });

    zipInput.addEventListener('change', (e) => {
      if (e.target.files.length > 0) {
        handleZipSelected(e.target.files[0]);
      }
    });
  }

  function handleZipSelected(file) {
    if (!file.name.toLowerCase().endsWith('.zip')) {
      showToast('Please select a valid .zip file', 'error');
      return;
    }
    selectedSiteZip = file;
    if (zipChip && zipName) {
      zipName.textContent = `${file.name} (${formatBytes(file.size)})`;
      zipChip.style.display = 'inline-flex';
    }
    // Auto-fill site name if empty
    const nameInput = document.getElementById('siteName');
    if (nameInput && !nameInput.value) {
      nameInput.value = file.name.replace(/\.zip$/i, '').toLowerCase().replace(/[^a-z0-9-_]/g, '-');
    }
  }

  if (btnRemoveZip) {
    btnRemoveZip.addEventListener('click', (e) => {
      e.stopPropagation();
      selectedSiteZip = null;
      zipInput.value = '';
      if (zipChip) zipChip.style.display = 'none';
    });
  }

  // Multi-files & Folder Upload
  const filesDropzone = document.getElementById('siteFilesDropzone');
  const filesInput = document.getElementById('inputSiteFiles');
  const folderInput = document.getElementById('inputSiteFolder');
  const btnPickFiles = document.getElementById('btnPickFiles');
  const btnPickFolder = document.getElementById('btnPickFolder');
  const summaryBox = document.getElementById('filesPickedSummary');
  const summaryCount = document.getElementById('filesPickedCount');

  if (btnPickFiles && filesInput) {
    btnPickFiles.addEventListener('click', (e) => {
      e.stopPropagation();
      filesInput.click();
    });
  }
  if (btnPickFolder && folderInput) {
    btnPickFolder.addEventListener('click', (e) => {
      e.stopPropagation();
      folderInput.click();
    });
  }

  const handleFilesSelected = (fileList) => {
    selectedSiteFiles = Array.from(fileList);
    if (summaryBox && summaryCount) {
      summaryCount.textContent = selectedSiteFiles.length;
      summaryBox.style.display = 'block';
    }
    const nameInput = document.getElementById('siteName');
    if (nameInput && !nameInput.value && selectedSiteFiles.length > 0) {
      nameInput.value = 'static-site';
    }
  };

  if (filesInput) filesInput.addEventListener('change', (e) => handleFilesSelected(e.target.files));
  if (folderInput) folderInput.addEventListener('change', (e) => handleFilesSelected(e.target.files));

  if (filesDropzone) {
    filesDropzone.addEventListener('dragover', (e) => {
      e.preventDefault();
      filesDropzone.classList.add('dragover');
    });
    filesDropzone.addEventListener('dragleave', () => filesDropzone.classList.remove('dragover'));
    filesDropzone.addEventListener('drop', (e) => {
      e.preventDefault();
      filesDropzone.classList.remove('dragover');
      if (e.dataTransfer.files.length > 0) {
        handleFilesSelected(e.dataTransfer.files);
      }
    });
  }

  // Suggest Free Port Button
  const btnSuggest = document.getElementById('btnSuggestPort');
  if (btnSuggest) {
    btnSuggest.addEventListener('click', suggestFreePort);
  }

  // Form Submit: Deploy Static Website
  const deployForm = document.getElementById('deploySiteForm');
  if (deployForm) {
    deployForm.addEventListener('submit', handleDeploySite);
  }

  // Site Filter Buttons
  document.querySelectorAll('[data-site-filter]').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('[data-site-filter]').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      currentSiteFilter = btn.dataset.siteFilter;
      renderSites();
    });
  });

  // Explorer Modal Close Handlers
  const closeExplorer = document.getElementById('closeSiteFilesModal');
  const closeExplorerBtn = document.getElementById('closeSiteFilesModalBtn');
  const modalExplorer = document.getElementById('siteFilesModal');
  if (closeExplorer && modalExplorer) closeExplorer.addEventListener('click', () => modalExplorer.classList.remove('active'));
  if (closeExplorerBtn && modalExplorer) closeExplorerBtn.addEventListener('click', () => modalExplorer.classList.remove('active'));

  // Export Zip in Explorer
  const btnExport = document.getElementById('btnExportSiteZip');
  if (btnExport) {
    btnExport.addEventListener('click', () => {
      if (activeExplorerSite) {
        window.location.href = `/api/sites/${activeExplorerSite.id}/export`;
      }
    });
  }

  // Update Assets Modal
  const modalUpdate = document.getElementById('siteUpdateModal');
  const closeUpdate = document.getElementById('closeSiteUpdateModal');
  const cancelUpdate = document.getElementById('cancelSiteUpdateModal');
  const updateForm = document.getElementById('siteUpdateForm');
  const updateDropzone = document.getElementById('updateDropzone');
  const updateInput = document.getElementById('inputUpdateZip');
  const updateChip = document.getElementById('updateZipChip');
  const updateName = document.getElementById('updateZipName');

  if (closeUpdate && modalUpdate) closeUpdate.addEventListener('click', () => modalUpdate.classList.remove('active'));
  if (cancelUpdate && modalUpdate) cancelUpdate.addEventListener('click', () => modalUpdate.classList.remove('active'));
  if (updateDropzone && updateInput) {
    updateDropzone.addEventListener('click', () => updateInput.click());
    updateInput.addEventListener('change', (e) => {
      if (e.target.files.length > 0) {
        updateName.textContent = e.target.files[0].name;
        updateChip.style.display = 'inline-flex';
      }
    });
  }
  if (updateForm) updateForm.addEventListener('submit', handleSiteUpdate);
}

// Fetch all static sites from REST API
async function fetchSites() {
  try {
    const res = await fetch('/api/sites');
    if (!res.ok) return;
    sites = await res.json();
    updateSitesKPI();
    renderSites();
  } catch (err) {
    console.error('Error fetching sites:', err);
  }
}

function updateSitesKPI() {
  const runningSites = sites.filter(s => s.status === 'running');
  const statRunning = document.getElementById('statSitesRunning');
  const statTotal = document.getElementById('statSitesTotal');
  const statPorts = document.getElementById('statSitesPorts');
  const statDomains = document.getElementById('statSitesDomains');
  const statStorage = document.getElementById('statSitesStorage');
  const statFiles = document.getElementById('statSitesFiles');
  const badge = document.getElementById('navSitesBadge');
  const countBadge = document.getElementById('sitesCountBadge');
  const countAll = document.getElementById('countSitesAll');
  const countRun = document.getElementById('countSitesRunning');
  const countStop = document.getElementById('countSitesStopped');

  if (statRunning) statRunning.textContent = runningSites.length;
  if (statTotal) statTotal.textContent = `${sites.length} total hosted`;

  const uniquePorts = new Set(sites.map(s => s.port));
  if (statPorts) statPorts.textContent = uniquePorts.size;

  const withDomain = sites.filter(s => s.domain && s.domain !== '*');
  if (statDomains) statDomains.textContent = withDomain.length;

  let totalBytes = 0;
  let totalFiles = 0;
  sites.forEach(s => {
    totalBytes += (s.total_size || 0);
    totalFiles += (s.file_count || 0);
  });
  if (statStorage) statStorage.textContent = formatBytes(totalBytes);
  if (statFiles) statFiles.textContent = `${totalFiles} files across sites`;

  if (badge) badge.textContent = runningSites.length;
  if (countBadge) countBadge.textContent = `${sites.length} sites`;
  if (countAll) countAll.textContent = sites.length;
  if (countRun) countRun.textContent = runningSites.length;
  if (countStop) countStop.textContent = sites.length - runningSites.length;
}

function updateSiteFilterButtons() {
  document.querySelectorAll('[data-site-filter]').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.siteFilter === currentSiteFilter);
  });
}

function renderSites() {
  const container = document.getElementById('sitesContainer');
  const tableBody = document.getElementById('sitesTableBody');
  if (!container && !tableBody) return;

  // Filter sites
  let filtered = sites.filter(s => {
    if (currentSiteFilter === 'running') return s.status === 'running';
    if (currentSiteFilter === 'stopped') return s.status !== 'running';
    return true;
  });

  if (searchKeyword) {
    filtered = filtered.filter(s =>
      s.name.toLowerCase().includes(searchKeyword) ||
      s.id.toLowerCase().includes(searchKeyword) ||
      (s.domain && s.domain.toLowerCase().includes(searchKeyword)) ||
      String(s.port).includes(searchKeyword)
    );
  }

  if (filtered.length === 0) {
    const emptyHTML = `
      <div class="empty-state" style="grid-column: 1 / -1; padding: 4rem 1rem; text-align: center;">
        <div style="margin-bottom: 1rem; color: var(--accent-blue);">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
        </div>
        <h3 style="color: var(--text-primary); margin-bottom: 0.5rem;">No Static Websites Found</h3>
        <p style="color: var(--text-muted); max-width: 440px; margin: 0 auto 1.5rem;">
          Host HTML/static files or ZIP bundles with high-performance native Go serving and virtual host domain routing.
        </p>
        <button class="btn btn-primary" onclick="document.getElementById('deploySiteModal').classList.add('active'); suggestFreePort();">
          Deploy Your First Website
        </button>
      </div>`;
    if (container) container.innerHTML = emptyHTML;
    if (tableBody) tableBody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding: 3rem; color: var(--text-muted);">No websites found</td></tr>`;
    return;
  }

  // 1. Render Cards
  if (container) {
    container.innerHTML = filtered.map(site => {
      const isRunning = site.status === 'running';
      const statusPill = isRunning
        ? `<span class="stat-badge-pill green"><span class="pulse-dot"></span>Online</span>`
        : `<span class="stat-badge-pill" style="background:#334155; color:#94a3b8;">Stopped</span>`;

      const domainDisplay = site.domain && site.domain !== '*' ? site.domain : 'localhost';
      const accessURL = site.port === 80
        ? `http://${domainDisplay}`
        : `http://${domainDisplay}:${site.port}`;

      return `
        <div class="site-card">
          <div class="site-card-header">
            <div class="site-title-area">
              <div class="site-avatar">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="2" y1="12" x2="22" y2="12"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
              </div>
              <div>
                <div class="site-name-heading">${escapeHtml(site.name)}</div>
                <div class="site-id-sub">ID: ${escapeHtml(site.id)}</div>
              </div>
            </div>
            ${statusPill}
          </div>

          <div class="site-url-box">
            <a href="${accessURL}" target="_blank" rel="noopener" class="site-url-link" title="Open website in new tab">
              ${accessURL}
            </a>
            <div class="site-url-actions">
              <button class="icon-btn-sm" onclick="copyToClipboard('${accessURL}')" title="Copy URL">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
              </button>
              <a href="${accessURL}" target="_blank" rel="noopener" class="icon-btn-sm" title="Visit">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>
              </a>
            </div>
          </div>

          <div class="site-specs-row">
            <div class="site-spec-item">
              <span class="site-spec-label">Port</span>
              <span class="site-spec-val">${site.port}</span>
            </div>
            <div class="site-spec-item">
              <span class="site-spec-label">Domain</span>
              <span class="site-spec-val" style="font-size:11px; overflow:hidden; text-overflow:ellipsis;">${site.domain ? escapeHtml(site.domain) : '*'}</span>
            </div>
            <div class="site-spec-item">
              <span class="site-spec-label">Files / Size</span>
              <span class="site-spec-val" style="font-size:11px;">${site.file_count || 0} (${formatBytes(site.total_size || 0)})</span>
            </div>
          </div>

          <div class="site-feature-tags">
            <span class="site-tag ${site.spa_mode ? 'active' : ''}">SPA Fallback: ${site.spa_mode ? 'ON' : 'OFF'}</span>
            <span class="site-tag ${site.directory_listing ? 'active' : ''}">Dir Browsing: ${site.directory_listing ? 'ON' : 'OFF'}</span>
          </div>

          <div class="site-actions-footer">
            <div class="site-actions-left">
              ${isRunning
                ? `<button class="btn btn-secondary btn-sm" onclick="stopSite('${site.id}', '${escapeHtml(site.name)}')">Stop</button>`
                : `<button class="btn btn-primary btn-sm" onclick="startSite('${site.id}', '${escapeHtml(site.name)}')">Start</button>`
              }
              <button class="btn btn-secondary btn-sm" onclick="openSiteFilesModal('${site.id}', '${escapeHtml(site.name)}', ${site.port}, '${escapeHtml(site.root_dir)}')">Files</button>
              <button class="btn btn-secondary btn-sm" onclick="openSiteUpdateModal('${site.id}', '${escapeHtml(site.name)}')">Update</button>
            </div>
            <div class="site-actions-right">
              <a href="/api/sites/${site.id}/export" class="icon-btn-sm" title="Download .zip bundle">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/></svg>
              </a>
              <button class="icon-btn-sm" style="color:var(--accent-rose);" onclick="deleteSite('${site.id}', '${escapeHtml(site.name)}')" title="Delete site">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          </div>
        </div>`;
    }).join('');
  }

  // 2. Render Table Rows
  if (tableBody) {
    tableBody.innerHTML = filtered.map(site => {
      const isRunning = site.status === 'running';
      const statusPill = isRunning
        ? `<span class="stat-badge-pill green"><span class="pulse-dot"></span>Online</span>`
        : `<span class="stat-badge-pill" style="background:#334155; color:#94a3b8;">Stopped</span>`;

      const domainDisplay = site.domain && site.domain !== '*' ? site.domain : 'localhost';
      const accessURL = site.port === 80
        ? `http://${domainDisplay}`
        : `http://${domainDisplay}:${site.port}`;

      return `
        <tr>
          <td>
            <div style="display:flex; align-items:center; gap:8px;">
              <span style="display:inline-flex; align-items:center; color:var(--accent-blue);">${Icons.globe}</span>
              <strong>${escapeHtml(site.name)}</strong>
            </div>
          </td>
          <td>${statusPill}</td>
          <td><code>${site.port}</code></td>
          <td><code>${site.domain ? escapeHtml(site.domain) : '*'}</code></td>
          <td>${site.file_count || 0}</td>
          <td>${formatBytes(site.total_size || 0)}</td>
          <td><a href="${accessURL}" target="_blank" rel="noopener" style="color:var(--accent-cyan); font-family:var(--font-mono); font-size:12px;">${accessURL}</a></td>
          <td style="text-align:right;">
            <div style="display:inline-flex; gap:6px;">
              ${isRunning
                ? `<button class="btn btn-secondary btn-sm" onclick="stopSite('${site.id}', '${escapeHtml(site.name)}')">Stop</button>`
                : `<button class="btn btn-primary btn-sm" onclick="startSite('${site.id}', '${escapeHtml(site.name)}')">Start</button>`
              }
              <button class="btn btn-secondary btn-sm" onclick="openSiteFilesModal('${site.id}', '${escapeHtml(site.name)}', ${site.port}, '${escapeHtml(site.root_dir)}')">Files</button>
              <button class="btn btn-secondary btn-sm" onclick="openSiteUpdateModal('${site.id}', '${escapeHtml(site.name)}')">Update</button>
              <a href="/api/sites/${site.id}/export" class="btn btn-secondary btn-sm" title="Export .zip">Zip</a>
              <button class="btn btn-danger btn-sm" onclick="deleteSite('${site.id}', '${escapeHtml(site.name)}')">Delete</button>
            </div>
          </td>
        </tr>`;
    }).join('');
  }
}

// Deploy Static Site Form Submission
async function handleDeploySite(e) {
  e.preventDefault();

  const nameInput = document.getElementById('siteName');
  const portInput = document.getElementById('sitePort');
  const domainInput = document.getElementById('siteDomain');
  const spaInput = document.getElementById('siteSpaMode');
  const dirInput = document.getElementById('siteDirListing');
  const autoStartInput = document.getElementById('siteAutoStart');
  const btnSubmit = document.getElementById('btnSubmitDeploySite');

  const name = nameInput ? nameInput.value.trim() : '';
  const port = portInput ? parseInt(portInput.value, 10) : 8080;
  const domain = domainInput ? domainInput.value.trim() : '';
  const spa = spaInput ? spaInput.checked : false;
  const dirListing = dirInput ? dirInput.checked : false;
  const autoStart = autoStartInput ? autoStartInput.checked : true;

  if (!name) {
    showToast('Please specify a website name', 'error');
    return;
  }
  if (!port || port <= 0 || port > 65535) {
    showToast('Please specify a valid port (1-65535)', 'error');
    return;
  }

  btnSubmit.disabled = true;
  btnSubmit.innerHTML = '<span>Deploying Website...</span>';

  try {
    const formData = new FormData();
    formData.append('name', name);
    formData.append('port', String(port));
    formData.append('domain', domain);
    formData.append('spa', String(spa));
    formData.append('dir_listing', String(dirListing));
    formData.append('auto_start', String(autoStart));

    if (currentSiteTab === 'zip' && selectedSiteZip) {
      formData.append('bundle', selectedSiteZip);
    } else if (currentSiteTab === 'files' && selectedSiteFiles.length > 0) {
      selectedSiteFiles.forEach(f => {
        // preserve webkitRelativePath if folder was selected
        const path = f.webkitRelativePath || f.name;
        formData.append('files', f, path);
      });
    } else if (currentSiteTab === 'html') {
      const htmlCode = document.getElementById('inputSiteHtmlCode').value;
      formData.append('html', htmlCode);
    }

    const res = await fetch('/api/sites', {
      method: 'POST',
      body: formData
    });

    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to deploy website');

    showToast(`Static website '${data.name}' deployed on port ${data.port}!`, 'success');
    document.getElementById('deploySiteModal').classList.remove('active');

    // Reset form
    document.getElementById('deploySiteForm').reset();
    selectedSiteZip = null;
    selectedSiteFiles = [];
    const zipChip = document.getElementById('zipFileChip');
    if (zipChip) zipChip.style.display = 'none';
    const filesSummary = document.getElementById('filesPickedSummary');
    if (filesSummary) filesSummary.style.display = 'none';

    fetchSites();
  } catch (err) {
    showToast(`Deployment error: ${err.message}`, 'error');
  } finally {
    btnSubmit.disabled = false;
    btnSubmit.innerHTML = '<span>Deploy & Host Website</span>';
  }
}

async function startSite(siteId, siteName) {
  try {
    const res = await fetch(`/api/sites/${siteId}/start`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to start site');
    showToast(`Started site '${siteName}'`, 'success');
    fetchSites();
  } catch (err) {
    showToast(`Start error: ${err.message}`, 'error');
  }
}

async function stopSite(siteId, siteName) {
  try {
    const res = await fetch(`/api/sites/${siteId}/stop`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to stop site');
    showToast(`Stopped site '${siteName}'`, 'info');
    fetchSites();
  } catch (err) {
    showToast(`Stop error: ${err.message}`, 'error');
  }
}

async function deleteSite(siteId, siteName) {
  if (!confirm(`Are you sure you want to delete static site '${siteName}' and all its files?`)) {
    return;
  }
  try {
    const res = await fetch(`/api/sites/${siteId}`, { method: 'DELETE' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to delete site');
    showToast(`Deleted static site '${siteName}'`, 'info');
    fetchSites();
  } catch (err) {
    showToast(`Delete error: ${err.message}`, 'error');
  }
}

// File Explorer Modal
async function openSiteFilesModal(siteId, siteName, port, rootDir) {
  activeExplorerSite = { id: siteId, name: siteName, port: port, rootDir: rootDir };
  const title = document.getElementById('explorerSiteTitle');
  const meta = document.getElementById('explorerSiteMeta');
  const modal = document.getElementById('siteFilesModal');

  if (title) title.textContent = `Files: ${siteName}`;
  if (meta) meta.textContent = `Port ${port} • ${rootDir || 'Public Root'}`;
  if (modal) modal.classList.add('active');

  loadSiteFiles(siteId, '');
}

async function loadSiteFiles(siteId, subpath) {
  const fileList = document.getElementById('explorerFileList');
  const countBadge = document.getElementById('explorerFileCountBadge');
  const previewContent = document.getElementById('explorerPreviewContent');
  const previewName = document.getElementById('explorerSelectedFileName');
  const previewActions = document.getElementById('explorerPreviewActions');

  if (!fileList) return;
  fileList.innerHTML = '<div style="padding:16px; color:var(--text-muted); font-size:12px;">Loading files...</div>';
  if (previewContent) previewContent.innerHTML = '<div class="empty-preview-msg">Select a file from the list to preview</div>';
  if (previewName) previewName.textContent = 'No file selected';
  if (previewActions) previewActions.style.display = 'none';

  try {
    const url = subpath ? `/api/sites/${siteId}/files?subpath=${encodeURIComponent(subpath)}` : `/api/sites/${siteId}/files`;
    const res = await fetch(url);
    const files = await res.json();
    if (!res.ok) throw new Error('Failed to load files');

    if (countBadge) countBadge.textContent = `${files.length} items`;

    if (files.length === 0) {
      fileList.innerHTML = '<div style="padding:16px; color:var(--text-muted); font-size:12px;">No files found in this site.</div>';
      return;
    }

    fileList.innerHTML = files.map(f => {
      const iconSvg = getFileIconSvg(f.name, f.is_dir);
      const sizeDisplay = f.is_dir ? 'DIR' : formatBytes(f.size);
      const clickHandler = f.is_dir ? '' : `onclick="previewSiteFile('${siteId}', '${f.path}', '${f.name}', ${f.size})"`;

      return `
        <div class="explorer-file-item" ${clickHandler}>
          <span class="explorer-file-name">
            <span style="display:inline-flex; align-items:center; color:var(--text-secondary);">${iconSvg}</span>
            <span>${escapeHtml(f.name)}</span>
          </span>
          <span class="explorer-file-size">${sizeDisplay}</span>
        </div>`;
    }).join('');

    // Auto-preview index.html if present
    const indexFile = files.find(f => f.name === 'index.html');
    if (indexFile) {
      previewSiteFile(siteId, indexFile.path, indexFile.name, indexFile.size);
    }
  } catch (err) {
    fileList.innerHTML = `<div style="padding:16px; color:var(--accent-rose); font-size:12px;">Error: ${err.message}</div>`;
  }
}

async function previewSiteFile(siteId, path, name, size) {
  const previewName = document.getElementById('explorerSelectedFileName');
  const previewSize = document.getElementById('explorerFileSizeBadge');
  const previewActions = document.getElementById('explorerPreviewActions');
  const previewContent = document.getElementById('explorerPreviewContent');

  if (previewName) previewName.textContent = name;
  if (previewSize) previewSize.textContent = formatBytes(size);
  if (previewActions) previewActions.style.display = 'block';

  // Highlight active item in list
  document.querySelectorAll('.explorer-file-item').forEach(el => {
    el.classList.toggle('active', el.querySelector('.explorer-file-name span:last-child')?.textContent === name);
  });

  if (previewContent) {
    previewContent.innerHTML = '<div style="color:var(--text-muted);">Loading preview...</div>';

    // Image preview
    if (/\.(png|jpg|jpeg|svg|webp|gif|ico)$/i.test(name)) {
      previewContent.innerHTML = `
        <div style="display:flex; align-items:center; justify-content:center; height:100%; padding:20px;">
          <img src="/api/sites/${siteId}/files?path=${encodeURIComponent(path)}&content=true" style="max-width:100%; max-height:100%; border-radius:8px; box-shadow:var(--shadow-md);">
        </div>`;
      return;
    }

    try {
      const res = await fetch(`/api/sites/${siteId}/files?path=${encodeURIComponent(path)}`);
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Failed to read file');

      // Syntax display
      previewContent.textContent = data.content || '';
    } catch (err) {
      previewContent.innerHTML = `<div style="color:var(--accent-rose);">Failed to preview file: ${err.message}</div>`;
    }
  }
}

// Site Update Assets Modal
function openSiteUpdateModal(siteId, siteName) {
  const modal = document.getElementById('siteUpdateModal');
  const targetSpan = document.getElementById('updateSiteTarget');
  const idInput = document.getElementById('updateSiteID');
  const chip = document.getElementById('updateZipChip');
  const input = document.getElementById('inputUpdateZip');

  if (targetSpan) targetSpan.textContent = siteName;
  if (idInput) idInput.value = siteId;
  if (chip) chip.style.display = 'none';
  if (input) input.value = '';
  if (modal) modal.classList.add('active');
}

async function handleSiteUpdate(e) {
  e.preventDefault();
  const idInput = document.getElementById('updateSiteID');
  const zipInput = document.getElementById('inputUpdateZip');
  const btn = document.getElementById('btnSubmitSiteUpdate');

  const siteId = idInput ? idInput.value : '';
  if (!siteId) return;

  if (!zipInput || !zipInput.files || zipInput.files.length === 0) {
    showToast('Please select a .zip archive to upload', 'error');
    return;
  }

  btn.disabled = true;
  btn.textContent = 'Uploading assets...';

  try {
    const formData = new FormData();
    formData.append('bundle', zipInput.files[0]);

    const res = await fetch(`/api/sites/${siteId}/upload`, {
      method: 'POST',
      body: formData
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Update failed');

    showToast(`Updated assets for website '${data.name}'!`, 'success');
    document.getElementById('siteUpdateModal').classList.remove('active');
    fetchSites();
  } catch (err) {
    showToast(`Update error: ${err.message}`, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Upload & Redeploy';
  }
}

// Helper: suggest free port for static site
async function suggestFreePort() {
  const portInput = document.getElementById('sitePort');
  if (!portInput) return;

  // Find used ports in existing apps and sites
  const usedPorts = new Set();
  apps.forEach(a => {
    if (a.app.port > 0) usedPorts.add(a.app.port);
  });
  sites.forEach(s => {
    if (s.port > 0) usedPorts.add(s.port);
  });

  const candidates = [8080, 8081, 8082, 8085, 3000, 3001, 5000, 5173, 8000, 9000];
  for (const p of candidates) {
    if (!usedPorts.has(p)) {
      portInput.value = p;
      return;
    }
  }

  // Fallback: random between 8080 and 8999
  const randomPort = Math.floor(8080 + Math.random() * 900);
  portInput.value = randomPort;
}

// Helper: Copy URL to clipboard
function copyToClipboard(text) {
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).then(() => {
      showToast(`Copied URL: ${text}`, 'success');
    }).catch(() => {
      prompt('Copy URL:', text);
    });
  } else {
    prompt('Copy URL:', text);
  }
}

// Hook up initStaticWebsites during DOM ready
setTimeout(initStaticWebsites, 0);

// ============================================================================
// Cluster & Pods Distributed Mesh Logic
// ============================================================================

let clusterState = {
  status: null,
  nodes: [],
  deployments: [],
  pods: []
};
let currentPodFilter = 'all';
let activeLogPodID = null;
let podLogsAutoRefresh = null;

function initClusterMesh() {
  // Topbar and Section Action Buttons
  const btnTopDeploy = document.getElementById('btnDeployPodTop');
  const btnSectionDeploy = document.getElementById('btnOpenDeployPodModal');
  const btnTopSetup = document.getElementById('btnClusterSetupTop');
  const btnBannerSetup = document.getElementById('btnClusterSetupModal');
  const btnRefreshNodes = document.getElementById('btnRefreshNodes');
  const btnCopyJoin = document.getElementById('btnCopyJoinCmd');
  const btnLeave = document.getElementById('btnClusterLeave');

  // Modals
  const modalDeployPod = document.getElementById('deployPodModal');
  const closeDeployPod = document.getElementById('closeDeployPodModal');
  const cancelDeployPod = document.getElementById('cancelDeployPodModal');

  const modalSetup = document.getElementById('clusterSetupModal');
  const closeSetup = document.getElementById('closeClusterSetupModal');
  const cancelInit = document.getElementById('cancelClusterInit');
  const cancelJoin = document.getElementById('cancelClusterJoin');

  const modalLogs = document.getElementById('podLogsModal');
  const closeLogs = document.getElementById('closePodLogsModal');
  const closeLogsBtn = document.getElementById('closePodLogsBtn');
  const btnRefreshLogs = document.getElementById('btnRefreshPodLogs');
  const selectLogLines = document.getElementById('selectPodLogLines');

  // Modal Open Handlers
  const openDeployPod = () => {
    if (modalDeployPod) modalDeployPod.classList.add('active');
  };
  const openSetup = () => {
    if (modalSetup) modalSetup.classList.add('active');
  };

  if (btnTopDeploy) btnTopDeploy.addEventListener('click', openDeployPod);
  if (btnSectionDeploy) btnSectionDeploy.addEventListener('click', openDeployPod);
  if (btnTopSetup) btnTopSetup.addEventListener('click', openSetup);
  if (btnBannerSetup) btnBannerSetup.addEventListener('click', openSetup);

  if (btnRefreshNodes) btnRefreshNodes.addEventListener('click', () => {
    fetchClusterData();
    showToast('Refreshed cluster nodes', 'info');
  });

  // Modal Close Handlers
  if (closeDeployPod) closeDeployPod.addEventListener('click', () => modalDeployPod.classList.remove('active'));
  if (cancelDeployPod) cancelDeployPod.addEventListener('click', () => modalDeployPod.classList.remove('active'));

  if (closeSetup) closeSetup.addEventListener('click', () => modalSetup.classList.remove('active'));
  if (cancelInit) cancelInit.addEventListener('click', () => modalSetup.classList.remove('active'));
  if (cancelJoin) cancelJoin.addEventListener('click', () => modalSetup.classList.remove('active'));

  const closeLogsHandler = () => {
    if (modalLogs) modalLogs.classList.remove('active');
    if (podLogsAutoRefresh) {
      clearInterval(podLogsAutoRefresh);
      podLogsAutoRefresh = null;
    }
  };
  if (closeLogs) closeLogs.addEventListener('click', closeLogsHandler);
  if (closeLogsBtn) closeLogsBtn.addEventListener('click', closeLogsHandler);

  if (btnRefreshLogs) {
    btnRefreshLogs.addEventListener('click', () => {
      if (activeLogPodID) fetchPodLogs(activeLogPodID);
    });
  }

  if (selectLogLines) {
    selectLogLines.addEventListener('change', () => {
      if (activeLogPodID) fetchPodLogs(activeLogPodID);
    });
  }

  // Setup Modal Tabs (Init vs Join)
  const tabInit = document.getElementById('tabClusterInit');
  const tabJoin = document.getElementById('tabClusterJoin');
  const formInit = document.getElementById('clusterInitForm');
  const formJoin = document.getElementById('clusterJoinForm');

  if (tabInit && tabJoin) {
    tabInit.addEventListener('click', () => {
      tabInit.classList.add('active');
      tabJoin.classList.remove('active');
      if (formInit) formInit.style.display = 'block';
      if (formJoin) formJoin.style.display = 'none';
    });
    tabJoin.addEventListener('click', () => {
      tabJoin.classList.add('active');
      tabInit.classList.remove('active');
      if (formInit) formInit.style.display = 'none';
      if (formJoin) formJoin.style.display = 'block';
    });
  }

  // Deploy Pod Replicas Slider / Stepper sync
  const inputReplicas = document.getElementById('inputPodReplicas');
  const sliderReplicas = document.getElementById('sliderPodReplicas');
  const badgeReplicas = document.getElementById('badgePodReplicasPreview');

  const updateReplicasUI = (val) => {
    if (inputReplicas) inputReplicas.value = val;
    if (sliderReplicas) sliderReplicas.value = val;
    if (badgeReplicas) badgeReplicas.textContent = `${val} Replicas`;
  };

  if (inputReplicas) {
    inputReplicas.addEventListener('input', (e) => updateReplicasUI(e.target.value));
  }
  if (sliderReplicas) {
    sliderReplicas.addEventListener('input', (e) => updateReplicasUI(e.target.value));
  }

  // Copy Join Command
  if (btnCopyJoin) {
    btnCopyJoin.addEventListener('click', () => {
      const cmdEl = document.getElementById('clusterJoinCmd');
      if (cmdEl) {
        copyToClipboard(cmdEl.textContent);
      }
    });
  }

  // Cluster Leave
  if (btnLeave) {
    btnLeave.addEventListener('click', () => {
      if (confirm('Disconnect this node and leave the cluster? Local pods will be terminated.')) {
        fetch('/api/cluster/leave', { method: 'POST' })
          .then(res => res.json())
          .then(() => {
            showToast('Left cluster successfully', 'success');
            fetchClusterData();
          })
          .catch(err => showToast(`Error leaving cluster: ${err}`, 'error'));
      }
    });
  }

  // Cluster Init Form Submit
  if (formInit) {
    formInit.addEventListener('submit', (e) => {
      e.preventDefault();
      const name = document.getElementById('inputClusterName').value.trim();
      const port = parseInt(document.getElementById('inputClusterPort').value.trim(), 10) || 9090;

      fetch('/api/cluster/init', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, port })
      })
      .then(res => res.json())
      .then(data => {
        if (data.error) throw new Error(data.error);
        showToast(`Cluster '${name}' initialized as Control Plane!`, 'success');
        modalSetup.classList.remove('active');
        fetchClusterData();
      })
      .catch(err => showToast(`Cluster init failed: ${err.message}`, 'error'));
    });
  }

  // Cluster Join Form Submit
  if (formJoin) {
    formJoin.addEventListener('submit', (e) => {
      e.preventDefault();
      const leader_addr = document.getElementById('inputJoinLeaderAddr').value.trim();
      const token = document.getElementById('inputJoinToken').value.trim();
      const node_name = document.getElementById('inputJoinNodeName').value.trim();

      fetch('/api/cluster/join', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ leader_addr, token, node_name })
      })
      .then(res => res.json())
      .then(data => {
        if (data.error) throw new Error(data.error);
        showToast('Successfully joined cluster mesh as Worker node!', 'success');
        modalSetup.classList.remove('active');
        fetchClusterData();
      })
      .catch(err => showToast(`Failed to join cluster: ${err.message}`, 'error'));
    });
  }

  // Deploy Pod Form Submit
  const deployPodForm = document.getElementById('deployPodForm');
  if (deployPodForm) {
    deployPodForm.addEventListener('submit', (e) => {
      e.preventDefault();
      const name = document.getElementById('inputPodName').value.trim();
      const replicas = parseInt(document.getElementById('inputPodReplicas').value, 10) || 1;
      const command = document.getElementById('inputPodCommand').value.trim();
      const port = parseInt(document.getElementById('inputPodPort').value, 10) || 0;
      const service_port = parseInt(document.getElementById('inputPodServicePort').value, 10) || 0;
      const engine = document.getElementById('selectPodEngine').value;
      const selectorRaw = document.getElementById('inputPodSelector').value.trim();
      const envRaw = document.getElementById('inputPodEnv').value.trim();

      const env = {};
      if (envRaw) {
        envRaw.split('\n').forEach(line => {
          const idx = line.indexOf('=');
          if (idx > 0) {
            env[line.slice(0, idx).trim()] = line.slice(idx + 1).trim();
          }
        });
      }

      const selector = {};
      if (selectorRaw) {
        selectorRaw.split(',').forEach(part => {
          const idx = part.indexOf('=');
          if (idx > 0) {
            selector[part.slice(0, idx).trim()] = part.slice(idx + 1).trim();
          }
        });
      }

      const payload = {
        name: name || undefined,
        namespace: 'default',
        replicas,
        service_port,
        template: {
          name: name || 'pod-app',
          command,
          port,
          engine,
          env,
          node_selector: Object.keys(selector).length > 0 ? selector : undefined,
          restart_policy: 'always'
        }
      };

      fetch('/api/cluster/deployments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      })
      .then(res => res.json())
      .then(data => {
        if (data.error) throw new Error(data.error);
        showToast(`Deployment created! Scaling ${replicas} replicas across nodes...`, 'success');
        modalDeployPod.classList.remove('active');
        deployPodForm.reset();
        fetchClusterData();
      })
      .catch(err => showToast(`Deployment failed: ${err.message}`, 'error'));
    });
  }

  // Pods Filter Buttons
  const filterBtns = ['All', 'Running', 'Pending', 'Failed'];
  filterBtns.forEach(f => {
    const btn = document.getElementById(`btnPodsFilter${f}`);
    if (btn) {
      btn.addEventListener('click', () => {
        currentPodFilter = f.toLowerCase();
        updatePodFilterButtons();
        renderClusterPods(clusterState.pods);
      });
    }
  });

  // Background refresh interval
  setInterval(() => {
    if (currentWorkspace === 'cluster') {
      fetchClusterData(true);
    }
  }, 5000);

  // Initial load
  fetchClusterData();
}

function updatePodFilterButtons() {
  const filterBtns = ['all', 'running', 'pending', 'failed'];
  filterBtns.forEach(f => {
    const btn = document.getElementById(`btnPodsFilter${f.charAt(0).toUpperCase() + f.slice(1)}`);
    if (btn) {
      if (currentPodFilter === f) {
        btn.classList.add('active');
      } else {
        btn.classList.remove('active');
      }
    }
  });
}

function fetchClusterData(silent = false) {
  Promise.all([
    fetch('/api/cluster/status').then(r => r.json()).catch(() => ({ initialized: false })),
    fetch('/api/cluster/nodes').then(r => r.json()).catch(() => []),
    fetch('/api/cluster/deployments').then(r => r.json()).catch(() => []),
    fetch('/api/cluster/pods').then(r => r.json()).catch(() => [])
  ])
  .then(([status, nodes, deployments, pods]) => {
    clusterState = {
      status,
      nodes: Array.isArray(nodes) ? nodes : [],
      deployments: Array.isArray(deployments) ? deployments : [],
      pods: Array.isArray(pods) ? pods : []
    };

    // Update Nav Badge
    const navBadge = document.getElementById('navClusterBadge');
    if (navBadge) {
      navBadge.textContent = clusterState.nodes.length;
    }

    renderClusterBanner(clusterState.status);
    renderClusterKPIs(clusterState.status, clusterState.nodes, clusterState.deployments, clusterState.pods);
    renderClusterNodes(clusterState.nodes);
    renderClusterDeployments(clusterState.deployments, clusterState.pods);
    renderClusterPods(clusterState.pods);
  })
  .catch(err => {
    if (!silent) console.error('fetchClusterData error:', err);
  });
}

function renderClusterBanner(status) {
  const nameEl = document.getElementById('clusterBannerName');
  const roleBadge = document.getElementById('clusterRoleBadge');
  const subtitleEl = document.getElementById('clusterBannerSubtitle');
  const leaderAddr = document.getElementById('clusterLeaderAddr');
  const joinBox = document.getElementById('clusterJoinBox');
  const joinCmd = document.getElementById('clusterJoinCmd');
  const btnLeave = document.getElementById('btnClusterLeave');
  const onlineDot = document.getElementById('clusterOnlineDot');

  if (!status || !status.initialized) {
    if (nameEl) nameEl.textContent = 'Cluster Mesh (Offline)';
    if (roleBadge) {
      roleBadge.textContent = 'Standalone';
      roleBadge.className = 'cluster-role-pill offline';
    }
    if (subtitleEl) subtitleEl.innerHTML = 'Initialize a control-plane or join an existing Docroid cluster';
    if (joinBox) joinBox.style.display = 'none';
    if (btnLeave) btnLeave.style.display = 'none';
    if (onlineDot) onlineDot.className = 'status-indicator stopped';
    return;
  }

  if (onlineDot) onlineDot.className = 'status-indicator online';

  if (nameEl) nameEl.textContent = status.cluster_name || 'docroid-mesh';
  if (roleBadge) {
    roleBadge.textContent = status.role === 'control-plane' ? 'Control Plane' : 'Worker Node';
    roleBadge.className = `cluster-role-pill ${status.role}`;
  }

  if (leaderAddr) leaderAddr.textContent = status.leader_addr || '127.0.0.1:9090';
  if (btnLeave) btnLeave.style.display = 'inline-flex';

  // If Control Plane and join token available, show join command
  if (status.role === 'control-plane' && status.join_token && joinBox && joinCmd) {
    joinBox.style.display = 'flex';
    const leaderHost = status.leader_addr || `${window.location.hostname}:${window.location.port || 9090}`;
    joinCmd.textContent = `docroid cluster join http://${leaderHost} --token ${status.join_token}`;
  } else if (joinBox) {
    joinBox.style.display = 'none';
  }
}

function renderClusterKPIs(status, nodes, deployments, pods) {
  const statNodes = document.getElementById('statClusterNodes');
  const statNodesSub = document.getElementById('statClusterNodesSub');
  const statPods = document.getElementById('statClusterPods');
  const statPodsSub = document.getElementById('statClusterPodsSub');
  const statReplicas = document.getElementById('statClusterReplicas');
  const statReplicasSub = document.getElementById('statClusterReplicasSub');
  const statCompute = document.getElementById('statClusterCompute');
  const statMemory = document.getElementById('statClusterMemory');

  const readyNodes = nodes.filter(n => n.status === 'Ready').length;
  if (statNodes) statNodes.textContent = `${readyNodes} / ${nodes.length}`;
  if (statNodesSub) statNodesSub.textContent = `${readyNodes} Healthy devices`;

  const runningPods = pods.filter(p => p.status === 'Running').length;
  if (statPods) statPods.textContent = `${runningPods} / ${pods.length}`;
  if (statPodsSub) statPodsSub.textContent = `${runningPods} Running across nodes`;

  const totalDesired = deployments.reduce((acc, d) => acc + (d.replicas || 0), 0);
  if (statReplicas) statReplicas.textContent = `${totalDesired}`;
  if (statReplicasSub) statReplicasSub.textContent = `${deployments.length} Active services`;

  let totalCores = 0;
  let totalMem = 0;
  nodes.forEach(n => {
    if (n.capacity) {
      totalCores += n.capacity.cpu_cores || 0;
      totalMem += n.capacity.memory_mb || 0;
    }
  });
  if (statCompute) statCompute.textContent = `${totalCores || 4} Cores`;
  if (statMemory) statMemory.textContent = `${totalMem || 4096} MB Aggregate Memory`;

  // Count Badges
  const nodesBadge = document.getElementById('nodesCountBadge');
  if (nodesBadge) nodesBadge.textContent = `${nodes.length} Devices`;

  const depsBadge = document.getElementById('deploymentsCountBadge');
  if (depsBadge) depsBadge.textContent = `${deployments.length} Services`;

  const podsBadge = document.getElementById('podsCountBadge');
  if (podsBadge) podsBadge.textContent = `${pods.length} Pods`;
}

function renderClusterNodes(nodes) {
  const container = document.getElementById('clusterNodesGrid');
  if (!container) return;

  if (nodes.length === 0) {
    container.innerHTML = `
      <div class="empty-preview-msg" style="grid-column: 1 / -1; padding: 2rem;">
        No nodes connected yet. Click 'Cluster Setup' above to initialize or join a mesh.
      </div>`;
    return;
  }

  container.innerHTML = nodes.map(node => {
    let iconSvg = Icons.monitor;
    if (node.os === 'android') iconSvg = Icons.smartphone;
    else if (node.os === 'darwin') iconSvg = Icons.laptop;

    const isReady = node.status === 'Ready';
    const statusClass = isReady ? 'running' : 'stopped';
    const isControlPlane = node.role === 'control-plane';

    return `
      <div class="device-node-card">
        <div class="device-node-header">
          <div class="device-node-identity">
            <div class="device-type-icon">${iconSvg}</div>
            <div>
              <div class="device-node-name">${escapeHtml(node.name || 'node')} ${node.is_local ? '<span class="device-chip" style="color:#60a5fa;">Local</span>' : ''}</div>
              <div class="device-node-addr">${escapeHtml(node.address || '')}</div>
            </div>
          </div>
          <span class="status-badge ${statusClass}">
            <span class="status-dot"></span>
            ${node.status}
          </span>
        </div>

        <div class="device-node-chips">
          <span class="cluster-role-pill ${node.role}">${isControlPlane ? 'Control Plane' : 'Worker'}</span>
          <span class="device-chip">${escapeHtml(node.os || '')} / ${escapeHtml(node.arch || '')}</span>
          ${node.labels && node.labels['device.docroid/type'] ? `<span class="device-chip">${node.labels['device.docroid/type']}</span>` : ''}
        </div>

        <div class="device-node-stats">
          <div>
            <div class="device-stat-val">${node.capacity ? node.capacity.cpu_cores : 2}</div>
            <div class="device-stat-lbl">CPUs</div>
          </div>
          <div>
            <div class="device-stat-val">${node.capacity ? (node.capacity.memory_mb / 1024).toFixed(1) : 2}G</div>
            <div class="device-stat-lbl">RAM</div>
          </div>
          <div>
            <div class="device-stat-val" style="color: #60a5fa;">${node.allocated_pods || 0}</div>
            <div class="device-stat-lbl">Pods</div>
          </div>
        </div>

        <div class="device-node-footer">
          <span>Joined: ${formatRelativeTime(node.joined_at)}</span>
          ${!node.is_local ? `
            <button class="btn btn-sm btn-icon" title="Remove Node" onclick="removeClusterNode('${node.id}')" style="color:var(--accent-rose);">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>` : ''}
        </div>
      </div>
    `;
  }).join('');
}

function removeClusterNode(nodeID) {
  if (confirm(`Remove node ${nodeID} from the cluster? Running pods on this node will be rescheduled.`)) {
    fetch(`/api/cluster/nodes/${encodeURIComponent(nodeID)}`, { method: 'DELETE' })
      .then(res => res.json())
      .then(() => {
        showToast(`Node ${nodeID} removed`, 'info');
        fetchClusterData();
      })
      .catch(err => showToast(`Error removing node: ${err}`, 'error'));
  }
}

function renderClusterDeployments(deployments, pods) {
  const container = document.getElementById('clusterDeploymentsContainer');
  if (!container) return;

  if (deployments.length === 0) {
    container.innerHTML = `
      <div class="empty-preview-msg" style="padding: 2rem;">
        No scalable deployments created. Click '+ Deploy Service / Pod' to launch multi-replica workloads across devices.
      </div>`;
    return;
  }

  container.innerHTML = deployments.map(dep => {
    // Count active running pods for this deployment
    const depPods = pods.filter(p => p.deployment === dep.name);
    const runningCount = depPods.filter(p => p.status === 'Running').length;
    const targetCount = dep.replicas || 0;

    // Distribution across devices
    const nodeCounts = {};
    depPods.forEach(p => {
      const nodeName = p.node_name || p.node_id || 'unknown';
      nodeCounts[nodeName] = (nodeCounts[nodeName] || 0) + 1;
    });

    const distributionHtml = Object.entries(nodeCounts).map(([nodeName, count]) => {
      return `<span class="distribution-chip">${Icons.phone} <span>${escapeHtml(nodeName)}: <strong>${count}</strong></span></span>`;
    }).join('');

    const isHealthy = runningCount === targetCount && targetCount > 0;
    const statusText = `${runningCount}/${targetCount} Ready`;

    return `
      <div class="deployment-card">
        <div class="deployment-card-top">
          <div class="deployment-meta">
            <h4>
              <span style="display:inline-flex; align-items:center; gap:6px; color:var(--text-primary);">${Icons.cluster} <span>${escapeHtml(dep.name)}</span></span>
              <span class="badge" style="font-weight:normal; font-size:11px;">${escapeHtml(dep.namespace || 'default')}</span>
            </h4>
            <div class="deployment-cmd">$ ${escapeHtml(dep.template.command)} ${escapeHtml((dep.template.args || []).join(' '))}</div>
          </div>

          <div style="display:flex; align-items:center; gap:0.75rem;">
            ${dep.service_port ? `
              <div class="deployment-service-badge">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
                <a href="http://${window.location.hostname}:${dep.service_port}" target="_blank">Cluster Service :${dep.service_port}</a>
              </div>` : ''}
            <span class="status-badge ${isHealthy ? 'running' : 'stopped'}">${statusText}</span>
            <button class="btn btn-sm btn-icon" title="Delete Deployment" onclick="deleteClusterDeployment('${escapeHtml(dep.name)}')">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>
          </div>
        </div>

        <!-- Node Distribution -->
        <div class="deployment-distribution">
          <span style="color:var(--text-muted);">Replicas Placed:</span>
          ${distributionHtml || '<span style="color:var(--text-muted);">No pods scheduled</span>'}
        </div>

        <!-- Real-Time Horizontal Scaling Bar -->
        <div class="scaling-control-panel">
          <div class="scaling-info">
            <span style="font-size:0.85rem; font-weight:600; color:var(--text-primary);">Horizontal Scale:</span>
            <span class="scaling-replica-badge" id="badgeReplica_${escapeHtml(dep.name)}">${targetCount} Replicas</span>
          </div>

          <div class="scaling-interactive-controls">
            <div class="scaling-stepper">
              <button type="button" onclick="adjustScale('${escapeHtml(dep.name)}', -1)">−</button>
              <input type="number" id="scaleInput_${escapeHtml(dep.name)}" min="0" max="20" value="${targetCount}" oninput="syncScaleSlider('${escapeHtml(dep.name)}', this.value)">
              <button type="button" onclick="adjustScale('${escapeHtml(dep.name)}', 1)">+</button>
            </div>
            <input type="range" class="scaling-slider" id="scaleSlider_${escapeHtml(dep.name)}" min="0" max="20" value="${targetCount}" oninput="syncScaleInput('${escapeHtml(dep.name)}', this.value)">
            <button class="btn btn-sm btn-primary" onclick="applyDeploymentScale('${escapeHtml(dep.name)}')">
              Scale
            </button>
          </div>
        </div>
      </div>
    `;
  }).join('');
}

function syncScaleSlider(name, val) {
  const slider = document.getElementById(`scaleSlider_${name}`);
  const badge = document.getElementById(`badgeReplica_${name}`);
  if (slider) slider.value = val;
  if (badge) badge.textContent = `${val} Replicas`;
}

function syncScaleInput(name, val) {
  const input = document.getElementById(`scaleInput_${name}`);
  const badge = document.getElementById(`badgeReplica_${name}`);
  if (input) input.value = val;
  if (badge) badge.textContent = `${val} Replicas`;
}

function adjustScale(name, delta) {
  const input = document.getElementById(`scaleInput_${name}`);
  if (!input) return;
  let val = parseInt(input.value, 10) || 0;
  val = Math.max(0, Math.min(20, val + delta));
  input.value = val;
  syncScaleSlider(name, val);
}

function applyDeploymentScale(name) {
  const input = document.getElementById(`scaleInput_${name}`);
  if (!input) return;
  const count = parseInt(input.value, 10) || 0;

  fetch('/api/cluster/scale', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deployment: name, replicas: count })
  })
  .then(res => res.json())
  .then(data => {
    if (data.error) throw new Error(data.error);
    showToast(`Scaling ${name} to ${count} replicas...`, 'success');
    fetchClusterData();
  })
  .catch(err => showToast(`Scaling failed: ${err.message}`, 'error'));
}

function deleteClusterDeployment(name) {
  if (confirm(`Delete deployment '${name}' and terminate all associated pods across all devices?`)) {
    fetch(`/api/cluster/deployments/${encodeURIComponent(name)}`, { method: 'DELETE' })
      .then(res => res.json())
      .then(() => {
        showToast(`Deployment ${name} deleted`, 'info');
        fetchClusterData();
      })
      .catch(err => showToast(`Failed to delete deployment: ${err}`, 'error'));
  }
}

function renderClusterPods(pods) {
  const tbody = document.getElementById('clusterPodsTableBody');
  if (!tbody) return;

  // Filter pods
  let filtered = pods;
  if (currentPodFilter !== 'all') {
    filtered = pods.filter(p => (p.status || '').toLowerCase() === currentPodFilter);
  }

  // Update filter count numbers
  const allCount = pods.length;
  const runningCount = pods.filter(p => (p.status || '').toLowerCase() === 'running').length;
  const pendingCount = pods.filter(p => (p.status || '').toLowerCase() === 'pending').length;
  const failedCount = pods.filter(p => (p.status || '').toLowerCase() === 'failed').length;

  const countAll = document.getElementById('countPodsAll');
  if (countAll) countAll.textContent = allCount;
  const countRunning = document.getElementById('countPodsRunning');
  if (countRunning) countRunning.textContent = runningCount;
  const countPending = document.getElementById('countPodsPending');
  if (countPending) countPending.textContent = pendingCount;
  const countFailed = document.getElementById('countPodsFailed');
  if (countFailed) countFailed.textContent = failedCount;

  if (filtered.length === 0) {
    tbody.innerHTML = `
      <tr>
        <td colspan="8" style="text-align: center; color: var(--text-muted); padding: 2rem;">
          No pods matching filter '${currentPodFilter}'
        </td>
      </tr>`;
    return;
  }

  tbody.innerHTML = filtered.map(pod => {
    const statusLow = (pod.status || 'pending').toLowerCase();
    const shortID = pod.id ? pod.id.slice(-8) : '';
    const age = formatRelativeTime(pod.created_at);

    return `
      <tr>
        <td>
          <div style="font-weight: 600; color: var(--text-primary);">${escapeHtml(pod.name || pod.id)}</div>
          <div style="font-family: var(--font-mono); font-size: 11px; color: var(--text-muted);">${escapeHtml(shortID)}</div>
        </td>
        <td>
          <span class="badge" style="font-size: 11px;">${escapeHtml(pod.deployment || 'standalone')}</span>
        </td>
        <td>
          <div style="font-weight: 500; display:flex; align-items:center; gap:6px;">${Icons.phone} <span>${escapeHtml(pod.node_name || 'node')}</span></div>
          <div style="font-size: 11px; color: var(--text-muted); font-family: var(--font-mono);">${escapeHtml(pod.node_address || '')}</div>
        </td>
        <td>
          <span class="pod-status-pill ${statusLow}">
            <span class="status-dot"></span>
            ${pod.status}
          </span>
        </td>
        <td style="font-family: var(--font-mono); font-size: 12px;">
          ${pod.host_port ? `:${pod.host_port}` : '—'}
        </td>
        <td>${pod.restart_count || 0}</td>
        <td>${age}</td>
        <td style="text-align: right;">
          <button class="btn btn-sm btn-secondary" onclick="viewPodLogs('${pod.id}', '${escapeHtml(pod.name)}')">
            Logs
          </button>
          <button class="btn btn-sm btn-outline-danger" onclick="deleteClusterPod('${pod.id}')" title="Terminate Pod">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="vertical-align:-1px;"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </td>
      </tr>
    `;
  }).join('');
}

function deleteClusterPod(podID) {
  if (confirm(`Terminate pod ${podID}?`)) {
    fetch(`/api/cluster/pods/${encodeURIComponent(podID)}`, { method: 'DELETE' })
      .then(res => res.json())
      .then(() => {
        showToast(`Pod terminated`, 'info');
        fetchClusterData();
      })
      .catch(err => showToast(`Failed to terminate pod: ${err}`, 'error'));
  }
}

function viewPodLogs(podID, podName) {
  activeLogPodID = podID;
  const modal = document.getElementById('podLogsModal');
  const subtitle = document.getElementById('podLogsSubtitle');
  if (subtitle) subtitle.textContent = `Viewing logs for ${podName || podID}`;
  if (modal) modal.classList.add('active');

  fetchPodLogs(podID);

  if (podLogsAutoRefresh) clearInterval(podLogsAutoRefresh);
  podLogsAutoRefresh = setInterval(() => {
    if (activeLogPodID) fetchPodLogs(activeLogPodID, true);
  }, 3000);
}

function fetchPodLogs(podID, silent = false) {
  const content = document.getElementById('podLogsContent');
  const selectLines = document.getElementById('selectPodLogLines');
  const lines = selectLines ? selectLines.value : 100;

  fetch(`/api/cluster/pods/${encodeURIComponent(podID)}/logs?lines=${lines}`)
    .then(res => res.text())
    .then(text => {
      if (content) {
        content.textContent = text || '(Pod log stream is empty)';
        content.scrollTop = content.scrollHeight;
      }
    })
    .catch(err => {
      if (!silent && content) content.textContent = `Error fetching logs: ${err}`;
    });
}

// Hook up initClusterMesh during DOM ready
setTimeout(initClusterMesh, 0);

/* ====================================================
   Virtual Networks & Inter-Container Controller
==================================================== */

function initNetworks() {
  const btnTopCreate = document.getElementById('btnCreateNetworkTop');
  const btnSectionCreate = document.getElementById('btnOpenCreateNetworkModal');
  const modalCreate = document.getElementById('createNetworkModal');
  const closeCreate = document.getElementById('closeCreateNetworkModal');
  const cancelCreate = document.getElementById('cancelCreateNetworkModal');
  const formCreate = document.getElementById('createNetworkForm');

  const openCreateModal = () => {
    if (formCreate) formCreate.reset();
    if (modalCreate) modalCreate.classList.add('active');
  };

  const closeCreateModal = () => {
    if (modalCreate) modalCreate.classList.remove('active');
  };

  if (btnTopCreate) btnTopCreate.addEventListener('click', openCreateModal);
  if (btnSectionCreate) btnSectionCreate.addEventListener('click', openCreateModal);
  if (closeCreate) closeCreate.addEventListener('click', closeCreateModal);
  if (cancelCreate) cancelCreate.addEventListener('click', closeCreateModal);

  if (formCreate) {
    formCreate.addEventListener('submit', async (e) => {
      e.preventDefault();
      const btn = document.getElementById('btnSubmitCreateNetwork');
      if (btn) {
        btn.disabled = true;
        btn.textContent = 'Creating...';
      }

      const name = document.getElementById('inputNetworkName').value.trim();
      const driver = document.getElementById('selectNetworkDriver').value;
      const subnet = document.getElementById('inputNetworkSubnet').value.trim();

      try {
        const res = await fetch('/api/networks', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name, driver, subnet })
        });
        const data = await res.json();
        if (!res.ok) {
          throw new Error(data.error || 'Failed to create network');
        }
        showToast(`Virtual network '${data.name}' created successfully!`, 'success');
        closeCreateModal();
        fetchNetworks();
      } catch (err) {
        showToast(err.message, 'error');
      } finally {
        if (btn) {
          btn.disabled = false;
          btn.textContent = 'Create Network';
        }
      }
    });
  }

  // Connect Container Modal
  const modalConnect = document.getElementById('connectContainerModal');
  const closeConnect = document.getElementById('closeConnectContainerModal');
  const cancelConnect = document.getElementById('cancelConnectContainerModal');
  const formConnect = document.getElementById('connectContainerForm');

  const closeConnectModal = () => {
    if (modalConnect) modalConnect.classList.remove('active');
  };

  if (closeConnect) closeConnect.addEventListener('click', closeConnectModal);
  if (cancelConnect) cancelConnect.addEventListener('click', closeConnectModal);

  if (formConnect) {
    formConnect.addEventListener('submit', async (e) => {
      e.preventDefault();
      const btn = document.getElementById('btnSubmitConnectContainer');
      if (btn) {
        btn.disabled = true;
        btn.textContent = 'Connecting...';
      }

      const networkID = document.getElementById('connectTargetNetworkID').value;
      const containerID = document.getElementById('selectConnectContainer').value;
      const aliasesRaw = document.getElementById('inputConnectAliases').value.trim();
      const aliases = aliasesRaw ? aliasesRaw.split(',').map(s => s.trim()).filter(Boolean) : [];

      try {
        const res = await fetch(`/api/networks/${encodeURIComponent(networkID)}/connect`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ container_id: containerID, aliases })
        });
        const data = await res.json();
        if (!res.ok) {
          throw new Error(data.error || 'Failed to connect container');
        }
        showToast(`Container connected with virtual IP: ${data.ipv4_address}`, 'success');
        closeConnectModal();
        fetchNetworks();
      } catch (err) {
        showToast(err.message, 'error');
      } finally {
        if (btn) {
          btn.disabled = false;
          btn.textContent = 'Connect';
        }
      }
    });
  }

  // Network filter buttons
  const btnFilterAll = document.getElementById('btnNetFilterAll');
  const btnFilterBridge = document.getElementById('btnNetFilterBridge');
  const btnFilterHost = document.getElementById('btnNetFilterHost');

  if (btnFilterAll) {
    btnFilterAll.addEventListener('click', () => {
      currentNetFilter = 'all';
      updateNetFilterButtons();
      renderNetworks();
    });
  }
  if (btnFilterBridge) {
    btnFilterBridge.addEventListener('click', () => {
      currentNetFilter = 'bridge';
      updateNetFilterButtons();
      renderNetworks();
    });
  }
  if (btnFilterHost) {
    btnFilterHost.addEventListener('click', () => {
      currentNetFilter = 'host';
      updateNetFilterButtons();
      renderNetworks();
    });
  }
}

function updateNetFilterButtons() {
  const btnFilterAll = document.getElementById('btnNetFilterAll');
  const btnFilterBridge = document.getElementById('btnNetFilterBridge');
  const btnFilterHost = document.getElementById('btnNetFilterHost');

  if (btnFilterAll) btnFilterAll.classList.toggle('active', currentNetFilter === 'all');
  if (btnFilterBridge) btnFilterBridge.classList.toggle('active', currentNetFilter === 'bridge');
  if (btnFilterHost) btnFilterHost.classList.toggle('active', currentNetFilter === 'host');
}

async function fetchNetworks() {
  try {
    const res = await fetch('/api/networks');
    if (!res.ok) return;
    const data = await res.json();
    networks = data || [];

    // Update KPI counters
    const statNetworksCount = document.getElementById('statNetworksCount');
    const statNetworksSub = document.getElementById('statNetworksSub');
    const statNetworkContainers = document.getElementById('statNetworkContainers');
    const navNetworksBadge = document.getElementById('navNetworksBadge');
    const networksCountBadge = document.getElementById('networksCountBadge');
    const countNetAll = document.getElementById('countNetAll');

    if (statNetworksCount) statNetworksCount.textContent = networks.length;
    if (navNetworksBadge) navNetworksBadge.textContent = networks.length;
    if (countNetAll) countNetAll.textContent = networks.length;
    if (networksCountBadge) networksCountBadge.textContent = `${networks.length} network${networks.length === 1 ? '' : 's'}`;

    let totalContainers = 0;
    networks.forEach(net => {
      if (net.endpoints) {
        totalContainers += Object.keys(net.endpoints).length;
      }
    });
    if (statNetworkContainers) statNetworkContainers.textContent = totalContainers;
    if (statNetworksSub) {
      const bridgeCount = networks.filter(n => n.driver === 'bridge').length;
      const hostCount = networks.filter(n => n.driver === 'host').length;
      statNetworksSub.textContent = `${bridgeCount} bridge, ${hostCount} host`;
    }

    // Populate Deploy App modal network selector
    populateAppNetworkSelect();

    renderNetworks();
  } catch (err) {
    console.error('Error fetching networks:', err);
  }
}

function populateAppNetworkSelect() {
  const select = document.getElementById('appNetwork');
  if (!select) return;

  const currentVal = select.value || 'bridge';
  select.innerHTML = networks.map(n => {
    const isSelected = n.name === currentVal || (currentVal === '' && n.name === 'bridge');
    return `<option value="${escapeHtml(n.name)}" ${isSelected ? 'selected' : ''}>${escapeHtml(n.name)} (${escapeHtml(n.driver)} - ${escapeHtml(n.subnet || 'host')})</option>`;
  }).join('');
}

function openConnectModalForNetwork(netNameOrID) {
  const modalConnect = document.getElementById('connectContainerModal');
  const inputHidden = document.getElementById('connectTargetNetworkID');
  const selectContainer = document.getElementById('selectConnectContainer');
  const inputAliases = document.getElementById('inputConnectAliases');

  if (inputHidden) inputHidden.value = netNameOrID;
  if (inputAliases) inputAliases.value = '';

  if (selectContainer) {
    if (!apps || apps.length === 0) {
      selectContainer.innerHTML = '<option value="" disabled>No deployed containers found</option>';
    } else {
      selectContainer.innerHTML = apps.map(item => {
        const a = item.app;
        return `<option value="${a.id}">${escapeHtml(a.name)} (${a.engine}, port: ${a.port || 'none'}) [ID: ${a.id}]</option>`;
      }).join('');
    }
  }

  if (modalConnect) modalConnect.classList.add('active');
}

async function disconnectContainerFromNetwork(netNameOrID, containerID, containerName) {
  if (!confirm(`Are you sure you want to disconnect container '${containerName || containerID}' from network '${netNameOrID}'?`)) {
    return;
  }
  try {
    const res = await fetch(`/api/networks/${encodeURIComponent(netNameOrID)}/disconnect`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ container_id: containerID })
    });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Failed to disconnect container');
    }
    showToast(`Disconnected container '${containerName || containerID}' from network`, 'info');
    fetchNetworks();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function deleteNetwork(netNameOrID) {
  if (!confirm(`Are you sure you want to delete virtual network '${netNameOrID}'? Connected containers will be unlinked.`)) {
    return;
  }
  try {
    const res = await fetch(`/api/networks/${encodeURIComponent(netNameOrID)}`, {
      method: 'DELETE'
    });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Failed to delete network');
    }
    showToast(`Virtual network '${netNameOrID}' deleted`, 'info');
    fetchNetworks();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

function renderNetworks() {
  const container = document.getElementById('networksContainer');
  if (!container) return;

  let filtered = networks;
  if (currentNetFilter !== 'all') {
    filtered = networks.filter(n => n.driver === currentNetFilter);
  }

  if (filtered.length === 0) {
    container.innerHTML = `
      <div class="loading-placeholder" style="grid-column: 1 / -1; padding: 48px; text-align: center;">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="color:var(--text-muted); margin-bottom:12px;"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
        <p style="font-size:15px; font-weight:600; color:var(--text-primary);">No virtual networks found</p>
        <p style="font-size:13px; color:var(--text-muted); max-width:380px; margin:6px auto 16px;">Create a user-defined network to interconnect services with automatic DNS and service discovery.</p>
        <button class="btn btn-primary btn-sm" onclick="document.getElementById('createNetworkModal').classList.add('active')">
          + Create Network
        </button>
      </div>
    `;
    return;
  }

  container.innerHTML = filtered.map(net => {
    const endpoints = net.endpoints || {};
    const epCount = Object.keys(endpoints).length;
    const isSystem = net.name === 'bridge' || net.name === 'host';
    const driverClass = net.driver === 'bridge' ? 'tag-bridge' : 'tag-host';

    let endpointsHTML = '';
    if (epCount === 0) {
      endpointsHTML = `
        <div style="font-size:12px; color:var(--text-muted); font-style:italic; padding:8px 0;">
          No containers attached to this network yet.
        </div>
      `;
    } else {
      endpointsHTML = `
        <div class="network-endpoints-list" style="margin-top:10px; display:flex; flex-direction:column; gap:6px;">
          ${Object.entries(endpoints).map(([id, ep]) => {
            const aliasBadges = (ep.aliases || []).map(a => 
              `<span class="tag" style="background:rgba(99,102,241,0.15); color:#818cf8; font-size:10px; padding:1px 5px;">${escapeHtml(a)}</span>`
            ).join(' ');

            return `
              <div class="network-endpoint-row" style="display:flex; align-items:center; justify-content:space-between; background:var(--bg-card-hover); padding:7px 10px; border-radius:6px; font-size:12px; border:1px solid var(--border-color);">
                <div style="display:flex; align-items:center; gap:8px; overflow:hidden;">
                  <span style="font-weight:600; color:var(--text-primary);">${escapeHtml(ep.name)}</span>
                  <span style="font-family:var(--font-mono); color:var(--accent-cyan); font-size:11px;">${ep.ipv4_address || '127.0.0.1'}${ep.port ? ':' + ep.port : ''}</span>
                  ${aliasBadges ? `<div style="display:flex; gap:3px; flex-wrap:wrap;">${aliasBadges}</div>` : ''}
                </div>
                <button class="btn btn-secondary btn-sm" style="padding:2px 6px; font-size:11px; color:#ef4444; border-color:rgba(239,68,68,0.3);" title="Disconnect Container" onclick="disconnectContainerFromNetwork('${escapeHtml(net.name)}', '${escapeHtml(id)}', '${escapeHtml(ep.name)}')">
                  ${Icons.trash}
                </button>
              </div>
            `;
          }).join('')}
        </div>
      `;
    }

    return `
      <div class="app-card" style="display:flex; flex-direction:column; justify-content:space-between;">
        <div>
          <div class="card-top">
            <div class="app-identity">
              <span class="app-name" style="display:flex; align-items:center; gap:6px;">
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color:var(--accent-primary);"><circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><line x1="8.59" y1="13.51" x2="15.42" y2="17.49"/><line x1="15.41" y1="6.51" x2="8.59" y2="10.49"/></svg>
                ${escapeHtml(net.name)}
              </span>
              <span class="app-id-tag">ID: ${(net.id || '').substring(0, 12)}</span>
            </div>
            <div class="card-badges">
              <span class="tag ${driverClass}" style="background:rgba(6,182,212,0.15); color:#22d3ee; font-weight:600;">${escapeHtml(net.driver)}</span>
              ${isSystem ? '<span class="tag" style="background:rgba(148,163,184,0.15); color:#94a3b8;">system</span>' : '<span class="tag" style="background:rgba(16,185,129,0.15); color:#34d399;">custom</span>'}
            </div>
          </div>

          <div class="app-telemetry" style="margin: 12px 0;">
            <div class="telemetry-item">
              <span class="telemetry-label">Subnet</span>
              <span class="telemetry-val" style="font-family:var(--font-mono);">${net.subnet || 'Host IP'}</span>
            </div>
            <div class="telemetry-item">
              <span class="telemetry-label">Gateway</span>
              <span class="telemetry-val" style="font-family:var(--font-mono);">${net.gateway || '127.0.0.1'}</span>
            </div>
            <div class="telemetry-item">
              <span class="telemetry-label">Endpoints</span>
              <span class="telemetry-val" style="color:var(--accent-cyan); font-weight:600;">${epCount} attached</span>
            </div>
          </div>

          <div style="margin-top:10px;">
            <div style="font-size:11px; text-transform:uppercase; letter-spacing:0.5px; color:var(--text-muted); font-weight:600; margin-bottom:4px;">
              Connected Services (${epCount})
            </div>
            ${endpointsHTML}
          </div>
        </div>

        <div class="card-actions" style="margin-top:14px; padding-top:10px; border-top:1px solid var(--border-color); display:flex; justify-content:space-between;">
          <button class="btn btn-secondary btn-sm" onclick="openConnectModalForNetwork('${escapeHtml(net.name)}')">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:4px;"><path d="M12 5v14M5 12h14"/></svg>
            Connect Container
          </button>
          ${!isSystem ? `
            <button class="btn btn-danger btn-sm" title="Delete virtual network" onclick="deleteNetwork('${escapeHtml(net.name)}')">
              ${Icons.trash} Delete
            </button>
          ` : ''}
        </div>
      </div>
    `;
  }).join('');
}

// Hook up initNetworks during DOM ready
setTimeout(initNetworks, 0);

/* ==========================================================================
   Comprehensive In-App Documentation Functionality
   ========================================================================== */

function copyDocsCode(btn) {
  if (!btn) return;
  const codeBox = btn.closest('.docs-code-box');
  if (!codeBox) return;
  const codeEl = codeBox.querySelector('code');
  if (!codeEl) return;
  const text = codeEl.innerText.trim();

  function triggerSuccess() {
    const originalHtml = btn.innerHTML;
    btn.innerHTML = `<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg> Copied!`;
    btn.classList.add('copied');
    showToast('Command copied to clipboard!', 'success');
    setTimeout(() => {
      btn.innerHTML = originalHtml;
      btn.classList.remove('copied');
    }, 2000);
  }

  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(text).then(triggerSuccess).catch(() => {
      fallbackCopy(text, triggerSuccess);
    });
  } else {
    fallbackCopy(text, triggerSuccess);
  }
}

function fallbackCopy(text, onSuccess) {
  try {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-9999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    const successful = document.execCommand('copy');
    document.body.removeChild(textArea);
    if (successful && onSuccess) onSuccess();
    else showToast('Please manually copy the text', 'info');
  } catch (err) {
    showToast('Failed to copy code snippet', 'error');
  }
}

function initDocs() {
  const searchInput = document.getElementById('docsSearchInput');
  const tocLinks = document.querySelectorAll('.docs-toc-link');
  const sectionCards = document.querySelectorAll('.docs-section-card');

  // TOC click smooth scroll & active state
  tocLinks.forEach(link => {
    link.addEventListener('click', (e) => {
      e.preventDefault();
      const targetId = link.getAttribute('data-target') || (link.getAttribute('href') ? link.getAttribute('href').replace('#', '') : null);
      if (!targetId) return;
      const targetSection = document.getElementById(targetId);
      if (targetSection) {
        tocLinks.forEach(l => l.classList.remove('active'));
        link.classList.add('active');
        targetSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
        if (history.pushState) {
          history.pushState(null, null, '#' + targetId);
        }
      }
    });
  });

  // Docs live search / filter
  if (searchInput) {
    searchInput.addEventListener('input', (e) => {
      const q = e.target.value.toLowerCase().trim();
      if (!q) {
        sectionCards.forEach(card => card.style.display = '');
        tocLinks.forEach(link => link.style.display = '');
        return;
      }

      sectionCards.forEach(card => {
        const text = card.innerText.toLowerCase();
        const matches = text.includes(q);
        card.style.display = matches ? '' : 'none';

        const cardId = card.id;
        const matchingTocLink = document.querySelector(`.docs-toc-link[data-target="${cardId}"]`);
        if (matchingTocLink) {
          matchingTocLink.style.display = matches ? '' : 'none';
        }
      });
    });
  }

  // Highlight active TOC item on scroll inside main container
  const scrollContainer = document.querySelector('.zoho-content-scroll');
  if (scrollContainer) {
    scrollContainer.addEventListener('scroll', () => {
      const sectionDocs = document.getElementById('sectionDocs');
      if (!sectionDocs || sectionDocs.style.display === 'none') return;

      let currentActiveId = null;
      sectionCards.forEach(card => {
        const rect = card.getBoundingClientRect();
        if (rect.top <= 250 && rect.bottom >= 100) {
          currentActiveId = card.id;
        }
      });

      if (currentActiveId) {
        tocLinks.forEach(link => {
          if (link.getAttribute('data-target') === currentActiveId) {
            link.classList.add('active');
          } else {
            link.classList.remove('active');
          }
        });
      }
    }, { passive: true });
  }
}

function handleUrlHash() {
  const hash = window.location.hash ? window.location.hash.substring(1) : '';
  if (!hash) return;

  if (hash === 'docs' || hash.startsWith('doc-')) {
    const navDocs = document.getElementById('navDocs');
    if (navDocs) setActiveNavItem(navDocs);
    switchWorkspaceView('docs');
    if (hash.startsWith('doc-')) {
      setTimeout(() => {
        const targetEl = document.getElementById(hash);
        if (targetEl) targetEl.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 150);
    }
  } else if (hash === 'sites') {
    const navSites = document.getElementById('navSites');
    if (navSites) setActiveNavItem(navSites);
    switchWorkspaceView('sites');
    fetchSites();
  } else if (hash === 'cluster') {
    const navCluster = document.getElementById('navCluster');
    if (navCluster) setActiveNavItem(navCluster);
    switchWorkspaceView('cluster');
    fetchClusterData();
  } else if (hash === 'networks') {
    const navNetworks = document.getElementById('navNetworks');
    if (navNetworks) setActiveNavItem(navNetworks);
    switchWorkspaceView('networks');
    fetchNetworks();
  } else if (hash === 'tunnels') {
    const navTunnels = document.getElementById('navTunnels');
    if (navTunnels) setActiveNavItem(navTunnels);
    switchWorkspaceView('tunnels');
    fetchTunnels();
  } else if (hash === 'apps') {
    const navApps = document.getElementById('navApps');
    if (navApps) setActiveNavItem(navApps);
    switchWorkspaceView('apps');
  }
}

setTimeout(initDocs, 0);

// ==========================================
// Cloudflare Tunnel Management Subsystem
// ==========================================

let tunnels = [];
let currentTunnelFilter = 'all';
let tunnelLogsTimer = null;
let activeLogsTunnelID = null;
let selectedTunnelTargetType = 'app';
let selectedTunnelMode = 'quick';
let discoverableTargetsCache = [];

async function fetchTunnels() {
  try {
    const [tunnelsRes, statusRes] = await Promise.all([
      fetch('/api/tunnels'),
      fetch('/api/tunnels/status')
    ]);

    if (tunnelsRes.ok) {
      tunnels = await tunnelsRes.json();
    }
    if (statusRes.ok) {
      const status = await statusRes.json();
      updateCloudflaredStatus(status);
    }

    updateTunnelStats();
    renderTunnels();
  } catch (err) {
    console.error('Error fetching tunnels:', err);
  }
}

function updateCloudflaredStatus(status) {
  const pill = document.getElementById('statCloudflaredStatusPill');
  const verEl = document.getElementById('statCloudflaredVersion');
  const pathEl = document.getElementById('statCloudflaredPath');

  if (status.installed) {
    if (pill) {
      pill.textContent = 'Active';
      pill.className = 'stat-badge-pill green';
    }
    if (verEl) verEl.textContent = `v${status.version || 'installed'}`;
    if (pathEl) pathEl.textContent = status.path || 'System binary ready';
  } else {
    if (pill) {
      pill.textContent = 'Missing';
      pill.className = 'stat-badge-pill amber';
    }
    if (verEl) verEl.textContent = 'Not Found';
    if (pathEl) pathEl.textContent = 'Run: pkg install cloudflared / brew install cloudflared';
  }
}

function updateTunnelStats() {
  const activeCount = tunnels.filter(t => t.status === 'running' || t.status === 'starting').length;
  const quickCount = tunnels.filter(t => t.mode === 'quick').length;
  const tokenCount = tunnels.filter(t => t.mode === 'token').length;

  const statActive = document.getElementById('statTunnelsActive');
  const statTotal = document.getElementById('statTunnelsTotal');
  const statQuick = document.getElementById('statTunnelsQuick');
  const statToken = document.getElementById('statTunnelsToken');
  const navBadge = document.getElementById('navTunnelsBadge');
  const countBadge = document.getElementById('tunnelsCountBadge');
  const countAll = document.getElementById('countTunnelAll');
  const countActive = document.getElementById('countTunnelActive');

  if (statActive) statActive.textContent = activeCount;
  if (statTotal) statTotal.textContent = `${tunnels.length} configured tunnels`;
  if (statQuick) statQuick.textContent = quickCount;
  if (statToken) statToken.textContent = tokenCount;
  if (navBadge) navBadge.textContent = activeCount;
  if (countBadge) countBadge.textContent = `${tunnels.length} tunnels`;
  if (countAll) countAll.textContent = tunnels.length;
  if (countActive) countActive.textContent = activeCount;
}

function renderTunnels() {
  const container = document.getElementById('tunnelsContainer');
  if (!container) return;

  let filtered = tunnels;
  if (currentTunnelFilter === 'active') {
    filtered = tunnels.filter(t => t.status === 'running' || t.status === 'starting');
  } else if (currentTunnelFilter === 'stopped') {
    filtered = tunnels.filter(t => t.status === 'stopped' || t.status === 'error');
  } else if (currentTunnelFilter === 'quick') {
    filtered = tunnels.filter(t => t.mode === 'quick');
  } else if (currentTunnelFilter === 'token') {
    filtered = tunnels.filter(t => t.mode === 'token');
  }

  if (searchKeyword) {
    filtered = filtered.filter(t =>
      (t.name || '').toLowerCase().includes(searchKeyword) ||
      (t.id || '').toLowerCase().includes(searchKeyword) ||
      (t.target || '').toLowerCase().includes(searchKeyword) ||
      (t.public_url || '').toLowerCase().includes(searchKeyword) ||
      (t.app_name || '').toLowerCase().includes(searchKeyword)
    );
  }

  if (filtered.length === 0) {
    container.innerHTML = `
      <div class="loading-placeholder" style="grid-column: 1 / -1; padding: 48px; text-align: center;">
        <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#f38020" stroke-width="1.5" style="margin-bottom:14px;">
          <path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/>
          <polyline points="13 11 9 15 13 19"/>
          <line x1="9" y1="15" x2="20" y2="15"/>
        </svg>
        <p style="font-size:16px; font-weight:700; color:var(--text-primary); margin-bottom:6px;">No Cloudflare Tunnels Configured</p>
        <p style="font-size:13px; color:var(--text-muted); max-width:440px; margin:0 auto 18px; line-height:1.5;">
          Directly expose any local port, Docroid container, or web service through Cloudflare's global edge network with instant public URLs.
        </p>
        <button class="btn btn-primary" onclick="openTunnelModal()" style="background: linear-gradient(135deg, #f38020 0%, #faad3f 100%); border-color: #f38020; color: #fff;">
          + Direct Configure Tunnel
        </button>
      </div>
    `;
    return;
  }

  container.innerHTML = filtered.map(t => {
    const isRunning = t.status === 'running';
    const isStarting = t.status === 'starting';
    const isError = t.status === 'error';
    const isQuick = t.mode === 'quick';

    let statusPill = '';
    if (isRunning) {
      statusPill = `<span class="tunnel-status-pill running"><span class="pulse-dot"></span> RUNNING</span>`;
    } else if (isStarting) {
      statusPill = `<span class="tunnel-status-pill starting"><span class="pulse-dot-orange"></span> CONNECTING...</span>`;
    } else if (isError) {
      statusPill = `<span class="tunnel-status-pill error">ERROR</span>`;
    } else {
      statusPill = `<span class="tunnel-status-pill stopped">STOPPED</span>`;
    }

    let urlBannerHTML = '';
    if (t.public_url) {
      urlBannerHTML = `
        <div class="tunnel-url-banner">
          <div style="display:flex; align-items:center; gap:8px; overflow:hidden;">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#f38020" stroke-width="2" style="flex-shrink:0;"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/></svg>
            <a href="${escapeHtml(t.public_url)}" target="_blank" class="tunnel-url-text" title="Open ${escapeHtml(t.public_url)}">
              ${escapeHtml(t.public_url)}
            </a>
          </div>
          <div class="tunnel-url-actions">
            <button class="btn btn-secondary btn-xs" onclick="copyTunnelURL('${escapeHtml(t.public_url)}', this)" title="Copy public URL">
              Copy URL
            </button>
            <a href="${escapeHtml(t.public_url)}" target="_blank" class="btn btn-primary btn-xs" style="background:#f38020; border-color:#f38020; color:#fff;" title="Visit Public URL">
              Visit &rarr;
            </a>
          </div>
        </div>
      `;
    } else if (isStarting) {
      urlBannerHTML = `
        <div class="tunnel-url-banner" style="background:rgba(245, 158, 11, 0.1); border-color:rgba(245, 158, 11, 0.3);">
          <div style="font-size:12px; color:#fbbf24; display:flex; align-items:center; gap:8px;">
            <span class="pulse-dot-orange"></span>
            Negotiating tunnel route with Cloudflare Edge...
          </div>
        </div>
      `;
    } else if (isError) {
      urlBannerHTML = `
        <div class="tunnel-url-banner" style="background:rgba(239, 68, 68, 0.1); border-color:rgba(239, 68, 68, 0.3);">
          <div style="font-size:12px; color:#f87171;">
            <strong>Error:</strong> ${escapeHtml(t.last_error || 'Cloudflared terminated unexpectedly')}
          </div>
        </div>
      `;
    } else {
      urlBannerHTML = `
        <div class="tunnel-url-banner" style="background:rgba(148, 163, 184, 0.08); border-color:rgba(148, 163, 184, 0.2);">
          <div style="font-size:12px; color:var(--text-muted);">
            Tunnel is stopped. Click Start to reconnect.
          </div>
        </div>
      `;
    }

    return `
      <div class="tunnel-card">
        <div>
          <div class="tunnel-card-header">
            <div class="tunnel-identity">
              <div class="tunnel-title">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#f38020" stroke-width="2"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/><polyline points="13 11 9 15 13 19"/><line x1="9" y1="15" x2="20" y2="15"/></svg>
                <span>${escapeHtml(t.name)}</span>
              </div>
              <span class="tunnel-id-badge">ID: ${t.id}</span>
            </div>
            <div style="display:flex; align-items:center; gap:8px;">
              <span class="tunnel-mode-tag ${isQuick ? 'quick' : 'token'}">
                ${isQuick ? 'Quick Tunnel' : 'Zero Trust'}
              </span>
              ${statusPill}
            </div>
          </div>

          <div style="display:flex; gap:8px; align-items:center; margin-bottom:8px; flex-wrap:wrap;">
            <div style="font-size:12px; color:var(--text-secondary); display:flex; align-items:center; gap:6px;">
              <span style="font-weight:600; color:var(--text-muted);">TARGET:</span>
              <span style="font-family:var(--font-mono); color:var(--text-primary); font-weight:600;">${escapeHtml(t.target)}</span>
            </div>
            ${t.app_name ? `
              <span class="tag" style="background:rgba(220, 38, 38, 0.15); color:#f87171; font-size:11px;">
                App: ${escapeHtml(t.app_name)}
              </span>
            ` : ''}
          </div>

          ${urlBannerHTML}

          <div class="app-telemetry" style="margin: 10px 0;">
            <div class="telemetry-item">
              <span class="telemetry-label">Port</span>
              <span class="telemetry-val" style="font-family:var(--font-mono); font-weight:700; color:var(--accent-cyan);">${t.target_port || '-'}</span>
            </div>
            <div class="telemetry-item">
              <span class="telemetry-label">PID</span>
              <span class="telemetry-val">${t.pid || '-'}</span>
            </div>
            <div class="telemetry-item">
              <span class="telemetry-label">Created</span>
              <span class="telemetry-val" style="font-size:11px;">${new Date(t.created_at).toLocaleTimeString([], {hour:'2-digit', minute:'2-digit'})}</span>
            </div>
          </div>
        </div>

        <div class="card-actions" style="margin-top:14px; padding-top:12px; border-top:1px solid var(--border-card); display:flex; justify-content:space-between; align-items:center; gap:6px; flex-wrap:wrap;">
          <div style="display:flex; gap:6px;">
            ${(isRunning || isStarting) ? `
              <button class="btn btn-secondary btn-sm" onclick="stopTunnel('${t.id}', '${escapeHtml(t.name)}')">Stop</button>
              <button class="btn btn-secondary btn-sm" onclick="restartTunnel('${t.id}', '${escapeHtml(t.name)}')">Restart</button>
            ` : `
              <button class="btn btn-primary btn-sm" onclick="startTunnel('${t.id}', '${escapeHtml(t.name)}')">Start</button>
            `}
            <button class="btn btn-secondary btn-sm" onclick="openTunnelLogs('${t.id}', '${escapeHtml(t.name)}', '${escapeHtml(t.public_url || '')}')">Logs</button>
          </div>
          <button class="btn btn-danger btn-sm" onclick="deleteTunnel('${t.id}', '${escapeHtml(t.name)}')">
            Delete
          </button>
        </div>
      </div>
    `;
  }).join('');
}

function updateTunnelFilterButtons() {
  const filterBtns = document.querySelectorAll('[data-tunnel-filter]');
  filterBtns.forEach(btn => {
    if (btn.getAttribute('data-tunnel-filter') === currentTunnelFilter) {
      btn.classList.add('active');
    } else {
      btn.classList.remove('active');
    }
  });
}

async function openTunnelModal(prefillAppID, prefillAppName, prefillPort) {
  const modal = document.getElementById('tunnelModal');
  if (!modal) return;

  const targetSelect = document.getElementById('tunnelTargetSelect');
  const customPort = document.getElementById('tunnelCustomPort');
  const customHost = document.getElementById('tunnelCustomHost');
  const nameInput = document.getElementById('tunnelNameInput');
  const tokenInput = document.getElementById('tunnelTokenInput');
  const hostnameInput = document.getElementById('tunnelHostnameInput');

  // Reset form
  if (tokenInput) tokenInput.value = '';
  if (hostnameInput) hostnameInput.value = '';

  // Set mode to Quick Tunnel by default
  setTunnelMode('quick');

  // Load targets
  try {
    targetSelect.innerHTML = '<option value="">Loading running apps...</option>';
    const res = await fetch('/api/tunnels/targets');
    if (res.ok) {
      discoverableTargetsCache = await res.json();
      if (discoverableTargetsCache.length === 0) {
        targetSelect.innerHTML = '<option value="">(No running apps detected — use Custom Port)</option>';
        setTunnelTargetSource('custom');
      } else {
        targetSelect.innerHTML = discoverableTargetsCache.map(tgt => `
          <option value="${tgt.port}" data-app-id="${tgt.id}" data-app-name="${escapeHtml(tgt.name)}" data-target-url="${tgt.url}">
            ${escapeHtml(tgt.name)} — Port ${tgt.port} (${tgt.type})
          </option>
        `).join('');
        setTunnelTargetSource('app');
      }
    }
  } catch (err) {
    console.error('Failed to load targets:', err);
    setTunnelTargetSource('custom');
  }

  // Handle prefill if passed
  if (prefillAppID || prefillPort) {
    if (prefillAppName && nameInput) {
      nameInput.value = `${prefillAppName.toLowerCase().replace(/[^a-z0-9]/g, '-')}-tunnel`;
    }
    if (prefillPort && customPort) {
      customPort.value = prefillPort;
    }

    // Try selecting in dropdown
    if (prefillAppID && targetSelect) {
      for (let i = 0; i < targetSelect.options.length; i++) {
        if (targetSelect.options[i].getAttribute('data-app-id') === prefillAppID) {
          targetSelect.selectedIndex = i;
          setTunnelTargetSource('app');
          break;
        }
      }
    }
  } else {
    if (nameInput) {
      nameInput.value = `tunnel-${Math.random().toString(36).substring(2, 7)}`;
    }
  }

  modal.classList.add('active');
}

function openTunnelModalForApp(appID, appName, port) {
  openTunnelModal(appID, appName, port);
}

function setTunnelTargetSource(type) {
  selectedTunnelTargetType = type;
  const btnApp = document.getElementById('btnSourceApp');
  const btnCustom = document.getElementById('btnSourceCustom');
  const appGroup = document.getElementById('tunnelAppSourceGroup');
  const customGroup = document.getElementById('tunnelCustomSourceGroup');

  if (type === 'custom') {
    if (btnApp) btnApp.classList.remove('active');
    if (btnCustom) btnCustom.classList.add('active');
    if (appGroup) appGroup.style.display = 'none';
    if (customGroup) customGroup.style.display = 'block';
  } else {
    if (btnApp) btnApp.classList.add('active');
    if (btnCustom) btnCustom.classList.remove('active');
    if (appGroup) appGroup.style.display = 'block';
    if (customGroup) customGroup.style.display = 'none';
  }
}

function setTunnelMode(mode) {
  selectedTunnelMode = mode;
  const cardQuick = document.getElementById('cardModeQuick');
  const cardToken = document.getElementById('cardModeToken');
  const tokenGroup = document.getElementById('tunnelTokenGroup');

  if (mode === 'token') {
    if (cardQuick) cardQuick.classList.remove('active');
    if (cardToken) cardToken.classList.add('active');
    if (tokenGroup) tokenGroup.style.display = 'block';
  } else {
    if (cardQuick) cardQuick.classList.add('active');
    if (cardToken) cardToken.classList.remove('active');
    if (tokenGroup) tokenGroup.style.display = 'none';
  }
}

async function submitCreateTunnel(e) {
  if (e) e.preventDefault();

  const nameInput = document.getElementById('tunnelNameInput');
  const tokenInput = document.getElementById('tunnelTokenInput');
  const hostnameInput = document.getElementById('tunnelHostnameInput');
  const autoStartCheck = document.getElementById('tunnelAutoStart');
  const autoRestartCheck = document.getElementById('tunnelAutoRestart');
  const submitBtn = document.getElementById('btnSubmitTunnel');

  let targetPort = 0;
  let targetStr = '';
  let appID = '';
  let appName = '';

  if (selectedTunnelTargetType === 'app') {
    const targetSelect = document.getElementById('tunnelTargetSelect');
    if (targetSelect && targetSelect.selectedOptions.length > 0) {
      const opt = targetSelect.selectedOptions[0];
      targetPort = parseInt(opt.value, 10) || 0;
      targetStr = opt.getAttribute('data-target-url') || `http://127.0.0.1:${targetPort}`;
      appID = opt.getAttribute('data-app-id') || '';
      appName = opt.getAttribute('data-app-name') || '';
    }
  } else {
    const portInput = document.getElementById('tunnelCustomPort');
    const hostInput = document.getElementById('tunnelCustomHost');
    targetPort = parseInt(portInput.value, 10) || 0;
    const host = (hostInput && hostInput.value.trim()) ? hostInput.value.trim() : '127.0.0.1';
    targetStr = `http://${host}:${targetPort}`;
  }

  if (selectedTunnelMode !== 'token' && targetPort <= 0 && !targetStr) {
    showToast('Please select an application or specify a valid port number.', 'error');
    return;
  }

  if (selectedTunnelMode === 'token' && (!tokenInput || !tokenInput.value.trim())) {
    showToast('Cloudflare Tunnel Token is required for Zero Trust mode.', 'error');
    return;
  }

  const payload = {
    name: nameInput ? nameInput.value.trim() : '',
    app_id: appID,
    app_name: appName,
    target: targetStr,
    target_port: targetPort,
    mode: selectedTunnelMode,
    token: tokenInput ? tokenInput.value.trim() : '',
    hostname: hostnameInput ? hostnameInput.value.trim() : '',
    auto_start: autoStartCheck ? autoStartCheck.checked : true,
    auto_restart: autoRestartCheck ? autoRestartCheck.checked : false
  };

  const originalBtnHTML = submitBtn ? submitBtn.innerHTML : '';
  if (submitBtn) {
    submitBtn.disabled = true;
    submitBtn.innerHTML = `<span>Connecting to Cloudflare...</span>`;
  }

  try {
    const res = await fetch('/api/tunnels', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Failed to configure tunnel');
    }

    showToast(`Tunnel '${data.name}' configured and starting!`, 'success');
    document.getElementById('tunnelModal').classList.remove('active');
    
    // Switch to tunnels view if not already there
    switchWorkspaceView('tunnels');
    const navTunnels = document.getElementById('navTunnels');
    if (navTunnels) setActiveNavItem(navTunnels);

    await fetchTunnels();

    // Poll a few times for instant URL display
    let pollCount = 0;
    const quickPoll = setInterval(async () => {
      pollCount++;
      await fetchTunnels();
      const current = tunnels.find(t => t.id === data.id);
      if (current && (current.public_url || current.status === 'running' || pollCount > 10)) {
        clearInterval(quickPoll);
      }
    }, 1500);

  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    if (submitBtn) {
      submitBtn.disabled = false;
      submitBtn.innerHTML = originalBtnHTML;
    }
  }
}

async function startTunnel(id, name) {
  try {
    const res = await fetch(`/api/tunnels/${encodeURIComponent(id)}/start`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to start tunnel');
    showToast(`Tunnel '${name || id}' starting...`, 'info');
    await fetchTunnels();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function stopTunnel(id, name) {
  try {
    const res = await fetch(`/api/tunnels/${encodeURIComponent(id)}/stop`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to stop tunnel');
    showToast(`Tunnel '${name || id}' stopped.`, 'success');
    await fetchTunnels();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function restartTunnel(id, name) {
  try {
    showToast(`Restarting tunnel '${name || id}'...`, 'info');
    const res = await fetch(`/api/tunnels/${encodeURIComponent(id)}/restart`, { method: 'POST' });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Failed to restart tunnel');
    showToast(`Tunnel '${name || id}' restarted.`, 'success');
    await fetchTunnels();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

async function deleteTunnel(id, name) {
  if (!confirm(`Are you sure you want to delete tunnel '${name || id}'? This will stop the cloudflared process and remove the configuration.`)) {
    return;
  }
  try {
    const res = await fetch(`/api/tunnels/${encodeURIComponent(id)}`, { method: 'DELETE' });
    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || 'Failed to delete tunnel');
    }
    showToast(`Tunnel '${name || id}' deleted.`, 'success');
    await fetchTunnels();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

function copyTunnelURL(url, btn) {
  if (!url) return;
  navigator.clipboard.writeText(url).then(() => {
    if (btn) {
      const orig = btn.textContent;
      btn.textContent = 'Copied!';
      btn.style.borderColor = '#10b981';
      btn.style.color = '#34d399';
      setTimeout(() => {
        btn.textContent = orig;
        btn.style.borderColor = '';
        btn.style.color = '';
      }, 2000);
    }
    showToast('Public Cloudflare URL copied to clipboard!', 'success');
  }).catch(() => {
    prompt('Copy public URL:', url);
  });
}

async function openTunnelLogs(id, name, publicURL) {
  activeLogsTunnelID = id;
  const modal = document.getElementById('tunnelLogsModal');
  const title = document.getElementById('tunnelLogsTitle');
  const subtitle = document.getElementById('tunnelLogsSubtitle');
  const content = document.getElementById('tunnelLogsContent');
  const urlBanner = document.getElementById('tunnelLogsUrlBanner');
  const publicLink = document.getElementById('tunnelLogsPublicLink');

  if (title) title.textContent = `Logs: ${name || id}`;
  if (subtitle) subtitle.textContent = `Live stdout & stderr for tunnel ${id}`;
  if (content) content.textContent = 'Loading logs...';

  if (publicURL && urlBanner && publicLink) {
    urlBanner.style.display = 'flex';
    publicLink.textContent = publicURL;
    publicLink.href = publicURL;
  } else if (urlBanner) {
    urlBanner.style.display = 'none';
  }

  if (modal) modal.classList.add('active');

  await refreshTunnelLogs();

  if (tunnelLogsTimer) clearInterval(tunnelLogsTimer);
  tunnelLogsTimer = setInterval(() => {
    const autoRefresh = document.getElementById('tunnelLogsAutoRefresh');
    if (autoRefresh && autoRefresh.checked && modal && modal.classList.contains('active')) {
      refreshTunnelLogs();
    }
  }, 2000);
}

async function refreshTunnelLogs() {
  if (!activeLogsTunnelID) return;
  try {
    const res = await fetch(`/api/tunnels/${encodeURIComponent(activeLogsTunnelID)}/logs?lines=150`);
    if (res.ok) {
      const data = await res.json();
      const content = document.getElementById('tunnelLogsContent');
      if (content) {
        content.textContent = data.logs || '(No logs recorded yet)';
        content.scrollTop = content.scrollHeight;
      }
    }
  } catch (err) {
    console.error('Failed to refresh tunnel logs:', err);
  }
}

function closeTunnelLogs() {
  if (tunnelLogsTimer) {
    clearInterval(tunnelLogsTimer);
    tunnelLogsTimer = null;
  }
  activeLogsTunnelID = null;
  const modal = document.getElementById('tunnelLogsModal');
  if (modal) modal.classList.remove('active');
}

async function checkCloudflaredBinary() {
  try {
    const res = await fetch('/api/tunnels/status');
    const status = await res.json();
    updateCloudflaredStatus(status);
    if (status.installed) {
      showToast(`✓ cloudflared is installed (${status.version}) at ${status.path}`, 'success');
    } else {
      showToast(`⚠️ cloudflared not found. Install via 'pkg install cloudflared' (Termux) or 'brew install cloudflared' (macOS).`, 'info');
    }
  } catch (err) {
    showToast(`Failed to check cloudflared: ${err.message}`, 'error');
  }
}

// Wire Cloudflare Tunnel UI Event Listeners
document.addEventListener('DOMContentLoaded', () => {
  // Topbar + Section open modal buttons
  const btnTop = document.getElementById('btnConfigureTunnelTop');
  const btnOpen = document.getElementById('btnOpenConfigureTunnelModal');
  if (btnTop) btnTop.addEventListener('click', () => openTunnelModal());
  if (btnOpen) btnOpen.addEventListener('click', () => openTunnelModal());

  // Close modals
  const closeTunnel = document.getElementById('closeTunnelModal');
  const cancelTunnel = document.getElementById('cancelTunnelModal');
  if (closeTunnel) closeTunnel.addEventListener('click', () => document.getElementById('tunnelModal').classList.remove('active'));
  if (cancelTunnel) cancelTunnel.addEventListener('click', () => document.getElementById('tunnelModal').classList.remove('active'));

  // Target Source toggle
  const btnSrcApp = document.getElementById('btnSourceApp');
  const btnSrcCustom = document.getElementById('btnSourceCustom');
  if (btnSrcApp) btnSrcApp.addEventListener('click', () => setTunnelTargetSource('app'));
  if (btnSrcCustom) btnSrcCustom.addEventListener('click', () => setTunnelTargetSource('custom'));

  // Mode Cards toggle
  const cardQuick = document.getElementById('cardModeQuick');
  const cardToken = document.getElementById('cardModeToken');
  if (cardQuick) cardQuick.addEventListener('click', () => setTunnelMode('quick'));
  if (cardToken) cardToken.addEventListener('click', () => setTunnelMode('token'));

  // Target select change auto-updates default tunnel name
  const targetSelect = document.getElementById('tunnelTargetSelect');
  if (targetSelect) {
    targetSelect.addEventListener('change', () => {
      const opt = targetSelect.selectedOptions[0];
      const nameInput = document.getElementById('tunnelNameInput');
      if (opt && nameInput && !nameInput.value.includes('custom')) {
        const appName = opt.getAttribute('data-app-name') || '';
        if (appName) {
          nameInput.value = `${appName.toLowerCase().replace(/[^a-z0-9]/g, '-')}-tunnel`;
        }
      }
    });
  }

  // Form submit
  const form = document.getElementById('tunnelForm');
  if (form) form.addEventListener('submit', submitCreateTunnel);

  // Logs modal controls
  const closeLogs = document.getElementById('closeTunnelLogsModal');
  if (closeLogs) closeLogs.addEventListener('click', closeTunnelLogs);

  const btnCopyLogs = document.getElementById('btnCopyTunnelLogs');
  if (btnCopyLogs) {
    btnCopyLogs.addEventListener('click', () => {
      const content = document.getElementById('tunnelLogsContent');
      if (content) {
        navigator.clipboard.writeText(content.textContent).then(() => {
          showToast('Diagnostic logs copied to clipboard!', 'success');
        });
      }
    });
  }

  const btnRefreshLogs = document.getElementById('btnRefreshTunnelLogs');
  if (btnRefreshLogs) btnRefreshLogs.addEventListener('click', refreshTunnelLogs);

  const btnCopyLogsPublic = document.getElementById('btnCopyLogsPublicURL');
  if (btnCopyLogsPublic) {
    btnCopyLogsPublic.addEventListener('click', () => {
      const publicLink = document.getElementById('tunnelLogsPublicLink');
      if (publicLink && publicLink.textContent) {
        copyTunnelURL(publicLink.textContent, btnCopyLogsPublic);
      }
    });
  }

  // Filter Buttons
  const filterBtns = document.querySelectorAll('[data-tunnel-filter]');
  filterBtns.forEach(btn => {
    btn.addEventListener('click', () => {
      currentTunnelFilter = btn.getAttribute('data-tunnel-filter');
      updateTunnelFilterButtons();
      renderTunnels();
    });
  });

  // Check binary button
  const btnCheck = document.getElementById('btnCheckCloudflared');
  if (btnCheck) btnCheck.addEventListener('click', checkCloudflaredBinary);
});



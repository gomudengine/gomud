const { app, BrowserWindow, Menu } = require('electron');
const path = require('path');

// Required for Steam overlay compatibility with Chromium
app.commandLine.appendSwitch('in-process-gpu');

// Initialize Steam SDK (graceful fallback for non-Steam launches)
let steamClient = null;
try {
  const steamworks = require('steamworks.js');
  steamClient = steamworks.init(YOUR_APP_ID); // Replace YOUR_APP_ID with your numeric Steam App ID
} catch (err) {
  console.warn('Steam SDK not available:', err.message);
}

function createWindow() {
  const win = new BrowserWindow({
    width: 1920,
    height: 1080,
    fullscreen: true,
    title: 'GoMud Client',
    webPreferences: {
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  win.loadFile(path.join(__dirname, 'build_cache', 'webclient-pure.html'));

  if (process.argv.includes('--dev')) {
    win.webContents.openDevTools();
  }
}

function buildMenu() {
  const template = [
    ...(process.platform === 'darwin' ? [{
      label: app.name,
      submenu: [
        { role: 'about' },
        { type: 'separator' },
        { role: 'quit' },
      ],
    }] : []),
    {
      label: 'View',
      submenu: [
        { role: 'togglefullscreen' },
        { role: 'resetZoom' },
        { role: 'zoomIn' },
        { role: 'zoomOut' },
      ],
    },
    {
      label: 'Edit',
      submenu: [
        { role: 'copy' },
        { role: 'paste' },
        { role: 'selectAll' },
      ],
    },
  ];

  Menu.setApplicationMenu(Menu.buildFromTemplate(template));
}

app.whenReady().then(() => {
  buildMenu();
  createWindow();
});

app.on('window-all-closed', () => {
  app.quit();
});

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow();
  }
});

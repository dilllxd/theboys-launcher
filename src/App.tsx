import React, { useEffect, useState } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';

import { Layout } from './components/layout';
import { HomePage } from './pages/HomePage';
import { ModpacksPage } from './pages/ModpacksPage';
import { InstancesPage } from './pages/InstancesPage';
import { DownloadsPage } from './pages/DownloadsPage';
import { SettingsPage } from './pages/SettingsPage';
import { UpdatesPage } from './pages/UpdatesPage';
import LaunchPage from './pages/LaunchPage';
import { LoadingScreen } from './components/LoadingScreen';
import { ErrorBoundary } from './components/ErrorBoundary';
import { UpdateNotification } from './components/updates';
import { initializeTheme } from './utils/theme';
import { api } from './utils/api';

import './styles/App.css';


// App component with error boundary and loading states
const AppContent: React.FC = () => {
  const [isInitialized, setIsInitialized] = useState(false);
  const [initError, setInitError] = useState<string | null>(null);

  // Initialize theme and perform health check
  useEffect(() => {
    const initializeApp = async () => {
      try {
        // Initialize theme
        initializeTheme();

        // Perform health check
        await api.healthCheck();

        setIsInitialized(true);
      } catch (error) {
        console.error('Failed to initialize app:', error);
        setInitError(error instanceof Error ? error.message : 'Unknown error');
        setIsInitialized(true); // Still show the app even if health check fails
      }
    };

    initializeApp();
  }, []);

  // Show loading screen during initialization
  if (!isInitialized) {
    return <LoadingScreen message="Initializing TheBoys Launcher..." />;
  }

  // Show error screen if initialization failed
  if (initError) {
    return (
      <div className="error-screen">
        <div className="error-content">
          <h1>Initialization Error</h1>
          <p>Failed to initialize the launcher:</p>
          <p className="error-message">{initError}</p>
          <button onClick={() => window.location.reload()}>
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <ErrorBoundary>
      <AnimatePresence mode="wait">
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.3 }}
          className="app-container"
        >
          <Layout>
            <Routes>
              <Route path="/" element={<Navigate to="/home" replace />} />
              <Route path="/home" element={<HomePage />} />
              <Route path="/modpacks" element={<ModpacksPage />} />
              <Route path="/instances" element={<InstancesPage />} />
              <Route path="/launch" element={<LaunchPage />} />
              <Route path="/downloads" element={<DownloadsPage />} />
              <Route path="/updates" element={<UpdatesPage />} />
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="*" element={<Navigate to="/home" replace />} />
            </Routes>
          </Layout>
          <UpdateNotification />
        </motion.div>
      </AnimatePresence>
    </ErrorBoundary>
  );
};

// Main App component
const App: React.FC = () => {
  return <AppContent />;
};

export default App;
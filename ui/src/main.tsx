/**
 * @fileoverview The Stem - Application Entry Point
 * @description Bootstraps the React application and mounts it to the DOM.
 */

import '@fontsource-variable/inter';
import '@fontsource-variable/jetbrains-mono';
import { QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { ErrorBoundary } from './components/ErrorBoundary';
import { initThemeFromStorage } from './hooks/useTheme';
import { getQueryClient } from './lib/queryClient';
import './index.css';

// Initialize i18n before rendering
import './i18n';

// Apply stored theme before first paint to avoid light/dark flash.
initThemeFromStorage();

const rootElement: HTMLElement | null = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <React.StrictMode>
      <ErrorBoundary>
        <QueryClientProvider client={getQueryClient()}>
          <App />
        </QueryClientProvider>
      </ErrorBoundary>
    </React.StrictMode>,
  );
}

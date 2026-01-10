/**
 * @fileoverview React Error Boundary Component
 * @description Catches JavaScript errors in child component tree and displays fallback UI.
 *              Prevents entire app from crashing due to component errors.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Component, type ErrorInfo, type ReactNode } from 'react';

interface ErrorBoundaryProps {
  /** Child components to render */
  children: ReactNode;
  /** Optional fallback UI to show when an error occurs */
  fallback?: ReactNode;
  /** Optional callback when an error is caught */
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

/**
 * Error Boundary component that catches JavaScript errors in its child component tree.
 *
 * Usage:
 * ```tsx
 * <ErrorBoundary>
 *   <App />
 * </ErrorBoundary>
 * ```
 *
 * With custom fallback:
 * ```tsx
 * <ErrorBoundary fallback={<div>Something went wrong</div>}>
 *   <App />
 * </ErrorBoundary>
 * ```
 */
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    // Update state so the next render shows the fallback UI
    return { hasError: true, error };
  }

  override componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    // Update state with error info
    this.setState({ errorInfo });

    // Call optional error callback
    this.props.onError?.(error, errorInfo);
  }

  handleRetry = (): void => {
    // Reset error state and try rendering children again
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  handleReload = (): void => {
    // Full page reload as last resort
    window.location.reload();
  };

  override render(): ReactNode {
    if (this.state.hasError) {
      // If custom fallback provided, use it
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Default fallback UI
      return (
        <div className="min-h-screen flex items-center justify-center bg-[var(--color-surface-base)] p-4">
          <div className="max-w-md w-full rounded-2xl border border-[var(--color-surface-border)] bg-[var(--color-surface-raised)] p-6 shadow-lg">
            <div className="flex items-center gap-3 mb-4">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-[var(--color-status-error)]/10">
                <AlertTriangle className="h-5 w-5 text-[var(--color-status-error)]" />
              </div>
              <div>
                <h2 className="text-lg font-semibold text-[var(--color-text-primary)]">
                  Something went wrong
                </h2>
                <p className="text-sm text-[var(--color-text-muted)]">
                  An unexpected error occurred
                </p>
              </div>
            </div>

            {this.state.error && (
              <div className="mb-4 p-3 rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
                <p className="text-sm font-medium text-[var(--color-text-primary)] mb-1">
                  Error Details
                </p>
                <p className="text-sm text-[var(--color-status-error)] font-mono break-all">
                  {this.state.error.message}
                </p>
              </div>
            )}

            <div className="flex gap-3">
              <button
                type="button"
                onClick={this.handleRetry}
                className="btn btn-primary flex-1 justify-center"
              >
                <RefreshCw className="h-4 w-4" />
                Try Again
              </button>
              <button
                type="button"
                onClick={this.handleReload}
                className="btn btn-secondary flex-1 justify-center"
              >
                Reload Page
              </button>
            </div>

            <p className="mt-4 text-xs text-center text-[var(--color-text-muted)]">
              If this problem persists, please contact support.
            </p>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

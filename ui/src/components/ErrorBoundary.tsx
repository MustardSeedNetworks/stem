// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

/**
 * Error Boundary Component
 *
 * A React Error Boundary class component for catching and handling errors in child component trees.
 * Provides graceful error handling with an optional custom fallback UI or a default error alert.
 *
 * Key Features:
 * - Catches JavaScript errors in child components
 * - Logs errors via structured logging utility
 * - Supports optional onError callback for external integrations (Sentry, etc.)
 * - Displays error UI with retry and reload options
 * - Styled with theme tokens for consistency
 * - Supports custom fallback UI
 *
 * @example
 * ```tsx
 * <ErrorBoundary fallback={<CustomErrorUI />}>
 *   <YourComponent />
 * </ErrorBoundary>
 * ```
 *
 * @example
 * ```tsx
 * // With logging callback
 * <ErrorBoundary onError={(error, errorInfo) => sendToSentry(error, errorInfo)}>
 *   <App />
 * </ErrorBoundary>
 * ```
 */

import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Translation } from 'react-i18next';
import { button, cn, icon as iconTokens, radius, spacing } from '../styles/theme';
import { logError } from '../utils/logger';

/**
 * Props for the ErrorBoundary component
 */
interface ErrorBoundaryProps {
  /** Child components to be protected by error boundary */
  children: ReactNode;
  /** Optional custom fallback UI to display when error occurs */
  fallback?: ReactNode;
  /** Optional callback when an error is caught - useful for external logging services */
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

/**
 * State for the ErrorBoundary component
 */
interface ErrorBoundaryState {
  /** Flag indicating if an error has been caught */
  hasError: boolean;
  /** The caught error object, or null if no error */
  error: Error | null;
  /** Error info with component stack, or null if no error */
  errorInfo: ErrorInfo | null;
}

/**
 * Alert triangle icon for error display
 */
function AlertIcon() {
  return (
    <svg
      className={cn(iconTokens.size.md, 'text-status-error shrink-0')}
      fill="currentColor"
      viewBox="0 0 20 20"
      aria-hidden="true"
    >
      <path
        fillRule="evenodd"
        d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z"
        clipRule="evenodd"
      />
    </svg>
  );
}

/**
 * Refresh icon for retry button
 */
function RefreshIcon() {
  return (
    <svg
      className={iconTokens.size.sm}
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
      />
    </svg>
  );
}

/**
 * Error Boundary Class Component
 *
 * Implements React's Error Boundary pattern to gracefully handle errors in the component tree.
 * - Catches JavaScript errors in child components
 * - Logs error information via structured logger
 * - Calls optional onError callback for external integrations
 * - Displays error UI with retry functionality
 * - Supports custom fallback UI as alternative to default error display
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

  /**
   * React lifecycle method called when an error is thrown in a child component.
   * Updates component state to indicate error state and stores the error object.
   */
  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { hasError: true, error };
  }

  /**
   * React lifecycle method called after an error has been caught.
   * Used for logging and error reporting to services.
   */
  override componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    // Update state with error info for display
    this.setState({ errorInfo });

    // Log to structured logger
    logError(error, {
      component: 'ErrorBoundary',
      action: 'componentDidCatch',
      additionalData: {
        componentStack: errorInfo.componentStack,
      },
    });

    // Call optional external error callback (e.g., for Sentry)
    this.props.onError?.(error, errorInfo);
  }

  /**
   * Handler to reset error state and attempt recovery.
   */
  handleRetry = (): void => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  /**
   * Handler to reload the page as last resort.
   */
  handleReload = (): void => {
    window.location.reload();
  };

  /**
   * Render method that returns either error UI or child components.
   */
  override render(): ReactNode {
    if (this.state.hasError) {
      // Use custom fallback if provided
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Render default error UI
      return (
        <Translation ns="common">
          {(t) => (
            <div
              role="alert"
              className="min-h-screen flex items-center justify-center bg-surface-base p-4"
            >
              <div
                className={cn(
                  'max-w-md w-full',
                  'border border-surface-border bg-surface-raised',
                  spacing.pad.lg,
                  'shadow-lg',
                  radius.xl,
                )}
              >
                {/* Header with icon */}
                <div className={cn('flex items-center', spacing.gap.default, 'mb-4')}>
                  <div
                    className={cn(
                      'flex items-center justify-center',
                      'h-10 w-10 rounded-full',
                      'bg-status-error/10',
                    )}
                  >
                    <AlertIcon />
                  </div>
                  <div>
                    <h2 className="text-lg font-semibold text-text-primary">
                      {t('errorBoundary.title')}
                    </h2>
                    <p className="text-sm text-text-muted">{t('errorBoundary.defaultMessage')}</p>
                  </div>
                </div>

                {/* Error details */}
                {this.state.error && (
                  <div
                    className={cn(
                      'mb-4',
                      spacing.pad.sm,
                      radius.lg,
                      'bg-surface-base border border-surface-border',
                    )}
                  >
                    <p className="text-sm font-medium text-text-primary mb-1">
                      {t('errorBoundary.errorDetails')}
                    </p>
                    <p className="text-sm text-status-error font-mono break-all">
                      {this.state.error.message}
                    </p>
                  </div>
                )}

                {/* Action buttons */}
                <div className={cn('flex', spacing.gap.default)}>
                  <button
                    type="button"
                    onClick={this.handleRetry}
                    className={cn(
                      button.base,
                      button.variant.primary,
                      button.size.md,
                      'flex-1 justify-center',
                    )}
                  >
                    <RefreshIcon />
                    {t('errorBoundary.tryAgain')}
                  </button>
                  <button
                    type="button"
                    onClick={this.handleReload}
                    className={cn(
                      button.base,
                      button.variant.secondary,
                      button.size.md,
                      'flex-1 justify-center',
                    )}
                  >
                    {t('errorBoundary.reload')}
                  </button>
                </div>

                {/* Help text */}
                <p className="mt-4 text-xs text-center text-text-muted">
                  {t('errorBoundary.persistMessage')}
                </p>
              </div>
            </div>
          )}
        </Translation>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;

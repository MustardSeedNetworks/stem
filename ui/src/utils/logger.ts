/**
 * @fileoverview Development Logger Utility
 * @description Provides logging utilities that are no-ops in production.
 *              Uses proper logging service hooks for future monitoring integration.
 */

/**
 * Check if running in development mode.
 * In production builds, Vite replaces import.meta.env.DEV with false
 * and tree-shakes the logging code.
 */
const isDev: boolean = import.meta.env.DEV;

/**
 * Error context for logging.
 */
interface ErrorContext {
  component?: string;
  action?: string;
  additionalData?: Record<string, unknown>;
}

/**
 * Log an error with optional context.
 * In production, this is a no-op.
 * In development, logs to console.error.
 *
 * @param error - The error to log
 * @param context - Additional context about where/why the error occurred
 */
export function logError(error: unknown, context?: ErrorContext): void {
  if (!isDev) {
    return;
  }

  const errorMessage = error instanceof Error ? error.message : String(error);
  const errorStack = error instanceof Error ? error.stack : undefined;

  // In development, use console.error for visibility
  console.error('[STEM Error]', {
    message: errorMessage,
    stack: errorStack,
    ...context,
    timestamp: new Date().toISOString(),
  });
}

/**
 * Log a warning with optional context.
 * In production, this is a no-op.
 *
 * @param message - Warning message
 * @param context - Additional context
 */
export function logWarn(message: string, context?: ErrorContext): void {
  if (!isDev) {
    return;
  }

  console.warn('[STEM Warning]', {
    message,
    ...context,
    timestamp: new Date().toISOString(),
  });
}

/**
 * Toast context — kept apart from Toast.tsx so the provider file exports only
 * components (react-refresh).
 */
import { type Context, createContext } from 'react';
import type { AlertTone } from '../../shared/Alert';

export type ToastType = AlertTone;

export interface ToastContextType {
  /** Show a toast. `duration` is in ms; 0 keeps it until dismissed. */
  addToast: (message: string, type?: ToastType, duration?: number) => void;
  removeToast: (id: string) => void;
}

export const ToastContext: Context<ToastContextType | null> =
  createContext<ToastContextType | null>(null);

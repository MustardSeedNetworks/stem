/**
 * Toast — non-modal feedback for an action that finished somewhere the
 * operator is not looking (a failed Start, a settings save that was refused).
 *
 * Ported from seed's Toast/useToast. A toast is an Alert on a raised surface,
 * so the two read as one family; the queue and auto-dismiss are the only new
 * behaviour. Toasts stack bottom-right and clear themselves after `duration`.
 */
import {
  type ReactElement,
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';
import { useTranslation } from 'react-i18next';
import { Alert } from './Alert';
import { ToastContext, type ToastType } from './toastContext';

const DEFAULT_DURATION_MS = 5000;

interface ToastEntry {
  id: string;
  message: string;
  type: ToastType;
  duration: number;
}

function ToastItem({
  toast,
  onClose,
}: {
  toast: ToastEntry;
  onClose: (id: string) => void;
}): ReactElement {
  const { id, duration } = toast;
  useEffect(() => {
    if (duration <= 0) {
      return;
    }
    const timer = setTimeout(() => onClose(id), duration);
    return () => clearTimeout(timer);
  }, [id, duration, onClose]);

  return (
    <Alert
      status={toast.type}
      onDismiss={() => onClose(id)}
      className="bg-surface-raised shadow-lg"
      data-testid="toast"
    >
      {toast.message}
    </Alert>
  );
}

export function ToastProvider({ children }: { children: ReactNode }): ReactElement {
  const { t } = useTranslation('common');
  const [toasts, setToasts] = useState<ToastEntry[]>([]);

  const removeToast = useCallback((id: string): void => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }, []);

  const addToast = useCallback(
    (message: string, type: ToastType = 'info', duration: number = DEFAULT_DURATION_MS): void => {
      const id = crypto.randomUUID();
      setToasts((prev) => [...prev, { id, message, type, duration }]);
    },
    [],
  );

  const value = useMemo(() => ({ addToast, removeToast }), [addToast, removeToast]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <section
        aria-live="polite"
        aria-label={t('accessibility.notifications')}
        className="fixed bottom-4 right-4 z-[60] flex w-full max-w-sm flex-col gap-compact"
      >
        {toasts.map((toast) => (
          <ToastItem key={toast.id} toast={toast} onClose={removeToast} />
        ))}
      </section>
    </ToastContext.Provider>
  );
}

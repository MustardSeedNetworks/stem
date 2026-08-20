/**
 * FieldError — the message under a field that failed validation.
 *
 * All seven config forms declared their own identical copy of this. Seven
 * copies of four lines is not a crisis, but it is seven places for the icon,
 * the spacing or the colour token to drift, and the forms are meant to read
 * as one surface.
 */
import { AlertTriangle } from 'lucide-react';
import type { ReactElement } from 'react';

export function FieldError({ message }: { message?: string }): ReactElement | null {
  if (!message) {
    return null;
  }
  return (
    <div className="mt-tight text-xs text-status-error flex items-center gap-tight">
      <AlertTriangle className="w-3 h-3" />
      {message}
    </div>
  );
}

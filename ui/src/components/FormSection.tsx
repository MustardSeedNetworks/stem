/**
 * FormSection — a titled group of fields inside a test-configuration form.
 *
 * The seven config forms each hand-rolled this: a `stack` div wrapping an
 * uppercase `text-xs` heading. Seven copies agreed by copy-paste rather than
 * by anything enforcing it, which is how a heading drifts a class at a time
 * and nobody notices until two forms sit side by side in Certify.
 *
 * A section's heading may carry a HelpIcon; pass it as `help` rather than
 * building the row again.
 */
import type { ReactElement, ReactNode } from 'react';

interface FormSectionProps {
  title: string;
  /** Usually a <HelpIcon>, rendered beside the heading. */
  help?: ReactNode;
  children: ReactNode;
}

export function FormSection({ title, help, children }: FormSectionProps): ReactElement {
  return (
    <div className="stack">
      <div className="text-xs font-semibold text-text-muted uppercase tracking-wide flex items-center gap-tight">
        {title}
        {help}
      </div>
      {children}
    </div>
  );
}

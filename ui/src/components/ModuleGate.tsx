/**
 * ModuleGate — a Pro module on Free says so before the run, not after it.
 *
 * A Free stem could configure a whole RFC 2544 run and only learn the module
 * was not included when Start came back 402 (UI-STEM-15). The module now
 * renders its pitch above the form, and the form below it is the preview:
 * real, readable, and inert.
 *
 * `inert` and not `aria-hidden`: the form is the thing an operator reads to
 * decide whether to buy it, so it stays in the accessibility tree and loses
 * only its interaction and its tab stops. `aria-describedby` ties it to the
 * pitch that explains why it cannot be filled in.
 *
 * The gate wraps the form, not the page: the role banner's switch action, the
 * empty state's Settings link and the Start control are all things an operator
 * on Free must still be able to use, and an inert page body would kill them.
 *
 * The run-time 402 stays exactly where it was. This is the explanation the
 * operator was missing, not a replacement for the server's answer — the UI
 * gate and the licence gate are two different things and only one of them is
 * authoritative.
 */

import type { ReactElement, ReactNode } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { type GatedModulePath, MODULE_FEATURES } from '../constants/moduleFeatures';
import { useLicense } from '../contexts/LicenseContext';

interface ModuleGateProps {
  path: GatedModulePath;
  /**
   * Whether these children are the module's form, and so a picture. False
   * means they are something the operator still needs to act on — the empty
   * state's Settings link is how they configure an interface in the first
   * place — so the pitch appears above them and they stay live.
   */
  preview?: boolean;
  children: ReactNode;
}

export function ModuleGate({ path, preview = true, children }: ModuleGateProps): ReactElement {
  const { hasFeature, loading, status } = useLicense();
  const { t } = useTranslation('common');
  const { features, i18nKey } = MODULE_FEATURES[path];

  // While the status is unknown the form is the form. A null flash of an
  // already-visible page is worse than a pitch that arrives a moment late,
  // and `status === null` after the fetch settles means the daemon did not
  // answer — an unreachable daemon is not an entitlement, and must never put
  // a paying operator behind a pitch. The server's 402 is the authority here;
  // this is only the explanation that used to be missing.
  if (loading || status === null || features.some(hasFeature)) {
    return <>{children}</>;
  }

  const pitchId = `module-gate-pitch-${i18nKey}`;

  return (
    <div className="stack-sm" data-testid="module-gate" data-module={i18nKey}>
      <section
        id={pitchId}
        data-testid="module-gate-pitch"
        className="w-full rounded-lg border border-brand-primary/30 bg-brand-primary/5 pad-sm stack-xs"
      >
        <h2 className="body-large font-semibold text-text-primary">
          {t('license.gated.title', { module: t(`license.gated.modules.${i18nKey}.name`) })}
        </h2>
        <p className="body-small text-text-secondary">
          {t(`license.gated.modules.${i18nKey}.pitch`)}
        </p>
        <p className="caption text-text-secondary">
          <Trans
            i18nKey="license.gated.actions"
            ns="common"
            components={{
              code: <code className="mx-1 px-1 rounded bg-surface-raised" />,
              code2: <code className="ml-tight px-1 rounded bg-surface-raised" />,
            }}
          />
        </p>
      </section>

      {preview ? (
        <div inert={true} aria-describedby={pitchId} className="opacity-60">
          {children}
        </div>
      ) : (
        children
      )}
    </div>
  );
}

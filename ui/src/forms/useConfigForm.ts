/**
 * useConfigForm — shared react-hook-form wiring for stem's test-config forms.
 *
 * The seven ConfigForms (RFC 2544 / 2889 / 6349, Y.1564 / 1731, TSN,
 * TrafficGen) all hold their canonical state in app context (so the
 * test-runner page can read it on dispatch) and accept a `setConfig`
 * callback to push changes back. react-hook-form normally owns form
 * state internally — we keep both worlds happy by:
 *
 *   1. seeding useForm() with defaultValues = the current external config,
 *   2. subscribing to form.watch() and pushing every accepted change
 *      back to the parent via setConfig,
 *   3. running validation onChange so the UI shows inline errors at
 *      the field where the bad value was typed.
 *
 * Callers register inputs with form.register('fieldName', { valueAsNumber:
 * true }) for numeric fields. The schema then coerces strings to numbers
 * (HTML inputs always emit strings) before the resolver runs.
 *
 * Forms with cross-field rules (Y.1564 FDV ≤ FD, RFC6349 minRTT ≤ maxRTT,
 * TSN jitter ≤ latency) get the cross-field error from form.formState.
 * errors.root — display it once at the form footer.
 */
import { valibotResolver } from '@hookform/resolvers/valibot';
import { useEffect, useRef } from 'react';
import {
  type DefaultValues,
  type FieldValues,
  type Resolver,
  type UseFormReturn,
  useForm,
} from 'react-hook-form';

interface UseConfigFormOptions<TConfig extends FieldValues> {
  // The schema type is intentionally unknown — valibot's GenericSchema
  // generics don't compose with react-hook-form's FieldValues constraint
  // cleanly. The runtime contract is enforced by the schema itself; the
  // type-erased shape here just keeps callers from importing the valibot
  // type machinery at every call site.
  schema: unknown;
  config: TConfig;
  setConfig: (next: TConfig) => void;
}

/**
 * useConfigForm wires a parent-owned config object to react-hook-form
 * with valibot validation. Returns the standard react-hook-form API so
 * callers can register inputs and read errors normally.
 */
export function useConfigForm<TConfig extends FieldValues>(
  options: UseConfigFormOptions<TConfig>,
): UseFormReturn<TConfig> {
  const { schema, config, setConfig } = options;

  const form = useForm<TConfig>({
    resolver: valibotResolver(schema as never) as Resolver<TConfig>,
    defaultValues: config as DefaultValues<TConfig>,
    mode: 'onChange',
  });

  // Push every form change back to parent context, but only for
  // shape-valid values. If the resolver rejects a value, the field
  // stays in error state in react-hook-form's internal state and the
  // parent context is *not* updated — that prevents the test-runner
  // from seeing a bad config on dispatch.
  const setConfigRef = useRef(setConfig);
  setConfigRef.current = setConfig;

  useEffect(() => {
    const subscription = form.watch((values, { type }) => {
      // Only forward on actual changes (not the initial subscription event).
      if (type !== 'change') return;
      // Run an extra parse to be sure cross-field constraints pass before
      // forwarding. Skipping invalid intermediate states keeps the parent
      // store consistent.
      const trigger = form.trigger;
      void trigger().then((isValid) => {
        if (isValid) {
          setConfigRef.current(values as TConfig);
        }
      });
    });
    return () => subscription.unsubscribe();
  }, [form]);

  return form;
}

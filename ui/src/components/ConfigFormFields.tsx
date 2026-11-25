/**
 * Field adapters between react-hook-form and the shared form controls.
 *
 * `register` binds a native input through a ref and a change event; the shared
 * Select and Checkbox are controlled and report a value, so these two bind them
 * with `Controller`. Everything visible (label, hint, error, width cap) is the
 * shared FormField and the shared control; nothing here draws anything.
 */
import type { ReactElement } from 'react';
import { type Control, Controller, type FieldPath, type FieldValues } from 'react-hook-form';
import { Checkbox } from '../shared/Checkbox';
import { FormField } from '../shared/FormField';
import { Select } from '../shared/Select';
import { HelpIcon } from './HelpIcon';

interface NumberSelectFieldProps<T extends FieldValues> {
  control: Control<T>;
  name: FieldPath<T>;
  label: string;
  hint?: string;
  error?: string | undefined;
  options: readonly { value: number; label: string }[];
}

/** A Select whose value is a number in the form (every choice list here is numeric). */
export function NumberSelectField<T extends FieldValues>({
  control,
  name,
  label,
  hint,
  error,
  options,
}: NumberSelectFieldProps<T>): ReactElement {
  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <FormField label={label} {...(hint ? { hint } : {})} {...(error ? { error } : {})}>
          {(controlProps) => (
            <Select
              {...controlProps}
              name={field.name}
              ref={field.ref}
              onBlur={field.onBlur}
              value={String(field.value)}
              onValueChange={(value): void => field.onChange(Number(value))}
              options={options.map((option) => ({
                value: String(option.value),
                label: option.label,
              }))}
            />
          )}
        </FormField>
      )}
    />
  );
}

interface CheckboxFieldProps<T extends FieldValues> {
  control: Control<T>;
  name: FieldPath<T>;
  label: string;
  /** One line under the label. */
  description?: string;
  /** The help text: shown on the help icon beside the checkbox. */
  help?: string;
}

/** A boolean field as the shared Checkbox, with its help beside it. */
export function CheckboxField<T extends FieldValues>({
  control,
  name,
  label,
  description,
  help,
}: CheckboxFieldProps<T>): ReactElement {
  return (
    <Controller
      control={control}
      name={name}
      render={({ field }) => (
        <div className="flex items-start gap-1">
          <Checkbox
            className="min-w-0 flex-1"
            checked={Boolean(field.value)}
            onChange={field.onChange}
            label={label}
            {...(description ? { description } : {})}
          />
          {help ? <HelpIcon tooltip={help} /> : null}
        </div>
      )}
    />
  );
}

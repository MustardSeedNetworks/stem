/**
 * Valibot schemas for stem's auth / MFA / recovery / setup forms.
 *
 * These cover the user-input boundary of the auth flow: login,
 * password change, MFA TOTP setup/verify/disable, recovery code
 * entry, and the initial setup wizard.
 *
 * The Go side validates again on receipt — these schemas exist for the
 * UI to show inline per-field errors before the network call, not as
 * the security boundary.
 */
import * as v from 'valibot';

/**
 * 6-digit TOTP code. Whitespace is trimmed (users often paste with
 * spaces from authenticator apps). Stored without separators.
 */
export const TotpCodeSchema = v.pipe(
  v.string('Code is required'),
  v.trim(),
  v.regex(/^\d{6}$/, 'Code must be exactly 6 digits'),
);

/** Username for login + recovery flows. */
export const UsernameSchema = v.pipe(
  v.string('Username is required'),
  v.trim(),
  v.minLength(1, 'Username is required'),
  v.maxLength(128, 'Username is too long'),
);

/** Password for login + MFA-disable flows. Length-only check; the Go
 * side enforces complexity rules. */
export const PasswordSchema = v.pipe(
  v.string('Password is required'),
  v.minLength(1, 'Password is required'),
  v.maxLength(512, 'Password is too long'),
);

/** Recovery code: 16 hex characters in groups of 4. Accept with or
 * without separators on input; normalize before posting. */
export const RecoveryCodeSchema = v.pipe(
  v.string('Recovery code is required'),
  v.trim(),
  v.transform((s) => s.replace(/[-\s]/g, '').toLowerCase()),
  v.regex(/^[0-9a-f]{16}$/, 'Recovery code must be 16 hex characters'),
);

// =============================================================================
// Form schemas
// =============================================================================

export const LoginSchema = v.object({
  username: UsernameSchema,
  password: PasswordSchema,
});

export const TotpSetupVerifySchema = v.object({
  code: TotpCodeSchema,
});

export const TotpDisableSchema = v.object({
  password: PasswordSchema,
  code: TotpCodeSchema,
});

export const MfaVerifySchema = v.object({
  code: TotpCodeSchema,
});

export const RecoveryEnterSchema = v.object({
  username: UsernameSchema,
  recoveryCode: RecoveryCodeSchema,
});

/**
 * Setup wizard: initial-admin creation. Password confirmation is a
 * cross-field check; the resolver surfaces under formState.errors.root.
 */
export const SetupWizardSchema = v.pipe(
  v.object({
    username: UsernameSchema,
    password: v.pipe(
      v.string('Password is required'),
      v.minLength(12, 'Password must be at least 12 characters'),
      v.maxLength(512, 'Password is too long'),
    ),
    confirmPassword: v.string(),
  }),
  v.check((c) => c.password === c.confirmPassword, 'Passwords do not match'),
);

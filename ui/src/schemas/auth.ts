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
 * Setup wizard: initial-admin creation. The wizard fixes the username
 * (from setup status), so the schema only validates the password the
 * user actually types. Password confirmation is a cross-field check;
 * the resolver surfaces it under formState.errors.root.
 */
export const SetupWizardSchema = v.pipe(
  v.object({
    password: v.pipe(
      v.string('Password is required'),
      v.minLength(12, 'Password must be at least 12 characters'),
      v.maxLength(512, 'Password is too long'),
    ),
    confirmPassword: v.string(),
  }),
  // Forwarded onto confirmPassword deliberately. A bare v.check() produces an
  // issue with no path, and the react-hook-form resolver cannot attach a
  // path-less issue to a field: the form did not block, submitted with no
  // values at all, and the message could never render anywhere.
  v.forward(
    v.check((c) => c.password === c.confirmPassword, 'Passwords do not match'),
    ['confirmPassword'],
  ),
);

/**
 * Recovery completion: filesystem-token-based password reset. The
 * operator writes the token to a file on the server, the user pastes
 * the token here, then enters and confirms a new password.
 */
/**
 * RecoveryInstructionsSchema — body of GET /api/v1/recovery/instructions.
 *
 * Validated rather than cast because the panel renders `steps` directly: a 200
 * whose body lacks it (a proxy page, a changed contract) crashed the whole
 * recovery form, which is the one way back into a locked-out account.
 */
export const RecoveryInstructionsSchema = v.object({
  triggerFile: v.string(),
  tokenFile: v.string(),
  expiryTime: v.string(),
  steps: v.array(v.string()),
});

export type RecoveryInstructions = v.InferOutput<typeof RecoveryInstructionsSchema>;

/**
 * parseRecoveryInstructions — returns the parsed instructions, or null when the
 * body is not the expected shape. Null renders no panel, which is the same as
 * the endpoint being unavailable: instructions are a convenience, and losing
 * them must not block recovery.
 */
export function parseRecoveryInstructions(body: unknown): RecoveryInstructions | null {
  const result = v.safeParse(RecoveryInstructionsSchema, body);
  return result.success ? result.output : null;
}

export const RecoveryCompleteSchema = v.pipe(
  v.object({
    token: v.pipe(
      v.string('Recovery token is required'),
      v.trim(),
      v.minLength(1, 'Recovery token is required'),
    ),
    password: v.pipe(
      v.string('Password is required'),
      v.minLength(12, 'Password must be at least 12 characters'),
      v.maxLength(512, 'Password is too long'),
    ),
    confirmPassword: v.string(),
  }),
  // Forwarded onto confirmPassword deliberately. A bare v.check() produces an
  // issue with no path, and the react-hook-form resolver cannot attach a
  // path-less issue to a field: the form did not block, submitted with no
  // values at all, and the message could never render anywhere.
  v.forward(
    v.check((c) => c.password === c.confirmPassword, 'Passwords do not match'),
    ['confirmPassword'],
  ),
);

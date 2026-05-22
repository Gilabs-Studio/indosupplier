export interface PasswordResetTokenPrefillPayload {
  userId: string;
  token: string;
  createdAt: number;
}

const RESET_TOKEN_PREFILL_KEY = "user-management:reset-token-prefill";

export function setPasswordResetTokenPrefill(payload: PasswordResetTokenPrefillPayload): void {
  if (typeof window === "undefined") return;
  try {
    window.sessionStorage.setItem(RESET_TOKEN_PREFILL_KEY, JSON.stringify(payload));
  } catch {
    // Ignore storage failures in restricted environments.
  }
}

export function getPasswordResetTokenPrefill(): PasswordResetTokenPrefillPayload | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.sessionStorage.getItem(RESET_TOKEN_PREFILL_KEY);
    if (!raw) return null;

    const parsed = JSON.parse(raw) as Partial<PasswordResetTokenPrefillPayload>;
    if (
      typeof parsed.userId !== "string" ||
      typeof parsed.token !== "string" ||
      typeof parsed.createdAt !== "number"
    ) {
      return null;
    }

    return {
      userId: parsed.userId,
      token: parsed.token,
      createdAt: parsed.createdAt,
    };
  } catch {
    return null;
  }
}

export function clearPasswordResetTokenPrefill(): void {
  if (typeof window === "undefined") return;
  try {
    window.sessionStorage.removeItem(RESET_TOKEN_PREFILL_KEY);
  } catch {
    // Ignore storage failures in restricted environments.
  }
}

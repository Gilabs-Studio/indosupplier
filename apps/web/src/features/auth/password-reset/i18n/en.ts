const _passwordReset = {
  forgotPassword: "Forgot Password",
  forgotPasswordDescription: "Enter your email address to submit a password reset request.",
  email: "Email",
  emailPlaceholder: "Enter your email",
  emailRequired: "Email is required",
  invalidEmail: "Please enter a valid email address",
  sendResetLink: "Send Reset Link",
  sending: "Sending...",
  cancel: "Cancel",
  resetLinkSent: "Reset Link Sent",
  resetLinkSentDescription: "Your password reset request has been sent to admin. You will be contacted for the next steps.",
  linkExpiresIn: "The link expires in 24 hours",
  checkSpamFolder: "Check your spam folder if you don't see the email",
  backToLogin: "Back to Login",
  noAccountYet: "Don't have an account?",
  signUp: "Sign Up",
  forgotPasswordSuccess: "Password reset request has been sent to admin.",
  forgotPasswordError: "Failed to process forgot password request. Please try again.",
  resetPassword: "Reset Password",
  resetPasswordDescription: "Enter your new password to complete the reset process.",
  newPassword: "New Password",
  passwordMinLength: "Password must be at least 6 characters",
  confirmPassword: "Confirm Password",
  passwordsDoNotMatch: "Passwords do not match",
  resetPasswordSuccess: "Password reset successfully. You can now log in with your new password.",
  resetPasswordError: "Failed to reset password. Please try again.",
  validateToken: "Validating reset link...",
  invalidToken: "This reset link is invalid or has expired. Please request a new one.",
  expiredToken: "This reset link has expired. Please request a new one.",
  alreadyUsedToken: "This reset link has already been used. Please request a new one.",
};

export const passwordResetEn = {
  // standard: messages.en.passwordReset.emailRequired
  passwordReset: {
    ..._passwordReset,
    // alias so code that mistakenly calls `passwordReset.passwordReset.*` still resolves
    passwordReset: _passwordReset,
  },
  // support older namespace usage like `auth.passwordReset`
  auth: {
    passwordReset: _passwordReset,
  },
};

export type PasswordResetTranslations = typeof passwordResetEn;

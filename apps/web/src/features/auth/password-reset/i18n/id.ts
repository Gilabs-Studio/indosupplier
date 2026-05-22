const _passwordReset = {
  forgotPassword: "Lupa Password",
  forgotPasswordDescription: "Masukkan alamat email Anda untuk mengirim permintaan reset password.",
  email: "Email",
  emailPlaceholder: "Masukkan email Anda",
  emailRequired: "Email harus diisi",
  invalidEmail: "Silahkan masukkan alamat email yang valid",
  sendResetLink: "Kirim Link Reset",
  sending: "Mengirim...",
  cancel: "Batal",
  resetLinkSent: "Link Reset Terkirim",
  resetLinkSentDescription: "Permintaan reset password Anda telah dikirim ke admin. Anda akan dihubungi untuk langkah selanjutnya.",
  linkExpiresIn: "Link akan berlaku selama 24 jam",
  checkSpamFolder: "Periksa folder spam jika Anda tidak melihat email",
  backToLogin: "Kembali ke Login",
  noAccountYet: "Tidak memiliki akun?",
  signUp: "Daftar",
  forgotPasswordSuccess: "Permintaan reset password telah dikirim ke admin.",
  forgotPasswordError: "Gagal memproses permintaan lupa password. Silahkan coba lagi.",
  resetPassword: "Reset Password",
  resetPasswordDescription: "Masukkan password baru Anda untuk menyelesaikan proses reset.",
  newPassword: "Password Baru",
  passwordMinLength: "Password harus minimal 6 karakter",
  confirmPassword: "Konfirmasi Password",
  passwordsDoNotMatch: "Password tidak cocok",
  resetPasswordSuccess: "Password berhasil direset. Anda sekarang dapat login dengan password baru Anda.",
  resetPasswordError: "Gagal mereset password. Silahkan coba lagi.",
  validateToken: "Memvalidasi link reset...",
  invalidToken: "Link reset ini tidak valid atau telah kadaluarsa. Silahkan minta yang baru.",
  expiredToken: "Link reset ini telah kadaluarsa. Silahkan minta yang baru.",
  alreadyUsedToken: "Link reset ini telah digunakan. Silahkan minta yang baru.",
};

export const passwordResetId = {
  passwordReset: {
    ..._passwordReset,
    passwordReset: _passwordReset,
  },
  auth: {
    passwordReset: _passwordReset,
  },
};

export type PasswordResetTranslationsId = typeof passwordResetId;

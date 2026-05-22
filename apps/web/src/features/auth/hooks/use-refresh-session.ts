import { useAuthStore } from "../stores/use-auth-store";
import { authService } from "../services/auth-service";
import { useLogout } from "./use-logout";

export function useRefreshSession() {
  const { setUser } = useAuthStore();
  const handleLogout = useLogout();

  const refreshSession = async () => {
    try {
      // Browser automatically sends HttpOnly refresh_token cookie
      const response = await authService.refreshToken();

      if (response.success && response.data) {
        const { user } = response.data;
        setUser(user);
        useAuthStore.setState({ isAuthenticated: true });
      }
    } catch {
      // Refresh failed, logout user
      await handleLogout();
      throw new Error("Session refresh failed");
    }
  };

  return { refreshSession };
}

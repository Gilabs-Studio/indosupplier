export interface SystemAdmin {
  id: string;
  email: string;
  name: string;
  permission_set: string;
  permissions: string[];
  status: string;
}

export interface SysadminLoginPayload {
  admin: SystemAdmin;
  token: string;
  refresh_token: string;
  expires_in: number;
}

export interface SysadminLoginResponse {
  success: boolean;
  data: SysadminLoginPayload;
}

export interface SysadminMeResponse {
  success: boolean;
  data: SystemAdmin;
}

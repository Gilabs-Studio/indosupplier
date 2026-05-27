export interface NavItem {
  id?: string;
  name: string;
  i18nKey?: string;
  icon: string;
  url: string;
  permission?: string;
  children?: NavItem[];
}

export const navigationConfig: NavItem[] = [
  {
    name: "Home",
    i18nKey: "home",
    icon: "home",
    url: "/",
  },
  {
    name: "Login",
    i18nKey: "login",
    icon: "log-in",
    url: "/login",
  },
];

"use client";

import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useSyncExternalStore,
} from "react";

type Theme = string;
type ThemeSetter = (theme: Theme | ((currentTheme: Theme) => Theme)) => void;

type Attribute = "class" | `data-${string}` | string | string[];

export type ThemeProviderProps = {
  children: React.ReactNode;
  attribute?: Attribute;
  defaultTheme?: Theme;
  enableSystem?: boolean;
  enableColorScheme?: boolean;
  disableTransitionOnChange?: boolean;
  forcedTheme?: Theme;
  storageKey?: string;
  themes?: string[];
  value?: Record<string, string>;
};

type ThemeContextValue = {
  theme: Theme | undefined;
  setTheme: ThemeSetter;
  forcedTheme?: Theme;
  resolvedTheme: Theme | undefined;
  themes: string[];
  systemTheme: "light" | "dark" | undefined;
};

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);
const THEME_EVENT = "gims-theme-storage-sync";

function getSystemTheme(): "light" | "dark" {
  if (typeof window === "undefined") {
    return "light";
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function disableTransitionsTemporarily() {
  const style = document.createElement("style");
  style.appendChild(
    document.createTextNode(
      "*,*::before,*::after{-webkit-transition:none!important;-moz-transition:none!important;-o-transition:none!important;transition:none!important}",
    ),
  );
  document.head.appendChild(style);
  return () => {
    window.getComputedStyle(document.body);
    setTimeout(() => {
      document.head.removeChild(style);
    }, 1);
  };
}

export function ThemeProvider({
  children,
  attribute = "class",
  defaultTheme = "system",
  enableSystem = true,
  enableColorScheme = true,
  disableTransitionOnChange = false,
  forcedTheme,
  storageKey = "theme",
  themes = ["light", "dark"],
  value,
}: ThemeProviderProps) {
  const subscribeTheme = useCallback(
    (onStoreChange: () => void) => {
      if (typeof window === "undefined") {
        return () => {};
      }

      const onStorage = (event: Event) => {
        if (event instanceof StorageEvent && event.key !== null && event.key !== storageKey) {
          return;
        }
        onStoreChange();
      };

      window.addEventListener("storage", onStorage);
      window.addEventListener(THEME_EVENT, onStorage);
      return () => {
        window.removeEventListener("storage", onStorage);
        window.removeEventListener(THEME_EVENT, onStorage);
      };
    },
    [storageKey],
  );

  const getThemeSnapshot = useCallback(() => {
    if (typeof window === "undefined") {
      return defaultTheme;
    }
    try {
      return localStorage.getItem(storageKey) ?? defaultTheme;
    } catch {
      return defaultTheme;
    }
  }, [defaultTheme, storageKey]);

  const theme = useSyncExternalStore(subscribeTheme, getThemeSnapshot, () => defaultTheme);

  const setTheme = useCallback<ThemeSetter>(
    (nextThemeOrFn) => {
      const nextTheme =
        typeof nextThemeOrFn === "function" ? nextThemeOrFn(theme) : nextThemeOrFn;

      if (typeof window === "undefined") {
        return;
      }

      try {
        localStorage.setItem(storageKey, nextTheme);
      } catch {
        return;
      }

      window.dispatchEvent(new Event(THEME_EVENT));
    },
    [storageKey, theme],
  );

  const systemTheme = useSyncExternalStore<"light" | "dark">(
    (onStoreChange) => {
      if (typeof window === "undefined") {
        return () => {};
      }
      const media = window.matchMedia("(prefers-color-scheme: dark)");
      const listener = () => onStoreChange();
      if (typeof media.addEventListener === "function") {
        media.addEventListener("change", listener);
        return () => media.removeEventListener("change", listener);
      }
      media.addListener(listener);
      return () => media.removeListener(listener);
    },
    getSystemTheme,
    () => "light",
  );

  const resolvedTheme = useMemo(() => {
    if (forcedTheme) {
      return forcedTheme;
    }
    if (theme === "system") {
      return enableSystem ? systemTheme : "light";
    }
    return theme;
  }, [enableSystem, forcedTheme, systemTheme, theme]);

  useEffect(() => {
    if (typeof document === "undefined") {
      return;
    }

    const root = document.documentElement;
    const nextTheme = resolvedTheme;
    if (!nextTheme) {
      return;
    }

    const mappedTheme = value?.[nextTheme] ?? nextTheme;
    const allThemeClasses = value ? [...themes.map((t) => value[t] ?? t), ...themes] : themes;
    const cleanupTransitions = disableTransitionOnChange ? disableTransitionsTemporarily() : null;

    const applyAttribute = (attr: string) => {
      if (attr === "class") {
        root.classList.remove(...allThemeClasses);
        root.classList.add(mappedTheme);
        return;
      }
      root.setAttribute(attr, mappedTheme);
    };

    if (Array.isArray(attribute)) {
      attribute.forEach(applyAttribute);
    } else {
      applyAttribute(attribute);
    }

    if (enableColorScheme) {
      if (mappedTheme === "dark" || mappedTheme === "light") {
        root.style.colorScheme = mappedTheme;
      }
    }

    cleanupTransitions?.();
  }, [
    attribute,
    disableTransitionOnChange,
    enableColorScheme,
    resolvedTheme,
    themes,
    value,
  ]);

  const availableThemes = useMemo(() => {
    if (!enableSystem) {
      return themes;
    }
    return themes.includes("system") ? themes : [...themes, "system"];
  }, [enableSystem, themes]);

  const contextValue = useMemo<ThemeContextValue>(
    () => ({
      theme,
      setTheme,
      forcedTheme,
      resolvedTheme,
      themes: availableThemes,
      systemTheme,
    }),
    [availableThemes, forcedTheme, resolvedTheme, setTheme, systemTheme, theme],
  );

  return <ThemeContext.Provider value={contextValue}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  return (
    useContext(ThemeContext) ?? {
      theme: undefined,
      setTheme: () => {},
      forcedTheme: undefined,
      resolvedTheme: undefined,
      themes: [],
      systemTheme: undefined,
    }
  );
}


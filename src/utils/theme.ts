import { ThemeMode } from '../types/launcher';

export const themes = {
  light: 'light',
  dark: 'dark',
  system: 'system',
} as const;

export const getSystemTheme = (): 'light' | 'dark' => {
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
    return 'dark';
  }
  return 'light';
};

export const applyTheme = (theme: ThemeMode): void => {
  const root = document.documentElement;

  if (theme === 'system') {
    const systemTheme = getSystemTheme();
    root.setAttribute('data-theme', systemTheme);
  } else {
    root.setAttribute('data-theme', theme);
  }

  // Store preference
  localStorage.setItem('theme-preference', theme);
};

export const getStoredTheme = (): ThemeMode => {
  const stored = localStorage.getItem('theme-preference');
  if (stored && isValidTheme(stored)) {
    return stored;
  }
  return 'system';
};

export const isValidTheme = (theme: string): theme is ThemeMode => {
  return theme === 'light' || theme === 'dark' || theme === 'system';
};

export const initializeTheme = (): void => {
  const storedTheme = getStoredTheme();
  applyTheme(storedTheme);

  // Listen for system theme changes if using system preference
  if (storedTheme === 'system') {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      applyTheme('system');
    });
  }
};

export const toggleTheme = (currentTheme: ThemeMode): ThemeMode => {
  const nextTheme = currentTheme === 'light' ? 'dark' : 'light';
  applyTheme(nextTheme);
  return nextTheme;
};
/**
 * User preference & recommendation tracking utility.
 * Captures user search and category browsing intent to personalize catalog views.
 */

const STORAGE_KEY = "indosupplier_user_interest_v1";
export const INTEREST_CHANGE_EVENT = "indosupplier:interest-updated";

export interface SearchInterestItem {
  term: string;
  count: number;
  lastUsed: number;
}

export interface UserSearchPreferences {
  activeInterest?: string;
  activeCategory?: string;
  queries: SearchInterestItem[];
  categories: SearchInterestItem[];
}

function getStoredPreferences(): UserSearchPreferences {
  if (typeof window === "undefined") {
    return { queries: [], categories: [] };
  }
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { queries: [], categories: [] };
    const parsed = JSON.parse(raw);
    return {
      activeInterest: parsed.activeInterest,
      activeCategory: parsed.activeCategory,
      queries: Array.isArray(parsed.queries) ? parsed.queries : [],
      categories: Array.isArray(parsed.categories) ? parsed.categories : [],
    };
  } catch {
    return { queries: [], categories: [] };
  }
}

function saveStoredPreferences(pref: UserSearchPreferences): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(pref));
    window.dispatchEvent(new CustomEvent(INTEREST_CHANGE_EVENT, { detail: pref }));
  } catch (err) {
    console.warn("Failed to persist user search preferences:", err);
  }
}

export function recordUserSearch(query?: string, categorySlug?: string): void {
  if (typeof window === "undefined") return;
  const pref = getStoredPreferences();
  const now = Date.now();

  let hasChanged = false;

  if (query && query.trim().length >= 2) {
    const cleanQuery = query.trim().toLowerCase();
    pref.activeInterest = cleanQuery;

    const existingIdx = pref.queries.findIndex((q) => q.term === cleanQuery);
    if (existingIdx >= 0) {
      pref.queries[existingIdx].count += 1;
      pref.queries[existingIdx].lastUsed = now;
    } else {
      pref.queries.unshift({ term: cleanQuery, count: 1, lastUsed: now });
    }
    // Cap to top 10
    pref.queries = pref.queries.slice(0, 10);
    hasChanged = true;
  }

  if (categorySlug && categorySlug.trim()) {
    const cleanCat = categorySlug.trim().toLowerCase();
    pref.activeCategory = cleanCat;

    const existingIdx = pref.categories.findIndex((c) => c.term === cleanCat);
    if (existingIdx >= 0) {
      pref.categories[existingIdx].count += 1;
      pref.categories[existingIdx].lastUsed = now;
    } else {
      pref.categories.unshift({ term: cleanCat, count: 1, lastUsed: now });
    }
    pref.categories = pref.categories.slice(0, 10);
    hasChanged = true;
  }

  if (hasChanged) {
    saveStoredPreferences(pref);
  }
}

export function getUserTopInterest(): { query?: string; category?: string } | null {
  const pref = getStoredPreferences();
  if (pref.activeInterest) {
    return { query: pref.activeInterest, category: pref.activeCategory };
  }
  if (pref.queries.length > 0) {
    // Sort by count and recency
    const sorted = [...pref.queries].sort((a, b) => b.count * 2 + (b.lastUsed > a.lastUsed ? 1 : -1));
    return { query: sorted[0]?.term, category: pref.activeCategory };
  }
  if (pref.activeCategory) {
    return { category: pref.activeCategory };
  }
  return null;
}

export function dismissUserInterest(): void {
  const pref = getStoredPreferences();
  delete pref.activeInterest;
  delete pref.activeCategory;
  saveStoredPreferences(pref);
}

export function clearAllUserPreferences(): void {
  if (typeof window === "undefined") return;
  try {
    localStorage.removeItem(STORAGE_KEY);
    window.dispatchEvent(new CustomEvent(INTEREST_CHANGE_EVENT, { detail: null }));
  } catch (err) {
    console.warn("Failed to clear preferences:", err);
  }
}

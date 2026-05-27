import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { formatInTimeZone } from "date-fns-tz";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatCurrency(
  value: number | string | null | undefined,
  locale: string = "id-ID",
  currency: string = "IDR",
): string {
  if (value === null || value === undefined || value === "") {
    return "Rp 0";
  }
  const numValue = typeof value === "string" ? parseFloat(value) : value;
  if (isNaN(numValue)) {
    return "Rp 0";
  }
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(numValue);
}

export function formatDate(
  date: Date | string | null | undefined,
  locale: string = "en-GB",
): string {
  if (!date) return "";
  let dateObj: Date;
  if (typeof date === "string") {
    // Parse YYYY-MM-DD as local midnight to avoid UTC timezone shift
    if (/^\d{4}-\d{2}-\d{2}$/.test(date)) {
      const [y, m, d] = date.split("-").map(Number);
      dateObj = new Date(y, m - 1, d);
    } else {
      dateObj = new Date(date);
    }
  } else {
    dateObj = date;
  }
  if (isNaN(dateObj.getTime())) return "";

  return new Intl.DateTimeFormat(locale, {
    day: "2-digit",
    month: "short",
    year: "numeric",
  }).format(dateObj);
}

/** Parse a YYYY-MM-DD string as local midnight (avoids UTC timezone shift). */
export function parseLocalDate(dateStr: string): Date {
  const [y, m, d] = dateStr.split("-").map(Number);
  return new Date(y, m - 1, d);
}

/** Format a local Date to YYYY-MM-DD without UTC conversion. */
export function toLocalDateString(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

export function formatTime(
  date: Date | string | null | undefined,
  locale: string = "id-ID",
): string {
  if (!date) return "";
  const dateObj = typeof date === "string" ? new Date(date) : date;
  if (isNaN(dateObj.getTime())) return "";

  return new Intl.DateTimeFormat(locale, {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(dateObj);
}

export function resolveImageUrl(
  url: string | null | undefined,
): string | undefined {
  if (!url) return undefined;
  if (
    url.startsWith("http") ||
    url.startsWith("blob:") ||
    url.startsWith("data:")
  ) {
    return url;
  }
  // Try to use NEXT_PUBLIC_API_URL, fallback to localhost:8088
  // Note: This needs to match the backend's static file serving URL
  const baseUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8088";

  // Clean up double slashes if url starts with /
  const cleanPath = url.startsWith("/") ? url.substring(1) : url;

  // If the path already has "uploads/" prefix and API serves from root, good.
  // Assuming API serves static files typically under /uploads or similar route.
  // The user paths were like apps/api/uploads/..., so likely just appending to base is correct if backend routes it.
  // If backend is Go Fiber/Echo/Gin usually Static("/uploads", "./uploads")
  return `${baseUrl}/${cleanPath}`;
}

/**
 * Sorts an array of items alphabetically based on a label extracted from each item.
 * @param data The array of items to sort.
 * @param getLabel A function that extracts the string label to sort by from an item.
 * @returns A new sorted array.
 */
export function sortOptions<T>(
  data: readonly T[],
  getLabel: (item: T) => string,
): T[] {
  return [...data].sort((a, b) => {
    const labelA = getLabel(a);
    const labelB = getLabel(b);
    return labelA.localeCompare(labelB);
  });
}

/**
 * Formats a phone number into a WhatsApp link.
 * @param phone The phone number string.
 * @param message Optional message to pre-fill.
 * @returns A formatted wa.me link.
 */
export function formatWhatsAppLink(phone?: string, message?: string): string {
  if (!phone) return "#";
  // Remove non-digit characters
  const digits = phone.replace(/\D/g, "");
  // Basic normalization for ID numbers (08... -> 628...)
  const normalized = digits.startsWith("0")
    ? "62" + digits.substring(1)
    : digits;

  if (message) {
    // Normalize string to NFC and encode
    const sanitizedMessage = message.trim().normalize("NFC");
    return `https://api.whatsapp.com/send?phone=${normalized}&text=${encodeURIComponent(sanitizedMessage)}`;
  }

  return `https://wa.me/${normalized}`;
}

/**
 * Default timezone for Indonesia (WIB - Western Indonesian Time)
 * Can be overridden based on user/company location
 */
export const DEFAULT_TIMEZONE = "Asia/Jakarta";

/**
 * Indonesia timezones based on region
 */
export const INDONESIA_TIMEZONES = {
  WIB: "Asia/Jakarta", // UTC+7 - Sumatra, Java, etc.
  WITA: "Asia/Makassar", // UTC+8 - Sulawesi, Bali, Nusa Tenggara, etc.
  WIT: "Asia/Jayapura", // UTC+9 - Papua, Maluku, etc.
} as const;

/**
 * Get timezone based on longitude (approximate for Indonesia)
 * WIB: < 120°E, WITA: 120°E - 135°E, WIT: > 135°E
 */
export function getTimezoneFromLongitude(longitude: number): string {
  if (longitude < 120) {
    return INDONESIA_TIMEZONES.WIB;
  } else if (longitude < 135) {
    return INDONESIA_TIMEZONES.WITA;
  } else {
    return INDONESIA_TIMEZONES.WIT;
  }
}

/**
 * Format a UTC datetime string to local time with timezone
 * @param utcDate - UTC datetime string from backend (e.g., "2026-03-30T17:43:27Z")
 * @param timezone - Target timezone (default: Asia/Jakarta)
 * @param format - Output format (default: "HH:mm" for time only)
 * @returns Formatted local time string
 */
export function formatUTCToLocal(
  utcDate: string | null | undefined,
  timezone: string = DEFAULT_TIMEZONE,
  format: string = "HH:mm",
): string {
  if (!utcDate) return "-";

  try {
    return formatInTimeZone(utcDate, timezone, format);
  } catch {
    return "-";
  }
}

/**
 * Format attendance time (check_in_time/check_out_time) from UTC to local
 * This handles the time-only format "HH:mm:ss" from backend
 * @param timeStr - Time string from backend (e.g., "17:43:27" or "2026-03-30T17:43:27Z")
 * @param dateStr - Optional date string if timeStr doesn't include date
 * @param timezone - Target timezone
 * @returns Formatted local time (HH:mm)
 */
export function formatAttendanceTime(
  timeStr: string | null | undefined,
  dateStr?: string,
  timezone: string = DEFAULT_TIMEZONE,
): string {
  if (!timeStr) return "-";

  try {
    // If timeStr includes date (ISO format), use it directly
    if (timeStr.includes("T") || timeStr.includes("Z")) {
      return formatInTimeZone(timeStr, timezone, "HH:mm");
    }

    // If timeStr is just time (HH:mm:ss), combine with date
    if (dateStr) {
      const dateTimeStr = `${dateStr}T${timeStr}`;
      return formatInTimeZone(dateTimeStr, timezone, "HH:mm");
    }

    // Fallback: assume today and append timezone
    const today = new Date().toISOString().split("T")[0];
    const dateTimeStr = `${today}T${timeStr}`;
    return formatInTimeZone(dateTimeStr, timezone, "HH:mm");
  } catch {
    return timeStr.substring(0, 5); // Fallback to original HH:mm
  }
}

/**
 * Get current timezone based on user's company
 * In a real app, this would come from user context/company settings
 */
export function getUserTimezone(): string {
  // TODO: Get from user context/company settings
  // For now, default to WIB (Jakarta)
  return DEFAULT_TIMEZONE;
}

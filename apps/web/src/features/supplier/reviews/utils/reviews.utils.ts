export function formatDate(isoString: string, locale: string = "id"): string {
  if (!isoString) return "-";
  try {
    const date = new Date(isoString);
    return new Intl.DateTimeFormat(locale === "en" ? "en-US" : "id-ID", {
      day: "numeric",
      month: "short",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  } catch {
    return isoString.substring(0, 10);
  }
}

export function getInitials(name: string): string {
  if (!name) return "B";
  const parts = name.trim().split(" ");
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return name.substring(0, 2).toUpperCase();
}

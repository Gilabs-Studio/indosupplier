import { NextRequest, NextResponse } from "next/server";
import dynamicIconImports from "lucide-react/dynamicIconImports";

const DEFAULT_PAGE = 1;
const DEFAULT_PER_PAGE = 20;
const MAX_PER_PAGE = 20;

interface IconCatalogItem {
  name: string;
  label: string;
}

function toLabel(iconName: string): string {
  return iconName
    .split("-")
    .filter(Boolean)
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(" ");
}

function parsePositiveInt(value: string | null, fallback: number): number {
  if (!value) return fallback;
  const parsed = Number.parseInt(value, 10);
  if (Number.isNaN(parsed) || parsed < 1) return fallback;
  return parsed;
}

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams;
  const searchQuery = (searchParams.get("search") ?? "").trim().toLowerCase();
  const page = parsePositiveInt(searchParams.get("page"), DEFAULT_PAGE);
  const requestedPerPage = parsePositiveInt(searchParams.get("per_page"), DEFAULT_PER_PAGE);
  const perPage = Math.min(requestedPerPage, MAX_PER_PAGE);

  const allIcons = Object.keys(dynamicIconImports).sort((a, b) => a.localeCompare(b));
  const filteredIcons = allIcons.filter((iconName) => {
    if (searchQuery.length === 0) return true;
    const iconLabel = toLabel(iconName).toLowerCase();
    return iconName.includes(searchQuery) || iconLabel.includes(searchQuery);
  });

  const total = filteredIcons.length;
  const totalPages = total === 0 ? 0 : Math.ceil(total / perPage);
  const start = (page - 1) * perPage;
  const end = start + perPage;

  const data: IconCatalogItem[] = filteredIcons.slice(start, end).map((iconName) => ({
    name: iconName,
    label: toLabel(iconName),
  }));

  return NextResponse.json({
    success: true,
    data,
    meta: {
      pagination: {
        page,
        per_page: perPage,
        total,
        total_pages: totalPages,
        has_next: totalPages > 0 && page < totalPages,
        has_prev: totalPages > 0 && page > 1,
      },
    },
    timestamp: new Date().toISOString(),
    request_id: crypto.randomUUID(),
  });
}

"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { DataTablePagination } from "@/components/ui/data-table-pagination";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export interface AuditTrailUser {
  id?: string;
  name?: string;
  email?: string;
}

export interface AuditTrailEntry {
  id: string;
  action: string;
  permission_code?: string;
  metadata?: Record<string, unknown> | null;
  user?: AuditTrailUser | null;
  created_at: string;
}

export interface AuditTrailPagination {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface AuditTrailApiResponse {
  success: boolean;
  data: AuditTrailEntry[];
  meta?: {
    pagination?: AuditTrailPagination;
  };
  error?: string;
}

export interface AuditTrailLabels {
  empty: string;
  columns: {
    action: string;
    user: string;
    time: string;
    details: string;
  };
}

interface AuditTrailTableProps {
  readonly entries: AuditTrailEntry[];
  readonly labels: AuditTrailLabels;
  readonly isLoading?: boolean;
  readonly errorText?: string;
  readonly pagination?: AuditTrailPagination;
  readonly onPageChange?: (page: number) => void;
  readonly onPageSizeChange?: (pageSize: number) => void;
}

interface FallbackAuditEvent {
  id: string;
  action: string;
  at?: string | null;
  user?: string | null;
  metadata?: Record<string, unknown> | null;
  permissionCode?: string;
}

interface MetadataEntry {
  key: string;
  value: string;
}

interface MetadataSummary {
  heading?: string;
  entries?: MetadataEntry[];
  plain?: string;
}

function safeDateTime(value?: string | null): string {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return !!value && typeof value === "object" && !Array.isArray(value);
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return "-";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);

  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function prettifyKey(key: string): string {
  return key
    .replace(/[_-]+/g, " ")
    .trim()
    .replace(/\b\w/g, (char) => char.toUpperCase());
}

function isUuidString(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
}

const VISIBLE_REFERENCE_KEYS = new Set([
  "payment_terms_id",
  "supplier_id",
  "customer_id",
  "business_unit_id",
  "employee_id",
  "sales_rep_id",
]);

const REFERENCE_KEY_CANDIDATES: Record<string, string[]> = {
  payment_terms_id: ["payment_terms_name", "payment_terms_name_snapshot"],
  supplier_id: ["supplier_name", "supplier_name_snapshot"],
  customer_id: ["customer_name", "customer_name_snapshot"],
  business_unit_id: ["business_unit_name", "business_unit_name_snapshot"],
  employee_id: ["employee_name", "employee_name_snapshot"],
  sales_rep_id: ["sales_rep_name", "sales_rep_name_snapshot"],
};

const REFERENCE_NAME_KEYS = new Set(Object.values(REFERENCE_KEY_CANDIDATES).flat());

function shouldHideMetadataKey(key: string): boolean {
  if (VISIBLE_REFERENCE_KEYS.has(key)) {
    return false;
  }

  if (REFERENCE_NAME_KEYS.has(key)) {
    return true;
  }

  if (/(^id$|_id$|request_id$|created_by$|updated_by$|deleted_by$)/i.test(key)) {
    return true;
  }

  return /(_name_snapshot|_code_snapshot|_days_snapshot)$/i.test(key);
}

function formatMetadataValue(key: string, value: unknown): string {
  if (typeof value === "number") {
    if (/(amount|total|subtotal|price|cost|tax|discount|paid|remaining)/i.test(key)) {
      return new Intl.NumberFormat("id-ID").format(value);
    }
    return String(value);
  }

  if (typeof value === "string") {
    if (isUuidString(value)) {
      return value;
    }

    if (/status/i.test(key)) {
      return value
        .replace(/[_-]+/g, " ")
        .toLowerCase()
        .replace(/\b\w/g, (char) => char.toUpperCase());
    }

    if (/(date|_at)$/.test(key) || /(date| at)$/i.test(key)) {
      const parsedDate = new Date(value);
      if (!Number.isNaN(parsedDate.getTime())) {
        return parsedDate.toLocaleString();
      }
    }
    return value;
  }

  return formatValue(value);
}

function valuesEqual(left: unknown, right: unknown): boolean {
  try {
    return JSON.stringify(left) === JSON.stringify(right);
  } catch {
    return String(left) === String(right);
  }
}

function parseMaybeJsonObject(value: unknown): unknown {
  if (typeof value !== "string") return value;

  const trimmed = value.trim();
  if (!trimmed) return value;
  if (!(trimmed.startsWith("{") || trimmed.startsWith("["))) return value;

  try {
    return JSON.parse(trimmed);
  } catch {
    return value;
  }
}

function toSnapshotObject(value: unknown): Record<string, unknown> | null {
  const parsed = parseMaybeJsonObject(value);
  return isPlainObject(parsed) ? parsed : null;
}

function toObjectArray(value: unknown): Record<string, unknown>[] {
  if (!Array.isArray(value)) return [];
  return value.filter((entry): entry is Record<string, unknown> => isPlainObject(entry));
}

function firstNonEmptyString(...values: Array<unknown>): string {
  for (const value of values) {
    if (typeof value !== "string") continue;
    const trimmed = value.trim();
    if (trimmed.length > 0) return trimmed;
  }
  return "";
}

function getNestedString(source: Record<string, unknown>, key: string, nestedKey: string): string {
  const nested = source[key];
  if (!isPlainObject(nested)) return "";
  return firstNonEmptyString(nested[nestedKey]);
}

function resolveReferenceLabel(
  snapshot: Record<string, unknown>,
  key: string,
  rawValue: unknown,
): string {
  const candidateKeys = REFERENCE_KEY_CANDIDATES[key] ?? [];
  for (const candidate of candidateKeys) {
    const value = firstNonEmptyString(snapshot[candidate]);
    if (value) return value;
  }

  if (key === "payment_terms_id") {
    const nestedName = getNestedString(snapshot, "payment_terms", "name");
    if (nestedName) return nestedName;
  }

  if (typeof rawValue === "string") {
    const trimmed = rawValue.trim();
    if (trimmed.length === 0 || isUuidString(trimmed)) return "";
    return trimmed;
  }

  if (typeof rawValue === "number" || typeof rawValue === "boolean") {
    return String(rawValue);
  }

  return "";
}

function formatReferenceChange(
  key: string,
  beforeSnapshot: Record<string, unknown>,
  afterSnapshot: Record<string, unknown>,
  previous: unknown,
  next: unknown,
): string {
  const beforeLabel = resolveReferenceLabel(beforeSnapshot, key, previous);
  const afterLabel = resolveReferenceLabel(afterSnapshot, key, next);

  if (beforeLabel && afterLabel && beforeLabel === afterLabel) {
    return "";
  }

  if (!beforeLabel && afterLabel) {
    return `+ ${afterLabel}`;
  }

  if (beforeLabel && !afterLabel) {
    return `- ${beforeLabel}`;
  }

  if (!beforeLabel && !afterLabel) {
    return "Updated";
  }

  return `${beforeLabel || "-"} -> ${afterLabel || "-"}`;
}

function metadataKeyLabel(key: string): string {
  const map: Record<string, string> = {
    payment_terms_id: "Payment Terms",
    supplier_id: "Supplier",
    customer_id: "Customer",
    business_unit_id: "Business Unit",
    employee_id: "Employee",
    sales_rep_id: "Sales Rep",
  };
  return map[key] ?? prettifyKey(key);
}

function isLineItemsKey(key: string): boolean {
  return key.toLowerCase() === "items";
}

interface AuditLineItem {
  id: string;
  label: string;
  quantity: number | null;
  price: number | null;
  discount: number | null;
  subtotal: number | null;
}

function toNullableNumber(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string" && value.trim().length > 0) {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
}

function mapLineItems(items: Record<string, unknown>[]): Map<string, AuditLineItem> {
  const mapped = new Map<string, AuditLineItem>();

  items.forEach((item, index) => {
    const id = firstNonEmptyString(
      item.product_id,
      item.id,
      item.product_code,
      item.product_code_snapshot,
      item.product_name,
      item.product_name_snapshot,
    ) || `line-${index}`;

    const label = firstNonEmptyString(
      item.product_name,
      item.product_name_snapshot,
      getNestedString(item, "product", "name"),
      item.product_code,
      item.product_code_snapshot,
      item.product_id,
      item.id,
    ) || id;

    mapped.set(id, {
      id,
      label,
      quantity: toNullableNumber(item.quantity ?? item.qty),
      price: toNullableNumber(item.price ?? item.purchase_price),
      discount: toNullableNumber(item.discount),
      subtotal: toNullableNumber(item.subtotal ?? item.sub_total),
    });
  });

  return mapped;
}

function compactLabelList(values: string[]): string {
  if (values.length <= 3) {
    return values.join(", ");
  }

  const preview = values.slice(0, 3).join(", ");
  return `${preview} (+${values.length - 3} more)`;
}

function summarizeLineItemChanges(beforeItemsRaw: unknown, afterItemsRaw: unknown): MetadataEntry[] {
  const beforeItems = mapLineItems(toObjectArray(beforeItemsRaw));
  const afterItems = mapLineItems(toObjectArray(afterItemsRaw));

  if (beforeItems.size === 0 && afterItems.size === 0) {
    return [];
  }

  const added: string[] = [];
  const removed: string[] = [];
  const itemUpdated: string[] = [];

  for (const [id, afterItem] of afterItems.entries()) {
    const beforeItem = beforeItems.get(id);
    if (!beforeItem) {
      added.push(afterItem.label);
      continue;
    }
    if (
      beforeItem.quantity !== null &&
      afterItem.quantity !== null &&
      beforeItem.quantity !== afterItem.quantity
    ) {
      itemUpdated.push(`${afterItem.label} (qty ${beforeItem.quantity} -> ${afterItem.quantity})`);
      continue;
    }

    if (
      beforeItem.price !== afterItem.price ||
      beforeItem.discount !== afterItem.discount ||
      beforeItem.subtotal !== afterItem.subtotal
    ) {
      itemUpdated.push(`${afterItem.label} (detail updated)`);
    }
  }

  for (const [id, beforeItem] of beforeItems.entries()) {
    if (!afterItems.has(id)) {
      removed.push(beforeItem.label);
    }
  }

  const entries: MetadataEntry[] = [];

  if (added.length > 0) {
    entries.push({
      key: "Product Items",
      value: `+ ${compactLabelList(added)}`,
    });
  }

  if (removed.length > 0) {
    entries.push({
      key: "Product Items",
      value: `- ${compactLabelList(removed)}`,
    });
  }

  if (itemUpdated.length > 0) {
    entries.push({
      key: "Product Items",
      value: `~ ${compactLabelList(itemUpdated)}`,
    });
  }

  return entries;
}

function renderChangedEntries(before: Record<string, unknown>, after: Record<string, unknown>): MetadataEntry[] {
  const keys = Array.from(new Set([...Object.keys(before), ...Object.keys(after)])).sort();
  const changed: MetadataEntry[] = [];

  for (const key of keys) {
    if (shouldHideMetadataKey(key)) {
      continue;
    }

    const previous = before[key];
    const next = after[key];

    if (valuesEqual(previous, next)) {
      continue;
    }

    if (isLineItemsKey(key)) {
      changed.push(...summarizeLineItemChanges(previous, next));
      continue;
    }

    if (Array.isArray(previous) || Array.isArray(next) || isPlainObject(previous) || isPlainObject(next)) {
      continue;
    }

    const value = VISIBLE_REFERENCE_KEYS.has(key)
      ? formatReferenceChange(key, before, after, previous, next)
      : `${formatMetadataValue(key, previous)} -> ${formatMetadataValue(key, next)}`;

    if (!value || value === "- -> -") {
      continue;
    }

    changed.push({
      key: metadataKeyLabel(key),
      value,
    });
  }

  return changed;
}

function toMetadataEntries(snapshot: Record<string, unknown>): MetadataEntry[] {
  const keys = Object.keys(snapshot).filter((key) => {
    if (shouldHideMetadataKey(key)) {
      return false;
    }

    if (!REFERENCE_NAME_KEYS.has(key)) {
      return true;
    }

    const hasRelatedReferenceID = Object.entries(REFERENCE_KEY_CANDIDATES).some(
      ([idKey, nameKeys]) => nameKeys.includes(key) && firstNonEmptyString(snapshot[idKey]) !== "",
    );

    return !hasRelatedReferenceID;
  });
  if (keys.length === 0) {
    return [];
  }

  const prioritizedKeys = [
    "code",
    "status",
    "customer_name",
    "supplier_name",
    "total_amount",
    "amount",
    "order_date",
    "invoice_date",
    "due_date",
  ];

  const sortedKeys = keys.sort((a, b) => {
    const aIndex = prioritizedKeys.indexOf(a);
    const bIndex = prioritizedKeys.indexOf(b);
    if (aIndex === -1 && bIndex === -1) return a.localeCompare(b);
    if (aIndex === -1) return 1;
    if (bIndex === -1) return -1;
    return aIndex - bIndex;
  });

  return sortedKeys
    .slice(0, 7)
    .map((key) => {
      if (VISIBLE_REFERENCE_KEYS.has(key)) {
        const referenceLabel = resolveReferenceLabel(snapshot, key, snapshot[key]);
        return {
          key: metadataKeyLabel(key),
          value: referenceLabel || formatMetadataValue(key, snapshot[key]),
        };
      }

      if (isLineItemsKey(key) && Array.isArray(snapshot[key])) {
        const totalItems = toObjectArray(snapshot[key]).length;
        return {
          key: "Product Items",
          value: `${totalItems} item(s)`,
        };
      }

      return {
        key: prettifyKey(key),
        value: formatMetadataValue(key, snapshot[key]),
      };
    })
    .filter((entry) => entry.value !== "-");
}

function renderMetadataSummary(metadata?: Record<string, unknown> | null): MetadataSummary {
  if (!metadata) return { plain: "-" };

  const normalizedMetadata: Record<string, unknown> = {
    ...metadata,
    before: parseMaybeJsonObject(metadata.before),
    after: parseMaybeJsonObject(metadata.after),
  };

  const before = toSnapshotObject(normalizedMetadata.before);
  const after = toSnapshotObject(normalizedMetadata.after);

  if (before && after) {
    const changed = renderChangedEntries(before, after);

    if (changed.length > 0) {
      return {
        heading: "Changes",
        entries: changed,
      };
    }

    return { plain: "-" };
  }

  if (after) {
    const entries = toMetadataEntries(after);
    if (entries.length > 0) {
      return {
        heading: "After",
        entries,
      };
    }
  }

  if (before) {
    const entries = toMetadataEntries(before);
    if (entries.length > 0) {
      return {
        heading: "Before",
        entries,
      };
    }
  }

  const beforeStatusRaw = normalizedMetadata.before_status ?? normalizedMetadata.beforeStatus;
  const afterStatusRaw = normalizedMetadata.after_status ?? normalizedMetadata.afterStatus;
  const beforeStatus = typeof beforeStatusRaw === "string" ? beforeStatusRaw : "";
  const afterStatus = typeof afterStatusRaw === "string" ? afterStatusRaw : "";
  if (beforeStatus || afterStatus) {
    const entries: MetadataEntry[] = [];

    if (beforeStatus && afterStatus) {
      entries.push({
        key: "Status",
        value: `${formatMetadataValue("status", beforeStatus)} -> ${formatMetadataValue("status", afterStatus)}`,
      });
    } else {
      entries.push({
        key: "Status",
        value: formatMetadataValue("status", afterStatus || beforeStatus),
      });
    }

    const reason = normalizedMetadata.reason;
    if (typeof reason === "string" && reason.trim().length > 0) {
      entries.push({
        key: "Reason",
        value: reason,
      });
    }

    return {
      heading: "Status Update",
      entries,
    };
  }

  const status = normalizedMetadata.status;
  if (typeof status === "string" && status.length > 0) {
    return {
      heading: "Status",
      entries: [{
        key: "Status",
        value: formatMetadataValue("status", status),
      }],
    };
  }

  const details = normalizedMetadata.details;
  if (typeof details === "string" && details.length > 0) {
    return { plain: details };
  }

  return { plain: formatValue(normalizedMetadata) };
}

function actionLabel(action: string): string {
  const normalized = (action ?? "").trim();
  if (!normalized) return "-";

  const key = normalized.split(".").pop() ?? normalized;
  return key
    .split("_")
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1).toLowerCase())
    .join(" ");
}

export function buildFallbackAuditTrailEntries(events: FallbackAuditEvent[]): AuditTrailEntry[] {
  return events
    .filter((event) => !!event.at)
    .map((event) => ({
      id: event.id,
      action: event.action,
      permission_code: event.permissionCode,
      metadata: event.metadata,
      user: event.user
        ? {
            name: event.user,
            email: event.user,
          }
        : null,
      created_at: event.at as string,
    }))
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
}

export function AuditTrailTable({
  entries,
  labels,
  isLoading = false,
  errorText,
  pagination,
  onPageChange,
  onPageSizeChange,
}: AuditTrailTableProps) {
  if (errorText && entries.length === 0 && !isLoading) {
    return <div className="text-center py-8 text-destructive">{errorText}</div>;
  }

  if (isLoading) {
    return (
      <div className="space-y-2 py-2">
        {Array.from({ length: 5 }).map((_, index) => (
          <Skeleton key={index} className="h-10 w-full" />
        ))}
      </div>
    );
  }

  if (entries.length === 0) {
    return <div className="text-center py-8 text-muted-foreground">{labels.empty}</div>;
  }

  return (
    <div className="space-y-4">
      <div className="border rounded-lg overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{labels.columns.action}</TableHead>
              <TableHead>{labels.columns.user}</TableHead>
              <TableHead>{labels.columns.time}</TableHead>
              <TableHead>{labels.columns.details}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.map((entry) => (
              <TableRow key={entry.id}>
                <TableCell className="font-medium text-xs">{actionLabel(entry.action)}</TableCell>
                <TableCell className="text-sm">
                  {entry.user?.name ?? entry.user?.email ? (
                    <div>
                      <p>{entry.user.name ?? entry.user.email}</p>
                      {entry.user?.id && (
                        <p className="text-xs text-muted-foreground font-mono">
                          {entry.user.id.substring(0, 8)}
                        </p>
                      )}
                    </div>
                  ) : (
                    "-"
                  )}
                </TableCell>
                <TableCell className="text-xs text-muted-foreground">
                  {safeDateTime(entry.created_at)}
                </TableCell>
                <TableCell className="text-xs text-muted-foreground max-w-[420px] whitespace-normal wrap-break-word">
                  {(() => {
                    const summary = renderMetadataSummary(entry.metadata);

                    if (summary.plain) {
                      return summary.plain;
                    }

                    const metadataEntries = summary.entries ?? [];
                    if (metadataEntries.length === 0) {
                      return "-";
                    }

                    return (
                      <div className="space-y-1 leading-relaxed">
                        {summary.heading ? (
                          <p className="font-medium text-foreground">{summary.heading}</p>
                        ) : null}
                        {metadataEntries.map((item) => (
                          <p key={`${item.key}-${item.value}`}>
                            <span className="text-foreground/80">{item.key}:</span>{" "}
                            <span>{item.value}</span>
                          </p>
                        ))}
                      </div>
                    );
                  })()}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      {pagination && onPageChange && onPageSizeChange ? (
        <DataTablePagination
          pageIndex={pagination.page}
          pageSize={pagination.per_page}
          rowCount={pagination.total}
          onPageChange={onPageChange}
          onPageSizeChange={onPageSizeChange}
        />
      ) : null}
    </div>
  );
}
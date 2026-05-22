const TOOL_MARKER_REGEX = /<(tool_call|(?:create|update|delete|query|list|approve|reject|generate)_[a-z0-9_]+)>/i;
const TRAILING_TOOL_LABEL_REGEX = /\*\*tool call\*\*\s*$/i;
const TRAILING_FENCE_REGEX = /```(?:json)?\s*$/i;
const MISSING_FIELDS_MARKER = "karena data belum lengkap/valid:";
const UNGROUNDED_FIELDS_MARKER = "karena ada nilai yang tidak ditemukan di percakapan:";

/**
 * Removes leaked internal tool-call fragments from assistant-visible content.
 * This keeps chat bubbles readable when models accidentally output raw tool XML/JSON.
 */
export function sanitizeToolCallArtifacts(content: string): string {
  if (!content) {
    return "";
  }

  const marker = TOOL_MARKER_REGEX.exec(content);
  if (!marker || marker.index == null) {
    return formatLegacyGuardMessage(content);
  }

  const beforeToolCall = content
    .slice(0, marker.index)
    .replace(TRAILING_TOOL_LABEL_REGEX, "")
    .replace(TRAILING_FENCE_REGEX, "")
    .trim();

  if (beforeToolCall.length > 0) {
    return formatLegacyGuardMessage(beforeToolCall);
  }

  return "Processing action...";
}

function formatLegacyGuardMessage(content: string): string {
  let formatted = content;
  formatted = formatLegacyMissingFieldsMessage(formatted);
  formatted = formatLegacyUngroundedFieldsMessage(formatted);
  return formatted;
}

function formatLegacyMissingFieldsMessage(content: string): string {
  const lower = content.toLowerCase();
  const markerIndex = lower.indexOf(MISSING_FIELDS_MARKER);
  if (markerIndex === -1 || !content.includes(" | ")) {
    return content;
  }

  const prefix = content
    .slice(0, markerIndex + MISSING_FIELDS_MARKER.length)
    .trim();
  const remainder = content.slice(markerIndex + MISSING_FIELDS_MARKER.length).trim();

  const suffixMatch = remainder.match(/\.\s*Mohon kirim[\s\S]*$/i);
  const detailsRaw = suffixMatch
    ? remainder.slice(0, suffixMatch.index).trim()
    : remainder;
  const suffix = suffixMatch
    ? suffixMatch[0].replace(/^\.\s*/, "").trim()
    : "Mohon kirim nilai yang faktual. Jika belum ada nilainya, isi null.";

  const details = detailsRaw
    .split("|")
    .map((item) => item.trim())
    .filter(Boolean);

  if (details.length === 0) {
    return content;
  }

  const bulletList = details.map((item) => `- ${item}`).join("\n");
  return `${prefix}\n\n${bulletList}\n\n${suffix}`;
}

function formatLegacyUngroundedFieldsMessage(content: string): string {
  const lower = content.toLowerCase();
  const markerIndex = lower.indexOf(UNGROUNDED_FIELDS_MARKER);
  if (markerIndex === -1) {
    return content;
  }

  const prefix = content
    .slice(0, markerIndex + UNGROUNDED_FIELDS_MARKER.length)
    .trim();
  const remainder = content.slice(markerIndex + UNGROUNDED_FIELDS_MARKER.length).trim();

  const suffixMatch = remainder.match(/\.\s*Untuk mencegah[\s\S]*$/i);
  const detailsRaw = suffixMatch
    ? remainder.slice(0, suffixMatch.index).trim()
    : remainder;
  const suffix = suffixMatch
    ? suffixMatch[0].replace(/^\.\s*/, "").trim()
    : "Mohon kirim data faktual dari Anda. Jika belum ada nilainya, isi null.";

  const details = detailsRaw
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

  if (details.length === 0) {
    return content;
  }

  const bulletList = details.map((item) => `- ${item}`).join("\n");
  return `${prefix}\n\n${bulletList}\n\n${suffix}`;
}

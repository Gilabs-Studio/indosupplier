import { apiClient } from "@/lib/api-client";
import { getCSRFToken } from "@/lib/api-client";
import type {
  SendMessagePayload,
  SendMessageResponse,
  ConfirmActionPayload,
  SessionsListResponse,
  SessionDetailResponse,
  SessionFilters,
  ActionLogsResponse,
  ActionLogFilters,
  IntentRegistryResponse,
  AIModelsResponse,
  StreamEvent,
} from "../types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8087";
const BASE_PATH = "/ai";

const AI_CHAT_STREAM_DEBUG =
  process.env.NEXT_PUBLIC_AI_CHAT_STREAM_DEBUG === "true" ||
  process.env.NODE_ENV === "development";

function summarizePayload(payload: SendMessagePayload) {
  return {
    session_id: payload.session_id ?? null,
    model: payload.model ?? null,
    message_length: payload.message.length,
    message_preview: payload.message.slice(0, 160),
  };
}

function debugStreamLog(stage: string, details: Record<string, unknown>) {
  if (!AI_CHAT_STREAM_DEBUG) return;
  console.debug(`[ai-chat][stream] ${stage}`, details);
}

export const aiChatService = {
  /** Get available AI models */
  async getModels(): Promise<AIModelsResponse> {
    const response = await apiClient.get<AIModelsResponse>(
      `${BASE_PATH}/models`
    );
    return response.data;
  },

  /** Send a message to the AI assistant */
  async sendMessage(
    payload: SendMessagePayload
  ): Promise<SendMessageResponse> {
    const response = await apiClient.post<SendMessageResponse>(
      `${BASE_PATH}/chat/send`,
      payload
    );
    return response.data;
  },

  /** Confirm or cancel a pending action */
  async confirmAction(
    payload: ConfirmActionPayload
  ): Promise<SendMessageResponse> {
    const response = await apiClient.post<SendMessageResponse>(
      `${BASE_PATH}/chat/confirm`,
      payload
    );
    return response.data;
  },

  /** List chat sessions for the current user */
  async getSessions(
    filters?: SessionFilters
  ): Promise<SessionsListResponse> {
    const response = await apiClient.get<SessionsListResponse>(
      `${BASE_PATH}/sessions`,
      { params: filters }
    );
    return response.data;
  },

  /** Get a session with full message history */
  async getSessionDetail(id: string): Promise<SessionDetailResponse> {
    const response = await apiClient.get<SessionDetailResponse>(
      `${BASE_PATH}/sessions/${id}`
    );
    return response.data;
  },

  /** Build a full sessions export payload for debugging purposes. */
  async exportAllSessionsForDebug(): Promise<{
    source: string;
    exported_at: string;
    session_count: number;
    sessions: SessionDetailResponse["data"][];
  }> {
    const allSessionIDs: string[] = [];
    let page = 1;

    for (;;) {
      const listResp = await aiChatService.getSessions({ page, per_page: 100 });
      if (listResp.data?.length) {
        allSessionIDs.push(...listResp.data.map((session) => session.id));
      }

      const hasNext = listResp.meta?.pagination?.has_next ?? false;
      if (!hasNext) {
        break;
      }
      page += 1;
    }

    const details = await Promise.all(
      allSessionIDs.map(async (sessionID) => {
        const detailResp = await aiChatService.getSessionDetail(sessionID);
        return detailResp.data;
      }),
    );

    return {
      source: "salesview-ai-devkit",
      exported_at: new Date().toISOString(),
      session_count: details.length,
      sessions: details,
    };
  },

  /** Delete a chat session */
  async deleteSession(id: string): Promise<void> {
    await apiClient.delete(`${BASE_PATH}/sessions/${id}`);
  },

  /** List all action logs (admin) */
  async getActionLogs(
    filters?: ActionLogFilters
  ): Promise<ActionLogsResponse> {
    const response = await apiClient.get<ActionLogsResponse>(
      `${BASE_PATH}/admin/actions`,
      { params: filters }
    );
    return response.data;
  },

  /** Get intent registry (admin) */
  async getIntentRegistry(): Promise<IntentRegistryResponse> {
    const response = await apiClient.get<IntentRegistryResponse>(
      `${BASE_PATH}/admin/intents`
    );
    return response.data;
  },

  /**
   * Send a message via the SSE streaming endpoint (v2).
   * Uses native fetch with ReadableStream for real-time event consumption.
   * Returns an AbortController so the caller can cancel the stream.
   */
  sendMessageStream(
    payload: SendMessagePayload,
    onEvent: (event: StreamEvent) => void,
    onError: (error: Error) => void,
    onComplete: () => void,
  ): AbortController {
    const controller = new AbortController();
    const startedAt = Date.now();
    const payloadSummary = summarizePayload(payload);

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      Accept: "text/event-stream",
    };
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers["X-CSRF-Token"] = csrfToken;
    }

    debugStreamLog("request_start", {
      endpoint: `${API_BASE_URL}/api/v1${BASE_PATH}/chat/v2/stream`,
      payload: payloadSummary,
      has_csrf: !!csrfToken,
    });

    fetch(`${API_BASE_URL}/api/v1${BASE_PATH}/chat/v2/stream`, {
      method: "POST",
      headers,
      body: JSON.stringify(payload),
      credentials: "include",
      signal: controller.signal,
    })
      .then(async (res) => {
        const requestId =
          res.headers.get("x-request-id") ??
          res.headers.get("X-Request-ID") ??
          null;

        debugStreamLog("response", {
          request_id: requestId,
          status: res.status,
          ok: res.ok,
          content_type: res.headers.get("content-type"),
          duration_ms: Date.now() - startedAt,
          payload: payloadSummary,
        });

        if (!res.ok) {
          const text = await res.text().catch(() => "Stream request failed");
          debugStreamLog("response_error", {
            request_id: requestId,
            status: res.status,
            body_preview: text.slice(0, 500),
          });
          throw new Error(text);
        }

        const reader = res.body?.getReader();
        if (!reader) {
          throw new Error("ReadableStream not supported");
        }

        const decoder = new TextDecoder();
        let buffer = "";
        let eventCount = 0;
        let contentDeltaCount = 0;
        let contentDeltaChars = 0;

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          // Keep the last (possibly incomplete) line in the buffer
          buffer = lines.pop() ?? "";

          for (const line of lines) {
            if (!line.startsWith("data:")) continue;
            const jsonStr = line.slice(5).trim();
            if (!jsonStr) continue;

            try {
              const event = JSON.parse(jsonStr) as StreamEvent;
              eventCount += 1;
              if (event.type === "content_delta") {
                contentDeltaCount += 1;
                contentDeltaChars += event.content?.length ?? 0;
              } else {
                debugStreamLog("event", {
                  type: event.type,
                  request_id: requestId,
                  has_data: !!event.data,
                  content_preview: event.content?.slice(0, 240) ?? null,
                });
              }
              onEvent(event);
            } catch (parseError) {
              debugStreamLog("event_parse_error", {
                request_id: requestId,
                line_preview: jsonStr.slice(0, 240),
                error:
                  parseError instanceof Error
                    ? parseError.message
                    : String(parseError),
              });
              // Skip malformed lines
            }
          }
        }

        // Process any remaining data in buffer
        if (buffer.startsWith("data:")) {
          const jsonStr = buffer.slice(5).trim();
          if (jsonStr) {
            try {
              const event = JSON.parse(jsonStr) as StreamEvent;
              eventCount += 1;
              if (event.type === "content_delta") {
                contentDeltaCount += 1;
                contentDeltaChars += event.content?.length ?? 0;
              } else {
                debugStreamLog("event", {
                  type: event.type,
                  request_id: requestId,
                  has_data: !!event.data,
                  content_preview: event.content?.slice(0, 240) ?? null,
                });
              }
              onEvent(event);
            } catch (parseError) {
              debugStreamLog("event_parse_error", {
                request_id: requestId,
                line_preview: jsonStr.slice(0, 240),
                error:
                  parseError instanceof Error
                    ? parseError.message
                    : String(parseError),
              });
              // Skip malformed
            }
          }
        }

        debugStreamLog("request_complete", {
          request_id: requestId,
          duration_ms: Date.now() - startedAt,
          event_count: eventCount,
          content_delta_count: contentDeltaCount,
          content_delta_chars: contentDeltaChars,
          payload: payloadSummary,
        });

        onComplete();
      })
      .catch((err: unknown) => {
        if (err instanceof DOMException && err.name === "AbortError") {
          debugStreamLog("request_aborted", {
            duration_ms: Date.now() - startedAt,
            payload: payloadSummary,
          });
          return;
        }

        debugStreamLog("request_failed", {
          duration_ms: Date.now() - startedAt,
          payload: payloadSummary,
          error: err instanceof Error ? err.message : String(err),
        });

        onError(err instanceof Error ? err : new Error(String(err)));
      });

    return controller;
  },

  /** Send message via v2 non-streaming engine endpoint */
  async sendMessageV2(
    payload: SendMessagePayload
  ): Promise<SendMessageResponse> {
    const response = await apiClient.post<SendMessageResponse>(
      `${BASE_PATH}/chat/v2/send`,
      payload
    );
    return response.data;
  },

  /** Confirm action via v2 engine endpoint */
  async confirmActionV2(
    payload: ConfirmActionPayload
  ): Promise<SendMessageResponse> {
    const response = await apiClient.post<SendMessageResponse>(
      `${BASE_PATH}/chat/v2/confirm`,
      payload
    );
    return response.data;
  },
};

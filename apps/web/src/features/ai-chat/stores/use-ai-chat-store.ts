"use client";

import { create } from "zustand";
import type { StreamEvent, StreamToolCallData, StreamToolResultData } from "../types";

interface ToolCallItem {
  name: string;
  parameters: Record<string, unknown>;
  result?: StreamToolResultData["result"];
  status: "pending" | "running" | "done" | "error";
}

interface AIChatStore {
  // UI state
  isOpen: boolean;
  activeSessionId: string | null;
  selectedModel: string | null;

  // Streaming state
  isStreaming: boolean;
  streamingContent: string;
  streamingError: string | null;
  streamingToolCalls: ToolCallItem[];
  streamAbortController: AbortController | null;

  // UI actions
  openChat: () => void;
  closeChat: () => void;
  toggleChat: () => void;
  setActiveSession: (sessionId: string | null) => void;
  setSelectedModel: (model: string) => void;
  startNewChat: () => void;

  // Streaming actions
  startStreaming: (controller: AbortController) => void;
  appendStreamContent: (chunk: string) => void;
  addToolCall: (data: StreamToolCallData) => void;
  updateToolResult: (data: StreamToolResultData) => void;
  endStreaming: () => void;
  cancelStreaming: () => void;
  handleStreamEvent: (event: StreamEvent) => void;
}

export const useAIChatStore = create<AIChatStore>((set, get) => ({
  // UI state
  isOpen: false,
  activeSessionId: null,
  selectedModel: null,

  // Streaming state
  isStreaming: false,
  streamingContent: "",
  streamingError: null,
  streamingToolCalls: [],
  streamAbortController: null,

  // UI actions
  openChat: () => set({ isOpen: true }),
  closeChat: () => set({ isOpen: false }),
  toggleChat: () => set((state) => ({ isOpen: !state.isOpen })),
  setActiveSession: (sessionId) => set({ activeSessionId: sessionId }),
  setSelectedModel: (model) => set({ selectedModel: model }),
  startNewChat: () => set({ activeSessionId: null }),

  // Streaming actions
  startStreaming: (controller) =>
    set({
      isStreaming: true,
      streamingContent: "",
      streamingError: null,
      streamingToolCalls: [],
      streamAbortController: controller,
    }),

  appendStreamContent: (chunk) =>
    set((state) => ({
      streamingContent: state.streamingContent + chunk,
    })),

  addToolCall: (data) =>
    set((state) => ({
      streamingToolCalls: [
        ...state.streamingToolCalls,
        {
          name: data.name,
          parameters: data.parameters,
          status: "running" as const,
        },
      ],
    })),

  updateToolResult: (data) =>
    set((state) => {
      const callName = data?.call?.name;
      const result = data?.result;

      if (!callName || !result) {
        return {
          streamingToolCalls: state.streamingToolCalls.map((tc) =>
            tc.status === "running"
              ? {
                  ...tc,
                  status: "error" as const,
                }
              : tc
          ),
        };
      }

      return {
        streamingToolCalls: state.streamingToolCalls.map((tc) =>
          tc.name === callName && tc.status === "running"
            ? {
                ...tc,
                result,
                status: result.success ? ("done" as const) : ("error" as const),
              }
            : tc
        ),
      };
    }),

  endStreaming: () =>
    set({
      isStreaming: false,
      streamingContent: "",
      streamingError: null,
      streamingToolCalls: [],
      streamAbortController: null,
    }),

  cancelStreaming: () => {
    const { streamAbortController } = get();
    streamAbortController?.abort();
    set({
      isStreaming: false,
      streamingContent: "",
      streamingError: null,
      streamingToolCalls: [],
      streamAbortController: null,
    });
  },

  handleStreamEvent: (event) => {
    const store = get();
    switch (event.type) {
      case "message_start": {
        const data = event.data as { session_id?: string } | undefined;
        if (data?.session_id && !store.activeSessionId) {
          set({ activeSessionId: data.session_id });
        }
        break;
      }
      case "content_delta":
        if (event.content) {
          store.appendStreamContent(event.content);
        }
        break;
      case "tool_call":
        if (event.data) {
          store.addToolCall(event.data as StreamToolCallData);
        }
        break;
      case "tool_result":
        if (event.data) {
          store.updateToolResult(event.data as StreamToolResultData);
        }
        break;
      case "message_end":
        // Streaming is done; the hook will handle cache invalidation
        break;
      case "error":
        // Surface the error message so the UI can render it instead of silently dropping it
        set({
          isStreaming: false,
          streamingError: event.content ?? "An unexpected error occurred",
          streamingToolCalls: [],
          streamAbortController: null,
        });
        break;
      default:
        break;
    }
  },
}));

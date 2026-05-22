"use client";

import { useCallback } from "react";
import {
  useQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { aiChatService } from "../services/ai-chat-service";
import { useAIChatStore } from "../stores/use-ai-chat-store";
import type {
  SendMessagePayload,
  ConfirmActionPayload,
  SessionFilters,
  ActionLogFilters,
} from "../types";

// Query keys factory for consistent cache management
export const aiChatKeys = {
  all: ["ai-chat"] as const,
  models: () => [...aiChatKeys.all, "models"] as const,
  sessions: () => [...aiChatKeys.all, "sessions"] as const,
  sessionList: (filters?: SessionFilters) =>
    [...aiChatKeys.sessions(), filters] as const,
  sessionDetails: () => [...aiChatKeys.all, "session-detail"] as const,
  sessionDetail: (id: string) =>
    [...aiChatKeys.sessionDetails(), id] as const,
  actionLogs: () => [...aiChatKeys.all, "action-logs"] as const,
  actionLogList: (filters?: ActionLogFilters) =>
    [...aiChatKeys.actionLogs(), filters] as const,
  intents: () => [...aiChatKeys.all, "intents"] as const,
};

/** Hook: Get available AI models */
export function useAIModels() {
  return useQuery({
    queryKey: aiChatKeys.models(),
    queryFn: () => aiChatService.getModels(),
    staleTime: 5 * 60 * 1000, // Cache for 5 minutes
  });
}

/** Hook: List chat sessions for current user */
export function useAIChatSessions(filters?: SessionFilters) {
  return useQuery({
    queryKey: aiChatKeys.sessionList(filters),
    queryFn: () => aiChatService.getSessions(filters),
  });
}

/** Hook: Get session detail with messages */
export function useAIChatSessionDetail(sessionId: string | null) {
  return useQuery({
    queryKey: aiChatKeys.sessionDetail(sessionId ?? ""),
    queryFn: () => aiChatService.getSessionDetail(sessionId!),
    enabled: !!sessionId,
  });
}

/** Hook: Send a chat message */
export function useSendMessage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: SendMessagePayload) =>
      aiChatService.sendMessage(payload),
    onSuccess: (data) => {
      // Invalidate session list to update last_activity and message counts
      queryClient.invalidateQueries({ queryKey: aiChatKeys.sessions() });
      // Invalidate the specific session detail
      if (data?.data?.session_id) {
        queryClient.invalidateQueries({
          queryKey: aiChatKeys.sessionDetail(data.data.session_id),
        });
      }
    },
  });
}

/** Hook: Confirm or cancel a pending action */
export function useConfirmAction() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: ConfirmActionPayload) =>
      aiChatService.confirmAction(payload),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: aiChatKeys.sessions() });
      if (data?.data?.session_id) {
        queryClient.invalidateQueries({
          queryKey: aiChatKeys.sessionDetail(data.data.session_id),
        });
      }
    },
  });
}

/** Hook: Delete a chat session */
export function useDeleteSession() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (sessionId: string) =>
      aiChatService.deleteSession(sessionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: aiChatKeys.sessions() });
    },
  });
}

/** Hook: List action logs (admin) */
export function useActionLogs(filters?: ActionLogFilters) {
  return useQuery({
    queryKey: aiChatKeys.actionLogList(filters),
    queryFn: () => aiChatService.getActionLogs(filters),
  });
}

/** Hook: Get intent registry (admin) */
export function useIntentRegistry() {
  return useQuery({
    queryKey: aiChatKeys.intents(),
    queryFn: () => aiChatService.getIntentRegistry(),
  });
}

/**
 * Hook: Send a streaming message via SSE.
 * Manages store streaming state and invalidates caches on completion.
 */
export function useSendMessageStream() {
  const queryClient = useQueryClient();
  const {
    activeSessionId,
    selectedModel,
    startStreaming,
    handleStreamEvent,
    endStreaming,
    isStreaming,
  } = useAIChatStore();

  const send = useCallback(
    (message: string) => {
      if (isStreaming) return;

      const payload: SendMessagePayload = {
        message,
        session_id: activeSessionId ?? undefined,
        model: selectedModel ?? undefined,
      };

      const controller = aiChatService.sendMessageStream(
        payload,
        (event) => {
          handleStreamEvent(event);
        },
        (error) => {
          handleStreamEvent({
            type: "error",
            content: error.message,
          });
        },
        () => {
          // On complete: invalidate caches, clear streaming state
          const state = useAIChatStore.getState();
          const sessionId = state.activeSessionId;
          const hasStreamingError = !!state.streamingError;

          // Preserve error message from EventError so users can see what happened.
          if (!hasStreamingError) {
            endStreaming();
          }

          queryClient.invalidateQueries({ queryKey: aiChatKeys.sessions() });
          if (sessionId) {
            queryClient.invalidateQueries({
              queryKey: aiChatKeys.sessionDetail(sessionId),
            });
          }
        },
      );

      startStreaming(controller);
    },
    [
      activeSessionId,
      selectedModel,
      isStreaming,
      startStreaming,
      handleStreamEvent,
      endStreaming,
      queryClient,
    ],
  );

  return { send, isStreaming };
}

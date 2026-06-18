import { useState, useEffect, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { chatService } from "../services/chat.service";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import type { ChatMessage } from "../types/chat.types";

export function useBuyerChat(initialRoomId?: string | null) {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();
  const [selectedRoomId, setSelectedRoomId] = useState<string | null>(initialRoomId ?? null);
  const [wsConnected, setWsConnected] = useState(false);

  const selectedRoomIdRef = useRef<string | null>(null);

  useEffect(() => {
    selectedRoomIdRef.current = selectedRoomId;
  }, [selectedRoomId]);

  // 1. Fetch Chat Rooms
  const roomsQuery = useQuery({
    queryKey: ["chat-rooms"],
    queryFn: () => chatService.getRooms(),
    enabled: isAuthenticated,
    refetchOnWindowFocus: false,
  });

  // 2. Fetch Messages for Selected Room
  const messagesQuery = useQuery({
    queryKey: ["chat-messages", selectedRoomId],
    queryFn: () => chatService.getMessages(selectedRoomId || ""),
    enabled: isAuthenticated && !!selectedRoomId,
    refetchOnWindowFocus: false,
  });

  // 3. Mark as Read Mutation
  const markReadMutation = useMutation({
    mutationFn: (roomId: string) => chatService.markAsRead(roomId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["chat-rooms"] });
    },
  });

  // 4. Send Message Mutation
  const sendMessageMutation = useMutation({
    mutationFn: ({ roomId, body }: { roomId: string; body: string }) =>
      chatService.sendMessage(roomId, body),
    onSuccess: (newMsg) => {
      // Optimistically insert message
      queryClient.setQueryData<ChatMessage[]>(["chat-messages", newMsg.chat_room_id], (old) => {
        if (!old) return [newMsg];
        if (old.some((m) => m.id === newMsg.id)) return old;
        return [...old, newMsg];
      });
      queryClient.invalidateQueries({ queryKey: ["chat-rooms"] });
    },
  });

  // Trigger mark as read when selecting a room
  const handleSelectRoom = (roomId: string) => {
    setSelectedRoomId(roomId);
    markReadMutation.mutate(roomId);
  };

  useEffect(() => {
    if (!selectedRoomId || !roomsQuery.data?.some((room) => room.id === selectedRoomId)) return;
    markReadMutation.mutate(selectedRoomId);
  }, [markReadMutation, roomsQuery.data, selectedRoomId]);

  // 5. Setup WebSocket connection
  useEffect(() => {
    if (!isAuthenticated) return;

    let socket: WebSocket | null = null;
    let reconnectTimeout: NodeJS.Timeout | null = null;

    const handleMessage = (event: MessageEvent) => {
      try {
        const payload = JSON.parse(event.data);
        if (payload.type === "chat.message_received") {
          const message = payload.data as ChatMessage;

          // Append incoming message to query cache
          queryClient.setQueryData<ChatMessage[]>(["chat-messages", message.chat_room_id], (old) => {
            if (!old) return [message];
            if (old.some((m) => m.id === message.id)) return old;
            return [...old, message];
          });

          // Refetch rooms to update preview text & unread badge
          queryClient.invalidateQueries({ queryKey: ["chat-rooms"] });

          // If incoming message belongs to currently opened room, clear unread status instantly
          const currentOpenedRoomId = selectedRoomIdRef.current;
          if (currentOpenedRoomId === message.chat_room_id) {
            chatService.markAsRead(message.chat_room_id).then(() => {
              queryClient.invalidateQueries({ queryKey: ["chat-rooms"] });
            }).catch(console.error);
          }
        }
      } catch (err) {
        console.error("failed to parse websocket chat event:", err);
      }
    };

    const connect = () => {
      const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8088";
      const wsProtocol = globalThis.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${wsProtocol}//${API_BASE_URL.replace(/^https?:\/\//, "")}/api/v1/chat/ws`;

      socket = new WebSocket(wsUrl);

      socket.onopen = () => {
        setWsConnected(true);
      };

      socket.onclose = () => {
        setWsConnected(false);
        reconnectTimeout = setTimeout(connect, 3000); // Reconnect in 3s
      };

      socket.onerror = (err) => {
        console.error("Chat WebSocket error:", err);
        socket?.close();
      };

      socket.onmessage = handleMessage;
    };

    connect();

    return () => {
      if (socket) {
        socket.close();
      }
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
    };
  }, [isAuthenticated, queryClient]);

  return {
    rooms: roomsQuery.data || [],
    isLoadingRooms: roomsQuery.isLoading && isAuthenticated,
    messages: messagesQuery.data || [],
    isLoadingMessages: messagesQuery.isLoading && !!selectedRoomId,
    selectedRoomId,
    setSelectedRoom: handleSelectRoom,
    sendMessage: (body: string) => {
      if (selectedRoomId) {
        return sendMessageMutation.mutateAsync({ roomId: selectedRoomId, body });
      }
      return Promise.reject(new Error("No room selected"));
    },
    isSending: sendMessageMutation.isPending,
    wsConnected,
  };
}

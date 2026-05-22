"use client";

import { useEffect, useRef } from "react";
import { useTranslations } from "next-intl";
import { Bot, Loader2, Cpu } from "lucide-react";

import { ScrollArea } from "@/components/ui/scroll-area";

import { MessageBubble } from "./message-bubble";
import { StreamingMessage } from "./streaming-message";
import { useAIChatStore } from "../stores/use-ai-chat-store";
import type { AIChatMessage, AIActionPreview } from "../types";

interface MessageListProps {
  messages: AIChatMessage[];
  action?: AIActionPreview | null;
  sessionId: string;
  isLoading?: boolean;
}

export function MessageList({
  messages,
  action,
  sessionId,
  isLoading,
}: MessageListProps) {
  const t = useTranslations("aiChat");
  const bottomRef = useRef<HTMLDivElement>(null);
  const { isStreaming, streamingContent, streamingError } = useAIChatStore();

  // Auto-scroll to bottom when messages change or streaming content updates
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages.length, isLoading, isStreaming, streamingContent, streamingError]);

  if (messages.length === 0 && !isLoading) {
    return (
      <div className="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-full bg-primary/10">
          <Bot className="h-7 w-7 text-primary" />
        </div>
        <div>
          <h4 className="text-sm font-semibold text-foreground">
            {t("title")}
          </h4>
          <p className="mt-1 text-xs text-muted-foreground">
            {t("noMessages")}
          </p>
        </div>
      </div>
    );
  }

  // Find the last assistant message to attach the action card
  const lastAssistantIndex = messages.findLastIndex(
    (m) => m.role === "assistant"
  );

  return (
    <ScrollArea className="flex-1 overflow-hidden">
      <div className="flex flex-col gap-4 px-4 py-4">
        {messages.map((message, index) => (
          <MessageBubble
            key={message.id}
            message={message}
            sessionId={sessionId}
            action={index === lastAssistantIndex ? action : null}
          />
        ))}

        {/* Real-time streaming message (v2 engine) */}
        {(isStreaming || !!streamingContent || !!streamingError) && (
          <StreamingMessage />
        )}

        {/* Legacy loading indicator (non-streaming v1 fallback) */}
        {isLoading && !isStreaming && (
          <div className="flex items-start gap-3">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10">
              <Bot className="h-4 w-4 text-primary" />
            </div>
            <div className="max-w-[80%] space-y-2">
              <div className="flex items-center gap-2 rounded-2xl rounded-bl-md bg-muted px-4 py-2.5">
                <Loader2 className="h-3.5 w-3.5 animate-spin text-muted-foreground" />
                <span className="text-xs text-muted-foreground">
                  {t("typing")}
                </span>
              </div>
              <div className="rounded-lg border border-primary/30 bg-primary/5 p-3">
                <div className="flex items-center gap-2">
                  <Cpu className="h-4 w-4 animate-pulse text-primary" />
                  <span className="text-sm font-medium text-primary">
                    {t("action.processing")}
                  </span>
                </div>
                <div className="mt-1.5 flex items-center gap-1.5">
                  <div className="h-1 w-1 animate-bounce rounded-full bg-primary/60 [animation-delay:0ms]" />
                  <div className="h-1 w-1 animate-bounce rounded-full bg-primary/60 [animation-delay:150ms]" />
                  <div className="h-1 w-1 animate-bounce rounded-full bg-primary/60 [animation-delay:300ms]" />
                  <span className="ml-1 text-xs text-muted-foreground">
                    {t("action.analyzing")}
                  </span>
                </div>
              </div>
            </div>
          </div>
        )}

        <div ref={bottomRef} />
      </div>
    </ScrollArea>
  );
}

"use client";

import { useCallback, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Bot, PanelLeftClose, PanelLeftOpen, StopCircle } from "lucide-react";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

import { useAIChatStore } from "@/features/ai-chat/stores/use-ai-chat-store";
import {
  useSendMessageStream,
  useAIChatSessionDetail,
} from "@/features/ai-chat/hooks/use-ai-chat";
import { SessionList } from "@/features/ai-chat/components/session-list";
import { MessageList } from "@/features/ai-chat/components/message-list";
import { MessageInput } from "@/features/ai-chat/components/message-input";
import { ModelSelector } from "@/features/ai-chat/components/model-selector";
import type {
  AIChatMessage,
  AIActionPreview,
} from "@/features/ai-chat/types";

export default function AIChatbotPage() {
  const t = useTranslations("aiChat");
  const {
    activeSessionId,
    closeChat,
    cancelStreaming,
  } = useAIChatStore();
  const [showSidebar, setShowSidebar] = useState(true);

  useEffect(() => {
    closeChat();
  }, [closeChat]);

  const { data: sessionDetail, isLoading: isLoadingSession } =
    useAIChatSessionDetail(activeSessionId);

  const { send, isStreaming } = useSendMessageStream();

  const messages: AIChatMessage[] = sessionDetail?.data?.messages ?? [];
  const pendingAction: AIActionPreview | null =
    sessionDetail?.data?.pending_action ?? null;

  const handleSend = useCallback(
    (content: string) => {
      send(content);
    },
    [send],
  );


  return (
    <div className="flex h-full w-full overflow-hidden">
      {/* Sidebar */}
      <div
        className={cn(
          "h-full shrink-0 transition-all duration-200",
          showSidebar ? "w-[280px]" : "w-0"
        )}
      >
        {showSidebar && <SessionList />}
      </div>

      {/* Main Chat */}
      <div className="flex min-w-0 flex-1 flex-col">
        {/* Header */}
        <div className="flex items-center gap-3 border-b border-border px-4 py-3">
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 cursor-pointer"
                onClick={() => setShowSidebar((prev) => !prev)}
              >
                {showSidebar ? (
                  <PanelLeftClose className="h-4 w-4" />
                ) : (
                  <PanelLeftOpen className="h-4 w-4" />
                )}
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t("sessions")}</TooltipContent>
          </Tooltip>

          <div className="flex items-center gap-2">
            <Bot className="h-5 w-5 text-primary" />
            <div>
              <div className="flex items-center gap-1">
                <h1 className="text-base font-semibold text-foreground">
                  {sessionDetail?.data?.title || t("title")}
                </h1>
              </div>
              <div className="-ml-1 mt-0.5">
                <ModelSelector />
              </div>
            </div>
          </div>
        </div>

        {/* Messages */}
        <MessageList
          messages={messages}
          action={pendingAction}
          sessionId={activeSessionId ?? ""}
          isLoading={isStreaming || isLoadingSession}
        />

        {/* Input */}
        <div className="relative">
          {isStreaming && (
            <div className="absolute -top-8 left-0 right-0 flex justify-center">
              <Button
                variant="outline"
                size="sm"
                className="h-7 cursor-pointer gap-1.5 rounded-full text-xs shadow-sm"
                onClick={cancelStreaming}
              >
                <StopCircle className="h-3 w-3" />
                {t("cancel")}
              </Button>
            </div>
          )}
          <MessageInput
            onSend={handleSend}
            isLoading={isStreaming}
          />
        </div>
      </div>
    </div>
  );
}

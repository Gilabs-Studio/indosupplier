"use client";

import { useCallback, useState } from "react";
import { formatDistanceToNow } from "date-fns";
import { id as idLocale, enUS } from "date-fns/locale";
import { useLocale, useTranslations } from "next-intl";
import { MessageSquare, Plus, Trash2, Loader2, Copy } from "lucide-react";
import { toast } from "sonner";

import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

import { useAIChatSessions, useDeleteSession } from "../hooks/use-ai-chat";
import { useAIChatStore } from "../stores/use-ai-chat-store";
import { aiChatService } from "../services/ai-chat-service";
import type { AIChatSession } from "../types";

interface SessionListProps {
  showHeaderActions?: boolean;
}

export function SessionList({ showHeaderActions = true }: SessionListProps) {
  const t = useTranslations("aiChat");
  const locale = useLocale();
  const dateLocale = locale === "id" ? idLocale : enUS;
  const [isCopyingSessions, setIsCopyingSessions] = useState(false);

  const { activeSessionId, setActiveSession, startNewChat } =
    useAIChatStore();
  const { data, isLoading } = useAIChatSessions({ page: 1, per_page: 50 });
  const deleteSession = useDeleteSession();

  const sessions: AIChatSession[] = data?.data ?? [];

  const handleCopyAllSessions = useCallback(async () => {
    setIsCopyingSessions(true);
    try {
      const exportPayload = await aiChatService.exportAllSessionsForDebug();

      if (exportPayload.session_count === 0) {
        toast.error(t("toast.copyAllSessionsEmpty"));
        return;
      }

      await navigator.clipboard.writeText(JSON.stringify(exportPayload, null, 2));
      toast.success(`${t("toast.copyAllSessionsSuccess")} (${exportPayload.session_count})`);
    } catch {
      toast.error(t("toast.copyAllSessionsFailed"));
    } finally {
      setIsCopyingSessions(false);
    }
  }, [t]);

  const handleDelete = (e: React.MouseEvent, sessionId: string) => {
    e.stopPropagation();
    deleteSession.mutate(sessionId, {
      onSuccess: () => {
        if (activeSessionId === sessionId) {
          startNewChat();
        }
      },
    });
  };

  return (
    <div className="flex h-full flex-col border-r border-border bg-muted/30">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-border p-3">
        <h3 className="text-sm font-semibold text-foreground">
          {t("sessions")}
        </h3>
        {showHeaderActions && (
          <div className="flex items-center gap-1">
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 cursor-pointer"
                  onClick={handleCopyAllSessions}
                  disabled={isCopyingSessions}
                >
                  {isCopyingSessions ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </TooltipTrigger>
              <TooltipContent>{t("copyAllSessions")}</TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 cursor-pointer"
                  onClick={startNewChat}
                >
                  <Plus className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>{t("newChat")}</TooltipContent>
            </Tooltip>
          </div>
        )}
      </div>

      {/* Session List */}
      <ScrollArea className="flex-1">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
          </div>
        ) : sessions.length === 0 ? (
          <div className="flex flex-col items-center gap-2 px-3 py-8 text-center">
            <MessageSquare className="h-8 w-8 text-muted-foreground/40" />
            <p className="text-xs text-muted-foreground">{t("noSessions")}</p>
          </div>
        ) : (
          <div className="space-y-0.5 p-2">
            {sessions.map((session) => (
              <div
                key={session.id}
                role="button"
                tabIndex={0}
                onClick={() => setActiveSession(session.id)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    setActiveSession(session.id);
                  }
                }}
                className={cn(
                  "group flex w-full cursor-pointer items-center gap-2 rounded-md px-2.5 py-2 text-left transition-colors",
                  "hover:bg-accent/60",
                  activeSessionId === session.id &&
                    "bg-accent text-accent-foreground"
                )}
              >
                <div className="min-w-0 flex-1 overflow-hidden">
                  <p className="truncate text-sm font-medium leading-tight">
                    {session.title || t("newChat")}
                  </p>
                  <p className="mt-0.5 truncate text-xs text-muted-foreground">
                    {formatDistanceToNow(new Date(session.last_activity), {
                      addSuffix: true,
                      locale: dateLocale,
                    })}
                  </p>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 shrink-0 cursor-pointer opacity-0 transition-opacity group-hover:opacity-100"
                  onClick={(e) => handleDelete(e, session.id)}
                  disabled={deleteSession.isPending}
                >
                  {deleteSession.isPending ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <Trash2 className="h-3 w-3 text-destructive" />
                  )}
                </Button>
              </div>
            ))}
          </div>
        )}
      </ScrollArea>
    </div>
  );
}

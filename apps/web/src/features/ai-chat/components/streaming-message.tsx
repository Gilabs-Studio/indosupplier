"use client";

import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Bot, Loader2, CheckCircle2, XCircle, Wrench, AlertCircle, ExternalLink } from "lucide-react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";

import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";

import { useAIChatStore } from "../stores/use-ai-chat-store";
import { sanitizeToolCallArtifacts } from "./tool-call-sanitizer";

// Shared link renderer: internal /path → SPA chip, external → new tab with icon.
const markdownComponents = {
  a: ({
    href,
    children,
  }: React.AnchorHTMLAttributes<HTMLAnchorElement>) => {
    if (href?.startsWith("/")) {
      return (
        <Link
          href={href as Parameters<typeof Link>[0]["href"]}
          className="inline-flex items-center gap-1 rounded-md border border-primary/40 bg-primary/8 px-2 py-0.5 text-xs font-medium text-primary no-underline transition-colors hover:bg-primary/15 cursor-pointer"
        >
          {children}
        </Link>
      );
    }
    return (
      <a
        href={href}
        target="_blank"
        rel="noopener noreferrer"
        className="inline-flex items-center gap-1 text-primary underline underline-offset-2"
      >
        {children}
        <ExternalLink className="h-3 w-3 shrink-0 opacity-70" />
      </a>
    );
  },
};

/**
 * StreamingMessage renders the in-progress assistant message during SSE streaming.
 * Shows accumulating text content + tool call status cards in real-time.
 */
export function StreamingMessage() {
  const t = useTranslations("aiChat");
  const { isStreaming, streamingContent, streamingError, streamingToolCalls } =
    useAIChatStore();
  const displayStreamingContent = sanitizeToolCallArtifacts(streamingContent);

  if (!isStreaming && !streamingContent && !streamingError) return null;

  return (
    <div className="flex w-full gap-3">
      {/* Avatar */}
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary/10">
        <Bot className="h-4 w-4 text-primary" />
      </div>

      <div className="max-w-[80%] min-w-0 space-y-2">
        {/* Tool call status cards */}
        {streamingToolCalls.length > 0 && (
          <div className="space-y-1.5">
            {streamingToolCalls.map((tc, i) => (
              <div
                key={`${tc.name}-${i}`}
                className={cn(
                  "flex items-center gap-2 rounded-lg border px-3 py-2 text-xs",
                  tc.status === "running" &&
                    "border-primary/30 bg-primary/5 text-primary",
                  tc.status === "done" &&
                    "border-success/30 bg-success/5 text-success",
                  tc.status === "error" &&
                    "border-destructive/30 bg-destructive/5 text-destructive"
                )}
              >
                {tc.status === "running" && (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                )}
                {tc.status === "done" && (
                  <CheckCircle2 className="h-3.5 w-3.5" />
                )}
                {tc.status === "error" && (
                  <XCircle className="h-3.5 w-3.5" />
                )}
                <Wrench className="h-3 w-3" />
                <span className="font-medium">{tc.name}</span>
                {tc.result?.message && (
                  <Badge variant="outline" className="ml-auto text-[10px]">
                    {tc.result.message}
                  </Badge>
                )}
              </div>
            ))}
          </div>
        )}

        {/* Error state */}
        {streamingError ? (
          <div className="flex items-center gap-2 rounded-2xl rounded-bl-md border border-destructive/30 bg-destructive/5 px-4 py-2.5 text-sm text-destructive">
            <AlertCircle className="h-3.5 w-3.5 shrink-0" />
            <span className="wrap-anywhere">{streamingError}</span>
          </div>
        ) : displayStreamingContent ? (
          <div className="overflow-hidden rounded-2xl rounded-bl-md bg-muted px-4 py-2.5 text-sm leading-relaxed text-foreground">
            <div className="ai-chat-markdown prose prose-sm dark:prose-invert max-w-none wrap-anywhere [&_p]:my-1 [&_ul]:my-1 [&_ol]:my-1 [&_li]:my-0.5 [&_h1]:text-base [&_h2]:text-sm [&_h3]:text-sm [&_pre]:my-1 [&_pre]:max-w-full [&_pre]:overflow-x-auto [&_pre]:whitespace-pre-wrap [&_pre]:wrap-break-word [&_pre]:rounded-md [&_pre]:bg-background/50 [&_pre]:p-2 [&_code]:wrap-break-word [&_code]:rounded [&_code]:bg-background/50 [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-xs [&_table]:my-2 [&_table]:w-full [&_table]:border-collapse [&_table]:text-xs [&_table]:overflow-hidden [&_table]:rounded-md [&_table]:border [&_table]:border-border/40 [&_thead]:bg-muted/50 [&_th]:border [&_th]:border-border/40 [&_th]:px-2.5 [&_th]:py-1.5 [&_th]:text-left [&_th]:font-semibold [&_th]:text-foreground/80 [&_td]:border [&_td]:border-border/40 [&_td]:px-2.5 [&_td]:py-1.5 [&_td]:text-foreground/70 [&_tr:hover]:bg-muted/30 [&_blockquote]:border-l-2 [&_blockquote]:border-primary/30 [&_blockquote]:pl-3 [&_blockquote]:italic [&_strong]:font-semibold">
              <div className="overflow-x-auto">
                <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                  {displayStreamingContent}
                </ReactMarkdown>
              </div>
            </div>
            {/* Blinking cursor */}
            {isStreaming && (
              <span className="ml-0.5 inline-block h-4 w-0.5 animate-pulse bg-primary" />
            )}
          </div>
        ) : isStreaming ? (
          <div className="flex items-center gap-2 rounded-2xl rounded-bl-md bg-muted px-4 py-2.5">
            <Loader2 className="h-3.5 w-3.5 animate-spin text-muted-foreground" />
            <span className="text-xs text-muted-foreground">
              {t("typing")}
            </span>
          </div>
        ) : null}
      </div>
    </div>
  );
}

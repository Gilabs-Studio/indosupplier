"use client";

import React, { useState, useEffect, useRef } from "react";
import { useTranslations } from "next-intl";
import { useSearchParams } from "next/navigation";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Send, Check, CheckCheck, MessageSquare, ShieldCheck, Loader2 } from "lucide-react";
import { cn, getDicebearUrl } from "@/lib/utils";
import { BuyerLayout } from "@/features/buyer/components/buyer-layout";
import { useBuyerChat } from "../hooks/useBuyerChat";

function formatTime(dateStr?: string, yesterdayLabel: string = "Kemarin"): string {
  if (!dateStr) return "";
  try {
    const date = new Date(dateStr);
    const now = new Date();

    if (date.toDateString() === now.toDateString()) {
      return date.toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" });
    }

    const yesterday = new Date(now);
    yesterday.setDate(now.getDate() - 1);
    if (date.toDateString() === yesterday.toDateString()) {
      return yesterdayLabel;
    }

    return date.toLocaleDateString("id-ID", { day: "numeric", month: "short" });
  } catch {
    return "";
  }
}

export function BuyerChatPage() {
  const t = useTranslations("buyerChat");
  const searchParams = useSearchParams();
  const [inputText, setInputText] = useState<string>("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const roomIdFromQuery = searchParams.get("roomId");

  const {
    rooms,
    isLoadingRooms,
    messages,
    isLoadingMessages,
    selectedRoomId,
    setSelectedRoom,
    sendMessage,
    isSending,
  } = useBuyerChat(roomIdFromQuery);

  const selectedRoom = rooms.find((r) => r.id === selectedRoomId);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputText.trim()) return;

    try {
      const text = inputText;
      setInputText(""); // Clear input early for responsive UI feel
      await sendMessage(text);
    } catch (err) {
      console.error("Failed to send message", err);
    }
  };

  const renderRoomList = () => {
    if (isLoadingRooms) {
      return (
        <div className="flex flex-col items-center justify-center p-8 gap-2">
          <Loader2 className="h-6 w-6 animate-spin text-primary" />
          <span className="text-xs text-muted-foreground font-semibold">{t("loadingChats")}</span>
        </div>
      );
    }

    if (rooms.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center p-8 gap-2 text-center text-muted-foreground">
          <MessageSquare className="h-8 w-8 opacity-40" />
          <span className="text-xs font-semibold">{t("noActiveChats")}</span>
        </div>
      );
    }

    return rooms.map((room) => {
      const isSelected = room.id === selectedRoomId;
      const avatarUrl = getDicebearUrl(room.company_name, "initials");

      return (
        <div
          key={room.id}
          onClick={() => setSelectedRoom(room.id)}
          onKeyDown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              setSelectedRoom(room.id);
            }
          }}
          role="button"
          tabIndex={0}
          className={cn(
            "p-4 flex items-start gap-3 transition-all duration-300 cursor-pointer focus-visible:outline-hidden focus-visible:bg-primary/5",
            isSelected ? "bg-primary/10" : "bg-card hover:bg-muted/30"
          )}
        >
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={avatarUrl}
            alt={room.company_name}
            className="h-10 w-10 rounded-lg shrink-0 border border-border"
          />
          <div className="min-w-0 flex-1 space-y-1">
            <div className="flex items-center justify-between gap-2">
              <span className="font-extrabold text-xs text-foreground flex items-center gap-1 truncate">
                {room.company_name}
                {room.is_verified && (
                  <ShieldCheck className="h-3.5 w-3.5 text-primary fill-primary/10 shrink-0" />
                )}
              </span>
              <span className="text-[10px] text-muted-foreground font-semibold shrink-0">
                {formatTime(room.last_message?.created_at || room.updated_at.toString(), t("yesterday"))}
              </span>
            </div>
            
            <div className="flex items-center justify-between gap-1">
              <p className="text-[11px] text-muted-foreground truncate font-semibold leading-relaxed flex-1">
                {room.last_message?.body || "Belum ada pesan."}
              </p>
              {room.unread_count > 0 && (
                <span className="h-4 min-w-4 px-1 rounded-full bg-primary text-[9px] font-bold text-primary-foreground flex items-center justify-center shrink-0">
                  {room.unread_count}
                </span>
              )}
            </div>
          </div>
        </div>
      );
    });
  };

  const renderMessageThread = () => {
    if (isLoadingMessages) {
      return (
        <div className="flex flex-col items-center justify-center h-full gap-2">
          <Loader2 className="h-6 w-6 animate-spin text-primary" />
          <span className="text-xs text-muted-foreground font-semibold">{t("loadingConversation")}</span>
        </div>
      );
    }

    if (messages.length === 0) {
      return (
        <div className="flex flex-col items-center justify-center h-full gap-2 text-center text-muted-foreground">
          <MessageSquare className="h-8 w-8 opacity-40" />
          <span className="text-xs font-semibold">{t("startConversation")}</span>
        </div>
      );
    }

    return messages.map((msg) => {
      const isMe = msg.sender_type === "buyer";
      return (
        <div
          key={msg.id}
          className={cn("flex items-end gap-2", isMe ? "justify-end" : "justify-start")}
        >
          <div className="max-w-[75%] space-y-1">
            <div
              className={cn(
                "p-3 text-sm font-semibold rounded-lg leading-relaxed shadow-2xs border transition-all duration-300",
                isMe
                  ? "bg-primary text-primary-foreground border-primary/20"
                  : "bg-muted text-foreground border-border"
              )}
            >
              {msg.body}
            </div>
            <div className="flex items-center justify-end gap-1 text-[10px] text-muted-foreground font-semibold">
              <span>{formatTime(msg.created_at, t("yesterday"))}</span>
              {isMe && (
                <span className="shrink-0">
                  {msg.is_read ? (
                    <CheckCheck className="h-3.5 w-3.5 text-primary" />
                  ) : (
                    <Check className="h-3.5 w-3.5 text-muted-foreground" />
                  )}
                </span>
              )}
            </div>
          </div>
        </div>
      );
    });
  };

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Page Title */}
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {/* Chat Container Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 bg-card border border-border rounded-xl overflow-hidden min-h-[580px] shadow-xs">
          {/* Left Side: Partner List */}
          <div className="border-r border-border flex flex-col h-full bg-muted/10">
            <div className="p-4 border-b border-border bg-card">
              <p className="text-xs font-extrabold text-foreground tracking-wider uppercase">{t("activeNegotiations")}</p>
            </div>
            
            <div className="flex-1 overflow-y-auto divide-y divide-border">
              {renderRoomList()}
            </div>
          </div>

          {/* Right Side: Message Room */}
          <div className="md:col-span-2 flex flex-col h-full bg-card min-h-[500px]">
            {selectedRoom ? (
              <>
                {/* Partner Header */}
                <div className="p-4 border-b border-border flex items-center justify-between shrink-0">
                  <div className="flex items-center gap-3">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img
                      src={getDicebearUrl(selectedRoom.company_name, "initials")}
                      alt={selectedRoom.company_name}
                      className="h-10 w-10 rounded-lg shrink-0 border border-border"
                    />
                    <div>
                      <h3 className="text-xs font-extrabold text-foreground flex items-center gap-1">
                        {selectedRoom.company_name}
                        {selectedRoom.is_verified && (
                          <ShieldCheck className="h-4 w-4 text-primary fill-primary/10 shrink-0" />
                        )}
                      </h3>
                      <span className="inline-flex items-center gap-1 mt-0.5 text-[10px] text-muted-foreground font-semibold">
                        <span className="h-1.5 w-1.5 rounded-full bg-success animate-pulse" />
                        <span>{t("online")}</span>
                      </span>
                    </div>
                  </div>
                </div>

                {/* Chat Thread */}
                <div className="flex-1 overflow-y-auto p-4 space-y-4 max-h-[400px]">
                  {renderMessageThread()}
                  <div ref={messagesEndRef} />
                </div>

                {/* Chat Input Field */}
                <div className="p-4 border-t border-border bg-card shrink-0">
                  <form onSubmit={handleSend} className="flex items-center gap-2">
                    <Input
                      type="text"
                      value={inputText}
                      onChange={(e) => setInputText(e.target.value)}
                      placeholder={t("placeholder")}
                      disabled={isSending}
                      className="flex-1 h-10 rounded-lg border-border focus-visible:ring-primary focus-visible:border-primary text-sm font-medium"
                    />
                    <Button
                      type="submit"
                      size="icon"
                      disabled={isSending || !inputText.trim()}
                      className="h-10 w-10 shrink-0 cursor-pointer bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/30 transition-all duration-300"
                    >
                      {isSending ? (
                        <Loader2 className="h-4.5 w-4.5 animate-spin" />
                      ) : (
                        <Send className="h-4.5 w-4.5" />
                      )}
                    </Button>
                  </form>
                </div>
              </>
            ) : (
              <div className="flex-1 flex flex-col items-center justify-center text-center p-8 gap-3 my-auto">
                <MessageSquare className="h-8 w-8 text-muted-foreground" />
                <p className="text-sm text-muted-foreground">{t("empty")}</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </BuyerLayout>
  );
}

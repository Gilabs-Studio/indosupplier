"use client";

import React, { useEffect, useState, useCallback } from "react";
import { supportService } from "@/features/sysadmin/support/services";
import type { SupportTicket } from "@/features/sysadmin/support/types";
import { useSysadminStore } from "@/features/sysadmin/auth/stores/use-sysadmin-store";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  Trash2,
  RefreshCw,
  Inbox,
  Search,
  Filter,
  Loader2,
  Headset,
  Clock,
  UserCheck,
  Send,
  MessageSquare,
  Lock,
  ArrowUpCircle,
  CheckCircle,
  HelpCircle,
  XCircle
} from "lucide-react";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function SupportTickets() {
  const t = useTranslations("sysadminSupport");
  const [tickets, setTickets] = useState<SupportTicket[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const { admin } = useSysadminStore();

  // Search & Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedStatus, setSelectedStatus] = useState<string>("all");
  const [selectedPriority, setSelectedPriority] = useState<string>("all");

  // Detail Console State
  const [isDetailOpen, setIsDetailOpen] = useState(false);
  const [activeTicket, setActiveTicket] = useState<SupportTicket | null>(null);

  // Send message form state
  const [typedMessage, setTypedMessage] = useState("");
  const [isInternalNote, setIsInternalNote] = useState(false);
  const [currentTime, setCurrentTime] = useState<number>(0);

  useEffect(() => {
    const timer = setTimeout(() => {
      setCurrentTime(Date.now());
    }, 0);
    return () => clearTimeout(timer);
  }, []);

  const fetchTickets = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await supportService.list();
      setTickets(data);
    } catch {
      toast.error(t("subtitle"));
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchTickets();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchTickets]);

  const handleOpenDetail = (ticket: SupportTicket) => {
    setActiveTicket(ticket);
    setTypedMessage("");
    setIsInternalNote(false);
    setIsDetailOpen(true);
  };

  const handleSelfAssign = async (ticketId: string) => {
    if (!admin) {
      toast.error("Auth required");
      return;
    }
    try {
      const updated = await supportService.update(ticketId, {
        assignedAgent: admin.name,
        status: "in_progress"
      });
      toast.success(t("successAssign"));
      setActiveTicket(updated);
      fetchTickets();
    } catch {
      toast.error("Error");
    }
  };

  const handleUpdateStatus = async (ticketId: string, newStatus: SupportTicket["status"]) => {
    try {
      const updated = await supportService.update(ticketId, { status: newStatus });
      toast.success(t("successStatus"));
      setActiveTicket(updated);
      fetchTickets();
    } catch {
      toast.error("Error");
    }
  };

  const handleSendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeTicket || !typedMessage.trim()) return;

    try {
      const senderName = admin ? admin.name : "System Admin";
      const newMsg = await supportService.addMessage(activeTicket.id, {
        sender: "agent",
        senderName: senderName,
        content: typedMessage,
        isInternalNote: isInternalNote
      });

      const updatedMessages = [...activeTicket.messages, newMsg];
      const updatedTicket = {
        ...activeTicket,
        messages: updatedMessages,
        updatedAt: new Date().toISOString()
      };
      
      setActiveTicket(updatedTicket);
      setTypedMessage("");
      fetchTickets();
      toast.success(t("successMessage"));
    } catch {
      toast.error("Error");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("delete") + "?")) return;
    try {
      await supportService.delete(id);
      toast.success(t("successDelete"));
      fetchTickets();
    } catch {
      toast.error("Error");
    }
  };

  // Filters logic
  const filteredTickets = tickets.filter(t => {
    const matchesSearch = t.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          t.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          t.userName.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          t.category.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = selectedStatus === "all" || t.status === selectedStatus;
    const matchesPriority = selectedPriority === "all" || t.priority === selectedPriority;
    return matchesSearch && matchesStatus && matchesPriority;
  });

  const getSLAStatus = (slaStr: string, currentMs: number) => {
    if (currentMs === 0) return { label: t("slaCalculating"), color: "bg-muted text-muted-foreground border-border" };
    const timeDiff = new Date(slaStr).getTime() - currentMs;
    if (timeDiff < 0) {
      return { label: t("slaLabelOverdue"), color: "bg-destructive/15 text-destructive border-destructive/20" };
    } else if (timeDiff < 6 * 3600 * 1000) {
      return { label: t("slaLabelWarning"), color: "bg-warning/15 text-warning border-warning/20" };
    }
    return { label: t("slaLabelSafe"), color: "bg-success/15 text-success border-success/20" };
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
        </div>
        <Button
          onClick={fetchTickets}
          variant="outline"
          size="sm"
          className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
          {t("refresh")}
        </Button>
      </div>

      {/* Filter Bar */}
      <Card className="p-4 border border-border bg-card flex flex-col md:flex-row gap-4 items-center justify-between">
        <div className="relative w-full md:w-80">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t("searchPlaceholder")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 bg-background border-border text-sm"
          />
        </div>
        <div className="flex flex-wrap w-full md:w-auto items-center gap-4">
          {/* Status Filter */}
          <div className="flex items-center gap-2">
            <Filter className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">{t("status")}</span>
          </div>
          <Select value={selectedStatus} onValueChange={setSelectedStatus}>
            <SelectTrigger className="w-[140px] bg-background border-border text-xs h-9 cursor-pointer">
              <SelectValue placeholder={t("all")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              <SelectItem value="open" className="cursor-pointer">Open</SelectItem>
              <SelectItem value="in_progress" className="cursor-pointer">In Progress</SelectItem>
              <SelectItem value="resolved" className="cursor-pointer">Resolved</SelectItem>
              <SelectItem value="closed" className="cursor-pointer">Closed</SelectItem>
            </SelectContent>
          </Select>

          {/* Priority Filter */}
          <div className="flex items-center gap-2">
            <ArrowUpCircle className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">{t("priority")}</span>
          </div>
          <Select value={selectedPriority} onValueChange={setSelectedPriority}>
            <SelectTrigger className="w-[130px] bg-background border-border text-xs h-9 cursor-pointer">
              <SelectValue placeholder={t("all")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              <SelectItem value="low" className="cursor-pointer">Low</SelectItem>
              <SelectItem value="medium" className="cursor-pointer">Medium</SelectItem>
              <SelectItem value="high" className="cursor-pointer">High</SelectItem>
              <SelectItem value="urgent" className="cursor-pointer">Urgent</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </Card>

      {/* Content Table Card */}
      <Card className="border border-border/80 shadow-sm overflow-hidden bg-card">
        {isLoading ? (
          <div className="py-24 text-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
            <span className="text-sm font-semibold text-muted-foreground">Loading...</span>
          </div>
        ) : filteredTickets.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">Empty</h3>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("ticketTitle")}</TableHead>
                <TableHead className="font-semibold">{t("sender")}</TableHead>
                <TableHead className="font-semibold">{t("sla")}</TableHead>
                <TableHead className="font-semibold">{t("ticketPriority")}</TableHead>
                <TableHead className="font-semibold">{t("ticketStatus")}</TableHead>
                <TableHead className="font-semibold">{t("agent")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTickets.map((tck) => {
                const sla = getSLAStatus(tck.slaDeadline, currentTime);
                return (
                  <TableRow key={tck.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                    <TableCell className="pl-6 py-4">
                      <div className="text-left space-y-1">
                        <div className="font-bold text-foreground hover:text-primary transition-colors flex items-center gap-1.5">
                          <Headset className="h-4 w-4 text-primary shrink-0" />
                          <span>{tck.title}</span>
                        </div>
                        <div className="text-[10px] text-muted-foreground flex items-center gap-1.5">
                          <span className="font-bold font-mono">{tck.id}</span>
                          <span>&bull;</span>
                          <span>{tck.category}</span>
                        </div>
                      </div>
                    </TableCell>

                    <TableCell className="py-4 text-left">
                      <div className="space-y-0.5">
                        <div className="font-bold text-foreground">{tck.userName}</div>
                        <Badge className="uppercase text-[9px] px-1 py-0.1 border-none bg-secondary text-secondary-foreground">
                          {tck.userType}
                        </Badge>
                      </div>
                    </TableCell>

                    <TableCell className="py-4 text-left">
                      <Badge className={`text-[10px] font-bold px-2 py-0.5 ${sla.color} border`}>
                        {sla.label}
                      </Badge>
                    </TableCell>

                    <TableCell className="py-4 text-left">
                      <Badge className={`capitalize font-bold text-[10px] px-2 py-0.5 ${
                        tck.priority === "urgent"
                          ? "bg-destructive/15 text-destructive border border-destructive/20"
                          : tck.priority === "high"
                          ? "bg-warning/15 text-warning border border-warning/20"
                          : tck.priority === "medium"
                          ? "bg-primary/10 text-primary border border-primary/20"
                          : "bg-muted text-muted-foreground border-none"
                      }`}>
                        {tck.priority}
                      </Badge>
                    </TableCell>

                    <TableCell className="py-4 text-left">
                      <Badge className={`capitalize font-bold text-[10px] px-2 py-0.5 ${
                        tck.status === "open"
                          ? "bg-destructive/15 text-destructive border border-destructive/20"
                          : tck.status === "in_progress"
                          ? "bg-warning/15 text-warning border border-warning/20"
                          : tck.status === "resolved"
                          ? "bg-success/15 text-success border border-success/20"
                          : "bg-muted text-muted-foreground border-border"
                      }`}>
                        {tck.status.replace("_", " ")}
                      </Badge>
                    </TableCell>

                    <TableCell className="py-4 text-left font-semibold text-xs text-muted-foreground">
                      {tck.assignedAgent ? (
                        <div className="flex items-center gap-1">
                          <UserCheck className="h-3.5 w-3.5 text-success" />
                          <span>{tck.assignedAgent}</span>
                        </div>
                      ) : (
                        <span className="text-[10px] text-muted-foreground/60 italic">{t("unassigned")}</span>
                      )}
                    </TableCell>

                    <TableCell className="pr-6 py-4 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleOpenDetail(tck)}
                          className="h-8 w-8 text-primary hover:text-primary/90 hover:bg-muted cursor-pointer"
                          title={t("openConsole")}
                        >
                          <MessageSquare className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(tck.id)}
                          className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 cursor-pointer"
                          title={t("delete")}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        )}
      </Card>

      {/* Ticket Chat Console Dialog */}
      <Dialog open={isDetailOpen} onOpenChange={setIsDetailOpen}>
        <DialogContent className="max-w-2xl bg-card border-border text-foreground flex flex-col h-[550px] p-0 overflow-hidden">
          {/* Header */}
          <div className="p-4 border-b border-border flex items-center justify-between shrink-0">
            <div className="text-left space-y-0.5">
              <div className="text-[10px] font-bold text-primary uppercase tracking-wider flex items-center gap-1">
                <Clock className="h-3.5 w-3.5" />
                SLA: {activeTicket ? new Date(activeTicket.slaDeadline).toLocaleTimeString() : ""}
              </div>
              <DialogTitle className="text-base font-bold text-foreground truncate max-w-[400px]">
                {activeTicket?.title}
              </DialogTitle>
            </div>
            <div className="flex items-center gap-2">
              <Badge className="bg-muted text-muted-foreground border-none font-mono text-[9px] px-1.5 py-0.5">
                {activeTicket?.id}
              </Badge>
              <Badge className="bg-success/15 text-success border border-success/30 capitalize font-bold text-[9px] px-1.5 py-0.5">
                {activeTicket?.status.replace("_", " ")}
              </Badge>
            </div>
          </div>

          {/* Console Body Layout */}
          <div className="flex flex-1 overflow-hidden">
            {/* Left: Chat history */}
            <div className="flex-1 flex flex-col bg-muted/10 border-r border-border h-full overflow-hidden">
              <div className="flex-1 p-4 overflow-y-auto space-y-3">
                {activeTicket?.messages.map((msg) => {
                  const isNote = msg.isInternalNote;
                  const isSystem = msg.sender === "system";
                  
                  if (isSystem) {
                    return (
                      <div key={msg.id} className="text-center my-3">
                        <span className="text-[9px] bg-muted text-muted-foreground font-bold px-2 py-0.5 rounded-full border border-border">
                          {msg.content}
                        </span>
                      </div>
                    );
                  }

                  return (
                    <div
                      key={msg.id}
                      className={`flex flex-col max-w-[85%] space-y-0.5 ${
                        msg.sender === "user" ? "text-left float-left" : "text-right float-right ml-auto"
                      }`}
                      style={{ clear: "both" }}
                    >
                      <span className="text-[9px] text-muted-foreground font-bold px-1">
                        {msg.senderName} {isNote && " (Catatan Internal)"}
                      </span>
                      <div className={`p-2.5 rounded-lg text-xs font-normal text-left ${
                        isNote
                          ? "bg-warning/15 text-warning-foreground border border-warning/30"
                          : msg.sender === "user"
                          ? "bg-background text-foreground border border-border"
                          : "bg-primary text-primary-foreground"
                      }`}>
                        {msg.content}
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Chat Input form */}
              <form onSubmit={handleSendMessage} className="p-3 border-t border-border bg-background flex flex-col gap-2 shrink-0">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-1.5">
                    <input
                      type="checkbox"
                      id="internal-checkbox"
                      checked={isInternalNote}
                      onChange={(e) => setIsInternalNote(e.target.checked)}
                      className="h-3.5 w-3.5 text-warning focus:ring-warning border-border rounded cursor-pointer"
                    />
                    <label htmlFor="internal-checkbox" className="text-[10px] font-bold text-muted-foreground flex items-center gap-1 select-none cursor-pointer">
                      <Lock className="h-3 w-3 text-warning" />
                      {t("internalNote")}
                    </label>
                  </div>
                </div>
                <div className="flex gap-2">
                  <Input
                    placeholder={isInternalNote ? t("internalNotePlaceholder") : t("chatInputPlaceholder")}
                    value={typedMessage}
                    onChange={(e) => setTypedMessage(e.target.value)}
                    className="flex-1 bg-background border-border text-xs"
                  />
                  <Button type="submit" size="sm" className="bg-primary hover:bg-primary/90 text-primary-foreground cursor-pointer h-9 px-3">
                    <Send className="h-3.5 w-3.5" />
                  </Button>
                </div>
              </form>
            </div>

            {/* Right: Actions / Info Panel */}
            <div className="w-[200px] shrink-0 p-4 space-y-4 text-left text-xs bg-card overflow-y-auto">
              <div>
                <h4 className="font-bold text-muted-foreground text-[10px] uppercase tracking-wider mb-2">{t("agent")}</h4>
                {activeTicket?.assignedAgent ? (
                  <div className="p-2 border border-border rounded-lg flex flex-col gap-1 items-start bg-muted/10">
                    <span className="font-bold text-foreground text-xs">{activeTicket.assignedAgent}</span>
                  </div>
                ) : (
                  <Button
                    type="button"
                    size="sm"
                    onClick={() => activeTicket && handleSelfAssign(activeTicket.id)}
                    className="w-full bg-primary hover:bg-primary/90 text-primary-foreground font-bold text-[10px] py-1 h-7"
                  >
                    {t("selfAssign")}
                  </Button>
                )}
              </div>

              <div>
                <h4 className="font-bold text-muted-foreground text-[10px] uppercase tracking-wider mb-2">{t("updateStatus")}</h4>
                <div className="flex flex-col gap-1.5">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => activeTicket && handleUpdateStatus(activeTicket.id, "resolved")}
                    className="w-full justify-start text-[10px] font-bold border-border h-7 text-success hover:bg-success/10 cursor-pointer"
                  >
                    <CheckCircle className="h-3 w-3 mr-1" />
                    {t("resolveTicket")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => activeTicket && handleUpdateStatus(activeTicket.id, "closed")}
                    className="w-full justify-start text-[10px] font-bold border-border h-7 text-muted-foreground hover:bg-muted/50 cursor-pointer"
                  >
                    <XCircle className="h-3 w-3 mr-1" />
                    {t("closeTicket")}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => activeTicket && handleUpdateStatus(activeTicket.id, "open")}
                    className="w-full justify-start text-[10px] font-bold border-border h-7 text-destructive hover:bg-destructive/10 cursor-pointer"
                  >
                    <HelpCircle className="h-3 w-3 mr-1" />
                    {t("reopenTicket")}
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

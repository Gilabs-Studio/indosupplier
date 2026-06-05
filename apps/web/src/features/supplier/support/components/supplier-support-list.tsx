"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";
import { Plus, Search, MessageSquare, ArrowUpRight, Calendar } from "lucide-react";

export function SupplierSupportList() {
  const t = useTranslations("supplier.support");
  const [search, setSearch] = useState("");

  const [tickets, setTickets] = useState([
    { id: "TCK-2026-908", subject: "Verification status update request", category: "Account Verification", status: "open", updated: "2026-06-04" },
    { id: "TCK-2026-892", subject: "Invoice print formatting error", category: "Billing & Invoices", status: "resolved", updated: "2026-05-20" },
    { id: "TCK-2026-814", subject: "Custom domain mapping request", category: "Technical Support", status: "resolved", updated: "2026-05-02" },
  ]);

  const [showModal, setShowModal] = useState(false);
  const [newTicket, setNewTicket] = useState({ subject: "", category: "General Inquiries" });

  const handleCreateTicket = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTicket.subject.trim()) {
      toast.error("Please provide a subject.");
      return;
    }
    const item = {
      id: `TCK-2026-${Math.floor(Math.random() * 900) + 100}`,
      subject: newTicket.subject,
      category: newTicket.category,
      status: "open",
      updated: "Just now",
    };
    setTickets([item, ...tickets]);
    setShowModal(false);
    setNewTicket({ subject: "", category: "General Inquiries" });
    toast.success("Support ticket opened successfully!");
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "open":
        return <Badge className="bg-primary/10 text-primary border border-primary/20 font-bold">{t("statusOpen")}</Badge>;
      default:
        return <Badge className="bg-success/15 text-success border border-success/30 font-bold">{t("statusResolved")}</Badge>;
    }
  };

  const filtered = tickets.filter(tck =>
    tck.subject.toLowerCase().includes(search.toLowerCase()) ||
    tck.id.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6 text-left">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("subtitle")}
          </p>
        </div>
        <Button onClick={() => setShowModal(true)} className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20 font-semibold">
          <Plus className="mr-2 h-4 w-4" /> {t("btnCreate")}
        </Button>
      </div>

      {/* Filter and Table */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="relative max-w-xs w-full">
          <Search className="absolute left-3 top-2.5 h-4.5 w-4.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search tickets..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-1.5 bg-card border border-border text-sm rounded-lg outline-hidden focus:border-primary transition-all text-left"
          />
        </div>
      </div>

      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-b border-border bg-muted/10">
                  <TableHead className="font-bold text-foreground py-3 px-6">{t("tableId")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableSubject")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableCategory")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableStatus")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("tableUpdated")}</TableHead>
                  <TableHead className="font-bold text-foreground text-right px-6">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((tck) => (
                  <TableRow key={tck.id} className="hover:bg-muted/10 border-b border-border transition-colors">
                    <TableCell className="py-4 px-6 font-bold text-muted-foreground">{tck.id}</TableCell>
                    <TableCell className="py-4">
                      <Link href={`/supplier/support/${tck.id}`} className="font-semibold text-foreground hover:text-primary transition-colors flex items-center gap-2">
                        <MessageSquare className="h-4.5 w-4.5 text-muted-foreground shrink-0" />
                        {tck.subject}
                      </Link>
                    </TableCell>
                    <TableCell className="py-4 font-semibold text-xs text-muted-foreground">{tck.category}</TableCell>
                    <TableCell className="py-4">{getStatusBadge(tck.status)}</TableCell>
                    <TableCell className="py-4 text-xs font-semibold text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <Calendar className="h-3.5 w-3.5" />
                        {tck.updated}
                      </span>
                    </TableCell>
                    <TableCell className="py-4 px-6 text-right">
                      <Button asChild size="sm" variant="outline" className="text-xs font-semibold h-8 cursor-pointer border-border">
                        <Link href={`/supplier/support/${tck.id}`}>
                          Chat
                          <ArrowUpRight className="ml-1 h-3.5 w-3.5" />
                        </Link>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Create modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs">
          <Card className="max-w-md w-full border border-border bg-card shadow-2xl rounded-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <CardHeader className="border-b border-border">
              <CardTitle className="text-base font-bold font-heading">{t("btnCreate")}</CardTitle>
              <CardDescription className="text-xs">Raise a support ticket to sysadmin team.</CardDescription>
            </CardHeader>
            <form onSubmit={handleCreateTicket}>
              <CardContent className="p-6 space-y-4">
                <div className="space-y-1">
                  <label className="text-xs font-bold text-muted-foreground uppercase">{t("tableSubject")}</label>
                  <Input
                    required
                    placeholder="Brief description of the problem..."
                    value={newTicket.subject}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewTicket({ ...newTicket, subject: e.target.value })}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-bold text-muted-foreground uppercase">{t("tableCategory")}</label>
                  <select
                    value={newTicket.category}
                    onChange={(e) => setNewTicket({ ...newTicket, category: e.target.value })}
                    className="w-full px-3 py-2 bg-card border border-border text-sm rounded-lg focus:outline-none focus:border-primary transition-all text-left text-foreground font-semibold"
                  >
                    <option>General Inquiries</option>
                    <option>Account Verification</option>
                    <option>Billing & Invoices</option>
                    <option>Technical Support</option>
                  </select>
                </div>
              </CardContent>
              <div className="p-4 border-t border-border bg-muted/10 flex justify-end gap-2">
                <Button variant="outline" onClick={() => setShowModal(false)} className="text-xs h-9 cursor-pointer border-border">
                  Cancel
                </Button>
                <Button type="submit" className="text-xs h-9 bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold">
                  Open Ticket
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}
    </div>
  );
}

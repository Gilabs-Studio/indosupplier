"use client";

import React, { useEffect, useState, useCallback } from "react";
import { supplierService } from "@/features/sysadmin/suppliers/services";
import type { Supplier } from "@/features/sysadmin/suppliers/types";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  Trash2,
  RefreshCw,
  Inbox,
  Search,
  Filter,
  Loader2,
  Building,
  ShieldCheck,
  CheckCircle,
  Eye,
  AlertOctagon,
  Ban,
  UserX,
  FileText
} from "lucide-react";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function SuppliersDirectory() {
  const t = useTranslations("sysadminSuppliers");
  const [suppliers, setSuppliers] = useState<Supplier[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Search & Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedLevel, setSelectedLevel] = useState<string>("all");
  const [selectedStatus, setSelectedStatus] = useState<string>("all");

  // Profile Detail modal state
  const [isDetailOpen, setIsDetailOpen] = useState(false);
  const [activeSupplier, setActiveSupplier] = useState<Supplier | null>(null);

  // Suspend/Ban modal state
  const [isActionOpen, setIsActionOpen] = useState(false);
  const [actionType, setActionType] = useState<"suspend" | "ban">("suspend");
  const [actionReason, setActionReason] = useState("");

  const fetchSuppliers = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await supplierService.list();
      setSuppliers(data);
    } catch {
      toast.error(t("subtitle"));
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchSuppliers();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchSuppliers]);

  const handleOpenDetail = (supplier: Supplier) => {
    setActiveSupplier(supplier);
    setIsDetailOpen(true);
  };

  const handleOpenAction = (supplier: Supplier, type: "suspend" | "ban") => {
    setActiveSupplier(supplier);
    setActionType(type);
    setActionReason("");
    setIsActionOpen(true);
  };

  const handleUpdateStatus = async (id: string, newStatus: Supplier["status"]) => {
    try {
      await supplierService.update(id, { status: newStatus });
      toast.success(t("successStatus"));
      setIsDetailOpen(false);
      fetchSuppliers();
    } catch {
      toast.error("Error");
    }
  };

  const handlePromoteLevel = async (id: string, currentLevel: number) => {
    if (currentLevel >= 3) return;
    const nextLevel = (currentLevel + 1) as Supplier["verificationLevel"];
    try {
      await supplierService.update(id, { verificationLevel: nextLevel });
      toast.success(t("successLevel"));
      setIsDetailOpen(false);
      fetchSuppliers();
    } catch {
      toast.error("Error");
    }
  };

  const handleActionSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeSupplier) return;
    if (!actionReason.trim()) {
      toast.error("Reason required");
      return;
    }

    const targetStatus = actionType === "suspend" ? "suspended" : "banned";
    try {
      await supplierService.update(activeSupplier.id, { status: targetStatus });
      toast.success(t("successAction"));
      setIsActionOpen(false);
      fetchSuppliers();
    } catch {
      toast.error("Error");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("delete") + "?")) return;
    try {
      await supplierService.delete(id);
      toast.success(t("successDelete"));
      fetchSuppliers();
    } catch {
      toast.error("Error");
    }
  };

  // Metrics
  const totalCount = suppliers.length;
  const verifiedCount = suppliers.filter(s => s.verificationLevel === 3).length;
  const suspendedCount = suppliers.filter(s => s.status === "suspended" || s.status === "banned").length;
  const pendingCount = suppliers.filter(s => s.verificationLevel === 2 && s.status === "active").length;

  // Filters logic
  const filteredSuppliers = suppliers.filter(s => {
    const matchesSearch = s.companyName.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          s.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          s.nib.includes(searchQuery);
    const matchesLevel = selectedLevel === "all" || s.verificationLevel.toString() === selectedLevel;
    const matchesStatus = selectedStatus === "all" || s.status === selectedStatus;
    return matchesSearch && matchesLevel && matchesStatus;
  });

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
        </div>
        <Button
          onClick={fetchSuppliers}
          variant="outline"
          size="sm"
          className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
          {t("refresh")}
        </Button>
      </div>

      {/* Metrics Section */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-4 text-left">
        <Card className="p-4 border border-border bg-card hover:shadow-md transition-shadow">
          <span className="text-[10px] uppercase font-bold text-muted-foreground tracking-wider">{t("totalSuppliers")}</span>
          <h2 className="text-2xl font-extrabold text-foreground mt-1">{totalCount}</h2>
        </Card>

        <Card className="p-4 border border-border bg-card hover:shadow-md transition-shadow">
          <span className="text-[10px] uppercase font-bold text-success tracking-wider">{t("verifiedLvl3")}</span>
          <h2 className="text-2xl font-extrabold text-success mt-1">{verifiedCount}</h2>
        </Card>

        <Card className="p-4 border border-border bg-card hover:shadow-md transition-shadow">
          <span className="text-[10px] uppercase font-bold text-warning tracking-wider">{t("pendingLvl2")}</span>
          <h2 className="text-2xl font-extrabold text-warning mt-1">{pendingCount}</h2>
        </Card>

        <Card className="p-4 border border-border bg-card hover:shadow-md transition-shadow">
          <span className="text-[10px] uppercase font-bold text-destructive tracking-wider">{t("suspended")}</span>
          <h2 className="text-2xl font-extrabold text-destructive mt-1">{suspendedCount}</h2>
        </Card>
      </div>

      {/* Filters Bar */}
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
          {/* Level Filter */}
          <div className="flex items-center gap-2">
            <Filter className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">{t("filterLevel")}</span>
          </div>
          <Select value={selectedLevel} onValueChange={setSelectedLevel}>
            <SelectTrigger className="w-[140px] bg-background border-border text-xs h-9 cursor-pointer">
              <SelectValue placeholder={t("all")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              <SelectItem value="1" className="cursor-pointer">Level 1 (Standard)</SelectItem>
              <SelectItem value="2" className="cursor-pointer">Level 2 (Doc Uploaded)</SelectItem>
              <SelectItem value="3" className="cursor-pointer">Level 3 (Verified)</SelectItem>
            </SelectContent>
          </Select>

          {/* Status Filter */}
          <div className="flex items-center gap-2">
            <Filter className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">{t("filterStatus")}</span>
          </div>
          <Select value={selectedStatus} onValueChange={setSelectedStatus}>
            <SelectTrigger className="w-[130px] bg-background border-border text-xs h-9 cursor-pointer">
              <SelectValue placeholder={t("all")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              <SelectItem value="active" className="cursor-pointer">Aktif</SelectItem>
              <SelectItem value="inactive" className="cursor-pointer">Non-aktif</SelectItem>
              <SelectItem value="suspended" className="cursor-pointer">Ditangguhkan</SelectItem>
              <SelectItem value="banned" className="cursor-pointer">Diblokir (Banned)</SelectItem>
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
        ) : filteredSuppliers.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">Empty</h3>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("companyName")}</TableHead>
                <TableHead className="font-semibold">{t("nibType")}</TableHead>
                <TableHead className="font-semibold">{t("verifLevel")}</TableHead>
                <TableHead className="font-semibold">{t("taxStatus")}</TableHead>
                <TableHead className="font-semibold">{t("accountStatus")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredSuppliers.map((sup) => (
                <TableRow key={sup.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  <TableCell className="pl-6 py-4">
                    <div className="text-left space-y-0.5">
                      <div className="font-bold text-foreground flex items-center gap-1.5">
                        <Building className="h-4 w-4 text-primary shrink-0" />
                        {sup.companyName}
                      </div>
                      <div className="text-xs text-muted-foreground">{sup.email}</div>
                    </div>
                  </TableCell>

                  <TableCell className="py-4 text-left font-normal text-xs text-muted-foreground">
                    <div className="space-y-0.5">
                      <div className="font-bold text-foreground">{sup.companyType}</div>
                      <div className="font-mono">NIB: {sup.nib}</div>
                    </div>
                  </TableCell>

                  <TableCell className="py-4 text-left">
                    <div className="flex items-center gap-1.5">
                      <ShieldCheck className={`h-4 w-4 ${
                        sup.verificationLevel === 3 ? "text-success" : sup.verificationLevel === 2 ? "text-warning" : "text-muted-foreground"
                      }`} />
                      <span className="font-bold text-xs">
                        Level {sup.verificationLevel}
                      </span>
                    </div>
                  </TableCell>

                  <TableCell className="py-4 text-left">
                    <Badge className="uppercase text-[10px] font-bold px-2 py-0.5 bg-secondary text-secondary-foreground">
                      {sup.taxStatus}
                    </Badge>
                  </TableCell>

                  <TableCell className="py-4 text-left">
                    <Badge className={`capitalize font-bold text-[10px] px-2 py-0.5 ${
                      sup.status === "active"
                        ? "bg-success/15 text-success border-success/30"
                        : sup.status === "inactive"
                        ? "bg-muted text-muted-foreground border-border"
                        : sup.status === "suspended"
                        ? "bg-warning/15 text-warning border-warning/30"
                        : "bg-destructive/15 text-destructive border-destructive/30"
                    }`}>
                      {sup.status}
                    </Badge>
                  </TableCell>

                  <TableCell className="pr-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleOpenDetail(sup)}
                        className="h-8 w-8 text-primary hover:text-primary/90 hover:bg-muted cursor-pointer"
                        title={t("detail")}
                      >
                        <Eye className="h-4 w-4" />
                      </Button>

                      {sup.status === "active" && (
                        <>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleOpenAction(sup, "suspend")}
                            className="h-8 w-8 text-warning hover:text-warning/90 hover:bg-warning/10 cursor-pointer"
                            title={t("suspend")}
                          >
                            <AlertOctagon className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleOpenAction(sup, "ban")}
                            className="h-8 w-8 text-destructive hover:text-destructive/90 hover:bg-destructive/10 cursor-pointer"
                            title={t("ban")}
                          >
                            <Ban className="h-4 w-4" />
                          </Button>
                        </>
                      )}

                      {sup.status !== "active" && sup.status !== "inactive" && (
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleUpdateStatus(sup.id, "active")}
                          className="h-8 w-8 text-success hover:text-success/90 hover:bg-success/10 cursor-pointer"
                          title={t("reactivate")}
                        >
                          <CheckCircle className="h-4 w-4" />
                        </Button>
                      )}

                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(sup.id)}
                        className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 cursor-pointer"
                        title={t("delete")}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </Card>

      {/* Supplier Profile Detail Dialog */}
      <Dialog open={isDetailOpen} onOpenChange={setIsDetailOpen}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left border-b border-border pb-3">
            <div className="flex items-center gap-2 text-primary font-bold mb-1">
              <Building className="h-5 w-5" />
              {t("detailTitle")}
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {activeSupplier?.companyName}
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              {t("joinedDate")} {activeSupplier ? new Date(activeSupplier.joinedDate).toLocaleDateString() : ""}
            </DialogDescription>
          </DialogHeader>

          {activeSupplier && (
            <div className="my-4 space-y-4 text-left text-sm max-h-[350px] overflow-y-auto pr-1 font-normal">
              <div className="grid grid-cols-3 gap-2 py-1.5 border-b border-border">
                <span className="text-xs text-muted-foreground font-bold uppercase col-span-1">{t("taxStatus")}</span>
                <span className="col-span-2 capitalize font-semibold">{activeSupplier.taxStatus}</span>
              </div>
              <div className="grid grid-cols-3 gap-2 py-1.5 border-b border-border">
                <span className="text-xs text-muted-foreground font-bold uppercase col-span-1">{t("companyType")}</span>
                <span className="col-span-2 font-semibold">{activeSupplier.companyType}</span>
              </div>
              <div className="grid grid-cols-3 gap-2 py-1.5 border-b border-border">
                <span className="text-xs text-muted-foreground font-bold uppercase col-span-1">{t("nib")}</span>
                <span className="col-span-2 font-mono font-semibold">{activeSupplier.nib}</span>
              </div>
              <div className="grid grid-cols-3 gap-2 py-1.5 border-b border-border">
                <span className="text-xs text-muted-foreground font-bold uppercase col-span-1">{t("npwp")}</span>
                <span className="col-span-2 font-mono font-semibold">{activeSupplier.npwp}</span>
              </div>
              <div className="grid grid-cols-3 gap-2 py-1.5 border-b border-border">
                <span className="text-xs text-muted-foreground font-bold uppercase col-span-1">{t("verifLevel")}</span>
                <span className="col-span-2 font-bold text-primary">Level {activeSupplier.verificationLevel}</span>
              </div>

              {/* Verification Queue Action */}
              {activeSupplier.verificationLevel < 3 && activeSupplier.verificationDocumentUrl && (
                <div className="p-3 bg-warning/15 dark:bg-warning/10 rounded border border-warning/30 mt-4 space-y-3">
                  <div className="flex items-center gap-1.5 text-warning font-bold text-xs uppercase">
                    <FileText className="h-4 w-4" />
                    {t("docReviewTitle")}
                  </div>
                  <p className="text-xs text-muted-foreground leading-relaxed">
                    {t("docReviewDesc", { level: activeSupplier.verificationLevel + 1 })}
                  </p>
                  <div className="flex gap-2 pt-1.5">
                    <Button
                      type="button"
                      size="sm"
                      onClick={() => handlePromoteLevel(activeSupplier.id, activeSupplier.verificationLevel)}
                      className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold text-xs"
                    >
                      {t("approvePromote")}
                    </Button>
                    <a
                      href={activeSupplier.verificationDocumentUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center justify-center px-3 py-1.5 text-xs font-bold border border-border bg-background rounded-lg hover:bg-muted/40 text-foreground transition-all cursor-pointer"
                    >
                      {t("viewDoc")}
                    </a>
                  </div>
                </div>
              )}
            </div>
          )}

          <DialogFooter className="border-t border-border pt-3">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsDetailOpen(false)}
              className="border-border cursor-pointer text-foreground"
            >
              {t("close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Suspend / Ban Action Dialog */}
      <Dialog open={isActionOpen} onOpenChange={setIsActionOpen}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-destructive font-bold mb-1">
              <UserX className="h-5 w-5" />
              {t("detailTitle")}
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {actionType === "suspend" ? t("suspend") : t("ban")}
            </DialogTitle>
          </DialogHeader>

          <form onSubmit={handleActionSubmit} className="space-y-4 my-2 text-left">
            <FieldGroup>
              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("actionReason")}</FieldLabel>
                <textarea
                  placeholder={t("actionReasonPlaceholder")}
                  value={actionReason}
                  onChange={(e) => setActionReason(e.target.value)}
                  rows={3}
                  className="w-full bg-background border border-border rounded-lg p-2.5 text-sm placeholder-muted-foreground focus:outline-none focus:border-destructive focus:ring-1 focus:ring-destructive transition-all"
                  required
                />
              </Field>
            </FieldGroup>

            <DialogFooter className="pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setIsActionOpen(false)}
                className="border-border cursor-pointer text-foreground"
              >
                {t("cancel")}
              </Button>
              <Button
                type="submit"
                className="bg-destructive hover:bg-destructive/90 text-destructive-foreground font-bold cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-transform"
              >
                {t("submitAction")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";
import { Award, Plus, Trash2, Calendar, Upload } from "lucide-react";
import { DeleteDialog } from "@/components/ui/delete-dialog";

export function SupplierCertifications() {
  const t = useTranslations("supplier.profile");
  const [certs, setCerts] = useState([
    { id: "CERT-01", name: "ISO 9001:2015 Quality Management", issuer: "TUV SUD", validUntil: "2027-12-15", status: "verified" },
    { id: "CERT-02", name: "Halal Product Assurance (MUI)", issuer: "BPJPH Indonesia", validUntil: "2028-06-20", status: "verified" },
    { id: "CERT-03", name: "SNI 15-0129-2004 Crystalline Silica Standard", issuer: "BSN Indonesia", validUntil: "2026-11-02", status: "pending" },
  ]);

  const [showModal, setShowModal] = useState(false);
  const [newCert, setNewCert] = useState({ name: "", issuer: "", validUntil: "" });

  const handleUpload = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCert.name || !newCert.issuer || !newCert.validUntil) {
      toast.error("Please fill out all fields.");
      return;
    }

    const item = {
      id: `CERT-0${certs.length + 1}`,
      name: newCert.name,
      issuer: newCert.issuer,
      validUntil: newCert.validUntil,
      status: "pending",
    };

    setCerts([...certs, item]);
    setShowModal(false);
    setNewCert({ name: "", issuer: "", validUntil: "" });
    toast.success("Certificate uploaded for compliance check!");
  };

  const [deleteId, setDeleteId] = useState<string | null>(null);

  const handleDelete = (id: string) => {
    setDeleteId(id);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "verified":
        return <Badge className="bg-success/15 text-success border border-success/30 font-bold">{t("statusVerified")}</Badge>;
      case "pending":
        return <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold">{t("statusPending")}</Badge>;
      default:
        return <Badge variant="secondary">{t("statusEmpty")}</Badge>;
    }
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("certsTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("certsSubtitle")}
          </p>
        </div>
        <Button onClick={() => setShowModal(true)} className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20 font-semibold">
          <Plus className="mr-2 h-4 w-4" /> {t("btnUpload")}
        </Button>
      </div>

      {/* Certs Table */}
      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-b border-border bg-muted/10">
                  <TableHead className="font-bold text-foreground py-3 px-6">ID</TableHead>
                  <TableHead className="font-bold text-foreground">Certificate Title</TableHead>
                  <TableHead className="font-bold text-foreground">Issuing Authority</TableHead>
                  <TableHead className="font-bold text-foreground">Expiration Date</TableHead>
                  <TableHead className="font-bold text-foreground">Status</TableHead>
                  <TableHead className="font-bold text-foreground text-right px-6">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {certs.map((c) => (
                  <TableRow key={c.id} className="hover:bg-muted/10 border-b border-border transition-colors">
                    <TableCell className="py-4 px-6 font-bold text-muted-foreground">{c.id}</TableCell>
                    <TableCell className="py-4 font-semibold text-foreground flex items-center gap-2">
                      <Award className="h-4.5 w-4.5 text-primary shrink-0" />
                      <span>{c.name}</span>
                    </TableCell>
                    <TableCell className="py-4 text-sm font-medium">{c.issuer}</TableCell>
                    <TableCell className="py-4 text-xs font-semibold text-muted-foreground">
                      <span className="flex items-center gap-1">
                        <Calendar className="h-3.5 w-3.5" />
                        {c.validUntil}
                      </span>
                    </TableCell>
                    <TableCell className="py-4">{getStatusBadge(c.status)}</TableCell>
                    <TableCell className="py-4 px-6 text-right">
                      <Button onClick={() => handleDelete(c.id)} variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-destructive cursor-pointer hover:bg-destructive/10 border border-border">
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Upload Modal simulation */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs">
          <Card className="max-w-md w-full border border-border bg-card shadow-2xl rounded-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <CardHeader className="border-b border-border">
              <CardTitle className="text-base font-bold font-heading">{t("btnUpload")}</CardTitle>
              <CardDescription className="text-xs">Provide details and scan copies of trade certificates.</CardDescription>
            </CardHeader>
            <form onSubmit={handleUpload}>
              <CardContent className="p-6 space-y-4">
                <div className="space-y-1">
                  <label className="text-xs font-bold text-muted-foreground uppercase">{t("certName")}</label>
                  <Input
                    required
                    placeholder="e.g. ISO 9001:2015 Standards"
                    value={newCert.name}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewCert({ ...newCert, name: e.target.value })}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-bold text-muted-foreground uppercase">{t("issuer")}</label>
                  <Input
                    required
                    placeholder="e.g. SGS Certification"
                    value={newCert.issuer}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewCert({ ...newCert, issuer: e.target.value })}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-bold text-muted-foreground uppercase">{t("validUntil")}</label>
                  <Input
                    required
                    type="date"
                    value={newCert.validUntil}
                    onChange={(e: React.ChangeEvent<HTMLInputElement>) => setNewCert({ ...newCert, validUntil: e.target.value })}
                  />
                </div>
                <div className="border-2 border-dashed border-border p-4 rounded-xl flex flex-col items-center justify-center gap-2 bg-muted/5">
                  <Upload className="h-5 w-5 text-muted-foreground" />
                  <span className="text-xs font-semibold">Select PDF or Image</span>
                </div>
              </CardContent>
              <div className="p-4 border-t border-border bg-muted/10 flex justify-end gap-2">
                <Button variant="outline" onClick={() => setShowModal(false)} className="text-xs h-9 cursor-pointer border-border">
                  Cancel
                </Button>
                <Button type="submit" className="text-xs h-9 bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold">
                  Upload & Save
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}

      <DeleteDialog
        open={!!deleteId}
        onOpenChange={(open) => !open && setDeleteId(null)}
        onConfirm={() => {
          if (deleteId) {
            setCerts(certs.filter((c) => c.id !== deleteId));
            toast.success("Certificate removed.");
          }
        }}
        itemName="certificate"
      />
    </div>
  );
}

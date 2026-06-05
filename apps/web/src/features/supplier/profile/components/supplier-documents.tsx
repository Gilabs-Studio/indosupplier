"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";
import { FileCheck, FileText, Upload, Trash2, CheckCircle2 } from "lucide-react";
import { DeleteDialog } from "@/components/ui/delete-dialog";

export function SupplierDocuments() {
  const t = useTranslations("supplier.profile");
  const [docs, setDocs] = useState([
    { id: "DOC-01", name: "Nomor Induk Berusaha (NIB) - Business License", number: "9120001234567", file: "NIB_PT_Nusantara.pdf", status: "verified" },
    { id: "DOC-02", name: "NPWP Scan - Corporate Tax ID Cert", number: "01.234.567.8-999.000", file: "NPWP_PT_Nusantara.pdf", status: "verified" },
    { id: "DOC-03", name: "Surat Izin Usaha Perdagangan (SIUP)", number: "503/123/SIUP/2021", file: "SIUP_Nusantara.pdf", status: "pending" },
    { id: "DOC-04", name: "Tanda Daftar Perusahaan (TDP)", number: "", file: "", status: "empty" },
  ]);

  const handleUploadSimulate = (id: string) => {
    toast.promise(
      new Promise((resolve) => setTimeout(resolve, 800)),
      {
        loading: "Uploading permit document...",
        success: () => {
          setDocs(docs.map(d => d.id === id ? { ...d, file: "Uploaded_Document.pdf", number: "TEMP-REF-1234", status: "pending" } : d));
          return "Document uploaded for verification!";
        },
        error: "Upload failed",
      }
    );
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
        return <Badge variant="secondary" className="bg-secondary/40 font-bold">{t("statusEmpty")}</Badge>;
    }
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("docsTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("docsSubtitle")}
          </p>
        </div>
      </div>

      {/* Docs Table */}
      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-b border-border bg-muted/10">
                  <TableHead className="font-bold text-foreground py-3 px-6">{t("docName")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("docNumber")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("file")}</TableHead>
                  <TableHead className="font-bold text-foreground">Status</TableHead>
                  <TableHead className="font-bold text-foreground text-right px-6">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {docs.map((d) => (
                  <TableRow key={d.id} className="hover:bg-muted/10 border-b border-border transition-colors">
                    <TableCell className="py-4 px-6 font-semibold text-foreground">
                      <div className="flex items-center gap-2">
                        {d.status === "verified" ? (
                          <CheckCircle2 className="h-4.5 w-4.5 text-success shrink-0" />
                        ) : (
                          <FileText className="h-4.5 w-4.5 text-muted-foreground shrink-0" />
                        )}
                        <span>{d.name}</span>
                      </div>
                    </TableCell>
                    <TableCell className="py-4 text-xs font-semibold font-mono text-muted-foreground">
                      {d.number || "—"}
                    </TableCell>
                    <TableCell className="py-4 text-xs font-medium truncate max-w-[150px]">
                      {d.file ? (
                        <span className="text-primary hover:underline cursor-pointer flex items-center gap-1 font-semibold">
                          <FileCheck className="h-3.5 w-3.5 text-primary shrink-0" />
                          {d.file}
                        </span>
                      ) : (
                        <span className="text-muted-foreground italic">No file attached</span>
                      )}
                    </TableCell>
                    <TableCell className="py-4">{getStatusBadge(d.status)}</TableCell>
                    <TableCell className="py-4 px-6 text-right space-x-1">
                      {d.status === "empty" ? (
                        <Button onClick={() => handleUploadSimulate(d.id)} variant="outline" size="sm" className="text-xs h-8 cursor-pointer border-border">
                          <Upload className="mr-1.5 h-3.5 w-3.5" /> Upload File
                        </Button>
                      ) : (
                        <Button onClick={() => handleDelete(d.id)} variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-destructive cursor-pointer hover:bg-destructive/10 border border-border">
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      <DeleteDialog
        open={!!deleteId}
        onOpenChange={(open) => !open && setDeleteId(null)}
        onConfirm={() => {
          if (deleteId) {
            setDocs(docs.map(d => d.id === deleteId ? { ...d, file: "", number: "", status: "empty" } : d));
            toast.success("Document removed.");
          }
        }}
        itemName="document reference"
      />
    </div>
  );
}

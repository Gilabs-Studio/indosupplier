"use client";

import React from "react";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { FileText, ShieldCheck, FileUp } from "lucide-react";

export function BuyerDocumentsPage() {

  const documents = [
    { name: "Nomor Induk Berusaha (NIB)", code: "NIB-912048123", status: "Verified", date: "2026-01-10" },
    { name: "Surat Izin Usaha Perdagangan (SIUP)", code: "SIUP-412/10-24/2022", status: "Verified", date: "2026-01-10" },
    { name: "NPWP Perusahaan", code: "NPWP-01.234.567.8-012.000", status: "Verified", date: "2026-01-10" },
  ];

  return (
    <BuyerLayout>
      <div className="space-y-6 max-w-3xl">
        {/* Header */}
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">Dokumen Legalitas Perusahaan</h1>
          <p className="text-sm text-muted-foreground">Kelola dokumen identitas perusahaan B2B Anda untuk mendapatkan limit kredit sourcing yang lebih tinggi.</p>
        </div>

        {/* Documents Card */}
        <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
          <CardContent className="p-6 space-y-6">
            <div className="divide-y divide-border">
              {documents.map((doc, idx) => (
                <div key={idx} className="py-4 first:pt-0 last:pb-0 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                  <div className="flex items-start gap-3">
                    <div className="p-2 bg-primary/10 text-primary rounded-lg mt-0.5">
                      <FileText className="h-5 w-5" />
                    </div>
                    <div className="space-y-0.5">
                      <h4 className="text-sm font-semibold text-foreground">{doc.name}</h4>
                      <p className="text-xs text-muted-foreground">ID: {doc.code} • Diupload pada {doc.date}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3 self-end sm:self-center">
                    <Badge className="bg-success text-white border-0 text-[10px] px-2.5 py-0.5 rounded-full flex items-center gap-0.5">
                      <ShieldCheck className="h-3 w-3" /> {doc.status}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>

            {/* Upload Area */}
            <div className="pt-6 border-t border-border space-y-3">
              <h4 className="text-sm font-semibold text-foreground">Upload Dokumen Tambahan (Akta Perusahaan / Rekening Koran)</h4>
              <div className="border border-dashed border-border hover:border-primary/50 transition-colors rounded-lg p-6 text-center cursor-pointer space-y-2">
                <FileUp className="mx-auto h-8 w-8 text-muted-foreground opacity-60" />
                <p className="text-xs font-semibold text-foreground">Klik untuk upload atau drag & drop file PDF</p>
                <p className="text-[10px] text-muted-foreground">Maksimal ukuran file: 10MB</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}

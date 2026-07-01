"use client";

import React, { useState } from "react";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { CenteredLoading } from "@/components/loading";
import { FileText, ShieldCheck, FileUp, AlertCircle } from "lucide-react";
import { useBuyerDocuments } from "../hooks/useBuyerProfile";
import { toast } from "sonner";

export function BuyerDocumentsPage() {
  const { documents, isLoading, uploadRawFile, uploadDocument, isUploadingFile, isUploadingDocument } = useBuyerDocuments();

  const [docType, setDocType] = useState("Nomor Induk Berusaha (NIB)");
  const [docNumber, setDocNumber] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const file = e.target.files[0];
      if (file.size > 10 * 1024 * 1024) {
        toast.error("Ukuran file maksimal 10MB!");
        return;
      }
      setSelectedFile(file);
    }
  };

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!docNumber) {
      toast.error("Harap isi nomor dokumen!");
      return;
    }
    if (!selectedFile) {
      toast.error("Harap pilih file dokumen!");
      return;
    }

    try {
      const fileRes = await uploadRawFile(selectedFile);
      
      uploadDocument({
        document_type: docType,
        document_number: docNumber,
        file_url: fileRes.url,
      });

      setDocNumber("");
      setSelectedFile(null);
    } catch (err) {
      console.error(err);
    }
  };

  if (isLoading) {
    return (
      <BuyerLayout>
        <CenteredLoading />
      </BuyerLayout>
    );
  }

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
              {documents.length === 0 ? (
                <p className="text-sm text-muted-foreground py-4 text-center">Belum ada dokumen yang diupload.</p>
              ) : (
                documents.map((doc, idx) => (
                  <div key={doc.id || idx} className="py-4 first:pt-0 last:pb-0 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                    <div className="flex items-start gap-3">
                      <div className="p-2 bg-primary/10 text-primary rounded-lg mt-0.5">
                        <FileText className="h-5 w-5" />
                      </div>
                      <div className="space-y-0.5">
                        <h4 className="text-sm font-semibold text-foreground">{doc.document_type}</h4>
                        <p className="text-xs text-muted-foreground">ID: {doc.document_number} • Diupload pada {doc.created_at ? new Date(doc.created_at).toLocaleDateString("id-ID") : "-"}</p>
                        {doc.review_reason && (
                          <p className="text-xs text-destructive flex items-center gap-1 mt-1">
                            <AlertCircle className="h-3 w-3" /> Alasan penolakan: {doc.review_reason}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center gap-3 self-end sm:self-center">
                      <Badge
                        variant="outline"
                        className={
                          doc.status === "verified"
                            ? "bg-success/10 text-success border-success/20 rounded-full text-[10px]"
                            : doc.status === "pending"
                            ? "bg-warning/10 text-warning border-warning/20 rounded-full text-[10px]"
                            : "bg-destructive/10 text-destructive border-destructive/20 rounded-full text-[10px]"
                        }
                      >
                        <ShieldCheck className="h-3 w-3" /> {doc.status.toUpperCase()}
                      </Badge>
                    </div>
                  </div>
                ))
              )}
            </div>

            {/* Upload Area */}
            <form onSubmit={handleUpload} className="pt-6 border-t border-border space-y-4">
              <h4 className="text-sm font-semibold text-foreground">Upload Dokumen Baru (NIB, SIUP, NPWP, Akta, Rekening Koran)</h4>
              
              <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Field className="space-y-1">
                  <FieldLabel htmlFor="doc_type">Jenis Dokumen</FieldLabel>
                  <select
                    id="doc_type"
                    value={docType}
                    onChange={(e) => setDocType(e.target.value)}
                    className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden cursor-pointer"
                  >
                    <option value="Nomor Induk Berusaha (NIB)">Nomor Induk Berusaha (NIB)</option>
                    <option value="Surat Izin Usaha Perdagangan (SIUP)">Surat Izin Usaha Perdagangan (SIUP)</option>
                    <option value="NPWP Perusahaan">NPWP Perusahaan</option>
                    <option value="Akta Pendirian Perusahaan">Akta Pendirian Perusahaan</option>
                    <option value="Rekening Koran 3 Bulan Terakhir">Rekening Koran 3 Bulan Terakhir</option>
                  </select>
                </Field>

                <Field className="space-y-1">
                  <FieldLabel htmlFor="doc_number">Nomor Dokumen</FieldLabel>
                  <Input
                    id="doc_number"
                    value={docNumber}
                    onChange={(e) => setDocNumber(e.target.value)}
                    placeholder="Contoh: 912048123"
                    className="cursor-pointer"
                  />
                </Field>
              </FieldGroup>

              <Field className="space-y-2">
                <FieldLabel>Pilih Berkas Dokumen (PDF / Image)</FieldLabel>
                <div className="relative border border-dashed border-border hover:border-primary/50 transition-colors rounded-lg p-6 text-center cursor-pointer space-y-2">
                  <input
                    type="file"
                    accept=".pdf,image/*"
                    onChange={handleFileChange}
                    className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                  />
                  <FileUp className="mx-auto h-8 w-8 text-muted-foreground opacity-60" />
                  <p className="text-xs font-semibold text-foreground">
                    {selectedFile ? `Berkas terpilih: ${selectedFile.name}` : "Klik untuk upload atau drag & drop file PDF"}
                  </p>
                  <p className="text-[10px] text-muted-foreground">Maksimal ukuran file: 10MB</p>
                </div>
              </Field>

              <div className="flex justify-end pt-2">
                <Button
                  type="submit"
                  disabled={isUploadingFile || isUploadingDocument}
                  className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer px-6 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20 font-semibold"
                >
                  {isUploadingFile || isUploadingDocument ? "Mengunggah..." : "Ajukan Verifikasi"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}

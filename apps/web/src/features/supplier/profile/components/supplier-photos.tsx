"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { Image as ImageIcon, Upload, Trash2, Eye } from "lucide-react";
import { DeleteDialog } from "@/components/ui/delete-dialog";

export function SupplierPhotos() {
  const t = useTranslations("supplier.profile");
  const [photos, setPhotos] = useState([
    { id: "PH-01", name: "Main Factory Production Line.jpg", url: "https://images.unsplash.com/photo-1581091226825-a6a2a5aee158?w=500&auto=format&fit=crop" },
    { id: "PH-02", name: "Bulk Raw Materials Warehouse.jpg", url: "https://images.unsplash.com/photo-1586528116311-ad8dd3c8310d?w=500&auto=format&fit=crop" },
    { id: "PH-03", name: "Office Headquarters Front Lobby.jpg", url: "https://images.unsplash.com/photo-1497366216548-37526070297c?w=500&auto=format&fit=crop" },
  ]);

  const [deleteId, setDeleteId] = useState<string | null>(null);

  const handleUploadClick = () => {
    // Simulate photo upload
    const mockPhoto = {
      id: `PH-0${photos.length + 1}`,
      name: `Site Photo ${photos.length + 1}.jpg`,
      url: "https://images.unsplash.com/photo-1565793298595-6a879b1d9492?w=500&auto=format&fit=crop",
    };
    setPhotos([...photos, mockPhoto]);
    toast.success("Site photo uploaded successfully!");
  };

  const handleDelete = (id: string) => {
    setDeleteId(id);
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {t("photosTitle")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("photosSubtitle")}
          </p>
        </div>
        <Button onClick={handleUploadClick} className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20 font-semibold">
          <PlusIcon className="mr-2 h-4 w-4" /> {t("btnUploadPhoto")}
        </Button>
      </div>

      {/* Grid List */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6">
        {/* Upload Card Trigger */}
        <div
          onClick={handleUploadClick}
          className="border-2 border-dashed border-border hover:border-primary/50 transition-all rounded-xl p-8 flex flex-col items-center justify-center gap-3 cursor-pointer bg-card hover:bg-muted/5 min-h-[220px]"
        >
          <div className="h-12 w-12 bg-primary/10 text-primary rounded-full flex items-center justify-center border border-primary/20">
            <Upload className="h-6 w-6" />
          </div>
          <div className="text-center">
            <span className="text-xs font-bold text-foreground block">Drag or Click to Upload</span>
            <span className="text-[10px] text-muted-foreground mt-1 block">Supports PNG, JPG up to 10MB</span>
          </div>
        </div>

        {/* Photos list */}
        {photos.map((p) => (
          <Card key={p.id} className="border border-border shadow-xs rounded-xl overflow-hidden bg-card group flex flex-col justify-between">
            <div className="aspect-video relative overflow-hidden bg-muted/40 border-b border-border flex items-center justify-center">
              {/* Image simulation */}
              <div className="absolute inset-0 bg-cover bg-center transition-transform duration-300 group-hover:scale-105" style={{ backgroundImage: `url(${p.url})` }} />
              <div className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                <Button size="icon" variant="secondary" className="h-8 w-8 cursor-pointer rounded-full" onClick={() => window.open(p.url, "_blank")}>
                  <Eye className="h-4 w-4" />
                </Button>
                <Button size="icon" variant="destructive" className="h-8 w-8 cursor-pointer rounded-full" onClick={() => handleDelete(p.id)}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
            </div>
            <CardContent className="p-3.5 flex items-center justify-between gap-2">
              <span className="text-xs font-semibold truncate text-foreground flex items-center gap-1.5">
                <ImageIcon className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                {p.name}
              </span>
              <span className="text-[9px] font-bold uppercase text-muted-foreground bg-muted px-1.5 py-0.5 rounded-sm shrink-0">
                {p.id}
              </span>
            </CardContent>
          </Card>
        ))}
      </div>

      <DeleteDialog
        open={!!deleteId}
        onOpenChange={(open) => !open && setDeleteId(null)}
        onConfirm={() => {
          if (deleteId) {
            setPhotos(photos.filter((p) => p.id !== deleteId));
            toast.success("Photo removed from gallery.");
          }
        }}
        itemName="photo"
      />
    </div>
  );
}

function PlusIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg
      {...props}
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M5 12h14" />
      <path d="M12 5v14" />
    </svg>
  );
}

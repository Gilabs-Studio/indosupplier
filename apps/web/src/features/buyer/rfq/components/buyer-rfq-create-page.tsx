"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { BuyerLayout } from "../../components/buyer-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Link } from "@/i18n/routing";
import { Field, FieldLabel, FieldGroup, FieldError } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { rfqSchema, type RfqFormData } from "../schemas/rfq.schema";
import { useCreateRfq } from "../hooks/useBuyerRfqs";
import { FileUp, ArrowLeft, Info } from "lucide-react";
import { toast } from "sonner";

export function BuyerRfqCreatePage() {
  const t = useTranslations("buyerRfq.rfqCreate");
  const { mutate: createRfq, isPending: isCreating } = useCreateRfq();

  const [isUploading, setIsUploading] = useState(false);
  const [uploadedFileName, setUploadedFileName] = useState("");
  const [uploadedFileSize, setUploadedFileSize] = useState(0);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<RfqFormData>({
    resolver: zodResolver(rfqSchema),
    defaultValues: {
      product_name: "",
      category: "manufacturing",
      quantity: "",
      unit: "Ton",
      target_port: "",
      description: "",
      attachment_url: "",
    },
  });

  const attachmentUrl = watch("attachment_url");

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const file = e.target.files[0];
      if (file.size > 10 * 1024 * 1024) {
        toast.error(t("toastSizeLimit"));
        return;
      }

      try {
        setIsUploading(true);
        const { rfqService } = await import("../services/rfq.service");
        const res = await rfqService.uploadSpecFile(file);
        
        setValue("attachment_url", res.url);
        setUploadedFileName(file.name);
        setUploadedFileSize(file.size);
        toast.success(t("toastUploadSuccess"));
      } catch (err) {
        console.error(err);
        toast.error(t("toastUploadError"));
      } finally {
        setIsUploading(false);
      }
    }
  };

  const onSubmit = (data: RfqFormData) => {
    createRfq({
      product_name: data.product_name,
      category: data.category,
      quantity: data.quantity,
      unit: data.unit,
      target_port: data.target_port,
      description: data.description,
      attachment_url: data.attachment_url,
      attachment_name: uploadedFileName || undefined,
      attachment_size: uploadedFileSize || undefined,
    });
  };

  return (
    <BuyerLayout>
      <div className="space-y-6 max-w-3xl">
        {/* Header */}
        <div className="space-y-2">
          <Link href="/rfq" className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:underline cursor-pointer">
            <ArrowLeft className="h-3.5 w-3.5" /> {t("backLink")}
          </Link>
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {/* Info Banner */}
        <div className="bg-primary/5 border border-primary/20 rounded-xl p-4 flex gap-3 text-sm text-foreground">
          <Info className="h-5 w-5 text-primary shrink-0 mt-0.5" />
          <p>
            <strong>{t("tipsTitle")}:</strong> {t("tipsDesc")}
          </p>
        </div>

        {/* Form Card */}
        <Card className="border border-border rounded-xl bg-card shadow-xs overflow-hidden">
          <CardContent className="p-6">
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              <FieldGroup className="space-y-5">
                {/* Product Name */}
                <Field className="space-y-2">
                  <FieldLabel htmlFor="product_name">
                    {t("labelProduct")} <span className="text-destructive">*</span>
                  </FieldLabel>
                  <Input
                    id="product_name"
                    placeholder={t("placeholderProduct")}
                    {...register("product_name")}
                    className="cursor-pointer"
                  />
                  {errors.product_name && (
                    <FieldError>{errors.product_name.message}</FieldError>
                  )}
                </Field>

                {/* Category & Unit */}
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="category">{t("labelCategory")}</FieldLabel>
                    <select
                      id="category"
                      {...register("category")}
                      className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden cursor-pointer"
                    >
                      <option value="manufacturing">{t("catManufacturing")}</option>
                      <option value="agriculture">{t("catAgriculture")}</option>
                      <option value="textile">{t("catTextile")}</option>
                      <option value="furniture">{t("catFurniture")}</option>
                    </select>
                    {errors.category && (
                      <FieldError>{errors.category.message}</FieldError>
                    )}
                  </Field>

                  <div className="grid grid-cols-2 gap-2">
                    <Field className="space-y-2">
                      <FieldLabel htmlFor="quantity">
                        {t("labelVolume")} <span className="text-destructive">*</span>
                      </FieldLabel>
                      <Input
                        id="quantity"
                        placeholder="Qty"
                        {...register("quantity")}
                        className="cursor-pointer"
                      />
                      {errors.quantity && (
                        <FieldError>{errors.quantity.message}</FieldError>
                      )}
                    </Field>
                    <Field className="space-y-2">
                      <FieldLabel htmlFor="unit">{t("labelUnit")}</FieldLabel>
                      <select
                        id="unit"
                        {...register("unit")}
                        className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-hidden cursor-pointer"
                      >
                        <option value="Ton">Ton</option>
                        <option value="Kg">Kg</option>
                        <option value="Pcs">Pcs</option>
                        <option value="Container">20ft Container</option>
                      </select>
                      {errors.unit && (
                        <FieldError>{errors.unit.message}</FieldError>
                      )}
                    </Field>
                  </div>
                </div>

                {/* Shipping Destination */}
                <Field className="space-y-2">
                  <FieldLabel htmlFor="target_port">
                    {t("labelDestination")} <span className="text-destructive">*</span>
                  </FieldLabel>
                  <Input
                    id="target_port"
                    placeholder={t("placeholderDestination")}
                    {...register("target_port")}
                    className="cursor-pointer"
                  />
                  {errors.target_port && (
                    <FieldError>{errors.target_port.message}</FieldError>
                  )}
                </Field>

                {/* Description Requirements */}
                <Field className="space-y-2">
                  <FieldLabel htmlFor="description">{t("labelDescription")}</FieldLabel>
                  <textarea
                    id="description"
                    rows={4}
                    placeholder={t("placeholderDescription")}
                    {...register("description")}
                    className="w-full rounded-lg border border-border bg-card px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-hidden"
                  />
                  {errors.description && (
                    <FieldError>{errors.description.message}</FieldError>
                  )}
                </Field>

                {/* File Attachment */}
                <Field className="space-y-2">
                  <FieldLabel>{t("labelAttachment")}</FieldLabel>
                  <div className="relative border border-dashed border-border hover:border-primary/50 transition-colors rounded-lg p-6 text-center cursor-pointer space-y-2">
                    <input
                      type="file"
                      accept=".pdf,.xlsx,.xls,image/*"
                      onChange={handleFileChange}
                      className="absolute inset-0 w-full h-full opacity-0 cursor-pointer"
                    />
                    <FileUp className="mx-auto h-8 w-8 text-muted-foreground opacity-60" />
                    <p className="text-xs font-semibold text-foreground">
                      {isUploading
                        ? t("toastUploading")
                        : attachmentUrl
                        ? t("toastUploadedFile", { name: uploadedFileName || "spesifikasi.pdf" })
                        : t("uploadPlaceholder")}
                    </p>
                    <p className="text-[10px] text-muted-foreground">{t("uploadLimit")}</p>
                  </div>
                </Field>
              </FieldGroup>

              {/* Actions */}
              <div className="pt-4 flex items-center justify-end gap-3 border-t border-border">
                <Button asChild variant="outline" className="cursor-pointer transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-xs">
                  <Link href="/rfq">{t("btnCancel")}</Link>
                </Button>
                <Button type="submit" disabled={isCreating || isUploading} className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer px-6 transition-all hover:-translate-y-0.5 active:translate-y-0 shadow-lg hover:shadow-primary/20">
                  {isCreating ? t("toastSubmitting") : t("btnSubmit")}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      </div>
    </BuyerLayout>
  );
}

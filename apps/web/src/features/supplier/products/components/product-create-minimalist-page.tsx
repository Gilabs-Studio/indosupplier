"use client";

import React, { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldError } from "@/components/ui/field";
import { Switch } from "@/components/ui/switch";
import { NumericInput } from "@/components/ui/numeric-input";
import { useForm, useFieldArray, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { productFormSchema, CURRENCY_OPTIONS, type ProductFormValues } from "../schemas/products.schema";
import {
  useCategories,
  useCreateProduct,
  useUploadProductImage,
} from "../hooks/useProducts";
import { MobileMenuButton } from "@/features/supplier/layout/components/supplier-layout";
import {
  ArrowLeft,
  Upload,
  Star,
  Loader2,
  Check,
  ChevronRight,
  ImagePlus,
  FileText,
  DollarSign,
  Sparkles,
  X,
  AlertCircle,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

/* ─────────────────────────────────────────────
   Step definition
───────────────────────────────────────────── */
const STEPS = [
  { id: "info", label: "Info Produk", icon: FileText },
  { id: "price", label: "Harga & Order", icon: DollarSign },
  { id: "photos", label: "Foto", icon: ImagePlus },
  { id: "settings", label: "Pengaturan", icon: Sparkles },
] as const;

type StepId = (typeof STEPS)[number]["id"];

export function ProductCreateMinimalistPage() {
  const router = useRouter();
  const t = useTranslations("supplier.products");
  const [currentStep, setCurrentStep] = useState<number>(0);
  const [submitting, setSubmitting] = useState(false);
  const [previewMap, setPreviewMap] = useState<Record<string, string>>({});

  const { data: categories } = useCategories();
  const { mutate: createProduct, isPending: isCreating } = useCreateProduct();
  const { mutateAsync: uploadImage, isPending: isUploading } = useUploadProductImage();

  const getStepLabel = (id: string) => {
    switch (id) {
      case "info":
        return t("stepInfo");
      case "price":
        return t("stepPrice");
      case "photos":
        return t("stepPhotos");
      case "settings":
        return t("stepSettings");
      default:
        return "";
    }
  };

  const {
    register,
    handleSubmit,
    control,
    watch,
    trigger,
    formState: { errors },
  } = useForm<ProductFormValues>({
    resolver: zodResolver(productFormSchema),
    defaultValues: {
      name: "",
      category_id: "",
      description: "",
      moq: "1",
      starting_price: 0,
      currency: "IDR",
      capacity_text: "",
      is_featured: false,
      sort_order: 0,
      photos: [],
    },
  });

  const {
    fields: photoFields,
    append: appendPhoto,
    remove: removePhoto,
  } = useFieldArray({ control, name: "photos" });

  const watchPhotos = watch("photos") || [];
  const watchIsFeatured = watch("is_featured");
  const watchName = watch("name");

  /* Step field groups for validation */
  const stepFields: Record<StepId, (keyof ProductFormValues)[]> = {
    info: ["name", "category_id", "description"],
    price: ["starting_price", "currency", "moq", "capacity_text"],
    photos: ["photos"],
    settings: [],
  };

  const goNext = async (e?: React.MouseEvent) => {
    e?.preventDefault();
    const stepId = STEPS[currentStep].id;
    if (stepId === "photos") {
      if (photoFields.length === 0) {
        toast.error(t("photoRequired"));
        return;
      }
    }
    const fieldList = stepFields[stepId];
    const valid = fieldList.length > 0 ? await trigger(fieldList) : true;
    if (valid && currentStep < STEPS.length - 1) {
      setCurrentStep((s) => s + 1);
    }
  };

  const goPrev = (e?: React.MouseEvent) => {
    e?.preventDefault();
    if (currentStep > 0) setCurrentStep((s) => s - 1);
  };

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files || files.length === 0) return;

    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      const localBlobUrl = URL.createObjectURL(file);

      try {
        const result = await uploadImage(file);
        setPreviewMap((prev) => ({
          ...prev,
          [result.url]: localBlobUrl,
        }));
        appendPhoto({
          file_url: result.url,
          caption: file.name,
          sort_order: watchPhotos.length,
        });
      } catch (err) {
        console.error("Upload failed:", err);
        toast.error("Gagal mengunggah foto.");
      }
    }
    e.target.value = "";
  };

  const onFormError = (formErrors: Partial<Record<keyof ProductFormValues, unknown>>) => {
    if (formErrors.photos || watchPhotos.length === 0) {
      toast.error(t("photoRequired"));
      setCurrentStep(2);
      return;
    }
    toast.error("Mohon lengkapi semua data wajib sebelum menyimpan.");
  };

  const onSubmit = (values: ProductFormValues) => {
    if (!values.photos || values.photos.length === 0) {
      toast.error(t("photoRequired"));
      setCurrentStep(2);
      return;
    }
    setSubmitting(true);
    createProduct(values, {
      onSuccess: () => router.push("/supplier/products"),
      onError: () => setSubmitting(false),
    });
  };

  const isLastStep = currentStep === STEPS.length - 1;
  const progress = ((currentStep + 1) / STEPS.length) * 100;

  return (
    <form
      onSubmit={handleSubmit(onSubmit, onFormError)}
      className="flex-1 flex flex-col min-w-0 h-screen overflow-hidden bg-background text-foreground relative"
    >
      {/* ── Progressive Header (Replacing layout header) ────────────────── */}
      <header className="shrink-0 z-20 bg-background/95 backdrop-blur h-16 w-full border-b border-border flex items-center justify-between px-4 md:px-6 relative select-none">
        <div className="flex items-center gap-3 min-w-0">
          <MobileMenuButton />

          <button
            type="button"
            onClick={() => router.push("/supplier/products")}
            className="p-1.5 rounded-lg border border-border flex items-center justify-center text-muted-foreground hover:text-foreground hover:bg-muted/40 transition-all cursor-pointer shrink-0"
            aria-label="Back to products"
          >
            <ArrowLeft className="h-4 w-4" />
          </button>

          <div className="flex flex-col min-w-0">
            <h1 className="text-sm font-extrabold text-foreground truncate leading-none">
              {watchName.trim() || t("addProduct")}
            </h1>
            <p className="text-[11px] text-muted-foreground mt-1 leading-none font-medium truncate">
              {t("stepOf", {
                current: currentStep + 1,
                total: STEPS.length,
                label: getStepLabel(STEPS[currentStep].id),
              })}
            </p>
          </div>
        </div>

        {/* Interactive Step Pills */}
        <div className="flex items-center gap-1.5 sm:gap-2 shrink-0">
          {STEPS.map((s, i) => {
            const isCurrent = i === currentStep;
            const isPassed = i < currentStep;
            const Icon = s.icon;
            return (
              <button
                key={s.id}
                type="button"
                onClick={async () => {
                  if (i < currentStep) {
                    setCurrentStep(i);
                  } else if (i === currentStep + 1) {
                    await goNext();
                  }
                }}
                className={`flex items-center gap-1.5 px-2 sm:px-3 py-1.5 rounded-lg text-xs font-semibold transition-all cursor-pointer ${
                  isCurrent
                    ? "bg-primary text-primary-foreground shadow-xs shadow-primary/25"
                    : isPassed
                    ? "bg-muted text-foreground hover:bg-muted/80"
                    : "text-muted-foreground/60 hover:text-muted-foreground hover:bg-muted/30"
                }`}
              >
                <Icon className="h-3.5 w-3.5 shrink-0" />
                <span className="hidden md:inline">{getStepLabel(s.id)}</span>
                <span className="md:hidden">{i + 1}</span>
              </button>
            );
          })}
        </div>

        {/* Animated Progress Bar at bottom border */}
        <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-border overflow-hidden">
          <motion.div
            className="h-full bg-primary"
            animate={{ width: `${progress}%` }}
            transition={{ duration: 0.35, ease: "easeInOut" }}
          />
        </div>
      </header>

      {/* ── Scrollable Form Body ────────────────────────────────────────── */}
      <main className="flex-1 min-h-0 overflow-y-auto px-4 md:px-6 py-8">
        <div className="max-w-2xl mx-auto space-y-6">
          <AnimatePresence mode="wait">
            {/* ── STEP 0: Info Produk ──────────────────────────────────── */}
            {currentStep === 0 && (
              <StepWrapper key="step-info">
                <StepHeader
                  title={t("productInfoTitle")}
                  description={t("productInfoDesc")}
                />

                <div className="space-y-6">
                  <MinimalField label={t("productName")} required error={errors.name?.message}>
                    <Input
                      id="name"
                      placeholder={t("namePlaceholder")}
                      className="h-11 text-sm border-border focus:ring-2 focus:ring-primary/15"
                      {...register("name")}
                    />
                  </MinimalField>

                  <MinimalField label={t("category")} required error={errors.category_id?.message}>
                    <select
                      id="category_id"
                      className="w-full h-11 px-3 bg-background border border-border text-sm rounded-lg outline-hidden focus:border-primary focus:ring-2 focus:ring-primary/15 transition-all cursor-pointer text-foreground"
                      {...register("category_id")}
                    >
                      <option value="">{t("selectCategory")}</option>
                      {categories?.map((cat) => (
                        <option key={cat.id} value={cat.id}>
                          {cat.name}
                        </option>
                      ))}
                    </select>
                  </MinimalField>

                  <MinimalField label={t("description")} error={errors.description?.message}>
                    <Textarea
                      id="description"
                      placeholder={t("descriptionPlaceholder")}
                      rows={6}
                      className="resize-none text-sm border-border focus:ring-2 focus:ring-primary/15"
                      {...register("description")}
                    />
                  </MinimalField>
                </div>
              </StepWrapper>
            )}

            {/* ── STEP 1: Harga & Order ────────────────────────────────── */}
            {currentStep === 1 && (
              <StepWrapper key="step-price">
                <StepHeader
                  title={t("pricingTerms")}
                  description={t("pricingTermsDesc")}
                />

                <div className="space-y-6">
                  <div className="grid grid-cols-3 gap-3">
                    <div className="col-span-2">
                      <MinimalField label={t("price")} required error={errors.starting_price?.message}>
                        <Controller
                          control={control}
                          name="starting_price"
                          render={({ field: { value, onChange, onBlur, ref } }) => (
                            <NumericInput
                              id="starting_price"
                              value={value as number | undefined}
                              onChange={onChange}
                              onBlur={onBlur}
                              ref={ref}
                              placeholder={t("pricePlaceholder")}
                              className="h-11 border-border focus:ring-2 focus:ring-primary/15"
                            />
                          )}
                        />
                      </MinimalField>
                    </div>
                    <div className="col-span-1">
                      <MinimalField label={t("currency")} error={errors.currency?.message}>
                        <select
                          id="currency"
                          className="h-11 w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-medium text-foreground focus:outline-none focus:ring-2 focus:ring-primary/15 cursor-pointer"
                          {...register("currency")}
                        >
                          {CURRENCY_OPTIONS.map((opt) => (
                            <option key={opt.value} value={opt.value} className="bg-popover text-popover-foreground">
                              {opt.label}
                            </option>
                          ))}
                        </select>
                      </MinimalField>
                    </div>
                  </div>

                  <MinimalField label={t("moq")} required error={errors.moq?.message}>
                    <Input
                      id="moq"
                      placeholder={t("moqPlaceholder")}
                      className="h-11 text-sm border-border focus:ring-2 focus:ring-primary/15"
                      {...register("moq")}
                    />
                  </MinimalField>

                  <MinimalField label={t("capacity")} error={errors.capacity_text?.message}>
                    <Input
                      id="capacity_text"
                      placeholder={t("capacityPlaceholder")}
                      className="h-11 text-sm border-border focus:ring-2 focus:ring-primary/15"
                      {...register("capacity_text")}
                    />
                  </MinimalField>
                </div>
              </StepWrapper>
            )}

            {/* ── STEP 2: Foto ─────────────────────────────────────────── */}
            {currentStep === 2 && (
              <StepWrapper key="step-photos">
                <StepHeader
                  title={t("photos")}
                  description={t("photosDesc")}
                />

                {/* Upload Drop Zone */}
                <label className="block cursor-pointer">
                  <div
                    className={`border-2 border-dashed rounded-xl p-10 flex flex-col items-center justify-center gap-3 transition-all ${
                      isUploading
                        ? "border-primary/40 bg-primary/5"
                        : photoFields.length === 0 && errors.photos
                        ? "border-destructive/50 bg-destructive/5"
                        : "border-border hover:border-primary/50 hover:bg-muted/30"
                    }`}
                  >
                    {isUploading ? (
                      <>
                        <Loader2 className="h-6 w-6 animate-spin text-primary" />
                        <p className="text-sm font-medium text-primary">{t("saving")}</p>
                      </>
                    ) : (
                      <>
                        <div className="h-12 w-12 rounded-xl bg-primary/10 flex items-center justify-center">
                          <Upload className="h-5 w-5 text-primary" />
                        </div>
                        <div className="text-center">
                          <p className="text-sm font-semibold text-foreground">
                            {t("uploadClick")}
                          </p>
                          <p className="text-[11px] text-muted-foreground mt-0.5">
                            {t("uploadHint")}
                          </p>
                        </div>
                      </>
                    )}
                  </div>
                  <input
                    type="file"
                    multiple
                    accept="image/*"
                    onChange={handleImageUpload}
                    disabled={isUploading}
                    className="hidden"
                  />
                </label>

                {/* Validation Warning Alert */}
                {photoFields.length === 0 && (
                  <div className="p-3.5 rounded-lg border border-destructive/30 bg-destructive/5 flex items-center gap-2.5 text-destructive">
                    <AlertCircle className="h-4 w-4 shrink-0" />
                    <p className="text-xs font-medium">{t("photoRequired")}</p>
                  </div>
                )}

                {/* Photo Grid */}
                {photoFields.length > 0 && (
                  <div className="mt-6">
                    <p className="text-xs text-muted-foreground mb-3 font-medium">
                      {t("photosUploaded", { count: photoFields.length })}
                    </p>
                    <div className="grid grid-cols-3 gap-3">
                      <AnimatePresence>
                        {photoFields.map((field, index) => {
                          const isCover = index === 0;
                          const displaySrc = previewMap[field.file_url] || field.file_url;
                          return (
                            <motion.div
                              key={field.id}
                              initial={{ opacity: 0, scale: 0.85 }}
                              animate={{ opacity: 1, scale: 1 }}
                              exit={{ opacity: 0, scale: 0.85 }}
                              className="relative group rounded-lg overflow-hidden border border-border aspect-square bg-muted/20"
                            >
                              {/* eslint-disable-next-line @next/next/no-img-element */}
                              <img
                                src={displaySrc}
                                alt={`Photo ${index + 1}`}
                                className="w-full h-full object-cover"
                              />
                              {isCover && (
                                <div className="absolute top-1.5 left-1.5 bg-primary text-primary-foreground text-[9px] font-bold px-1.5 py-0.5 rounded">
                                  {t("coverLabel")}
                                </div>
                              )}
                              <button
                                type="button"
                                onClick={() => removePhoto(index)}
                                className="absolute top-1.5 right-1.5 h-6 w-6 rounded bg-background/90 border border-border flex items-center justify-center text-muted-foreground hover:text-destructive transition-all cursor-pointer opacity-0 group-hover:opacity-100"
                              >
                                <X className="h-3 w-3" />
                              </button>
                            </motion.div>
                          );
                        })}
                      </AnimatePresence>
                    </div>
                  </div>
                )}
              </StepWrapper>
            )}

            {/* ── STEP 3: Pengaturan ───────────────────────────────────── */}
            {currentStep === 3 && (
              <StepWrapper key="step-settings">
                <StepHeader
                  title={t("settingsTitle")}
                  description={t("settingsDesc")}
                />

                {/* Warning if photos missing */}
                {watchPhotos.length === 0 && (
                  <div className="rounded-xl border border-destructive/30 bg-destructive/5 p-4 flex items-center justify-between gap-4 text-destructive">
                    <div className="flex items-center gap-2.5">
                      <AlertCircle className="h-4.5 w-4.5 shrink-0" />
                      <p className="text-xs font-semibold">{t("photoRequiredDesc")}</p>
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => setCurrentStep(2)}
                      className="cursor-pointer border-destructive/40 text-destructive hover:bg-destructive/10 text-xs shrink-0"
                    >
                      {t("goToPhotos")}
                    </Button>
                  </div>
                )}

                {/* Featured toggle */}
                <div className="rounded-xl border border-border bg-card p-5">
                  <div className="flex items-center justify-between gap-6">
                    <div className="flex items-start gap-3">
                      <div
                        className={`h-9 w-9 rounded-lg flex items-center justify-center shrink-0 transition-all ${
                          watchIsFeatured
                            ? "bg-warning/15 text-warning"
                            : "bg-muted text-muted-foreground"
                        }`}
                      >
                        <Star
                          className={`h-4 w-4 transition-all ${
                            watchIsFeatured ? "fill-warning" : ""
                          }`}
                        />
                      </div>
                      <div>
                        <p className="text-sm font-semibold text-foreground">
                          {t("isFeatured")}
                        </p>
                        <p className="text-xs text-muted-foreground mt-0.5 leading-relaxed">
                          {t("featuredDesc")}
                        </p>
                      </div>
                    </div>
                    <Controller
                      control={control}
                      name="is_featured"
                      render={({ field: { value, onChange } }) => (
                        <Switch
                          checked={!!value}
                          onCheckedChange={onChange}
                          className="cursor-pointer shrink-0"
                        />
                      )}
                    />
                  </div>
                </div>

                {/* Summary Preview Card */}
                <div className="mt-6 rounded-xl border border-border bg-card overflow-hidden">
                  <div className="px-5 py-3 border-b border-border bg-muted/30">
                    <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
                      {t("summaryTitle")}
                    </p>
                  </div>
                  <SummaryRow label={t("summaryName")} value={watch("name") || "—"} />
                  <SummaryRow
                    label={t("summaryCategory")}
                    value={categories?.find((c) => c.id === watch("category_id"))?.name || "—"}
                  />
                  <SummaryRow
                    label={t("summaryPrice")}
                    value={
                      (watch("starting_price") as number) > 0
                        ? `${watch("currency")} ${Number(watch("starting_price") as number).toLocaleString("id-ID")}`
                        : "—"
                    }
                  />
                  <SummaryRow
                    label={t("summaryMOQ")}
                    value={watch("moq") || "—"}
                  />
                  <SummaryRow
                    label={t("summaryPhotos")}
                    value={t("photosUploaded", { count: photoFields.length })}
                  />
                </div>
              </StepWrapper>
            )}
          </AnimatePresence>
        </div>
      </main>

      {/* ── Pinned Bottom Footer Navigation ─────────────────────────────── */}
      <footer className="shrink-0 z-20 bg-background/95 backdrop-blur h-16 w-full border-t border-border flex items-center justify-between px-4 md:px-6">
        {currentStep > 0 ? (
          <Button
            type="button"
            variant="outline"
            onClick={(e) => goPrev(e)}
            className="cursor-pointer border-border hover:bg-muted text-xs sm:text-sm font-medium"
          >
            <ArrowLeft className="h-3.5 w-3.5 mr-1.5" />
            {t("back")}
          </Button>
        ) : (
          <Button
            type="button"
            variant="ghost"
            onClick={() => router.push("/supplier/products")}
            className="cursor-pointer text-muted-foreground hover:text-foreground text-xs sm:text-sm"
          >
            {t("cancel")}
          </Button>
        )}

        <div className="flex items-center gap-2">
          {isLastStep ? (
            <Button
              type="submit"
              disabled={isCreating || submitting}
              className="cursor-pointer bg-primary text-primary-foreground font-bold px-6 sm:px-7 py-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/25 text-xs sm:text-sm"
            >
              {isCreating || submitting ? (
                <>
                  <Loader2 className="h-3.5 w-3.5 animate-spin mr-1.5" />
                  {t("saving")}
                </>
              ) : (
                <>
                  <Check className="h-3.5 w-3.5 mr-1.5" />
                  {t("save")}
                </>
              )}
            </Button>
          ) : (
            <Button
              type="button"
              onClick={(e) => goNext(e)}
              className="cursor-pointer bg-primary text-primary-foreground font-bold px-5 sm:px-6 py-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/25 text-xs sm:text-sm"
            >
              {t("next")}
              <ChevronRight className="h-3.5 w-3.5 ml-1.5" />
            </Button>
          )}
        </div>
      </footer>
    </form>
  );
}

/* ── Reusable sub-components ─────────────────────────────────────────────── */

function StepWrapper({ children }: { children: React.ReactNode }) {
  return (
    <motion.div
      initial={{ opacity: 0, x: 20 }}
      animate={{ opacity: 1, x: 0 }}
      exit={{ opacity: 0, x: -20 }}
      transition={{ duration: 0.25, ease: "easeOut" }}
      className="space-y-6"
    >
      {children}
    </motion.div>
  );
}

function StepHeader({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="space-y-1 pb-3.5 border-b border-border">
      <h2 className="text-xl font-extrabold text-foreground tracking-tight">{title}</h2>
      <p className="text-sm text-muted-foreground">{description}</p>
    </div>
  );
}

function MinimalField({
  label,
  required = false,
  error,
  children,
}: {
  label: string;
  required?: boolean;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <Field className="space-y-1.5">
      <FieldLabel className="text-xs font-semibold text-foreground/80">
        {label}
        {required && <span className="text-primary ml-0.5">*</span>}
      </FieldLabel>
      {children}
      {error && <FieldError>{error}</FieldError>}
    </Field>
  );
}

function SummaryRow({
  label,
  value,
}: {
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-center justify-between px-5 py-3.5 border-b border-border last:border-0">
      <span className="text-sm text-muted-foreground font-medium">
        {label}
      </span>
      <span className="text-sm font-semibold text-foreground max-w-[60%] text-right truncate">
        {value}
      </span>
    </div>
  );
}

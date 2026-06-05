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
import { productFormSchema } from "../schemas/products.schema";
import {
  useCategories,
  useCreateProduct,
  useUploadProductImage,
} from "../hooks/useProducts";
import {
  ArrowLeft,
  Upload,
  Star,
  Loader2,
  Check,
  ChevronRight,
  ImagePlus,
  Tag,
  FileText,
  DollarSign,
  Package,
  Sparkles,
  X,
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

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
  const [currentStep, setCurrentStep] = useState<number>(0);
  const [submitting, setSubmitting] = useState(false);

  const { data: categories } = useCategories();
  const { mutate: createProduct, isPending: isCreating } = useCreateProduct();
  const { mutateAsync: uploadImage, isPending: isUploading } =
    useUploadProductImage();

  const {
    register,
    handleSubmit,
    control,
    watch,
    trigger,
    formState: { errors },
  } = useForm({
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
      photos: [] as {
        file_url: string;
        caption?: string;
        sort_order: number;
        id?: string;
      }[],
    },
  });

  const {
    fields: photoFields,
    append: appendPhoto,
    remove: removePhoto,
  } = useFieldArray({ control, name: "photos" });

  const watchPhotos = (watch("photos") || []) as {
    file_url: string;
    caption?: string;
    sort_order: number;
    id?: string;
  }[];
  const watchIsFeatured = watch("is_featured") as boolean;
  const watchName = watch("name") as string;

  /* Step field groups for validation */
  const stepFields: Record<StepId, string[]> = {
    info: ["name", "category_id", "description"],
    price: ["starting_price", "currency", "moq", "capacity_text"],
    photos: [],
    settings: [],
  };

  const goNext = async () => {
    const stepId = STEPS[currentStep].id;
    const fieldList = stepFields[stepId];
    const valid =
      fieldList.length > 0
        ? // eslint-disable-next-line @typescript-eslint/no-explicit-any
          await trigger(fieldList as any)
        : true;
    if (valid && currentStep < STEPS.length - 1) setCurrentStep((s) => s + 1);
  };

  const goPrev = () => {
    if (currentStep > 0) setCurrentStep((s) => s - 1);
  };

  const handleImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files) return;
    for (let i = 0; i < files.length; i++) {
      try {
        const result = await uploadImage(files[i]);
        appendPhoto({
          file_url: result.url,
          caption: files[i].name,
          sort_order: watchPhotos.length,
        });
      } catch (err) {
        console.error("Upload failed:", err);
      }
    }
    // reset input
    e.target.value = "";
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onSubmit = (values: any) => {
    setSubmitting(true);
    createProduct(values, {
      onSuccess: () => router.push("/supplier/products"),
      onError: () => setSubmitting(false),
    });
  };

  const isLastStep = currentStep === STEPS.length - 1;
  const progress = ((currentStep + 1) / STEPS.length) * 100;

  return (
    <div className="min-h-screen bg-background text-foreground">
      {/* ── Slim Top Bar ─────────────────────────────────────────────────── */}
      <div className="sticky top-0 z-40 bg-background/90 backdrop-blur-md border-b border-border/60">
        <div className="max-w-2xl mx-auto px-4 h-14 flex items-center gap-4">
          <button
            onClick={() => router.push("/supplier/products")}
            className="h-8 w-8 rounded-lg border border-border/60 flex items-center justify-center text-muted-foreground hover:text-foreground hover:border-border hover:bg-muted/40 transition-all cursor-pointer shrink-0"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
          </button>

          <div className="flex-1 min-w-0">
            <p className="text-xs font-bold text-foreground truncate">
              {watchName || "Produk Baru"}
            </p>
            <p className="text-[10px] text-muted-foreground">
              Langkah {currentStep + 1} dari {STEPS.length} —{" "}
              {STEPS[currentStep].label}
            </p>
          </div>

          {/* Step dots */}
          <div className="flex items-center gap-1.5 shrink-0">
            {STEPS.map((s, i) => (
              <button
                key={s.id}
                onClick={() => setCurrentStep(i)}
                className={`transition-all cursor-pointer rounded-full ${
                  i === currentStep
                    ? "h-2 w-6 bg-primary"
                    : i < currentStep
                    ? "h-2 w-2 bg-primary/40"
                    : "h-2 w-2 bg-border"
                }`}
              />
            ))}
          </div>
        </div>

        {/* Progress bar */}
        <div className="h-0.5 bg-border/40">
          <motion.div
            className="h-full bg-primary"
            animate={{ width: `${progress}%` }}
            transition={{ duration: 0.4, ease: "easeInOut" }}
          />
        </div>
      </div>

      {/* ── Form Content ─────────────────────────────────────────────────── */}
      <form onSubmit={handleSubmit(onSubmit)}>
        <div className="max-w-2xl mx-auto px-4 py-10">
          <AnimatePresence mode="wait">
            {/* ── STEP 0: Info Produk ──────────────────────────────────── */}
            {currentStep === 0 && (
              <StepWrapper key="step-info">
                <StepHeader
                  icon={<FileText className="h-5 w-5 text-primary" />}
                  title="Informasi Produk"
                  description="Lengkapi nama, kategori, dan deskripsi dasar produk."
                />

                <div className="space-y-6">
                  <MinimalField label="Nama Produk" required error={errors.name?.message as string}>
                    <Input
                      id="name"
                      placeholder="Contoh: Beras Premium Putih 25kg"
                      className="h-11 text-sm border-border/80 focus:ring-2 focus:ring-primary/15"
                      {...register("name")}
                    />
                  </MinimalField>

                  <MinimalField label="Kategori" required error={errors.category_id?.message as string}>
                    <select
                      id="category_id"
                      className="w-full h-11 px-3 bg-background border border-border/80 text-sm rounded-lg outline-hidden focus:border-primary focus:ring-2 focus:ring-primary/15 transition-all cursor-pointer text-foreground"
                      {...register("category_id")}
                    >
                      <option value="">Pilih kategori…</option>
                      {categories?.map((cat) => (
                        <option key={cat.id} value={cat.id}>
                          {cat.name}
                        </option>
                      ))}
                    </select>
                  </MinimalField>

                  <MinimalField label="Deskripsi" error={errors.description?.message as string}>
                    <Textarea
                      id="description"
                      placeholder="Jelaskan spesifikasi, standar ekspor, kemasan, dan keunggulan produk Anda…"
                      rows={6}
                      className="resize-none text-sm border-border/80 focus:ring-2 focus:ring-primary/15"
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
                  icon={<DollarSign className="h-5 w-5 text-primary" />}
                  title="Harga & Ketentuan Pesanan"
                  description="Tetapkan harga, mata uang, dan jumlah minimum order (MOQ)."
                />

                <div className="space-y-6">
                  {/* Price + Currency in one row */}
                  <div className="grid grid-cols-3 gap-3">
                    <div className="col-span-2">
                      <MinimalField label="Harga Mulai" required error={errors.starting_price?.message as string}>
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
                              placeholder="Contoh: 150.000"
                              className="h-11 border-border/80 focus:ring-2 focus:ring-primary/15"
                            />
                          )}
                        />
                      </MinimalField>
                    </div>
                    <div className="col-span-1">
                      <MinimalField label="Mata Uang" error={errors.currency?.message as string}>
                        <Input
                          id="currency"
                          placeholder="IDR"
                          className="h-11 text-sm border-border/80 focus:ring-2 focus:ring-primary/15"
                          {...register("currency")}
                        />
                      </MinimalField>
                    </div>
                  </div>

                  <MinimalField label="Min. Order (MOQ)" required error={errors.moq?.message as string}>
                    <Input
                      id="moq"
                      placeholder="Contoh: 20 Ton, 100 Karung…"
                      className="h-11 text-sm border-border/80 focus:ring-2 focus:ring-primary/15"
                      {...register("moq")}
                    />
                  </MinimalField>

                  <MinimalField label="Kapasitas Pasokan" error={errors.capacity_text?.message as string}>
                    <Input
                      id="capacity_text"
                      placeholder="Contoh: 500 Ton / Bulan"
                      className="h-11 text-sm border-border/80 focus:ring-2 focus:ring-primary/15"
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
                  icon={<ImagePlus className="h-5 w-5 text-primary" />}
                  title="Foto Produk"
                  description="Tambahkan foto yang menarik. Foto pertama akan menjadi cover utama."
                />

                {/* Upload Drop Zone */}
                <label className="block cursor-pointer">
                  <div
                    className={`border-2 border-dashed rounded-xl p-10 flex flex-col items-center justify-center gap-3 transition-all ${
                      isUploading
                        ? "border-primary/40 bg-primary/5"
                        : "border-border/60 hover:border-primary/50 hover:bg-muted/30"
                    }`}
                  >
                    {isUploading ? (
                      <>
                        <Loader2 className="h-6 w-6 animate-spin text-primary" />
                        <p className="text-sm font-medium text-primary">Mengunggah…</p>
                      </>
                    ) : (
                      <>
                        <div className="h-12 w-12 rounded-xl bg-primary/10 flex items-center justify-center">
                          <Upload className="h-5 w-5 text-primary" />
                        </div>
                        <div className="text-center">
                          <p className="text-sm font-semibold text-foreground">
                            Klik untuk upload foto
                          </p>
                          <p className="text-[11px] text-muted-foreground mt-0.5">
                            PNG, JPG, WebP — maks. 5MB per file
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

                {/* Photo Grid */}
                {photoFields.length > 0 && (
                  <div className="mt-6">
                    <p className="text-xs text-muted-foreground mb-3 font-medium">
                      {photoFields.length} foto diunggah
                    </p>
                    <div className="grid grid-cols-3 gap-3">
                      <AnimatePresence>
                        {photoFields.map((field, index) => {
                          const isCover = index === 0;
                          return (
                            <motion.div
                              key={field.id}
                              initial={{ opacity: 0, scale: 0.85 }}
                              animate={{ opacity: 1, scale: 1 }}
                              exit={{ opacity: 0, scale: 0.85 }}
                              className="relative group rounded-lg overflow-hidden border border-border/80 aspect-square bg-muted/10"
                            >
                              {/* eslint-disable-next-line @next/next/no-img-element */}
                              <img
                                src={field.file_url}
                                alt={`Photo ${index + 1}`}
                                className="w-full h-full object-cover"
                              />
                              {isCover && (
                                <div className="absolute top-1.5 left-1.5 bg-primary text-primary-foreground text-[9px] font-bold px-1.5 py-0.5 rounded">
                                  Cover
                                </div>
                              )}
                              <button
                                type="button"
                                onClick={() => removePhoto(index)}
                                className="absolute top-1.5 right-1.5 h-6 w-6 rounded bg-background/90 border border-border/60 flex items-center justify-center text-muted-foreground hover:text-destructive transition-all cursor-pointer opacity-0 group-hover:opacity-100"
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
                  icon={<Sparkles className="h-5 w-5 text-primary" />}
                  title="Pengaturan Produk"
                  description="Atur visibilitas dan prioritas tampil produk Anda."
                />

                {/* Featured toggle */}
                <div className="rounded-xl border border-border/80 bg-card p-5">
                  <div className="flex items-center justify-between gap-6">
                    <div className="flex items-start gap-3">
                      <div
                        className={`h-9 w-9 rounded-lg flex items-center justify-center shrink-0 transition-all ${
                          watchIsFeatured
                            ? "bg-amber-500/15 text-amber-500"
                            : "bg-muted text-muted-foreground"
                        }`}
                      >
                        <Star
                          className={`h-4 w-4 transition-all ${
                            watchIsFeatured ? "fill-amber-500" : ""
                          }`}
                        />
                      </div>
                      <div>
                        <p className="text-sm font-semibold text-foreground">
                          Produk Unggulan
                        </p>
                        <p className="text-xs text-muted-foreground mt-0.5 leading-relaxed">
                          Tampilkan produk ini di posisi teratas profil supplier Anda untuk
                          meningkatkan visibilitas kepada pembeli potensial.
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
                <div className="mt-6 rounded-xl border border-border/60 bg-muted/20 overflow-hidden">
                  <div className="px-5 py-3 border-b border-border/60">
                    <p className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
                      Ringkasan Produk
                    </p>
                  </div>
                  <SummaryRow icon={<Tag className="h-3 w-3" />} label="Nama" value={watch("name") || "—"} />
                  <SummaryRow
                    icon={<Package className="h-3 w-3" />}
                    label="Kategori"
                    value={categories?.find((c) => c.id === watch("category_id"))?.name || "—"}
                  />
                  <SummaryRow
                    icon={<DollarSign className="h-3 w-3" />}
                    label="Harga Mulai"
                    value={
                      (watch("starting_price") as number) > 0
                        ? `${watch("currency")} ${Number(watch("starting_price") as number).toLocaleString("id-ID")}`
                        : "—"
                    }
                  />
                  <SummaryRow
                    icon={<FileText className="h-3 w-3" />}
                    label="MOQ"
                    value={watch("moq") || "—"}
                  />
                  <SummaryRow
                    icon={<ImagePlus className="h-3 w-3" />}
                    label="Foto"
                    value={`${photoFields.length} foto`}
                  />
                </div>
              </StepWrapper>
            )}
          </AnimatePresence>
        </div>

        {/* ── Sticky Navigation ──────────────────────────────────────────── */}
        <div className="sticky bottom-0 z-40 bg-background/90 backdrop-blur-md border-t border-border/60">
          <div className="max-w-2xl mx-auto px-4 py-4 flex items-center gap-3">
            {currentStep > 0 ? (
              <Button
                type="button"
                variant="outline"
                onClick={goPrev}
                className="cursor-pointer border-border/80 hover:bg-muted text-sm font-medium"
              >
                <ArrowLeft className="h-3.5 w-3.5 mr-1.5" />
                Kembali
              </Button>
            ) : (
              <Button
                type="button"
                variant="ghost"
                onClick={() => router.push("/supplier/products")}
                className="cursor-pointer text-muted-foreground hover:text-foreground text-sm"
              >
                Batal
              </Button>
            )}

            <div className="flex-1" />

            {isLastStep ? (
              <Button
                type="submit"
                disabled={isCreating || submitting}
                className="cursor-pointer bg-primary text-primary-foreground font-bold px-7 py-2.5 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/25 text-sm"
              >
                {isCreating || submitting ? (
                  <>
                    <Loader2 className="h-3.5 w-3.5 animate-spin mr-2" />
                    Menyimpan…
                  </>
                ) : (
                  <>
                    <Check className="h-3.5 w-3.5 mr-2" />
                    Simpan Produk
                  </>
                )}
              </Button>
            ) : (
              <Button
                type="button"
                onClick={goNext}
                className="cursor-pointer bg-primary text-primary-foreground font-bold px-6 py-2.5 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/25 text-sm"
              >
                Lanjut
                <ChevronRight className="h-3.5 w-3.5 ml-1.5" />
              </Button>
            )}
          </div>
        </div>
      </form>
    </div>
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
      className="space-y-8"
    >
      {children}
    </motion.div>
  );
}

function StepHeader({
  icon,
  title,
  description,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
}) {
  return (
    <div className="space-y-1 pb-2 border-b border-border/60">
      <div className="flex items-center gap-2.5">
        {icon}
        <h2 className="text-xl font-extrabold text-foreground tracking-tight">{title}</h2>
      </div>
      <p className="text-sm text-muted-foreground pl-7">{description}</p>
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
  icon,
  label,
  value,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-center justify-between px-5 py-3 border-b border-border/40 last:border-0">
      <span className="flex items-center gap-2 text-xs text-muted-foreground">
        <span className="text-primary/60">{icon}</span>
        {label}
      </span>
      <span className="text-xs font-semibold text-foreground max-w-[60%] text-right truncate">
        {value}
      </span>
    </div>
  );
}

"use client";

import React, { useEffect } from "react";
import { useRouter } from "@/i18n/routing";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup, FieldError } from "@/components/ui/field";
import { Switch } from "@/components/ui/switch";
import { NumericInput } from "@/components/ui/numeric-input";
import { Badge } from "@/components/ui/badge";
import { useForm, useFieldArray, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { productFormSchema } from "../schemas/products.schema";
import { useSupplierProduct, useCategories, useCreateProduct, useUpdateProduct, useUploadProductImage } from "../hooks/useProducts";
import { ArrowLeft, Save, Upload, Trash2, Star, Loader2, Layers, ShoppingBag, Eye } from "lucide-react";
import { useTranslations } from "next-intl";
import { Skeleton } from "@/components/ui/skeleton";
import { motion, AnimatePresence } from "framer-motion";

interface ProductDetailPageProps {
  id?: string;
  isCreate?: boolean;
}

export function ProductDetailPage({ id, isCreate = false }: ProductDetailPageProps) {
  const router = useRouter();
  const t = useTranslations("supplier.products");

  const { data: product, isLoading: isLoadingProduct } = useSupplierProduct(id || "");
  const { data: categories } = useCategories();
  const { mutate: createProduct, isPending: isCreating } = useCreateProduct();
  const { mutate: updateProduct, isPending: isUpdating } = useUpdateProduct();
  const { mutateAsync: uploadImage, isPending: isUploading } = useUploadProductImage();

  const {
    register,
    handleSubmit,
    control,
    watch,
    reset,
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
      photos: [] as { file_url: string; caption?: string; sort_order: number; id?: string }[],
    },
  });

  const { fields: photoFields, append: appendPhoto, remove: removePhoto } = useFieldArray({
    control,
    name: "photos",
  });

  // Watch values for real-time live preview
  const watchName = watch("name") as string;
  const watchCategoryId = watch("category_id") as string;
  const watchDescription = watch("description") as string;
  const watchMoq = watch("moq") as string;
  const watchStartingPrice = watch("starting_price") as number | undefined;
  const watchCurrency = watch("currency") as string;
  const watchCapacityText = watch("capacity_text") as string;
  const watchIsFeatured = watch("is_featured") as boolean;
  const watchPhotos = (watch("photos") || []) as { file_url: string; caption?: string; sort_order: number; id?: string }[];

  // Find selected category name for the live preview
  const selectedCategoryName = categories?.find(c => c.id === watchCategoryId)?.name || "Uncategorized";

  // Reset form with product values when editing
  useEffect(() => {
    if (!isCreate && product) {
      reset({
        name: product.name,
        category_id: product.category_id,
        description: product.description,
        moq: product.moq,
        starting_price: product.starting_price,
        currency: product.currency,
        capacity_text: product.capacity_text,
        is_featured: product.is_featured,
        sort_order: product.sort_order,
        photos: product.photos || [],
      });
    }
  }, [product, isCreate, reset]);

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
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const onSubmit = (values: any) => {
    if (isCreate) {
      createProduct(values, {
        onSuccess: () => router.push("/supplier/products"),
      });
    } else if (id) {
      updateProduct(
        { id, data: values },
        {
          onSuccess: () => router.push("/supplier/products"),
        }
      );
    }
  };

  if (isLoadingProduct && !isCreate) {
    return (
      <div className="space-y-6 text-left max-w-6xl mx-auto px-4 py-8">
        <Skeleton className="h-10 w-1/3" />
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2 space-y-6">
            <Skeleton className="h-[250px] w-full rounded-xl" />
            <Skeleton className="h-[150px] w-full rounded-xl" />
          </div>
          <div className="space-y-6">
            <Skeleton className="h-[300px] w-full rounded-xl" />
            <Skeleton className="h-[80px] w-full rounded-xl" />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto px-4 py-8 pb-32 text-left space-y-8 relative">
      {/* Header */}
      <div className="flex items-center gap-4 pb-5 border-b border-border/60">
        <Button
          type="button"
          variant="outline"
          size="icon"
          onClick={() => router.push("/supplier/products")}
          className="h-10 w-10 cursor-pointer border-border transition-all hover:bg-muted"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            {isCreate ? t("addProduct") : `${t("editProduct")}`}
          </h1>
          <p className="text-xs text-muted-foreground">
            {isCreate
              ? "Tambahkan produk baru ke dalam katalog grosir Anda dengan pratinjau instan."
              : "Ubah detail, spesifikasi, dan kelola galeri gambar produk Anda."}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-start">
          
          {/* Left Column (Forms) - Animates smoothly */}
          <motion.div 
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4 }}
            className="lg:col-span-2 space-y-8"
          >
            {/* Spesifikasi Utama Card */}
            <Card className="border border-border/80 bg-linear-to-b from-card to-card/98 shadow-sm rounded-xl overflow-hidden">
              <CardHeader className="pb-4 border-b border-border/40">
                <CardTitle className="text-sm font-bold font-heading tracking-tight flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-primary" />
                  Spesifikasi Utama
                </CardTitle>
                <CardDescription className="text-xs">Informasi umum mengenai produk Anda.</CardDescription>
              </CardHeader>
              <CardContent className="pt-6 space-y-4">
                <Field>
                  <FieldLabel htmlFor="name" className="text-xs font-semibold text-foreground/80">{t("productName")}</FieldLabel>
                  <Input
                    id="name"
                    placeholder="Masukkan nama produk lengkap"
                    className="focus:ring-2 focus:ring-primary/20 transition-all border-border/80 rounded-lg py-5.5 text-sm"
                    {...register("name")}
                  />
                  {errors.name && <FieldError>{errors.name.message}</FieldError>}
                </Field>

                <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel htmlFor="category_id" className="text-xs font-semibold text-foreground/80">{t("category")}</FieldLabel>
                    <select
                      id="category_id"
                      className="w-full px-3 py-2 bg-card border border-border/80 text-sm rounded-lg outline-hidden focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all cursor-pointer h-10"
                      {...register("category_id")}
                    >
                      <option value="">{t("selectCategory")}</option>
                      {categories?.map((cat) => (
                        <option key={cat.id} value={cat.id}>
                          {cat.name}
                        </option>
                      ))}
                    </select>
                    {errors.category_id && <FieldError>{errors.category_id.message}</FieldError>}
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="capacity_text" className="text-xs font-semibold text-foreground/80">{t("capacity")}</FieldLabel>
                    <Input
                      id="capacity_text"
                      placeholder="e.g. 500 Ton / Bulan"
                      className="focus:ring-2 focus:ring-primary/20 transition-all border-border/80 rounded-lg h-10"
                      {...register("capacity_text")}
                    />
                    {errors.capacity_text && <FieldError>{errors.capacity_text.message}</FieldError>}
                  </Field>
                </FieldGroup>

                <Field>
                  <FieldLabel htmlFor="description" className="text-xs font-semibold text-foreground/80">{t("description")}</FieldLabel>
                  <Textarea
                    id="description"
                    placeholder="Deskripsi produk, spesifikasi teknis, standar ekspor, kemasan, dll."
                    rows={6}
                    className="resize-none focus:ring-2 focus:ring-primary/20 transition-all border-border/80 rounded-lg"
                    {...register("description")}
                  />
                  {errors.description && <FieldError>{errors.description.message}</FieldError>}
                </Field>
              </CardContent>
            </Card>

            {/* Ketentuan Harga Card */}
            <Card className="border border-border/80 bg-linear-to-b from-card to-card/98 shadow-sm rounded-xl overflow-hidden">
              <CardHeader className="pb-4 border-b border-border/40">
                <CardTitle className="text-sm font-bold font-heading tracking-tight flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-primary" />
                  Ketentuan Harga & Minimum Order
                </CardTitle>
                <CardDescription className="text-xs">Atur harga mulai produk dan batas minimum kuantitas pesanan.</CardDescription>
              </CardHeader>
              <CardContent className="pt-6">
                <FieldGroup className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <Field>
                    <FieldLabel htmlFor="starting_price" className="text-xs font-semibold text-foreground/80">{t("price")}</FieldLabel>
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
                          placeholder="e.g. 15.000"
                          className="focus:ring-2 focus:ring-primary/20 border-border/80 rounded-lg h-10"
                        />
                      )}
                    />
                    {errors.starting_price && <FieldError>{errors.starting_price.message}</FieldError>}
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="currency" className="text-xs font-semibold text-foreground/80">{t("currency")}</FieldLabel>
                    <Input
                      id="currency"
                      placeholder="e.g. IDR"
                      className="focus:ring-2 focus:ring-primary/20 border-border/80 rounded-lg h-10"
                      {...register("currency")}
                    />
                    {errors.currency && <FieldError>{errors.currency.message}</FieldError>}
                  </Field>

                  <Field>
                    <FieldLabel htmlFor="moq" className="text-xs font-semibold text-foreground/80">{t("moq")}</FieldLabel>
                    <Input
                      id="moq"
                      placeholder="e.g. 20 Ton"
                      className="focus:ring-2 focus:ring-primary/20 border-border/80 rounded-lg h-10"
                      {...register("moq")}
                    />
                    {errors.moq && <FieldError>{errors.moq.message}</FieldError>}
                  </Field>
                </FieldGroup>
              </CardContent>
            </Card>
          </motion.div>

          {/* Right Column (Gallery & Status / Live Preview) */}
          <motion.div 
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.4, delay: 0.15 }}
            className="space-y-8"
          >
            {/* Foto & Galeri Card */}
            <Card className="border border-border/80 bg-linear-to-b from-card to-card/98 shadow-sm rounded-xl overflow-hidden">
              <CardHeader className="pb-4 border-b border-border/40">
                <CardTitle className="text-sm font-bold font-heading tracking-tight flex items-center gap-2">
                  <span className="h-2 w-2 rounded-full bg-primary" />
                  {t("photos")}
                </CardTitle>
                <CardDescription className="text-xs">Upload minimal 1 gambar. Gambar pertama akan menjadi cover utama.</CardDescription>
              </CardHeader>
              <CardContent className="pt-6 space-y-4">
                {/* Dashed Upload Button */}
                <label className="border-2 border-dashed border-border/80 hover:border-primary/50 transition-all rounded-lg p-5 flex flex-col items-center justify-center gap-1.5 cursor-pointer bg-muted/10 hover:bg-muted/20 relative group">
                  {isUploading ? (
                    <Loader2 className="h-5 w-5 animate-spin text-primary" />
                  ) : (
                    <div className="h-8 w-8 bg-primary/10 text-primary rounded-full flex items-center justify-center transition-all group-hover:scale-105">
                      <Upload className="h-4 w-4" />
                    </div>
                  )}
                  <span className="text-xs font-semibold text-foreground">
                    {isUploading ? "Mengunggah..." : t("addPhoto")}
                  </span>
                  <span className="text-[9px] text-muted-foreground">PNG, JPG, WebP (Max 5MB)</span>
                  <input
                    type="file"
                    multiple
                    accept="image/*"
                    onChange={handleImageUpload}
                    disabled={isUploading}
                    className="hidden"
                  />
                </label>

                {/* List of Photos with Animation */}
                <div className="space-y-3 max-h-[300px] overflow-y-auto pr-1">
                  <AnimatePresence initial={false}>
                    {photoFields.map((field, index) => {
                      const isCover = index === 0;
                      return (
                        <motion.div
                          key={field.id}
                          initial={{ opacity: 0, height: 0 }}
                          animate={{ opacity: 1, height: "auto" }}
                          exit={{ opacity: 0, height: 0 }}
                          className={`flex flex-col p-2.5 border rounded-lg bg-muted/5 relative group transition-all hover:bg-muted/10 ${
                            isCover ? "border-primary/60" : "border-border/60"
                          }`}
                        >
                          <div className="flex items-center gap-3">
                            <div className="relative h-11 w-11 rounded border border-border/80 overflow-hidden bg-background shrink-0">
                              {/* eslint-disable-next-line @next/next/no-img-element */}
                              <img
                                src={field.file_url}
                                alt={`Photo ${index}`}
                                className="h-full w-full object-cover"
                              />
                              {isCover && (
                                <div className="absolute inset-0 bg-primary/90 flex items-center justify-center">
                                  <span className="text-[7px] font-bold text-primary-foreground tracking-wider uppercase">Cover</span>
                                </div>
                              )}
                            </div>
                            
                            <div className="flex-1 min-w-0 space-y-1">
                              <Input
                                placeholder="Caption / Keterangan"
                                className="h-7 text-xs py-0.5 px-2 focus:ring-1 focus:ring-primary/20 border-border/80 rounded"
                                {...register(`photos.${index}.caption`)}
                              />
                            </div>
                            
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              onClick={() => removePhoto(index)}
                              className="h-7 w-7 text-muted-foreground hover:text-destructive hover:bg-destructive/10 shrink-0 cursor-pointer transition-all"
                            >
                              <Trash2 className="h-3.5 w-3.5" />
                            </Button>
                          </div>
                        </motion.div>
                      );
                    })}
                  </AnimatePresence>
                </div>
                {errors.photos && <FieldError>{errors.photos.message}</FieldError>}
              </CardContent>
            </Card>

            {/* Unggulkan Switch Card */}
            <Card className="border border-border/80 bg-gradient-to-r from-amber-500/5 to-card/50 shadow-xs rounded-xl overflow-hidden transition-all hover:border-amber-500/30">
              <CardContent className="p-4">
                <div className="flex items-center justify-between gap-4">
                  <div className="space-y-0.5 text-left flex-1">
                    <span className="text-xs font-bold text-foreground flex items-center gap-1.5">
                      <Star className={`h-3.5 w-3.5 transition-all ${
                        watchIsFeatured ? "text-amber-500 fill-amber-500 scale-110" : "text-muted-foreground"
                      }`} />
                      {t("isFeatured")}
                    </span>
                    <p className="text-[10px] text-muted-foreground leading-normal">
                      {t("featuredDesc")}
                    </p>
                  </div>
                  <Controller
                    control={control}
                    name="is_featured"
                    render={({ field: { value, onChange } }) => (
                      <Switch
                        checked={!!value}
                        onCheckedChange={onChange}
                        className="cursor-pointer"
                      />
                    )}
                  />
                </div>
              </CardContent>
            </Card>

            {/* Unlimited Budget Element: LIVE INTERACTIVE PREVIEW */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold tracking-wider text-muted-foreground flex items-center gap-1.5 uppercase">
                  <Eye className="h-3.5 w-3.5 text-primary" />
                  Pratinjau Live Pembeli
                </span>
                <span className="flex h-2 w-2 relative">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-success opacity-75"></span>
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-success"></span>
                </span>
              </div>

              {/* Renders a premium, dynamic copy of the card */}
              <Card className="overflow-hidden border border-border shadow-md rounded-xl bg-card flex flex-col justify-between h-[420px] relative group text-left transition-all duration-300">
                {/* Image Section */}
                <div className="h-[180px] bg-muted/20 relative overflow-hidden shrink-0 border-b border-border w-full">
                  {watchPhotos.length > 0 ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img
                      src={watchPhotos[0].file_url}
                      alt="Preview Cover"
                      className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-103"
                    />
                  ) : (
                    <div className="w-full h-full flex flex-col items-center justify-center text-muted-foreground gap-2">
                      <ShoppingBag className="h-8 w-8 opacity-40 animate-pulse" />
                      <span className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground/60">Upload Foto Produk</span>
                    </div>
                  )}

                  {watchIsFeatured && (
                    <Badge className="absolute top-3 left-3 bg-amber-500 hover:bg-amber-600 text-white border-0 flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-full select-none">
                      <Star className="h-3 w-3 fill-white" /> Featured
                    </Badge>
                  )}
                </div>

                {/* Content Section */}
                <CardContent className="p-5 flex-1 flex flex-col justify-between">
                  <div className="space-y-2">
                    <div className="flex items-center gap-1 text-[10px] text-muted-foreground font-semibold uppercase tracking-wider">
                      <Layers className="h-3 w-3" />
                      <span>{selectedCategoryName}</span>
                    </div>

                    <h3 className="font-bold text-foreground text-sm leading-tight tracking-tight line-clamp-2 min-h-[1.25rem]">
                      {watchName || "Nama Produk Anda"}
                    </h3>

                    <p className="text-[11px] text-muted-foreground line-clamp-2 min-h-[1.75rem]">
                      {watchDescription || "Deskripsi lengkap mengenai produk, spesifikasi ekspor, dan keunggulan akan tampil di sini..."}
                    </p>
                  </div>

                  {/* Info list */}
                  <div className="space-y-1.5 pt-2 border-t border-border mt-2">
                    <div className="flex justify-between items-center text-[11px]">
                      <span className="text-muted-foreground">Min. Order (MOQ)</span>
                      <Badge variant="outline" className="bg-primary/5 text-primary border-primary/20 text-[9px] font-bold rounded-full px-2 py-0.2">
                        {watchMoq || "1"}
                      </Badge>
                    </div>

                    <div className="flex justify-between items-center text-[11px]">
                      <span className="text-muted-foreground">Capacity</span>
                      <span className="font-semibold text-foreground truncate max-w-[130px]">{watchCapacityText || "-"}</span>
                    </div>

                    <div className="flex justify-between items-baseline pt-1">
                      <span className="text-[10px] text-muted-foreground uppercase font-bold tracking-wider">Starting Price</span>
                      <span className="font-extrabold text-foreground text-sm">
                        {watchStartingPrice !== undefined && watchStartingPrice > 0 ? (
                          <>
                            {watchCurrency} {Number(watchStartingPrice).toLocaleString("id-ID")}
                          </>
                        ) : (
                          "Hubungi Seller"
                        )}
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </motion.div>
        </div>

        {/* Sticky Bottom Actions Bar */}
        <div className="fixed bottom-0 left-0 right-0 z-40 bg-background/80 backdrop-blur-md border-t border-border py-4 shadow-md">
          <div className="max-w-6xl mx-auto px-4 flex items-center justify-between">
            <Button
              type="button"
              variant="outline"
              onClick={() => router.push("/supplier/products")}
              className="cursor-pointer border-border hover:bg-muted font-medium px-5 transition-all text-xs"
            >
              {t("cancel")}
            </Button>
            
            <Button
              type="submit"
              disabled={isCreating || isUpdating}
              className="bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-bold px-6 py-2.5 text-xs flex items-center gap-1.5 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-md hover:shadow-primary/20"
            >
              {isCreating || isUpdating ? (
                <>
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  {isCreate ? "Menyimpan..." : "Memperbarui..."}
                </>
              ) : (
                <>
                  <Save className="h-3.5 w-3.5" />
                  {isCreate ? t("save") : t("saveChanges")}
                </>
              )}
            </Button>
          </div>
        </div>
      </form>
    </div>
  );
}

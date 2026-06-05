"use client";

import React, { useMemo, useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useTranslations } from "next-intl";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup, FieldError } from "@/components/ui/field";
import { toast } from "sonner";
import { ArrowLeft, ArrowRight, Check, Sparkles } from "lucide-react";
import { authService } from "@/features/auth/services/auth-service";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";

const stepKeys = ["step1", "step2", "step3", "step4", "step5"] as const;

const getStepForField = (field: string): number => {
  if (["companyName", "primaryCategory", "subcategory"].includes(field)) return 1;
  if (["provinceId", "cityId", "address"].includes(field)) return 2;
  if (["phone", "whatsapp", "email", "website"].includes(field)) return 3;
  if (["companyType", "taxStatus", "npwp", "nib", "businessHours", "timezone"].includes(field)) return 4;
  if (["description", "firstProductName", "firstProductPrice"].includes(field)) return 5;
  return 1;
};

const inputClassName =
  "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground placeholder:opacity-75 focus-visible:outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 group-data-[invalid=true]/field:border-destructive group-data-[invalid=true]/field:focus-visible:ring-destructive group-data-[invalid=true]/field:focus-visible:shadow-[0_0_0_6px] group-data-[invalid=true]/field:focus-visible:shadow-destructive/10";

export function SupplierOnboardingPage() {
  const router = useRouter();
  const t = useTranslations("supplier.onboarding");
  const { user, setUser, setSessionVerified } = useAuthStore();
  const [step, setStep] = useState(1);
  const [isSaving, setIsSaving] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [form, setForm] = useState({
    companyName: "",
    primaryCategory: "",
    subcategory: "",
    provinceId: "",
    cityId: "",
    address: "",
    phone: "",
    whatsapp: "",
    email: user?.email ?? "",
    website: "",
    companyType: "",
    taxStatus: "PKP",
    npwp: "",
    nib: "",
    businessHours: "Monday-Friday 08:00-17:00",
    timezone: "Asia/Jakarta",
    description: "",
    firstProductName: "",
    firstProductPrice: "",
  });

  const stepTitles = useMemo(() => stepKeys.map((key) => t(key)), [t]);

  const updateForm = (key: keyof typeof form, value: string) => {
    setForm((current) => ({ ...current, [key]: value }));
    if (errors[key]) {
      setErrors((current) => {
        const next = { ...current };
        delete next[key];
        return next;
      });
    }
  };

  const validateStep = (targetStep: number) => {
    const newErrors: Record<string, string> = {};

    if (targetStep === 1) {
      if (!form.companyName.trim()) {
        newErrors.companyName = "Nama perusahaan wajib diisi";
      }
      if (!form.primaryCategory.trim()) {
        newErrors.primaryCategory = "Kategori utama wajib diisi";
      }
      if (!form.subcategory.trim()) {
        newErrors.subcategory = "Subkategori wajib diisi";
      }
    }

    if (targetStep === 2) {
      if (!form.provinceId.trim()) {
        newErrors.provinceId = "Provinsi wajib diisi";
      }
      if (!form.cityId.trim()) {
        newErrors.cityId = "Kota / Kabupaten wajib diisi";
      }
    }

    if (targetStep === 3) {
      if (!form.phone.trim()) {
        newErrors.phone = "Nomor telepon wajib diisi";
      }
      if (!form.whatsapp.trim()) {
        newErrors.whatsapp = "Nomor WhatsApp wajib diisi";
      }
      if (!form.email.trim()) {
        newErrors.email = "Email wajib diisi";
      } else if (!/\S+@\S+\.\S+/.test(form.email)) {
        newErrors.email = "Format email tidak valid";
      }
    }

    if (targetStep === 4) {
      if (!form.companyType) {
        newErrors.companyType = "Tipe perusahaan wajib diisi";
      }
      if (!form.taxStatus) {
        newErrors.taxStatus = "Status pajak wajib diisi";
      }
      if (form.npwp && form.npwp.trim().length > 0 && form.npwp.trim().length < 4) {
        newErrors.npwp = "NPWP minimal 4 karakter";
      }
      if (form.nib && form.nib.trim().length > 0 && form.nib.trim().length < 4) {
        newErrors.nib = "NIB minimal 4 karakter";
      }
    }

    if (targetStep === 5) {
      if (!form.description.trim()) {
        newErrors.description = "Deskripsi bisnis wajib diisi";
      } else if (form.description.trim().length < 10) {
        newErrors.description = "Deskripsi bisnis minimal 10 karakter";
      }
    }

    setErrors(newErrors);
    
    if (Object.keys(newErrors).length > 0) {
      toast.error("Harap lengkapi semua kolom yang wajib diisi pada langkah ini.");
      return false;
    }
    return true;
  };

  const handleNext = () => {
    if (!validateStep(step)) {
      return;
    }
    setStep((current) => current + 1);
  };

  const handleBack = () => {
    setStep((current) => current - 1);
  };

  const handleComplete = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateStep(5)) {
      return;
    }

    setIsSaving(true);
    try {
      const response = await authService.becomeSupplier({
        company_name: form.companyName.trim(),
        primary_category: form.primaryCategory.trim(),
        subcategory: form.subcategory.trim(),
        province_id: form.provinceId.trim(),
        city_id: form.cityId.trim(),
        address: form.address.trim() || undefined,
        phone: form.phone.trim(),
        whatsapp: form.whatsapp.trim(),
        email: form.email.trim(),
        website: form.website.trim() || undefined,
        company_type: form.companyType,
        tax_status: form.taxStatus,
        npwp: form.npwp.trim() || undefined,
        nib: form.nib.trim() || undefined,
        business_hours: form.businessHours.trim() || undefined,
        timezone: form.timezone.trim() || undefined,
        description: form.description.trim(),
        first_product_name: form.firstProductName.trim() || undefined,
        first_product_price: form.firstProductPrice.trim() || undefined,
      });

      if (response.success && response.data?.user) {
        setUser(response.data.user);
        setSessionVerified(true);
      }

      toast.success(t("success"));
      router.push("/supplier/dashboard");
    } catch (err: any) {
      const fieldErrors = err?.response?.data?.error?.field_errors;
      if (fieldErrors && Array.isArray(fieldErrors)) {
        const mappedErrors: Record<string, string> = {};
        let firstErrorStep = 5;
        fieldErrors.forEach((fe: any) => {
          let fieldKey = fe.field;
          if (fieldKey === "company_name") fieldKey = "companyName";
          else if (fieldKey === "primary_category") fieldKey = "primaryCategory";
          else if (fieldKey === "subcategory") fieldKey = "subcategory";
          else if (fieldKey === "province_id") fieldKey = "provinceId";
          else if (fieldKey === "city_id") fieldKey = "cityId";
          else if (fieldKey === "company_type" || fieldKey === "business_type") fieldKey = "companyType";
          else if (fieldKey === "tax_status") fieldKey = "taxStatus";
          else if (fieldKey === "npwp" || fieldKey === "tax_id") fieldKey = "npwp";
          else if (fieldKey === "nib") fieldKey = "nib";
          else if (fieldKey === "business_hours") fieldKey = "businessHours";
          else if (fieldKey === "first_product_name") fieldKey = "firstProductName";
          else if (fieldKey === "first_product_price") fieldKey = "firstProductPrice";

          mappedErrors[fieldKey] = fe.message || "Validasi gagal";
          
          const fieldStep = getStepForField(fieldKey);
          if (fieldStep < firstErrorStep) {
            firstErrorStep = fieldStep;
          }
        });
        setErrors(mappedErrors);
        setStep(firstErrorStep);
        toast.error("Validasi gagal. Silakan periksa kembali formulir Anda.");
      } else {
        toast.error(t("failed"));
      }
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="mx-auto max-w-4xl space-y-8 py-6">
      <div className="space-y-3 text-center">
        <div className="inline-flex items-center gap-1.5 rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-semibold text-primary">
          <Sparkles className="h-3.5 w-3.5" />
          Supplier Registration
        </div>
        <h1 className="text-2xl font-extrabold tracking-tight text-foreground md:text-3xl">
          {t("title")}
        </h1>
        <p className="mx-auto max-w-2xl text-sm text-muted-foreground">{t("subtitle")}</p>
      </div>

      <div className="relative mx-auto flex max-w-3xl items-start justify-between before:absolute before:left-6 before:right-6 before:top-[18px] before:h-0.5 before:bg-border">
        {stepTitles.map((title, index) => {
          const currentStep = index + 1;
          const isCompleted = currentStep < step;
          const isActive = currentStep === step;

          return (
            <div key={title} className="relative z-10 flex w-24 flex-col items-center gap-2 text-center">
              <div
                className={`flex h-9 w-9 items-center justify-center rounded-full border text-sm font-bold transition-all ${
                  isCompleted
                    ? "border-success bg-success text-success-foreground"
                    : isActive
                      ? "border-primary bg-primary text-primary-foreground ring-4 ring-primary/20"
                      : "border-border bg-card text-muted-foreground"
                }`}
              >
                {isCompleted ? <Check className="h-4 w-4" /> : currentStep}
              </div>
              <span className={`text-[10px] font-bold uppercase ${isActive ? "text-primary" : "text-muted-foreground"}`}>
                {title}
              </span>
            </div>
          );
        })}
      </div>

      <Card className="overflow-hidden rounded-2xl border border-border bg-card shadow-md">
        <CardContent className="p-6 md:p-8">
          {step === 1 && (
            <div className="space-y-4">
              <h3 className="text-base font-bold text-foreground">{t("step1")}</h3>
              <Field invalid={!!errors.companyName}>
                <FieldLabel>Nama perusahaan</FieldLabel>
                <Input value={form.companyName} onChange={(e) => updateForm("companyName", e.target.value)} placeholder="Contoh: PT Rempah Nusantara Abadi" />
                {errors.companyName && <FieldError>{errors.companyName}</FieldError>}
              </Field>
              <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field invalid={!!errors.primaryCategory}>
                  <FieldLabel>Kategori utama</FieldLabel>
                  <Input value={form.primaryCategory} onChange={(e) => updateForm("primaryCategory", e.target.value)} placeholder="Contoh: Makanan & Minuman" />
                  {errors.primaryCategory && <FieldError>{errors.primaryCategory}</FieldError>}
                </Field>
                <Field invalid={!!errors.subcategory}>
                  <FieldLabel>Subkategori</FieldLabel>
                  <Input value={form.subcategory} onChange={(e) => updateForm("subcategory", e.target.value)} placeholder="Contoh: Bumbu & Rempah" />
                  {errors.subcategory && <FieldError>{errors.subcategory}</FieldError>}
                </Field>
              </FieldGroup>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-4">
              <h3 className="text-base font-bold text-foreground">{t("step2")}</h3>
              <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field invalid={!!errors.provinceId}>
                  <FieldLabel>Provinsi</FieldLabel>
                  <Input value={form.provinceId} onChange={(e) => updateForm("provinceId", e.target.value)} placeholder="Contoh: Jawa Timur" />
                  {errors.provinceId && <FieldError>{errors.provinceId}</FieldError>}
                </Field>
                <Field invalid={!!errors.cityId}>
                  <FieldLabel>Kota / Kabupaten</FieldLabel>
                  <Input value={form.cityId} onChange={(e) => updateForm("cityId", e.target.value)} placeholder="Contoh: Surabaya" />
                  {errors.cityId && <FieldError>{errors.cityId}</FieldError>}
                </Field>
              </FieldGroup>
              <Field invalid={!!errors.address}>
                <FieldLabel>Alamat lengkap</FieldLabel>
                <Textarea value={form.address} onChange={(e) => updateForm("address", e.target.value)} placeholder="Opsional. Isi alamat produksi atau kantor jika ingin ditampilkan lebih detail." />
                {errors.address && <FieldError>{errors.address}</FieldError>}
              </Field>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-4">
              <h3 className="text-base font-bold text-foreground">{t("step3")}</h3>
              <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field invalid={!!errors.phone}>
                  <FieldLabel>Telepon</FieldLabel>
                  <Input value={form.phone} onChange={(e) => updateForm("phone", e.target.value)} placeholder="Contoh: 031-1234567" />
                  {errors.phone && <FieldError>{errors.phone}</FieldError>}
                </Field>
                <Field invalid={!!errors.whatsapp}>
                  <FieldLabel>WhatsApp aktif</FieldLabel>
                  <Input value={form.whatsapp} onChange={(e) => updateForm("whatsapp", e.target.value)} placeholder="Contoh: 6281234567890" />
                  {errors.whatsapp && <FieldError>{errors.whatsapp}</FieldError>}
                </Field>
              </FieldGroup>
              <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field invalid={!!errors.email}>
                  <FieldLabel>Email bisnis</FieldLabel>
                  <Input type="email" value={form.email} onChange={(e) => updateForm("email", e.target.value)} placeholder="sales@company.com" />
                  {errors.email && <FieldError>{errors.email}</FieldError>}
                </Field>
                <Field invalid={!!errors.website}>
                  <FieldLabel>Website</FieldLabel>
                  <Input value={form.website} onChange={(e) => updateForm("website", e.target.value)} placeholder="https://company.com" />
                  {errors.website && <FieldError>{errors.website}</FieldError>}
                </Field>
              </FieldGroup>
            </div>
          )}

          {step === 4 && (
            <div className="space-y-4">
              <h3 className="text-base font-bold text-foreground">{t("step4")}</h3>
              <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field invalid={!!errors.companyType}>
                  <FieldLabel>Tipe perusahaan</FieldLabel>
                  <select className={`${inputClassName} cursor-pointer`} value={form.companyType} onChange={(e) => updateForm("companyType", e.target.value)}>
                    <option value="">Pilih tipe perusahaan</option>
                    <option value="PT">PT</option>
                    <option value="CV">CV</option>
                    <option value="UD">UD</option>
                    <option value="Cooperative">Koperasi</option>
                    <option value="Other">Lainnya</option>
                  </select>
                  {errors.companyType && <FieldError>{errors.companyType}</FieldError>}
                </Field>
                <Field invalid={!!errors.taxStatus}>
                  <FieldLabel>Status pajak</FieldLabel>
                  <select className={`${inputClassName} cursor-pointer`} value={form.taxStatus} onChange={(e) => updateForm("taxStatus", e.target.value)}>
                    <option value="PKP">PKP</option>
                    <option value="Non-PKP">Non-PKP</option>
                  </select>
                  {errors.taxStatus && <FieldError>{errors.taxStatus}</FieldError>}
                </Field>
              </FieldGroup>
              <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field invalid={!!errors.npwp}>
                  <FieldLabel>NPWP</FieldLabel>
                  <Input value={form.npwp} onChange={(e) => updateForm("npwp", e.target.value)} placeholder="Opsional saat registrasi awal" />
                  {errors.npwp && <FieldError>{errors.npwp}</FieldError>}
                </Field>
                <Field invalid={!!errors.nib}>
                  <FieldLabel>NIB</FieldLabel>
                  <Input value={form.nib} onChange={(e) => updateForm("nib", e.target.value)} placeholder="Opsional saat registrasi awal" />
                  {errors.nib && <FieldError>{errors.nib}</FieldError>}
                </Field>
              </FieldGroup>
              <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field invalid={!!errors.businessHours}>
                  <FieldLabel>Jam operasional</FieldLabel>
                  <Input value={form.businessHours} onChange={(e) => updateForm("businessHours", e.target.value)} placeholder="Monday-Friday 08:00-17:00" />
                  {errors.businessHours && <FieldError>{errors.businessHours}</FieldError>}
                </Field>
                <Field invalid={!!errors.timezone}>
                  <FieldLabel>Timezone</FieldLabel>
                  <Input value={form.timezone} onChange={(e) => updateForm("timezone", e.target.value)} placeholder="Asia/Jakarta" />
                  {errors.timezone && <FieldError>{errors.timezone}</FieldError>}
                </Field>
              </FieldGroup>
            </div>
          )}

          {step === 5 && (
            <form onSubmit={handleComplete} className="space-y-4">
              <h3 className="text-base font-bold text-foreground">{t("step5")}</h3>
              <Field invalid={!!errors.description}>
                <FieldLabel>Deskripsi bisnis</FieldLabel>
                <Textarea
                  value={form.description}
                  onChange={(e) => updateForm("description", e.target.value)}
                  placeholder="Tulis 2-3 kalimat singkat tentang perusahaan, produk utama, kapasitas, atau pengalaman pasar Anda."
                  rows={5}
                />
                {errors.description && <FieldError>{errors.description}</FieldError>}
              </Field>
              <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field invalid={!!errors.firstProductName}>
                  <FieldLabel>Produk pertama</FieldLabel>
                  <Input value={form.firstProductName} onChange={(e) => updateForm("firstProductName", e.target.value)} placeholder="Opsional. Contoh: Bubuk kayu manis premium" />
                  {errors.firstProductName && <FieldError>{errors.firstProductName}</FieldError>}
                </Field>
                <Field invalid={!!errors.firstProductPrice}>
                  <FieldLabel>Harga awal / terms</FieldLabel>
                  <Input value={form.firstProductPrice} onChange={(e) => updateForm("firstProductPrice", e.target.value)} placeholder="Opsional. Contoh: Rp 95.000 / kg" />
                  {errors.firstProductPrice && <FieldError>{errors.firstProductPrice}</FieldError>}
                </Field>
              </FieldGroup>

              <div className="rounded-xl border border-border bg-muted/30 p-4 text-sm text-muted-foreground">
                Setelah profil supplier aktif, Anda tetap bisa melengkapi foto produk, portofolio, sertifikasi, dan dokumen verifikasi Level 2 dari dashboard supplier.
              </div>

              <div className="flex justify-between pt-4">
                <Button type="button" variant="outline" onClick={handleBack} className="h-9 cursor-pointer border-border text-xs font-semibold">
                  <ArrowLeft className="mr-1.5 h-4 w-4" />
                  {t("btnBack")}
                </Button>
                <Button
                  type="submit"
                  disabled={isSaving}
                  className="flex h-9 cursor-pointer items-center justify-center gap-1.5 bg-primary text-xs font-semibold text-primary-foreground transition-all duration-300 hover:-translate-y-0.5 hover:bg-primary/95 active:translate-y-0"
                >
                  <Check className="h-4 w-4" />
                  {isSaving ? "Memproses..." : t("btnComplete")}
                </Button>
              </div>
            </form>
          )}

          {step < 5 && (
            <div className="flex justify-between pt-6">
              <Button
                type="button"
                variant="outline"
                onClick={handleBack}
                disabled={step === 1}
                className="h-9 cursor-pointer border-border text-xs font-semibold"
              >
                <ArrowLeft className="mr-1.5 h-4 w-4" />
                {t("btnBack")}
              </Button>
              <Button type="button" onClick={handleNext} className="h-9 cursor-pointer text-xs font-semibold">
                {t("btnNext")}
                <ArrowRight className="ml-1.5 h-4 w-4" />
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { motion, AnimatePresence } from "framer-motion";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Field, FieldLabel, FieldError } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { ShieldCheck, User, Store, ArrowRight, ArrowLeft } from "lucide-react";

interface PublicRegistrationPageProps {
  locale: string;
}

const getRegSchema = (t: (key: string) => string) =>
  z.object({
    companyName: z.string().min(2, t("companyName")),
    contactName: z.string().min(2, t("contactName")),
    phone: z.string().min(6, t("phone")),
    email: z.string().email(),
    description: z.string().optional(),
  });

type RegFormData = z.infer<ReturnType<typeof getRegSchema>>;

export function PublicRegistrationPage({ locale }: PublicRegistrationPageProps) {
  const tReg = useTranslations("public.register");
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [accountType, setAccountType] = useState<"supplier" | "buyer" | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<RegFormData>({
    resolver: zodResolver(getRegSchema(tReg)),
    defaultValues: {
      companyName: "",
      contactName: "",
      phone: "",
      email: "",
      description: "",
    },
  });

  const handleNextStep = () => {
    if (!accountType) {
      toast.error("Please select how you want to join first.");
      return;
    }
    setStep(2);
  };

  const onSubmit = () => {
    setIsSubmitting(true);
    setTimeout(() => {
      setIsSubmitting(false);
      setStep(3);
      reset();
    }, 1500);
  };

  return (
    <PublicLayout locale={locale}>
      <div className="bg-muted py-16 md:py-24 min-h-[70vh] flex items-center justify-center">
        <div className="w-full max-w-[540px] mx-auto px-4">
          <AnimatePresence mode="wait">
             {step === 1 && (
              <motion.div
                key="step1"
                initial={{ opacity: 0, y: 15 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -15 }}
                className="bg-card border border-border shadow-sm rounded-2xl p-6 md:p-10 space-y-8"
              >
                <div className="text-center space-y-2">
                  <h1 className="text-3xl font-bold text-foreground font-heading tracking-tight">
                    {tReg("title")}
                  </h1>
                  <p className="text-sm text-muted-foreground max-w-sm mx-auto">
                    {tReg("subtitle")}
                  </p>
                </div>

                <div className="space-y-4">
                  <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    {tReg("joinAs")}
                  </h3>

                  <div className="grid gap-4">
                    {/* Buyer Selection Option */}
                    <button
                      onClick={() => setAccountType("buyer")}
                      className={`flex gap-4 p-5 rounded-xl border text-left transition-all duration-300 cursor-pointer ${
                        accountType === "buyer"
                          ? "border-primary bg-muted ring-1 ring-primary"
                          : "border-border bg-card hover:border-muted-foreground"
                      }`}
                    >
                      <div className="p-3 bg-primary text-primary-foreground rounded-lg shrink-0">
                        <User className="h-5 w-5" />
                      </div>
                      <div className="space-y-1">
                        <h4 className="font-bold text-foreground text-sm">{tReg("buyer")}</h4>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                          {tReg("buyerDesc")}
                        </p>
                      </div>
                    </button>

                    {/* Supplier Selection Option */}
                    <button
                      onClick={() => setAccountType("supplier")}
                      className={`flex gap-4 p-5 rounded-xl border text-left transition-all duration-300 cursor-pointer ${
                        accountType === "supplier"
                          ? "border-primary bg-muted ring-1 ring-primary"
                          : "border-border bg-card hover:border-muted-foreground"
                      }`}
                    >
                      <div className="p-3 bg-primary text-primary-foreground rounded-lg shrink-0">
                        <Store className="h-5 w-5" />
                      </div>
                      <div className="space-y-1">
                        <h4 className="font-bold text-foreground text-sm">{tReg("supplier")}</h4>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                          {tReg("supplierDesc")}
                        </p>
                      </div>
                    </button>
                  </div>
                </div>

                <Button
                  onClick={handleNextStep}
                  className="w-full bg-primary text-primary-foreground hover:bg-primary/95 py-6 font-semibold tracking-wider flex items-center justify-center gap-2 cursor-pointer"
                >
                  {tReg("btnNext")}
                  <ArrowRight className="h-4 w-4" />
                </Button>
              </motion.div>
            )}

            {step === 2 && (
              <motion.div
                key="step2"
                initial={{ opacity: 0, x: 25 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -25 }}
                className="bg-card border border-border shadow-sm rounded-2xl p-6 md:p-10 space-y-8"
              >
                <div className="flex items-center justify-between pb-4 border-b border-border">
                  <button
                    onClick={() => setStep(1)}
                    className="text-muted-foreground hover:text-foreground flex items-center gap-1 text-xs font-semibold cursor-pointer"
                  >
                    <ArrowLeft className="h-4 w-4" />
                    {tReg("backButton")}
                  </button>
                  <span className="text-xs font-bold text-muted-foreground uppercase tracking-widest">
                    {tReg("stepLabel", { current: 2, total: 2 })}
                  </span>
                </div>

                <div className="space-y-1">
                  <h2 className="text-2xl font-bold text-foreground font-heading">
                    {accountType === "supplier"
                      ? tReg("formTitleSupplier")
                      : tReg("formTitleBuyer")}
                  </h2>
                </div>

                <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
                  <Field invalid={!!errors.companyName}>
                    <FieldLabel>{tReg("companyName")}</FieldLabel>
                    <Input
                      type="text"
                      placeholder="e.g. PT Maju Jaya"
                      {...register("companyName")}
                      className="bg-card border-border focus-visible:border-muted-foreground focus-visible:ring-0"
                    />
                    {errors.companyName && <FieldError>{errors.companyName.message}</FieldError>}
                  </Field>

                  <Field invalid={!!errors.contactName}>
                    <FieldLabel>{tReg("contactName")}</FieldLabel>
                    <Input
                      type="text"
                      placeholder="e.g. John Doe"
                      {...register("contactName")}
                      className="bg-card border-border focus-visible:border-muted-foreground focus-visible:ring-0"
                    />
                    {errors.contactName && <FieldError>{errors.contactName.message}</FieldError>}
                  </Field>

                  <Field invalid={!!errors.phone}>
                    <FieldLabel>{tReg("phone")}</FieldLabel>
                    <Input
                      type="text"
                      placeholder="e.g. 08123456789"
                      {...register("phone")}
                      className="bg-card border-border focus-visible:border-muted-foreground focus-visible:ring-0"
                    />
                    {errors.phone && <FieldError>{errors.phone.message}</FieldError>}
                  </Field>

                  <Field invalid={!!errors.email}>
                    <FieldLabel>{tReg("email")}</FieldLabel>
                    <Input
                      type="email"
                      placeholder="e.g. info@company.com"
                      {...register("email")}
                      className="bg-card border-border focus-visible:border-muted-foreground focus-visible:ring-0"
                    />
                    {errors.email && <FieldError>{errors.email.message}</FieldError>}
                  </Field>

                  <Field>
                    <FieldLabel>{tReg("desc")}</FieldLabel>
                    <Textarea
                      rows={3}
                      placeholder={
                        accountType === "supplier"
                          ? tReg("descPlaceholderSupplier")
                          : tReg("descPlaceholderBuyer")
                      }
                      {...register("description")}
                      className="bg-card border-border focus-visible:border-muted-foreground resize-none"
                    />
                  </Field>

                  <Button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full bg-primary text-primary-foreground hover:bg-primary/95 py-6 font-semibold tracking-wider cursor-pointer"
                  >
                    {isSubmitting ? "Submitting..." : tReg("btnSubmit")}
                  </Button>
                </form>
              </motion.div>
            )}

            {step === 3 && (
              <motion.div
                key="step3"
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="bg-card border border-border shadow-sm rounded-2xl p-10 text-center space-y-6"
              >
                <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-500">
                  <ShieldCheck className="h-8 w-8" />
                </div>

                <div className="space-y-2">
                  <h2 className="text-2xl font-bold text-foreground font-heading">
                    {tReg("successTitle")}
                  </h2>
                  <p className="text-sm text-muted-foreground leading-relaxed max-w-sm mx-auto">
                    {tReg("successDesc")}
                  </p>
                </div>

                <Button
                  onClick={() => {
                    setAccountType(null);
                    setStep(1);
                  }}
                  className="bg-primary text-primary-foreground hover:bg-primary/95 font-semibold cursor-pointer px-8"
                >
                  {tReg("registerAnother")}
                </Button>
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </div>
    </PublicLayout>
  );
}

"use client";

import React, { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { motion, AnimatePresence } from "framer-motion";
import confetti from "canvas-confetti";
import { waitingListService } from "../services/waiting-list-service";
import { Field, FieldLabel, FieldError } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";

const getWaitlistSchema = (t: (key: string) => string) =>
  z.object({
    name: z.string().min(2, t("form.errors.nameRequired")),
    email: z.string().email(t("form.errors.emailInvalid")),
    company_name: z.string().min(2, t("form.errors.companyNameRequired")),
    company_type: z.enum(["supplier", "buyer", "other"], {
      message: t("form.errors.companyTypeRequired"),
    }),
    phone: z.string().optional(),
    notes: z.string().optional(),
  });

type WaitlistFormData = z.infer<ReturnType<typeof getWaitlistSchema>>;

export default function WaitingListForm() {
  const t = useTranslations("waitingList");
  const [isSuccess, setIsSuccess] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<WaitlistFormData>({
    resolver: zodResolver(getWaitlistSchema(t)),
    defaultValues: {
      name: "",
      email: "",
      company_name: "",
      company_type: "supplier",
      phone: "",
      notes: "",
    },
  });

  const onSubmit = async (data: WaitlistFormData) => {
    setIsSubmitting(true);
    try {
      await waitingListService.join(data);
      setIsSuccess(true);
      confetti({
        particleCount: 50,
        spread: 40,
        origin: { y: 0.6 },
        colors: ["#1c1917", "#737373", "#a3a3a3"],
      });
      reset();
    } catch (error: unknown) {
      const err = error as { response?: { data?: { error?: { code?: string; message?: string } } } };
      const apiCode = err?.response?.data?.error?.code;
      if (apiCode === "RESOURCE_ALREADY_EXISTS") {
        toast.error(t("form.errors.alreadyRegistered"));
      } else {
        toast.error(
          err?.response?.data?.error?.message ||
            "An unexpected error occurred. Please try again."
        );
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="w-full font-jost">
      <AnimatePresence mode="wait">
        {!isSuccess ? (
          <motion.div
            key="form"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="w-full max-w-[480px] mx-auto"
          >
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {/* Name */}
              <Field invalid={!!errors.name}>
                <FieldLabel>{t("form.name")}</FieldLabel>
                <Input
                  type="text"
                  placeholder={t("form.namePlaceholder")}
                  {...register("name")}
                />
                {errors.name && (
                  <FieldError>{errors.name.message}</FieldError>
                )}
              </Field>

              {/* Email */}
              <Field invalid={!!errors.email}>
                <FieldLabel>{t("form.email")}</FieldLabel>
                <Input
                  type="email"
                  placeholder={t("form.emailPlaceholder")}
                  {...register("email")}
                />
                {errors.email && (
                  <FieldError>{errors.email.message}</FieldError>
                )}
              </Field>

              {/* Company Details */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <Field invalid={!!errors.company_name}>
                  <FieldLabel>{t("form.companyName")}</FieldLabel>
                  <Input
                    type="text"
                    placeholder={t("form.companyNamePlaceholder")}
                    {...register("company_name")}
                  />
                  {errors.company_name && (
                    <FieldError>{errors.company_name.message}</FieldError>
                  )}
                </Field>

                <Field invalid={!!errors.company_type}>
                  <FieldLabel>{t("form.companyType")}</FieldLabel>
                  <div className="relative">
                    <select
                      {...register("company_type")}
                      className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:border-ring focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-200 group-data-[invalid=true]/field:border-destructive group-data-[invalid=true]/field:focus-visible:ring-destructive appearance-none bg-transparent"
                    >
                      <option value="supplier">{t("form.supplier")}</option>
                      <option value="buyer">{t("form.buyer")}</option>
                      <option value="other">{t("form.other")}</option>
                    </select>
                    <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-muted-foreground">
                      <svg className="fill-current h-4 w-4" viewBox="0 0 20 20">
                        <path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z" />
                      </svg>
                    </div>
                  </div>
                  {errors.company_type && (
                    <FieldError>{errors.company_type.message}</FieldError>
                  )}
                </Field>
              </div>

              {/* Phone */}
              <Field>
                <FieldLabel>{t("form.phone")}</FieldLabel>
                <Input
                  type="text"
                  placeholder={t("form.phonePlaceholder")}
                  {...register("phone")}
                />
              </Field>

              {/* Notes */}
              <Field>
                <FieldLabel>{t("form.notes")}</FieldLabel>
                <Textarea
                  rows={3}
                  placeholder={t("form.notesPlaceholder")}
                  {...register("notes")}
                  className="resize-none"
                />
              </Field>

              {/* Submit button */}
              <Button
                type="submit"
                disabled={isSubmitting}
                variant="cta"
                className="w-full"
              >
                {isSubmitting ? t("form.submitting") : t("form.submit")}
              </Button>
            </form>
          </motion.div>
        ) : (
          <motion.div
            key="success"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-12"
          >
            <h3 className="text-[24px] font-normal text-foreground mb-4 tracking-tight">
              {t("form.successTitle")}
            </h3>
            <p className="text-[15px] font-light text-muted-foreground leading-relaxed max-w-sm mx-auto mb-8">
              {t("form.successMessage")}
            </p>
            <Button
              onClick={() => setIsSuccess(false)}
              variant="outline"
              className="px-6 py-3 text-[13px] uppercase tracking-wider cursor-pointer"
            >
              {t("title")}
            </Button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

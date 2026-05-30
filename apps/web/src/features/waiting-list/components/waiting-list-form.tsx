"use client";

import React, { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { motion, AnimatePresence } from "framer-motion";
import { CheckCircle2, ChevronRight, Loader2, Sparkles, Building2, User, Mail, Phone, FileText } from "lucide-react";
import confetti from "canvas-confetti";
import { waitingListService } from "../services/waiting-list-service";

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
        particleCount: 100,
        spread: 70,
        origin: { y: 0.6 },
      });
      reset();
    } catch (error: any) {
      const apiCode = error?.response?.data?.error?.code;
      if (apiCode === "RESOURCE_ALREADY_EXISTS") {
        toast.error(t("form.errors.alreadyRegistered"));
      } else {
        toast.error(error?.response?.data?.error?.message || "An unexpected error occurred. Please try again.");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="w-full max-w-xl mx-auto">
      <AnimatePresence mode="wait">
        {!isSuccess ? (
          <motion.div
            key="form-card"
            initial={{ opacity: 0, y: 15 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -15 }}
            transition={{ duration: 0.4, ease: "easeOut" }}
            className="relative overflow-hidden rounded-2xl border border-neutral-200/80 dark:border-neutral-800/80 bg-white/70 dark:bg-neutral-900/70 backdrop-blur-xl p-8 shadow-2xl shadow-neutral-200/40 dark:shadow-black/40"
          >
            {/* Top decorative gradient bar */}
            <div className="absolute top-0 left-0 right-0 h-1.5 bg-gradient-to-r from-blue-600 via-indigo-600 to-purple-600" />

            <div className="flex items-center gap-2 mb-6">
              <span className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400">
                <Sparkles className="h-4 w-4" />
              </span>
              <h3 className="text-xl font-bold text-neutral-900 dark:text-white tracking-tight">
                {t("title")}
              </h3>
            </div>

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              {/* Name */}
              <div>
                <label className="block text-sm font-semibold text-neutral-700 dark:text-neutral-300 mb-1.5 flex items-center gap-1.5">
                  <User className="h-3.5 w-3.5 text-neutral-400" />
                  {t("form.name")}
                </label>
                <input
                  type="text"
                  placeholder={t("form.namePlaceholder")}
                  {...register("name")}
                  className={`w-full rounded-lg border px-3.5 py-2.5 bg-neutral-50/50 dark:bg-neutral-800/30 text-neutral-900 dark:text-white outline-none transition-all focus:ring-2 focus:ring-blue-500/20 ${
                    errors.name
                      ? "border-red-500/80 focus:border-red-500"
                      : "border-neutral-300/80 dark:border-neutral-700/80 focus:border-blue-500 dark:focus:border-blue-400"
                  }`}
                />
                {errors.name && (
                  <span className="text-xs font-medium text-red-500 mt-1 block">
                    {errors.name.message}
                  </span>
                )}
              </div>

              {/* Email */}
              <div>
                <label className="block text-sm font-semibold text-neutral-700 dark:text-neutral-300 mb-1.5 flex items-center gap-1.5">
                  <Mail className="h-3.5 w-3.5 text-neutral-400" />
                  {t("form.email")}
                </label>
                <input
                  type="email"
                  placeholder={t("form.emailPlaceholder")}
                  {...register("email")}
                  className={`w-full rounded-lg border px-3.5 py-2.5 bg-neutral-50/50 dark:bg-neutral-800/30 text-neutral-900 dark:text-white outline-none transition-all focus:ring-2 focus:ring-blue-500/20 ${
                    errors.email
                      ? "border-red-500/80 focus:border-red-500"
                      : "border-neutral-300/80 dark:border-neutral-700/80 focus:border-blue-500 dark:focus:border-blue-400"
                  }`}
                />
                {errors.email && (
                  <span className="text-xs font-medium text-red-500 mt-1 block">
                    {errors.email.message}
                  </span>
                )}
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* Company Name */}
                <div>
                  <label className="block text-sm font-semibold text-neutral-700 dark:text-neutral-300 mb-1.5 flex items-center gap-1.5">
                    <Building2 className="h-3.5 w-3.5 text-neutral-400" />
                    {t("form.companyName")}
                  </label>
                  <input
                    type="text"
                    placeholder={t("form.companyNamePlaceholder")}
                    {...register("company_name")}
                    className={`w-full rounded-lg border px-3.5 py-2.5 bg-neutral-50/50 dark:bg-neutral-800/30 text-neutral-900 dark:text-white outline-none transition-all focus:ring-2 focus:ring-blue-500/20 ${
                      errors.company_name
                        ? "border-red-500/80 focus:border-red-500"
                        : "border-neutral-300/80 dark:border-neutral-700/80 focus:border-blue-500 dark:focus:border-blue-400"
                    }`}
                  />
                  {errors.company_name && (
                    <span className="text-xs font-medium text-red-500 mt-1 block">
                      {errors.company_name.message}
                    </span>
                  )}
                </div>

                {/* Company Type */}
                <div>
                  <label className="block text-sm font-semibold text-neutral-700 dark:text-neutral-300 mb-1.5 flex items-center gap-1.5">
                    <Building2 className="h-3.5 w-3.5 text-neutral-400" />
                    {t("form.companyType")}
                  </label>
                  <select
                    {...register("company_type")}
                    className={`w-full rounded-lg border px-3.5 py-2.5 bg-neutral-50/50 dark:bg-neutral-800/30 text-neutral-900 dark:text-white outline-none transition-all focus:ring-2 focus:ring-blue-500/20 ${
                      errors.company_type
                        ? "border-red-500/80 focus:border-red-500"
                        : "border-neutral-300/80 dark:border-neutral-700/80 focus:border-blue-500 dark:focus:border-blue-400"
                    }`}
                  >
                    <option value="supplier">{t("form.supplier")}</option>
                    <option value="buyer">{t("form.buyer")}</option>
                    <option value="other">{t("form.other")}</option>
                  </select>
                  {errors.company_type && (
                    <span className="text-xs font-medium text-red-500 mt-1 block">
                      {errors.company_type.message}
                    </span>
                  )}
                </div>
              </div>

              {/* Phone */}
              <div>
                <label className="block text-sm font-semibold text-neutral-700 dark:text-neutral-300 mb-1.5 flex items-center gap-1.5">
                  <Phone className="h-3.5 w-3.5 text-neutral-400" />
                  {t("form.phone")}
                </label>
                <input
                  type="text"
                  placeholder={t("form.phonePlaceholder")}
                  {...register("phone")}
                  className="w-full rounded-lg border border-neutral-300/80 dark:border-neutral-700/80 px-3.5 py-2.5 bg-neutral-50/50 dark:bg-neutral-800/30 text-neutral-900 dark:text-white outline-none transition-all focus:border-blue-500 dark:focus:border-blue-400 focus:ring-2 focus:ring-blue-500/20"
                />
              </div>

              {/* Notes */}
              <div>
                <label className="block text-sm font-semibold text-neutral-700 dark:text-neutral-300 mb-1.5 flex items-center gap-1.5">
                  <FileText className="h-3.5 w-3.5 text-neutral-400" />
                  {t("form.notes")}
                </label>
                <textarea
                  rows={3}
                  placeholder={t("form.notesPlaceholder")}
                  {...register("notes")}
                  className="w-full rounded-lg border border-neutral-300/80 dark:border-neutral-700/80 px-3.5 py-2.5 bg-neutral-50/50 dark:bg-neutral-800/30 text-neutral-900 dark:text-white outline-none transition-all focus:border-blue-500 dark:focus:border-blue-400 focus:ring-2 focus:ring-blue-500/20 resize-none"
                />
              </div>

              {/* Submit */}
              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full relative group overflow-hidden rounded-lg bg-blue-600 py-3 px-4 font-semibold text-white shadow-lg transition-all duration-200 hover:bg-blue-700 active:scale-[0.98] disabled:opacity-85 disabled:pointer-events-none"
              >
                <span className="flex items-center justify-center gap-2">
                  {isSubmitting ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin" />
                      {t("form.submitting")}
                    </>
                  ) : (
                    <>
                      {t("form.submit")}
                      <ChevronRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
                    </>
                  )}
                </span>
              </button>
            </form>
          </motion.div>
        ) : (
          <motion.div
            key="success-card"
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            className="rounded-2xl border border-neutral-200/80 dark:border-neutral-800/80 bg-white dark:bg-neutral-900 p-10 text-center shadow-2xl"
          >
            <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-950/30 text-emerald-600 dark:text-emerald-400 mb-6">
              <CheckCircle2 className="h-8 w-8 animate-pulse" />
            </div>
            <h3 className="text-2xl font-bold text-neutral-900 dark:text-white mb-3">
              {t("form.successTitle")}
            </h3>
            <p className="text-neutral-600 dark:text-neutral-400 leading-relaxed max-w-md mx-auto">
              {t("form.successMessage")}
            </p>
            <button
              onClick={() => setIsSuccess(false)}
              className="mt-8 rounded-lg border border-neutral-300 dark:border-neutral-700 py-2.5 px-6 font-semibold text-neutral-700 dark:text-neutral-300 transition-colors hover:bg-neutral-50 dark:hover:bg-neutral-800/50"
            >
              {t("title")}
            </button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

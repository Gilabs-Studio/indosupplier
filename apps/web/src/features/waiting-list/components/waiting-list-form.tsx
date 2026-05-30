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

const inputClass =
  "w-full bg-white/50 border border-neutral-300 text-neutral-900 placeholder-neutral-400 px-4 py-3 text-[14px] font-jost font-light outline-none transition-all focus:border-neutral-900 focus:bg-white";

const inputErrorClass =
  "w-full bg-white/50 border border-red-500 text-neutral-900 placeholder-neutral-400 px-4 py-3 text-[14px] font-jost font-light outline-none transition-all focus:border-red-500 focus:bg-white";

const labelClass =
  "block text-[12px] uppercase tracking-wider font-jost font-medium text-neutral-500 mb-2";

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
        colors: ["#171717", "#737373", "#a3a3a3"],
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
              <div>
                <label className={labelClass}>{t("form.name")}</label>
                <input
                  type="text"
                  placeholder={t("form.namePlaceholder")}
                  {...register("name")}
                  className={errors.name ? inputErrorClass : inputClass}
                />
                {errors.name && (
                  <span className="text-[11px] text-red-500 mt-1 block font-light">
                    {errors.name.message}
                  </span>
                )}
              </div>

              {/* Email */}
              <div>
                <label className={labelClass}>{t("form.email")}</label>
                <input
                  type="email"
                  placeholder={t("form.emailPlaceholder")}
                  {...register("email")}
                  className={errors.email ? inputErrorClass : inputClass}
                />
                {errors.email && (
                  <span className="text-[11px] text-red-500 mt-1 block font-light">
                    {errors.email.message}
                  </span>
                )}
              </div>

              {/* Company Details */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div>
                  <label className={labelClass}>{t("form.companyName")}</label>
                  <input
                    type="text"
                    placeholder={t("form.companyNamePlaceholder")}
                    {...register("company_name")}
                    className={errors.company_name ? inputErrorClass : inputClass}
                  />
                  {errors.company_name && (
                    <span className="text-[11px] text-red-500 mt-1 block font-light">
                      {errors.company_name.message}
                    </span>
                  )}
                </div>

                <div>
                  <label className={labelClass}>{t("form.companyType")}</label>
                  <div className="relative">
                    <select
                      {...register("company_type")}
                      className={`${errors.company_type ? inputErrorClass : inputClass} appearance-none bg-transparent`}
                    >
                      <option value="supplier">{t("form.supplier")}</option>
                      <option value="buyer">{t("form.buyer")}</option>
                      <option value="other">{t("form.other")}</option>
                    </select>
                    <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-4 text-neutral-500">
                      <svg className="fill-current h-4 w-4" viewBox="0 0 20 20">
                        <path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z" />
                      </svg>
                    </div>
                  </div>
                  {errors.company_type && (
                    <span className="text-[11px] text-red-500 mt-1 block font-light">
                      {errors.company_type.message}
                    </span>
                  )}
                </div>
              </div>

              {/* Phone */}
              <div>
                <label className={labelClass}>{t("form.phone")}</label>
                <input
                  type="text"
                  placeholder={t("form.phonePlaceholder")}
                  {...register("phone")}
                  className={inputClass}
                />
              </div>

              {/* Notes */}
              <div>
                <label className={labelClass}>{t("form.notes")}</label>
                <textarea
                  rows={3}
                  placeholder={t("form.notesPlaceholder")}
                  {...register("notes")}
                  className={`${inputClass} resize-none`}
                />
              </div>

              {/* Submit button */}
              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full bg-neutral-900 text-white py-4 px-6 text-[14px] uppercase tracking-widest font-medium hover:bg-neutral-800 transition-all duration-300 disabled:opacity-50 cursor-pointer"
              >
                {isSubmitting ? t("form.submitting") : t("form.submit")}
              </button>
            </form>
          </motion.div>
        ) : (
          <motion.div
            key="success"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-12"
          >
            <h3 className="text-[24px] font-normal text-neutral-900 mb-4 tracking-tight">
              {t("form.successTitle")}
            </h3>
            <p className="text-[15px] font-light text-neutral-500 leading-relaxed max-w-sm mx-auto mb-8">
              {t("form.successMessage")}
            </p>
            <button
              onClick={() => setIsSuccess(false)}
              className="border border-neutral-300 text-neutral-600 px-6 py-3 text-[13px] uppercase tracking-wider hover:border-neutral-900 hover:text-neutral-900 transition-all duration-300 cursor-pointer"
            >
              {t("title")}
            </button>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}

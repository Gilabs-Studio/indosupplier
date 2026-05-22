"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { motion } from "framer-motion";
import { Mail } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Link } from "@/i18n/routing";
import { AuthLayout } from "./auth-layout";
import { ButtonLoading } from "@/components/loading";
import { getForgotPasswordSchema, type ForgotPasswordFormData } from "../password-reset/schemas/password-reset.schema";
import { useForgotPassword } from "../password-reset/hooks/use-password-reset";

export default function ForgotPasswordForm() {
  const t = useTranslations("passwordReset");
  const [submittedEmail, setSubmittedEmail] = useState<string | null>(null);
  const { mutate: forgotPassword, isPending } = useForgotPassword();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordFormData>({
    resolver: zodResolver(getForgotPasswordSchema(t)),
    defaultValues: {
      email: "",
    },
  });

  const onSubmit = (data: ForgotPasswordFormData) => {
    forgotPassword(data, {
      onSuccess: () => {
        setSubmittedEmail(data.email);
      },
    });
  };

  return (
    <AuthLayout>
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="w-full"
      >
        <Card className="border border-border/60 bg-card/90 shadow-sm">
          <CardHeader className="space-y-2 px-6 pb-2 pt-6">
            <CardTitle className="text-2xl">{t("forgotPassword")}</CardTitle>
            <CardDescription className="text-sm text-muted-foreground">
              {t("forgotPasswordDescription")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-5 px-6 pb-6 pt-2">
            {submittedEmail ? (
              <div className="space-y-4">
                <div className="rounded-md border border-success/20 bg-success/10 px-4 py-3 text-sm text-success">
                  {t("resetLinkSentDescription")}
                </div>
                <div className="space-y-1 text-sm text-muted-foreground">
                  <p>{t("linkExpiresIn")}</p>
                  <p>{t("checkSpamFolder")}</p>
                </div>
                <Link href="/login" className="inline-flex w-full">
                  <Button className="h-11 w-full text-sm font-semibold tracking-wide">
                    {t("backToLogin")}
                  </Button>
                </Link>
              </div>
            ) : (
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
                <FieldGroup className="space-y-4">
                  <Field className="space-y-2">
                    <FieldLabel htmlFor="email">{t("email")}</FieldLabel>
                    <Input
                      id="email"
                      type="email"
                      placeholder={t("emailPlaceholder")}
                      {...register("email")}
                      disabled={isPending}
                      aria-invalid={!!errors.email}
                      className="h-11"
                    />
                    {errors.email && <FieldError>{errors.email.message}</FieldError>}
                  </Field>

                  <Field className="pt-1">
                    <Button
                      type="submit"
                      disabled={isPending}
                      className="h-11 w-full text-sm font-semibold tracking-wide"
                    >
                      <ButtonLoading loading={isPending} loadingText={t("sending")}>
                        {t("sendResetLink")}
                      </ButtonLoading>
                    </Button>
                  </Field>
                </FieldGroup>

                <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
                  <Mail className="h-4 w-4" />
                  <Link href="/login" className="text-primary hover:underline cursor-pointer">
                    {t("backToLogin")}
                  </Link>
                </div>
              </form>
            )}
          </CardContent>
        </Card>
      </motion.div>
    </AuthLayout>
  );
}

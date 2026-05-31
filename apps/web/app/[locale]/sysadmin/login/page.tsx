"use client";

import React, { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { toast } from "sonner";
import { useSysadminStore } from "@/features/sysadmin/stores/use-sysadmin-store";
import { sysadminService } from "@/features/sysadmin/services/sysadmin-service";
import { Loader2, Lock, Mail, ShieldAlert } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel, FieldError } from "@/components/ui/field";
import { ThemeToggleButton } from "@/components/ui/theme-toggle";
import { LightRays } from "@/components/ui/light-rays";

const loginSchema = z.object({
  email: z.string().email("Please enter a valid email address"),
  password: z.string().min(1, "Password is required"),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function SysadminLoginPage() {
  const router = useRouter();
  const { setAdmin } = useSysadminStore();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    setIsSubmitting(true);
    try {
      const payload = await sysadminService.login(data);
      setAdmin(payload.admin);
      toast.success("Successfully logged in as admin!");
      router.push("/sysadmin");
    } catch (error: any) {
      const apiCode = error?.response?.data?.error?.code;
      if (apiCode === "INVALID_CREDENTIALS") {
        toast.error("Invalid admin email or password.");
      } else if (apiCode === "ACCOUNT_DISABLED") {
        toast.error("This administrator account has been disabled.");
      } else {
        toast.error(error?.response?.data?.error?.message || "Login failed. Please try again.");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background px-6 py-12 relative overflow-hidden transition-colors duration-300">
      {/* Light Rays Background for visual wow factor */}
      <LightRays count={5} blur={40} speed={12} className="opacity-70 dark:opacity-40" />

      {/* Grid Pattern */}
      <div className="absolute inset-0 bg-[linear-gradient(to_right,rgba(0,0,0,0.03)_1px,transparent_1px),linear-gradient(to_bottom,rgba(0,0,0,0.03)_1px,transparent_1px)] dark:bg-[linear-gradient(to_right,rgba(255,255,255,0.02)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.02)_1px,transparent_1px)] bg-[size:32px_32px] pointer-events-none" />

      {/* Theme Toggle Button */}
      <div className="absolute top-6 right-6 z-20">
        <ThemeToggleButton className="shadow-md border border-border bg-card hover:bg-muted" />
      </div>

      <div className="w-full max-w-md space-y-8 z-10">
        <div className="text-center space-y-3">
          <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-xl bg-primary/10 border border-primary/20 text-primary shadow-sm">
            <ShieldAlert className="h-7 w-7" />
          </div>
          <h2 className="text-3xl font-extrabold tracking-tight text-foreground font-heading">
            IndoSupplier Admin
          </h2>
          <p className="text-sm text-muted-foreground">
            Sign in to access system configurations & waiting lists
          </p>
        </div>

        <Card className="shadow-lg border border-border/80 bg-card/85 backdrop-blur-md">
          <CardHeader className="pb-2">
            <CardTitle className="text-lg font-bold">Admin Portal</CardTitle>
            <CardDescription className="text-xs">
              Authorized personnel only. Activities are logged.
            </CardDescription>
          </CardHeader>
          
          <CardContent className="pt-4">
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
              {/* Email Field */}
              <Field invalid={!!errors.email}>
                <FieldLabel className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  <Mail className="h-3.5 w-3.5" />
                  Admin Email
                </FieldLabel>
                <div className="relative">
                  <Input
                    type="email"
                    placeholder="admin@indosupplier.com"
                    {...register("email")}
                    className="pr-4"
                  />
                </div>
                {errors.email && (
                  <FieldError className="text-xs font-medium mt-1">
                    {errors.email.message}
                  </FieldError>
                )}
              </Field>

              {/* Password Field */}
              <Field invalid={!!errors.password}>
                <FieldLabel className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                  <Lock className="h-3.5 w-3.5" />
                  Password
                </FieldLabel>
                <div className="relative">
                  <Input
                    type="password"
                    placeholder="••••••••"
                    {...register("password")}
                    className="pr-4"
                  />
                </div>
                {errors.password && (
                  <FieldError className="text-xs font-medium mt-1">
                    {errors.password.message}
                  </FieldError>
                )}
              </Field>

              {/* Submit Button */}
              <Button
                type="submit"
                disabled={isSubmitting}
                className="w-full font-bold uppercase tracking-wider text-xs py-5 rounded-lg bg-primary hover:bg-primary/90 text-primary-foreground transition-all duration-200"
              >
                {isSubmitting ? (
                  <span className="flex items-center justify-center gap-2">
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Signing In...
                  </span>
                ) : (
                  "Sign In to Dashboard"
                )}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

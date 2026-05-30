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
    <div className="min-h-screen flex items-center justify-center bg-neutral-900 px-6 py-12 relative overflow-hidden">
      {/* Background radial glow */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(59,130,246,0.08),transparent_50%)] pointer-events-none" />
      <div className="absolute inset-0 bg-[linear-gradient(to_right,#ffffff02_1px,transparent_1px),linear-gradient(to_bottom,#ffffff02_1px,transparent_1px)] bg-[size:24px_24px] pointer-events-none" />

      <div className="w-full max-w-md space-y-8 z-10">
        <div className="text-center">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-blue-600/10 border border-blue-500/20 text-blue-400 mb-4 shadow-lg shadow-blue-500/5">
            <ShieldAlert className="h-6 w-6" />
          </div>
          <h2 className="text-3xl font-extrabold text-white tracking-tight">
            IndoSupplier Admin Portal
          </h2>
          <p className="mt-2 text-sm text-neutral-400">
            Sign in to access system configurations & waiting lists
          </p>
        </div>

        <div className="bg-neutral-900/60 backdrop-blur-md border border-neutral-800 rounded-2xl p-8 shadow-2xl">
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            {/* Email */}
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-400 mb-2 flex items-center gap-1.5">
                <Mail className="h-3.5 w-3.5 text-neutral-500" />
                Admin Email
              </label>
              <input
                type="email"
                placeholder="admin@indosupplier.com"
                {...register("email")}
                className={`w-full rounded-lg border bg-neutral-950 px-3.5 py-2.5 text-white outline-none transition-all focus:ring-2 focus:ring-blue-500/20 ${
                  errors.email
                    ? "border-red-500/80 focus:border-red-500"
                    : "border-neutral-800 focus:border-blue-500"
                }`}
              />
              {errors.email && (
                <span className="text-xs font-medium text-red-400 mt-1 block">
                  {errors.email.message}
                </span>
              )}
            </div>

            {/* Password */}
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-neutral-400 mb-2 flex items-center gap-1.5">
                <Lock className="h-3.5 w-3.5 text-neutral-500" />
                Password
              </label>
              <input
                type="password"
                placeholder="••••••••"
                {...register("password")}
                className={`w-full rounded-lg border bg-neutral-950 px-3.5 py-2.5 text-white outline-none transition-all focus:ring-2 focus:ring-blue-500/20 ${
                  errors.password
                    ? "border-red-500/80 focus:border-red-500"
                    : "border-neutral-800 focus:border-blue-500"
                }`}
              />
              {errors.password && (
                <span className="text-xs font-medium text-red-400 mt-1 block">
                  {errors.password.message}
                </span>
              )}
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full rounded-lg bg-blue-600 hover:bg-blue-700 py-3 px-4 font-semibold text-white shadow-lg shadow-blue-600/20 transition-all duration-200 active:scale-[0.98] disabled:opacity-80 disabled:pointer-events-none"
            >
              <span className="flex items-center justify-center gap-2">
                {isSubmitting ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin" />
                    Signing In...
                  </>
                ) : (
                  "Sign In to Dashboard"
                )}
              </span>
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}

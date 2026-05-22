"use client";

import { useTranslations } from "next-intl";
import { motion } from "framer-motion";
import {
  CheckCircle2,
  Circle,
  ChevronRight,
  Store,
  Package,
  type LucideIcon,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Link } from "@/i18n/routing";
import {
  useOnboardingState,
  useSetBusinessType,
  useMarkOnboardingComplete,
} from "../hooks/use-onboarding";
import type { OnboardingSteps } from "../services/onboarding-service";

interface ChecklistStep {
  key: keyof OnboardingSteps;
  labelKey: string;
  descriptionKey: string;
  href: string;
}

const FNB_CHECKLIST: ChecklistStep[] = [
  {
    key: "company",
    labelKey: "onboarding.steps.company.label",
    descriptionKey: "onboarding.steps.company.description",
    href: "/master-data/company",
  },
  {
    key: "fiscal_year",
    labelKey: "onboarding.steps.fiscalYear.label",
    descriptionKey: "onboarding.steps.fiscalYear.description",
    href: "/finance/settings/fiscal-years",
  },
  {
    key: "outlet",
    labelKey: "onboarding.steps.outlet.label",
    descriptionKey: "onboarding.steps.outlet.description",
    href: "/master-data/outlet",
  },
  {
    key: "floor_layout",
    labelKey: "onboarding.steps.floorLayout.label",
    descriptionKey: "onboarding.steps.floorLayout.description",
    href: "/pos/fb/floor-layout",
  },
  {
    key: "products",
    labelKey: "onboarding.steps.products.label",
    descriptionKey: "onboarding.steps.products.description",
    href: "/master-data/products",
  },
  {
    key: "users",
    labelKey: "onboarding.steps.users.label",
    descriptionKey: "onboarding.steps.users.description",
    href: "/master-data/users",
  },
];

const OTHER_CHECKLIST: ChecklistStep[] = [
  {
    key: "company",
    labelKey: "onboarding.steps.company.label",
    descriptionKey: "onboarding.steps.company.description",
    href: "/master-data/company",
  },
  {
    key: "fiscal_year",
    labelKey: "onboarding.steps.fiscalYear.label",
    descriptionKey: "onboarding.steps.fiscalYear.description",
    href: "/finance/settings/fiscal-years",
  },
  {
    key: "warehouse",
    labelKey: "onboarding.steps.warehouse.label",
    descriptionKey: "onboarding.steps.warehouse.description",
    href: "/master-data/warehouses",
  },
  {
    key: "products",
    labelKey: "onboarding.steps.products.label",
    descriptionKey: "onboarding.steps.products.description",
    href: "/master-data/products",
  },
  {
    key: "users",
    labelKey: "onboarding.steps.users.label",
    descriptionKey: "onboarding.steps.users.description",
    href: "/master-data/users",
  },
];

/** Full-page onboarding experience shown inside the dashboard before any widgets appear. */
export function OnboardingWizard() {
  const t = useTranslations("dashboard");
  const { data: state, isLoading } = useOnboardingState();
  const { mutate: setBusinessType, isPending: isSettingType } =
    useSetBusinessType();
  const { mutate: markComplete, isPending: isCompleting } =
    useMarkOnboardingComplete();

  if (isLoading) {
    return <OnboardingWizardSkeleton />;
  }

  if (state?.completed) return null;

  const businessType = state?.business_type ?? "";
  const steps = state?.steps;
  const checklist = businessType === "fnb" ? FNB_CHECKLIST : OTHER_CHECKLIST;
  const completedCount = steps
    ? checklist.filter((s) => steps[s.key]).length
    : 0;
  const allDone = checklist.length > 0 && completedCount === checklist.length;

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="flex w-full justify-center px-4 py-6"
    >
      <div className="w-full max-w-[440px] max-h-[calc(100dvh-3rem)] overflow-y-auto overscroll-contain pr-1">
        {!businessType ? (
          <StepChooseType
            t={t}
            isLoading={isSettingType}
            onSelect={setBusinessType}
          />
        ) : (
          <StepChecklist
            t={t}
            checklist={checklist}
            steps={steps}
            completedCount={completedCount}
            allDone={allDone}
            isCompleting={isCompleting}
            businessType={businessType}
            onComplete={() => markComplete()}
          />
        )}
      </div>
    </motion.div>
  );
}

function OnboardingWizardSkeleton() {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="flex w-full justify-center px-4 py-6"
    >
      <div className="w-full max-w-[440px] max-h-[calc(100dvh-3rem)] overflow-y-auto overscroll-contain pr-1">
        <div className="rounded-2xl border bg-card/80 p-6 shadow-sm backdrop-blur-sm">
          <div className="mb-8 text-center">
            <Skeleton className="mx-auto h-7 w-44" />
            <Skeleton className="mx-auto mt-3 h-4 w-60" />
          </div>

          <div className="space-y-3">
            {Array.from({ length: 2 }).map((_, index) => (
              <div
                key={index}
                className="flex items-center gap-4 rounded-xl border bg-background/60 px-5 py-4"
              >
                <Skeleton className="h-10 w-10 shrink-0 rounded-lg" />
                <div className="min-w-0 flex-1 space-y-2">
                  <Skeleton className="h-4 w-28" />
                  <Skeleton className="h-3 w-40" />
                </div>
                <Skeleton className="h-4 w-4 shrink-0 rounded-full" />
              </div>
            ))}
          </div>
        </div>
      </div>
    </motion.div>
  );
}

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

interface StepChooseTypeProps {
  t: ReturnType<typeof useTranslations<"dashboard">>;
  isLoading: boolean;
  onSelect: (type: string) => void;
}

function StepChooseType({ t, isLoading, onSelect }: StepChooseTypeProps) {
  return (
    <>
      <div className="mb-10 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          {t("onboarding.welcome")}
        </h1>
        <p className="text-muted-foreground mt-2 text-sm leading-relaxed">
          {t("onboarding.chooseTypePrompt")}
        </p>
      </div>

      <div className="space-y-3">
        <TypeButton
          Icon={Store}
          label={t("onboarding.types.fnb.label")}
          description={t("onboarding.types.fnb.description")}
          disabled={isLoading}
          onClick={() => onSelect("fnb")}
        />
        <TypeButton
          Icon={Package}
          label={t("onboarding.types.other.label")}
          description={t("onboarding.types.other.description")}
          disabled={isLoading}
          onClick={() => onSelect("other")}
        />
      </div>
    </>
  );
}

interface StepChecklistProps {
  t: ReturnType<typeof useTranslations<"dashboard">>;
  checklist: ChecklistStep[];
  steps: OnboardingSteps | undefined;
  completedCount: number;
  allDone: boolean;
  isCompleting: boolean;
  businessType: string;
  onComplete: () => void;
}

function StepChecklist({
  t,
  checklist,
  steps,
  completedCount,
  allDone,
  isCompleting,
  businessType,
  onComplete,
}: StepChecklistProps) {
  return (
    <>
      <div className="mb-2 text-center">
        <h1 className="text-2xl font-semibold tracking-tight">
          {t("onboarding.setupChecklist")}
        </h1>
        <p className="text-muted-foreground mt-2 text-sm">
          {t("onboarding.checklistPrompt", {
            done: completedCount,
            total: checklist.length,
          })}
        </p>
      </div>

      {/* Progress bar */}
      <div className="mb-8 mt-4 h-1 w-full rounded-full bg-muted">
        <motion.div
          className="h-1 rounded-full bg-primary"
          initial={{ width: 0 }}
          animate={{
            width: `${(completedCount / Math.max(checklist.length, 1)) * 100}%`,
          }}
          transition={{ duration: 0.5, ease: "easeOut" }}
        />
      </div>

      <div className="space-y-0.5">
        {checklist.map((step, idx) => {
          const isDone = steps?.[step.key] ?? false;
          return (
            <motion.div
              key={step.key}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: idx * 0.07 }}
            >
              <Link href={step.href}>
                <div
                  className={`group flex cursor-pointer items-center gap-4 rounded-lg px-3 py-3.5 transition-colors ${
                    isDone ? "opacity-50" : "hover:bg-accent"
                  }`}
                >
                  <span className="flex h-5 w-5 shrink-0 items-center justify-center">
                    {isDone ? (
                      <CheckCircle2 className="h-5 w-5 text-primary" />
                    ) : (
                      <Circle className="text-muted-foreground h-5 w-5" />
                    )}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p
                      className={`text-sm font-medium leading-none ${isDone ? "line-through" : ""}`}
                    >
                      {t(step.labelKey as Parameters<typeof t>[0])}
                    </p>
                    <p className="text-muted-foreground mt-1 text-xs">
                      {t(step.descriptionKey as Parameters<typeof t>[0])}
                    </p>
                  </div>
                  {!isDone && (
                    <ChevronRight className="text-muted-foreground h-4 w-4 shrink-0 opacity-0 transition-all group-hover:translate-x-0.5 group-hover:opacity-100" />
                  )}
                </div>
              </Link>
            </motion.div>
          );
        })}
      </div>

      <div className="mt-5 rounded-lg border bg-muted/20 p-3">
        <p className="text-sm font-medium">
          {t("onboarding.accessSetup.title")}
        </p>
        <p className="text-muted-foreground mt-1 text-xs">
          {t("onboarding.accessSetup.description")}
        </p>
        <div className="mt-3 space-y-1.5">
          <Link href="/master-data/users?tab=roles">
            <div className="group flex cursor-pointer items-center gap-2 rounded-md px-2 py-2 hover:bg-accent">
              <Circle className="text-muted-foreground h-4 w-4" />
              <span className="text-sm">{t("onboarding.accessSetup.rolePermissions")}</span>
              <ChevronRight className="text-muted-foreground ml-auto h-4 w-4 opacity-0 transition-all group-hover:translate-x-0.5 group-hover:opacity-100" />
            </div>
          </Link>
          <Link href="/master-data/employees">
            <div className="group flex cursor-pointer items-center gap-2 rounded-md px-2 py-2 hover:bg-accent">
              <Circle className="text-muted-foreground h-4 w-4" />
              <span className="text-sm">
                {businessType === "fnb"
                  ? t("onboarding.accessSetup.employeeAccessFnb")
                  : t("onboarding.accessSetup.employeeAccessOther")}
              </span>
              <ChevronRight className="text-muted-foreground ml-auto h-4 w-4 opacity-0 transition-all group-hover:translate-x-0.5 group-hover:opacity-100" />
            </div>
          </Link>
        </div>
      </div>

      <div className="mt-8 space-y-2 text-center">
        {allDone ? (
          <Button
            className="w-full"
            disabled={isCompleting}
            onClick={onComplete}
          >
            {t("onboarding.markDone")}
          </Button>
        ) : (
          <button
            type="button"
            disabled={isCompleting}
            onClick={onComplete}
            className="text-muted-foreground cursor-pointer text-xs underline-offset-4 hover:underline disabled:opacity-40"
          >
            {t("onboarding.dismiss")}
          </button>
        )}
      </div>
    </>
  );
}

interface TypeButtonProps {
  Icon: LucideIcon;
  label: string;
  description: string;
  disabled: boolean;
  onClick: () => void;
}

function TypeButton({ Icon, label, description, disabled, onClick }: TypeButtonProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="group flex w-full items-center gap-4 rounded-xl border bg-card px-5 py-4 text-left shadow-sm transition-all hover:border-primary/40 hover:shadow-md disabled:opacity-60"
    >
      <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-muted text-foreground transition-colors group-hover:bg-primary/10 group-hover:text-primary">
        <Icon className="h-5 w-5" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold">{label}</p>
        <p className="text-muted-foreground mt-0.5 text-xs">{description}</p>
      </div>
      <ChevronRight className="text-muted-foreground h-4 w-4 shrink-0 transition-transform group-hover:translate-x-0.5" />
    </button>
  );
}
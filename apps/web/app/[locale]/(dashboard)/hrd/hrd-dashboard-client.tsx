"use client";

import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { PageMotion } from "@/components/motion/page-motion";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Clock,
  Calendar,
  CalendarDays,
  Timer,
  Users,
  TrendingUp,
  ArrowRight,
  ClipboardList,
  UserCheck,
  Briefcase,
} from "lucide-react";
import { cn } from "@/lib/utils";

const hrdModules = [
  {
    id: "attendance",
    icon: Clock,
    href: "/hrd/attendance",
    color: "text-primary",
    bgColor: "bg-primary/10",
    stats: { label: "todayPresent", value: "-" },
  },
  {
    id: "workSchedules",
    icon: Calendar,
    href: "/hrd/work-schedules",
    color: "text-success",
    bgColor: "bg-success/10",
    stats: { label: "activeSchedules", value: "-" },
  },
  {
    id: "holidays",
    icon: CalendarDays,
    href: "/hrd/holidays",
    color: "text-warning",
    bgColor: "bg-warning/10",
    stats: { label: "thisYear", value: "-" },
  },
  {
    id: "overtime",
    icon: Timer,
    href: "/hrd/overtime",
    color: "text-accent",
    bgColor: "bg-accent/10",
    stats: { label: "pendingApproval", value: "-" },
  },
  {
    id: "leaves",
    icon: Briefcase,
    href: "/hrd/leaves",
    color: "text-accent",
    bgColor: "bg-accent/10",
    stats: { label: "pendingRequests", value: "-" },
    comingSoon: true,
  },
  {
    id: "employees",
    icon: Users,
    href: "/master-data/employees",
    color: "text-primary",
    bgColor: "bg-primary/10",
    stats: { label: "totalActive", value: "-" },
    comingSoon: true,
  },
];

const quickStats = [
  {
    label: "totalEmployees",
    value: "-",
    change: "",
    icon: Users,
    color: "text-primary",
  },
  {
    label: "presentToday",
    value: "-",
    change: "",
    icon: UserCheck,
    color: "text-success",
  },
  {
    label: "pendingOvertime",
    value: "-",
    change: "",
    icon: Timer,
    color: "text-warning",
  },
  {
    label: "upcomingHolidays",
    value: "-",
    change: "",
    icon: CalendarDays,
    color: "text-accent",
  },
];

export default function HrdDashboardClient() {
  const t = useTranslations("hrd");

  return (
    <PageMotion>
      <div className="space-y-8">
        {/* Header */}
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{t("title")}</h1>
          <p className="text-muted-foreground mt-2">{t("description")}</p>
        </div>

        {/* Quick Stats */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {quickStats.map((stat) => {
            const Icon = stat.icon;
            return (
              <Card key={stat.label}>
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    {t(`dashboard.${stat.label}`)}
                  </CardTitle>
                  <Icon className={cn("h-5 w-5", stat.color)} />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{stat.value}</div>
                  {stat.change && (
                    <p className="text-xs text-muted-foreground flex items-center gap-1 mt-1">
                      <TrendingUp className="h-3 w-3 text-success" />
                      {stat.change}
                    </p>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>

        {/* Module Navigation */}
        <div>
          <h2 className="text-xl font-semibold mb-4">{t("dashboard.modules")}</h2>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {hrdModules.map((module) => {
              const Icon = module.icon;
              const content = (
                <Card
                  className={cn(
                    "transition-all duration-200 group",
                    module.comingSoon
                      ? "opacity-60"
                      : "hover:shadow-md hover:border-primary/50 cursor-pointer"
                  )}
                >
                  <CardHeader>
                    <div className="flex items-center gap-4">
                      <div
                        className={cn(
                          "flex h-12 w-12 items-center justify-center rounded-lg",
                          module.bgColor
                        )}
                      >
                        <Icon className={cn("h-6 w-6", module.color)} />
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <CardTitle className="text-lg">
                            {t(`modules.${module.id}.title`)}
                          </CardTitle>
                          {module.comingSoon && (
                            <Badge variant="secondary" className="text-xs">
                              Coming Soon
                            </Badge>
                          )}
                        </div>
                        <CardDescription>
                          {t(`modules.${module.id}.description`)}
                        </CardDescription>
                      </div>
                      {!module.comingSoon && (
                        <ArrowRight className="h-5 w-5 text-muted-foreground group-hover:text-primary group-hover:translate-x-1 transition-all" />
                      )}
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        {t(`dashboard.stats.${module.stats.label}`)}
                      </span>
                      <span className="font-medium">{module.stats.value}</span>
                    </div>
                  </CardContent>
                </Card>
              );

              if (module.comingSoon) {
                return <div key={module.id}>{content}</div>;
              }

              return (
                <Link key={module.id} href={module.href} className="block">
                  {content}
                </Link>
              );
            })}
          </div>
        </div>

        {/* Recent Activity Section - Placeholder */}
        <div>
          <h2 className="text-xl font-semibold mb-4">{t("dashboard.recentActivity")}</h2>
          <Card>
            <CardContent className="py-8">
              <div className="flex flex-col items-center justify-center text-center">
                <ClipboardList className="h-12 w-12 text-muted-foreground/50 mb-4" />
                <p className="text-muted-foreground">
                  {t("dashboard.noRecentActivity")}
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </PageMotion>
  );
}

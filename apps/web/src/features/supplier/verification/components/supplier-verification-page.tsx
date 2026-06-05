"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ShieldCheck, CheckCircle2, Clock, MapPin, Building, Sparkles } from "lucide-react";

export function SupplierVerificationPage() {
  const t = useTranslations("supplier.verification");

  const steps = [
    { id: 1, label: t("stepEmail"), desc: "Primary corporate email address confirmed.", status: "complete" },
    { id: 2, label: t("stepDocs"), desc: "Valid corporate tax registration and NIB business licenses uploaded.", status: "complete" },
    { id: 3, label: t("stepReview"), desc: "Compliance administrators are inspecting document credentials validity.", status: "current" },
    { id: 4, label: t("stepVisit"), desc: "Physical on-site inspection visit to factory or warehouse facilities.", status: "pending" },
  ];

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="border-b border-border/80 pb-6">
        <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
          {t("title")}
        </h1>
        <p className="text-sm text-muted-foreground">
          {t("subtitle")}
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Timeline checklist */}
        <div className="lg:col-span-2 space-y-6">
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader className="border-b border-border bg-muted/20 py-4 px-6">
              <CardTitle className="text-sm font-bold text-foreground">{t("progress")}</CardTitle>
            </CardHeader>
            <CardContent className="p-6">
              <div className="space-y-8 relative before:absolute before:left-[17px] before:top-2 before:bottom-2 before:w-0.5 before:bg-border">
                {steps.map((step) => (
                  <div key={step.id} className="flex gap-4 relative">
                    <div className="shrink-0 z-10">
                      {step.status === "complete" ? (
                        <div className="h-9 w-9 rounded-full bg-success/20 text-success border border-success/30 flex items-center justify-center font-bold text-sm">
                          <CheckCircle2 className="h-4.5 w-4.5" />
                        </div>
                      ) : step.status === "current" ? (
                        <div className="h-9 w-9 rounded-full bg-warning/20 text-warning border border-warning/30 flex items-center justify-center font-bold text-sm">
                          <Clock className="h-4.5 w-4.5 animate-spin" />
                        </div>
                      ) : (
                        <div className="h-9 w-9 rounded-full bg-muted text-muted-foreground border border-border flex items-center justify-center font-bold text-xs">
                          {step.id}
                        </div>
                      )}
                    </div>
                    <div className="space-y-1 pt-1">
                      <h4 className="text-sm font-bold text-foreground flex items-center gap-2">
                        {step.label}
                        {step.status === "current" && (
                          <Badge className="bg-warning/15 text-warning border border-warning/30 text-[9px] font-bold uppercase">
                            Processing
                          </Badge>
                        )}
                      </h4>
                      <p className="text-xs text-muted-foreground leading-relaxed font-semibold">
                        {step.desc}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Verification Status Banner */}
        <div className="space-y-6">
          <Card className="border border-warning/30 bg-warning/5 rounded-xl text-left overflow-hidden">
            <CardHeader className="border-b border-warning/20 bg-warning/10 py-4 px-6">
              <CardTitle className="text-sm font-bold text-warning flex items-center gap-1.5">
                <Clock className="h-4.5 w-4.5 shrink-0" />
                Under Evaluation
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6 space-y-3">
              <p className="text-xs font-semibold text-muted-foreground leading-relaxed">
                Your business legal documents are currently undergoing review. Once verified, a local inspector will schedule a facility tour.
              </p>
              <div className="h-px bg-warning/20" />
              <div className="flex items-center gap-2 text-xs font-bold text-warning">
                <Sparkles className="h-4 w-4 shrink-0" />
                ETA: 3 business days
              </div>
            </CardContent>
          </Card>

          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-sm font-bold font-heading">Verified Badge Perks</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-xs font-semibold text-muted-foreground leading-relaxed">
              <p className="flex items-start gap-1.5">
                <ShieldCheck className="h-4.5 w-4.5 text-success shrink-0" />
                <span>Gain the Gold Badge label displayed on your search listings and catalogs.</span>
              </p>
              <p className="flex items-start gap-1.5">
                <Building className="h-4.5 w-4.5 text-success shrink-0" />
                <span>Attract multi-national enterprise buyers requiring strict compliance checks.</span>
              </p>
              <p className="flex items-start gap-1.5">
                <MapPin className="h-4.5 w-4.5 text-success shrink-0" />
                <span>Higher search positioning ranking than non-verified profiles.</span>
              </p>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

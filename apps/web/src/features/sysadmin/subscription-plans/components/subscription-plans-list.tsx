"use client";

import React, { useEffect, useState, useCallback } from "react";
import { subscriptionPlanService } from "@/features/sysadmin/subscription-plans/services";
import type { SubscriptionPlan } from "@/features/sysadmin/subscription-plans/types";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  Trash2,
  RefreshCw,
  Inbox,
  Plus,
  CreditCard,
  Edit2,
  Check,
  X,
  Loader2,
  CheckCircle2,
  ShieldCheck,
  Zap
} from "lucide-react";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogFooter } from "@/components/ui/dialog";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function SubscriptionPlansList() {
  const t = useTranslations("sysadminSubscriptionPlans");
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Filters state
  const [selectedCycle, setSelectedCycle] = useState<string>("all");

  // Create plan dialog state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [name, setName] = useState("");
  const [price, setPrice] = useState(0);
  const [billingCycle, setBillingCycle] = useState<"monthly" | "annually">("monthly");
  const [tier, setTier] = useState<SubscriptionPlan["tier"]>("basic");
  const [featuresInput, setFeaturesInput] = useState("");

  // Inline edit price state
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editPriceVal, setEditPriceVal] = useState<number>(0);

  const fetchPlans = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await subscriptionPlanService.list();
      setPlans(data);
    } catch {
      toast.error(t("subtitle"));
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchPlans();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchPlans]);

  const handleToggleActive = async (id: string, currentVal: boolean) => {
    try {
      await subscriptionPlanService.update(id, { active: !currentVal });
      toast.success(t("successSave"));
      fetchPlans();
    } catch {
      toast.error("Error");
    }
  };

  const handleSavePrice = async (id: string) => {
    try {
      await subscriptionPlanService.update(id, { price: editPriceVal });
      toast.success(t("successSave"));
      setEditingId(null);
      fetchPlans();
    } catch {
      toast.error("Error");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("delete") + "?")) return;
    try {
      await subscriptionPlanService.delete(id);
      toast.success(t("successDelete"));
      fetchPlans();
    } catch {
      toast.error("Error");
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error("Plan name required");
      return;
    }

    const featuresList = featuresInput
      .split("\n")
      .map((f) => f.trim())
      .filter((f) => f.length > 0);

    try {
      await subscriptionPlanService.create({
        name,
        price: Number(price),
        billingCycle,
        features: featuresList,
        active: true,
        tier
      });
      toast.success(t("successSave"));
      setIsCreateOpen(false);
      setName("");
      setPrice(0);
      setBillingCycle("monthly");
      setTier("basic");
      setFeaturesInput("");
      fetchPlans();
    } catch {
      toast.error("Error");
    }
  };

  // Filters logic
  const filteredPlans = plans.filter((plan) => {
    return selectedCycle === "all" || plan.billingCycle === selectedCycle;
  });

  const formatIDRCurrency = (val: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      maximumFractionDigits: 0
    }).format(val);
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            onClick={() => setIsCreateOpen(true)}
            className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold text-xs uppercase tracking-wider py-5 px-4 rounded-lg flex items-center gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md cursor-pointer"
          >
            <Plus className="h-4 w-4" />
            {t("newPlan")}
          </Button>
          <Button
            onClick={fetchPlans}
            variant="outline"
            size="sm"
            className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
            {t("refresh")}
          </Button>
        </div>
      </div>

      {/* Filter Bar */}
      <Card className="p-4 border border-border bg-card flex items-center justify-between">
        <div className="flex items-center gap-2 text-left">
          <CreditCard className="h-4 w-4 text-primary" />
          <span className="text-xs text-muted-foreground font-semibold">{t("filterCycle")}</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setSelectedCycle("all")}
            className={`px-3 py-1.5 text-xs font-bold rounded-lg border transition-all cursor-pointer ${
              selectedCycle === "all"
                ? "bg-primary border-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
                : "bg-background border-border text-muted-foreground hover:bg-muted/40"
            }`}
          >
            {t("allCycles")}
          </button>
          <button
            onClick={() => setSelectedCycle("monthly")}
            className={`px-3 py-1.5 text-xs font-bold rounded-lg border transition-all cursor-pointer ${
              selectedCycle === "monthly"
                ? "bg-primary border-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
                : "bg-background border-border text-muted-foreground hover:bg-muted/40"
            }`}
          >
            {t("monthly")}
          </button>
          <button
            onClick={() => setSelectedCycle("annually")}
            className={`px-3 py-1.5 text-xs font-bold rounded-lg border transition-all cursor-pointer ${
              selectedCycle === "annually"
                ? "bg-primary border-primary text-primary-foreground hover:bg-primary/90 shadow-sm"
                : "bg-background border-border text-muted-foreground hover:bg-muted/40"
            }`}
          >
            {t("annually")}
          </button>
        </div>
      </Card>

      {/* Plans List Table */}
      <Card className="border border-border/80 shadow-sm overflow-hidden bg-card">
        {isLoading ? (
          <div className="py-24 text-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
            <span className="text-sm font-semibold text-muted-foreground">Loading...</span>
          </div>
        ) : filteredPlans.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">Empty</h3>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("planName")}</TableHead>
                <TableHead className="font-semibold">{t("cycle")}</TableHead>
                <TableHead className="font-semibold">{t("price")}</TableHead>
                <TableHead className="font-semibold">{t("features")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredPlans.map((plan) => (
                <TableRow key={plan.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  <TableCell className="pl-6 py-4">
                    <div className="text-left space-y-1">
                      <div className="font-bold text-foreground flex items-center gap-1.5">
                        <Zap className="h-4 w-4 text-primary shrink-0" />
                        {plan.name}
                      </div>
                      <div>
                        <Badge className="uppercase text-[9px] font-bold px-1.5 py-0.2 bg-secondary text-secondary-foreground border-none">
                          {plan.tier}
                        </Badge>
                      </div>
                    </div>
                  </TableCell>

                  <TableCell className="py-4 text-left font-semibold text-xs capitalize text-muted-foreground">
                    {plan.billingCycle === "monthly" ? t("monthly") : t("annually")}
                  </TableCell>

                  <TableCell className="py-4 text-left">
                    {editingId === plan.id ? (
                      <div className="flex items-center gap-1.5">
                        <Input
                          type="number"
                          value={editPriceVal}
                          onChange={(e) => setEditPriceVal(Number(e.target.value))}
                          className="w-28 h-8 text-xs px-2 border-border"
                          min={0}
                        />
                        <button
                          onClick={() => handleSavePrice(plan.id)}
                          className="h-8 w-8 flex items-center justify-center bg-primary text-primary-foreground rounded hover:bg-primary/90 cursor-pointer"
                        >
                          <Check className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => setEditingId(null)}
                          className="h-8 w-8 flex items-center justify-center bg-muted text-muted-foreground rounded hover:bg-muted/90 cursor-pointer"
                        >
                          <X className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    ) : (
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-foreground text-sm">
                          {formatIDRCurrency(plan.price)}
                        </span>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => {
                            setEditingId(plan.id);
                            setEditPriceVal(plan.price);
                          }}
                          className="h-6 w-6 text-muted-foreground hover:text-primary cursor-pointer"
                        >
                          <Edit2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    )}
                  </TableCell>

                  <TableCell className="py-4 text-left max-w-xs">
                    <div className="space-y-1">
                      {plan.features.slice(0, 3).map((feat, idx) => (
                        <div key={idx} className="flex items-center gap-1 text-[11px] text-muted-foreground">
                          <CheckCircle2 className="h-3 w-3 text-success shrink-0" />
                          <span className="truncate">{feat}</span>
                        </div>
                      ))}
                      {plan.features.length > 3 && (
                        <span className="text-[9px] text-muted-foreground/60 italic font-semibold pl-4">
                          +{plan.features.length - 3} features
                        </span>
                      )}
                    </div>
                  </TableCell>

                  <TableCell className="py-4 text-left">
                    <button
                      onClick={() => handleToggleActive(plan.id, plan.active)}
                      className={`px-3 py-1 text-xs font-semibold rounded-full border cursor-pointer select-none transition-all ${
                        plan.active
                          ? "bg-success/15 text-success border-success/30"
                          : "bg-muted text-muted-foreground border-border"
                      }`}
                    >
                      {plan.active ? t("active") : t("cancel")}
                    </button>
                  </TableCell>

                  <TableCell className="pr-6 py-4 text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(plan.id)}
                      className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 cursor-pointer"
                      title={t("delete")}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </Card>

      {/* Create Plan Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-primary font-bold mb-1">
              <ShieldCheck className="h-5 w-5" />
              {t("createPlan")}
            </div>
          </DialogHeader>

          <form onSubmit={handleCreate} className="space-y-4 my-2 text-left">
            <FieldGroup className="space-y-3">
              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("planName")}</FieldLabel>
                <Input
                  type="text"
                  placeholder="Plan Name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="bg-background border-border"
                />
              </Field>

              <div className="grid grid-cols-2 gap-3">
                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("price")} (IDR)</FieldLabel>
                  <Input
                    type="number"
                    placeholder="Price"
                    value={price}
                    onChange={(e) => setPrice(Number(e.target.value))}
                    className="bg-background border-border"
                    min={0}
                  />
                </Field>

                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("cycle")}</FieldLabel>
                  <Select
                    value={billingCycle}
                    onValueChange={(val: "monthly" | "annually") => setBillingCycle(val)}
                  >
                    <SelectTrigger className="bg-background border-border cursor-pointer">
                      <SelectValue placeholder="Siklus" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="monthly" className="cursor-pointer">{t("monthly")}</SelectItem>
                      <SelectItem value="annually" className="cursor-pointer">{t("annually")}</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("planTier")}</FieldLabel>
                  <Select
                    value={tier}
                    onValueChange={(val: SubscriptionPlan["tier"]) => setTier(val)}
                  >
                    <SelectTrigger className="bg-background border-border cursor-pointer">
                      <SelectValue placeholder="Tier" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="free" className="cursor-pointer">Free</SelectItem>
                      <SelectItem value="basic" className="cursor-pointer">Basic</SelectItem>
                      <SelectItem value="premium" className="cursor-pointer">Premium</SelectItem>
                      <SelectItem value="enterprise" className="cursor-pointer">Enterprise</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
              </div>

              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("featuresInput")}</FieldLabel>
                <textarea
                  placeholder="Features..."
                  value={featuresInput}
                  onChange={(e) => setFeaturesInput(e.target.value)}
                  rows={4}
                  className="w-full bg-background border border-border rounded-lg p-2.5 text-sm placeholder-muted-foreground focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-all"
                />
              </Field>
            </FieldGroup>

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)} className="border-border cursor-pointer text-foreground">
                {t("cancel")}
              </Button>
              <Button type="submit" className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-transform">
                {t("save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

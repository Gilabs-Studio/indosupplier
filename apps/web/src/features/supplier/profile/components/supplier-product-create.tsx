"use client";

import React, { useState } from "react";
import { useRouter } from "@/i18n/routing";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { toast } from "sonner";
import { ArrowLeft, Save, Upload } from "lucide-react";

export function SupplierProductCreate() {
  const router = useRouter();
  const [isSaving, setIsSaving] = useState(false);

  const [form, setForm] = useState({
    name: "",
    category: "Industrial Minerals",
    moq: "",
    price: "",
    capacity: "",
    sku: "",
    description: "",
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    setTimeout(() => {
      setIsSaving(false);
      toast.success("Product added to catalog successfully!");
      router.push("/supplier/profile/products");
    }, 1000);
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="flex items-center gap-4 border-b border-border/80 pb-6">
        <Button
          variant="outline"
          size="icon"
          onClick={() => router.push("/supplier/profile/products")}
          className="h-9 w-9 cursor-pointer border-border"
        >
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            Add New Product
          </h1>
          <p className="text-sm text-muted-foreground">
            List a new product item in your wholesale marketplace catalog.
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Main Details */}
        <div className="lg:col-span-2 space-y-6">
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-base font-bold font-heading">Product Details</CardTitle>
              <CardDescription className="text-xs">Specify the main attributes of the product.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Field>
                <FieldLabel>Product Name</FieldLabel>
                <Input
                  placeholder="e.g. Garnet Sand Mesh 80"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  required
                />
              </Field>

              <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Field>
                  <FieldLabel>Category</FieldLabel>
                  <Input
                    placeholder="e.g. Industrial Minerals"
                    value={form.category}
                    onChange={(e) => setForm({ ...form, category: e.target.value })}
                    required
                  />
                </Field>
                <Field>
                  <FieldLabel>SKU / Item Code</FieldLabel>
                  <Input
                    placeholder="e.g. GRN-80-IND"
                    value={form.sku}
                    onChange={(e) => setForm({ ...form, sku: e.target.value })}
                  />
                </Field>
              </FieldGroup>

              <Field>
                <FieldLabel>Description</FieldLabel>
                <Textarea
                  placeholder="Describe your product specifications, quality grades, applications..."
                  rows={6}
                  value={form.description}
                  onChange={(e) => setForm({ ...form, description: e.target.value })}
                  required
                  className="resize-none"
                />
              </Field>
            </CardContent>
          </Card>

          {/* Pricing & Supply Terms */}
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-base font-bold font-heading">Pricing & Terms</CardTitle>
              <CardDescription className="text-xs">Wholesale quantities and transaction parameters.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <FieldGroup className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                <Field>
                  <FieldLabel>Price Terms (per Unit)</FieldLabel>
                  <Input
                    placeholder="e.g. Rp 3.500.000 / Ton"
                    value={form.price}
                    onChange={(e) => setForm({ ...form, price: e.target.value })}
                    required
                  />
                </Field>
                <Field>
                  <FieldLabel>Minimum Order Quantity</FieldLabel>
                  <Input
                    placeholder="e.g. 20 Ton"
                    value={form.moq}
                    onChange={(e) => setForm({ ...form, moq: e.target.value })}
                    required
                  />
                </Field>
                <Field>
                  <FieldLabel>Supply Capacity</FieldLabel>
                  <Input
                    placeholder="e.g. 500 Ton / Month"
                    value={form.capacity}
                    onChange={(e) => setForm({ ...form, capacity: e.target.value })}
                    required
                  />
                </Field>
              </FieldGroup>
            </CardContent>
          </Card>
        </div>

        {/* Sidebar Controls */}
        <div className="space-y-6">
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-base font-bold font-heading">Product Media</CardTitle>
              <CardDescription className="text-xs">Provide catalog images to attract buyers.</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="border-2 border-dashed border-border hover:border-primary/50 transition-all rounded-xl p-6 flex flex-col items-center justify-center gap-2 cursor-pointer bg-muted/5">
                <div className="h-10 w-10 bg-primary/10 text-primary rounded-full flex items-center justify-center">
                  <Upload className="h-5 w-5" />
                </div>
                <span className="text-xs font-semibold text-foreground">Upload product photos</span>
                <span className="text-[10px] text-muted-foreground">Supports PNG, JPG (Max 5MB)</span>
              </div>
            </CardContent>
          </Card>

          <Button
            type="submit"
            disabled={isSaving}
            className="w-full bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold py-6 text-sm flex items-center justify-center gap-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20"
          >
            {isSaving ? "Saving..." : (
              <>
                <Save className="h-4 w-4" /> Save Product
              </>
            )}
          </Button>
        </div>
      </form>
    </div>
  );
}

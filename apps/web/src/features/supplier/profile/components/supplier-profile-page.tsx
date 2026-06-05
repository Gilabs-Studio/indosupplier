"use client";

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { toast } from "sonner";
import { Save } from "lucide-react";

export function SupplierProfilePage() {
  const [isSaving, setIsSaving] = useState(false);

  const [form, setForm] = useState({
    companyName: "PT Nusantara Supplier Utama",
    businessType: "Manufacturer / Distributor",
    established: "2018",
    employees: "150 Employees",
    email: "info@nusantarasupplier.com",
    phone: "+62 811 2345 6789",
    website: "https://nusantarasupplier.com",
    taxId: "01.234.567.8-999.000 (NPWP)",
    nib: "9120001234567 (NIB)",
    overview: "We are the leading raw materials supplier in Indonesia, focusing on high-grade minerals, textiles, and agricultural products.",
  });

  const handleSave = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    setTimeout(() => {
      setIsSaving(false);
      toast.success("Profile saved successfully!");
    }, 800);
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-end md:justify-between gap-4 border-b border-border/80 pb-6 text-left">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            Company Profile
          </h1>
          <p className="text-sm text-muted-foreground">
            Manage your legal information and business details visible to potential buyers.
          </p>
        </div>
      </div>

      <form onSubmit={handleSave}>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
          {/* Main Info */}
          <div className="lg:col-span-2 space-y-6">
            <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
              <CardHeader>
                <CardTitle className="text-base font-bold font-heading">General Info</CardTitle>
                <CardDescription className="text-xs">Provide basic business context.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>Company Name</FieldLabel>
                    <Input
                      value={form.companyName}
                      onChange={(e) => setForm({ ...form, companyName: e.target.value })}
                      required
                    />
                  </Field>
                  <Field>
                    <FieldLabel>Business Type</FieldLabel>
                    <Input
                      value={form.businessType}
                      onChange={(e) => setForm({ ...form, businessType: e.target.value })}
                      required
                    />
                  </Field>
                </FieldGroup>

                <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>Established Year</FieldLabel>
                    <Input
                      value={form.established}
                      onChange={(e) => setForm({ ...form, established: e.target.value })}
                      required
                    />
                  </Field>
                  <Field>
                    <FieldLabel>Employee Count</FieldLabel>
                    <Input
                      value={form.employees}
                      onChange={(e) => setForm({ ...form, employees: e.target.value })}
                      required
                    />
                  </Field>
                </FieldGroup>

                <Field>
                  <FieldLabel>Company Overview</FieldLabel>
                  <Textarea
                    rows={4}
                    value={form.overview}
                    onChange={(e) => setForm({ ...form, overview: e.target.value })}
                    required
                    className="resize-none"
                  />
                </Field>
              </CardContent>
            </Card>

            <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
              <CardHeader>
                <CardTitle className="text-base font-bold font-heading">Legalities & Verification</CardTitle>
                <CardDescription className="text-xs">Verified tax identifiers for verified badges.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <FieldGroup className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <Field>
                    <FieldLabel>NPWP (Tax ID)</FieldLabel>
                    <Input
                      value={form.taxId}
                      onChange={(e) => setForm({ ...form, taxId: e.target.value })}
                      required
                    />
                  </Field>
                  <Field>
                    <FieldLabel>NIB (Registration Number)</FieldLabel>
                    <Input
                      value={form.nib}
                      onChange={(e) => setForm({ ...form, nib: e.target.value })}
                      required
                    />
                  </Field>
                </FieldGroup>
              </CardContent>
            </Card>
          </div>

          {/* Contact Details & Save Bar */}
          <div className="space-y-6">
            <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
              <CardHeader>
                <CardTitle className="text-base font-bold font-heading">Contact Details</CardTitle>
                <CardDescription className="text-xs">Channels for direct inquiries.</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <Field>
                  <FieldLabel>Official Email</FieldLabel>
                  <Input
                    type="email"
                    value={form.email}
                    onChange={(e) => setForm({ ...form, email: e.target.value })}
                    required
                  />
                </Field>
                <Field>
                  <FieldLabel>Office Phone</FieldLabel>
                  <Input
                    value={form.phone}
                    onChange={(e) => setForm({ ...form, phone: e.target.value })}
                    required
                  />
                </Field>
                <Field>
                  <FieldLabel>Website URL</FieldLabel>
                  <Input
                    type="url"
                    value={form.website}
                    onChange={(e) => setForm({ ...form, website: e.target.value })}
                    required
                  />
                </Field>
              </CardContent>
            </Card>

            <Button
              type="submit"
              disabled={isSaving}
              className="w-full bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold py-6 text-sm flex items-center justify-center gap-2 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg"
            >
              {isSaving ? "Saving..." : (
                <>
                  <Save className="h-4 w-4" /> Save Profile Details
                </>
              )}
            </Button>
          </div>
        </div>
      </form>
    </div>
  );
}

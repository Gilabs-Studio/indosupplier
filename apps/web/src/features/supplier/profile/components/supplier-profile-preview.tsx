"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Building2, MapPin, Mail, Phone, Globe, Calendar, Users, ShieldCheck } from "lucide-react";

export function SupplierProfilePreview() {
  const t = useTranslations("supplier.profile");
  const tDash = useTranslations("supplier.dashboard");

  const company = {
    name: "PT Nusantara Supplier Utama",
    type: "Manufacturer & Distributor",
    established: "2018",
    employees: "150 Employees",
    email: "info@nusantarasupplier.com",
    phone: "+62 811 2345 6789",
    website: "https://nusantarasupplier.com",
    location: "Kawasan Industri Jababeka, Cikarang, Jawa Barat, Indonesia",
    overview: "We are the leading raw materials supplier in Indonesia, focusing on high-grade minerals, industrial grade garnet sand, quartz powder, and agricultural bulk products.",
  };

  const products = [
    { id: "PROD-01", name: "Garnet Sand Mesh 80", category: "Industrial Minerals", moq: "20 Ton", price: "Rp 3.500.000 / Ton" },
    { id: "PROD-02", name: "Bentonite Clay Powder", category: "Industrial Minerals", moq: "10 Ton", price: "Rp 4.200.000 / Ton" },
    { id: "PROD-03", name: "Quartz Powder 325 Mesh", category: "Industrial Minerals", moq: "50 Ton", price: "Rp 2.800.000 / Ton" },
  ];

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="border-b border-border/80 pb-6">
        <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
          {t("previewTitle")}
        </h1>
        <p className="text-sm text-muted-foreground">
          {t("previewSubtitle")}
        </p>
      </div>

      {/* Banner/Header Block */}
      <Card className="border border-border shadow-md rounded-2xl overflow-hidden bg-card">
        <div className="h-32 bg-gradient-to-r from-primary/30 to-purple/30" />
        <CardContent className="p-6 relative">
          <div className="flex flex-col md:flex-row gap-5 items-start md:items-end -mt-16 mb-4">
            <div className="h-20 w-20 bg-primary border-4 border-card rounded-2xl flex items-center justify-center text-primary-foreground font-heading font-extrabold text-3xl shadow-lg">
              NS
            </div>
            <div className="space-y-1">
              <div className="flex flex-wrap items-center gap-2">
                <h2 className="text-xl font-bold text-foreground font-heading leading-none">
                  {company.name}
                </h2>
                <Badge className="bg-success/15 text-success border border-success/30 font-bold flex items-center gap-1">
                  <ShieldCheck className="h-3 w-3" /> Gold Supplier
                </Badge>
              </div>
              <p className="text-xs font-semibold text-muted-foreground flex items-center gap-1">
                <MapPin className="h-3.5 w-3.5" /> {company.location}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Grid Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
        {/* Company Overview & Products */}
        <div className="lg:col-span-2 space-y-6">
          {/* Overview */}
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-sm font-bold font-heading">{tDash("overview")}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground leading-relaxed">
                {company.overview}
              </p>
            </CardContent>
          </Card>

          {/* Products */}
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-sm font-bold font-heading">{tDash("products")}</CardTitle>
            </CardHeader>
            <CardContent className="p-0 border-t border-border">
              <div className="divide-y divide-border">
                {products.map((p) => (
                  <div key={p.id} className="p-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 hover:bg-secondary/10 transition-colors">
                    <div className="space-y-1">
                      <h4 className="text-sm font-bold text-foreground">{p.name}</h4>
                      <p className="text-xs text-muted-foreground">{p.category} • MOQ: <span className="font-semibold text-foreground">{p.moq}</span></p>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className="text-sm font-extrabold text-foreground">{p.price}</span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Business details sidebar */}
        <div className="space-y-6">
          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-sm font-bold font-heading">{tDash("businessInfo")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center justify-between text-xs py-1.5 border-b border-border/80">
                <span className="text-muted-foreground flex items-center gap-1.5 font-semibold"><Building2 className="h-3.5 w-3.5" /> {tDash("businessType")}</span>
                <span className="text-foreground font-bold">{company.type}</span>
              </div>
              <div className="flex items-center justify-between text-xs py-1.5 border-b border-border/80">
                <span className="text-muted-foreground flex items-center gap-1.5 font-semibold"><Calendar className="h-3.5 w-3.5" /> {tDash("established")}</span>
                <span className="text-foreground font-bold">{company.established}</span>
              </div>
              <div className="flex items-center justify-between text-xs py-1.5">
                <span className="text-muted-foreground flex items-center gap-1.5 font-semibold"><Users className="h-3.5 w-3.5" /> {tDash("employees")}</span>
                <span className="text-foreground font-bold">{company.employees}</span>
              </div>
            </CardContent>
          </Card>

          <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
            <CardHeader>
              <CardTitle className="text-sm font-bold font-heading">{tDash("contactInfo")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-xs">
              <div className="flex items-center gap-2 py-1.5 border-b border-border/80">
                <Mail className="h-4 w-4 text-muted-foreground shrink-0" />
                <span className="text-foreground font-semibold truncate">{company.email}</span>
              </div>
              <div className="flex items-center gap-2 py-1.5 border-b border-border/80">
                <Phone className="h-4 w-4 text-muted-foreground shrink-0" />
                <span className="text-foreground font-semibold">{company.phone}</span>
              </div>
              <div className="flex items-center gap-2 py-1.5">
                <Globe className="h-4 w-4 text-muted-foreground shrink-0" />
                <a href={company.website} target="_blank" rel="noreferrer" className="text-primary font-semibold hover:underline truncate">
                  {company.website}
                </a>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

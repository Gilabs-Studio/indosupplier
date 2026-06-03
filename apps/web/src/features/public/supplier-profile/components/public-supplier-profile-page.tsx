"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel } from "@/components/ui/field";
import { toast } from "sonner";
import {
  Building,
  Calendar,
  Users,
  MapPin,
  Mail,
  Phone,
  Globe,
  Award,
  Package,
  ChevronLeft,
} from "lucide-react";

interface PublicSupplierProfilePageProps {
  locale: string;
  slug: string;
}

export function PublicSupplierProfilePage({ locale, slug }: PublicSupplierProfilePageProps) {
  const tSup = useTranslations("public.supplier");

  // Mock RFQ form state
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [quantity, setQuantity] = useState("1");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Deriving readable company name from slug
  const companyName = slug
    .split("-")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");

  const handleSendRFQ = (e: React.FormEvent) => {
    e.preventDefault();
    if (!subject.trim() || !message.trim()) {
      toast.error("Please fill in all required fields.");
      return;
    }

    setIsSubmitting(true);
    setTimeout(() => {
      toast.success(tSup("quoteSuccess"));
      setSubject("");
      setMessage("");
      setQuantity("1");
      setIsSubmitting(false);
    }, 1000);
  };

  return (
    <PublicLayout locale={locale}>
      <div className="bg-muted py-8">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          {/* Back button */}
          <Link
            href="/demo/search"
            className="inline-flex items-center gap-1.5 text-sm font-semibold text-muted-foreground hover:text-foreground transition-colors mb-6 cursor-pointer"
          >
            <ChevronLeft className="h-4 w-4" />
            {tSup("backButton")}
          </Link>

          {/* Profile Header */}
          <div className="bg-card border border-border shadow-xs rounded-2xl overflow-hidden p-6 md:p-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-6">
            <div className="flex items-center gap-5">
              <div className="h-16 w-16 rounded-xl bg-primary text-primary-foreground font-heading font-bold text-2xl flex items-center justify-center">
                {companyName.substring(0, 2).toUpperCase()}
              </div>
              <div className="space-y-1.5">
                <div className="flex flex-wrap items-center gap-2">
                  <h1 className="text-2xl font-bold text-foreground font-heading tracking-tight leading-none">
                    {companyName}
                  </h1>
                  <Badge variant="outline" className="border-border text-muted-foreground font-semibold rounded-full px-2.5">
                    {tSup("notVerified")}
                  </Badge>
                </div>
                <div className="flex items-center gap-1.5 text-xs text-muted-foreground font-medium">
                  <MapPin className="h-4 w-4 text-muted-foreground" />
                  <span>{tSup("na")}</span>
                </div>
              </div>
            </div>
          </div>

          {/* Content Layout */}
          <div className="mt-8 grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Left/Main Column: Overview, Catalog, Certs */}
            <div className="lg:col-span-2 space-y-8">
              {/* Company Overview */}
              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("overview")}</CardTitle>
                </CardHeader>
                <CardContent className="text-sm text-muted-foreground leading-relaxed space-y-4">
                  <p>
                    {tSup("emptyOverview")}
                  </p>
                </CardContent>
              </Card>

              {/* Products Catalog */}
              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("products")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-center py-10 bg-muted border border-dashed border-border rounded-lg">
                    <Package className="mx-auto h-10 w-10 text-muted-foreground" />
                    <h3 className="mt-4 text-sm font-semibold text-foreground">{tSup("emptyProductsTitle")}</h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {tSup("emptyProductsDesc")}
                    </p>
                  </div>
                </CardContent>
              </Card>

              {/* Certifications */}
              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("certifications")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-center py-10 bg-muted border border-dashed border-border rounded-lg">
                    <Award className="mx-auto h-10 w-10 text-muted-foreground" />
                    <h3 className="mt-4 text-sm font-semibold text-foreground">{tSup("emptyCertsTitle")}</h3>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {tSup("emptyCertsDesc")}
                    </p>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Right Column: Business Details & Contact RFQ */}
            <div className="space-y-8">
              {/* Business Overview Stats */}
              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("businessInfo")}</CardTitle>
                </CardHeader>
                <CardContent className="p-0">
                  <div className="divide-y divide-border text-sm">
                    <div className="flex justify-between px-6 py-4">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Building className="h-4 w-4 text-muted-foreground" />
                        {tSup("businessType")}
                      </span>
                      <span className="font-semibold text-foreground">{tSup("na")}</span>
                    </div>
                    <div className="flex justify-between px-6 py-4">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Calendar className="h-4 w-4 text-muted-foreground" />
                        {tSup("established")}
                      </span>
                      <span className="font-semibold text-foreground">{tSup("na")}</span>
                    </div>
                    <div className="flex justify-between px-6 py-4">
                      <span className="text-muted-foreground font-medium flex items-center gap-2">
                        <Users className="h-4 w-4 text-muted-foreground" />
                        {tSup("employees")}
                      </span>
                      <span className="font-semibold text-foreground">{tSup("na")}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Contact Information */}
              <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("contactInfo")}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4 text-sm text-muted-foreground">
                  <div className="flex items-center gap-3">
                    <Mail className="h-4 w-4 text-muted-foreground" />
                    <span>{tSup("na")}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <Phone className="h-4 w-4 text-muted-foreground" />
                    <span>{tSup("na")}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <Globe className="h-4 w-4 text-muted-foreground" />
                    <span>{tSup("na")}</span>
                  </div>
                </CardContent>
              </Card>

              {/* Send RFQ Form */}
              <Card id="contact" className="border border-border shadow-xs rounded-xl overflow-hidden scroll-mt-20 bg-card">
                <CardHeader>
                  <CardTitle className="text-base font-bold font-heading">{tSup("requestQuote")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <form onSubmit={handleSendRFQ} className="space-y-4">
                    <Field>
                      <FieldLabel>{tSup("quoteSubject")}</FieldLabel>
                      <Input
                        type="text"
                        placeholder="e.g. Bulk Procurement for Teak Wood chairs"
                        value={subject}
                        onChange={(e) => setSubject(e.target.value)}
                        required
                        className="bg-card border-border focus-visible:border-muted-foreground"
                      />
                    </Field>

                    <Field>
                      <FieldLabel>{tSup("quoteQuantity")}</FieldLabel>
                      <Input
                        type="number"
                        min="1"
                        value={quantity}
                        onChange={(e) => setQuantity(e.target.value)}
                        required
                        className="bg-card border-border focus-visible:border-muted-foreground"
                      />
                    </Field>

                    <Field>
                      <FieldLabel>{tSup("quoteMessage")}</FieldLabel>
                      <Textarea
                        rows={4}
                        placeholder={tSup("quoteMessage")}
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        required
                        className="bg-card border-border focus-visible:border-muted-foreground resize-none"
                      />
                    </Field>

                    <Button
                      type="submit"
                      disabled={isSubmitting}
                      className="w-full bg-primary text-primary-foreground hover:bg-primary/95 font-semibold cursor-pointer"
                    >
                      {isSubmitting ? "Sending..." : tSup("btnSendQuote")}
                    </Button>
                  </form>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </div>
    </PublicLayout>
  );
}

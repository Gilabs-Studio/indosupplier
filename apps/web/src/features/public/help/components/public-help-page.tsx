"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Field, FieldLabel } from "@/components/ui/field";
import { toast } from "sonner";
import {
  Search,
  BookOpen,
  User,
  Store,
  CreditCard,
  HelpCircle,
} from "lucide-react";

interface PublicHelpPageProps {
  locale: string;
}

export function PublicHelpPage({ locale }: PublicHelpPageProps) {
  const tHelp = useTranslations("public.help");

  const [searchQuery, setSearchQuery] = useState("");
  const [subject, setSubject] = useState("");
  const [message, setMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const categories = [
    { id: "general", name: tHelp("general"), icon: HelpCircle, count: 5 },
    { id: "account", name: tHelp("account"), icon: User, count: 4 },
    { id: "supplier", name: tHelp("supplierHelp"), icon: Store, count: 8 },
    { id: "buyer", name: tHelp("buyerHelp"), icon: BookOpen, count: 6 },
    { id: "payments", name: tHelp("payments"), icon: CreditCard, count: 3 },
  ];

  const popularArticles = [
    tHelp("art1"),
    tHelp("art2"),
    tHelp("art3"),
    tHelp("art4"),
  ];

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!searchQuery.trim()) return;
    toast.success(`Searching help articles for: "${searchQuery}"`);
  };

  const handleContactSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!subject.trim() || !message.trim()) {
      toast.error("Please fill in all fields.");
      return;
    }

    setIsSubmitting(true);
    setTimeout(() => {
      toast.success(tHelp("messageSuccess"));
      setSubject("");
      setMessage("");
      setIsSubmitting(false);
    }, 1200);
  };

  return (
    <PublicLayout locale={locale}>
      {/* Help Hero search */}
      <section className="bg-neutral-950 text-white py-16 md:py-24 relative overflow-hidden">
        <div className="absolute inset-0 opacity-10 bg-[radial-gradient(circle_at_70%_75%,#3b82f6,transparent_60%)]" />
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center relative z-10 space-y-6">
          <h1 className="text-4xl font-bold tracking-tight font-heading sm:text-5xl">
            {tHelp("title")}
          </h1>
          <p className="mx-auto max-w-2xl text-neutral-400 text-sm sm:text-base">
            {tHelp("subtitle")}
          </p>

          <form onSubmit={handleSearch} className="mx-auto max-w-2xl mt-8">
            <div className="flex bg-card rounded-lg overflow-hidden border border-border p-1.5 shadow-md">
              <div className="flex items-center flex-1 px-3">
                <Search className="h-5 w-5 text-muted-foreground" />
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder={tHelp("placeholder")}
                  className="w-full px-3 py-3 bg-transparent text-foreground placeholder:text-muted-foreground border-0 outline-hidden focus:outline-hidden text-sm"
                />
              </div>
              <Button type="submit" className="bg-primary text-primary-foreground hover:bg-primary/95 px-6 cursor-pointer">
                {tHelp("title").split(" ")[0]}
              </Button>
            </div>
          </form>
        </div>
      </section>

      {/* Main categories */}
      <section className="bg-background py-16">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <h2 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-8">
            {tHelp("categories")}
          </h2>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-6">
            {categories.map((cat) => {
              const IconComponent = cat.icon;
              return (
                <Card
                  key={cat.id}
                  className="border border-border bg-card shadow-xs hover:shadow-md transition-all duration-300 rounded-xl overflow-hidden text-center cursor-pointer"
                >
                  <CardContent className="p-6 space-y-4">
                    <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-xl bg-muted text-foreground">
                      <IconComponent className="h-5 w-5" />
                    </div>
                    <div className="space-y-1">
                      <h3 className="font-semibold text-foreground text-sm">{cat.name}</h3>
                      <span className="text-xs text-muted-foreground font-medium">
                        {tHelp("articlesCount", { count: cat.count })}
                      </span>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>
      </section>

      {/* Popular articles & Contact Support Grid */}
      <section className="bg-muted py-16 border-t border-border">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 grid grid-cols-1 md:grid-cols-2 gap-12">
          {/* Popular Articles */}
          <div className="space-y-6">
            <h2 className="text-lg font-bold text-foreground font-heading">
              {tHelp("popularArticles")}
            </h2>

            <div className="divide-y divide-border bg-card border border-border rounded-xl overflow-hidden shadow-xs">
              {popularArticles.map((art, idx) => (
                <button
                  key={idx}
                  onClick={() => toast.info(`Opening help article: "${art}"`)}
                  className="flex items-center justify-between w-full text-left px-6 py-4 hover:bg-muted transition-colors text-sm font-medium text-foreground cursor-pointer"
                >
                  <span>{art}</span>
                  <BookOpen className="h-4 w-4 text-muted-foreground shrink-0" />
                </button>
              ))}
            </div>
          </div>

          {/* Contact Support Form */}
          <div className="space-y-6">
            <div className="space-y-2">
              <h2 className="text-lg font-bold text-foreground font-heading">
                {tHelp("contactUs")}
              </h2>
              <p className="text-xs text-muted-foreground max-w-sm">
                {tHelp("contactDesc")}
              </p>
            </div>

            <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
              <CardContent className="p-6">
                <form onSubmit={handleContactSubmit} className="space-y-4">
                  <Field>
                    <FieldLabel>{tHelp("subject")}</FieldLabel>
                    <Input
                      type="text"
                      placeholder="e.g. Account activation problem"
                      value={subject}
                      onChange={(e) => setSubject(e.target.value)}
                      required
                      className="bg-card border-border focus-visible:border-muted-foreground"
                    />
                  </Field>

                  <Field>
                    <FieldLabel>{tHelp("message")}</FieldLabel>
                    <Textarea
                      rows={4}
                      placeholder={tHelp("message")}
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
                    {isSubmitting ? "Sending..." : tHelp("btnSendMessage")}
                  </Button>
                </form>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>
    </PublicLayout>
  );
}

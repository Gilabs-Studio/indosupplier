"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Card } from "@/components/ui/card";
import { ChevronDown, HelpCircle } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";

interface PublicFaqPageProps {
  locale: string;
}

export function PublicFaqPage({ locale }: PublicFaqPageProps) {
  const tFaq = useTranslations("public.faq");
  const [openIndex, setOpenIndex] = useState<number | null>(0);

  const faqs = [
    { q: tFaq("q1"), a: tFaq("a1") },
    { q: tFaq("q2"), a: tFaq("a2") },
    { q: tFaq("q3"), a: tFaq("a3") },
    { q: tFaq("q4"), a: tFaq("a4") },
  ];

  const toggleFaq = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <PublicLayout locale={locale}>
      <section className="bg-muted py-16 md:py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center">
          <h1 className="text-4xl font-bold tracking-tight text-foreground font-heading sm:text-5xl">
            {tFaq("title")}
          </h1>
          <p className="mx-auto mt-4 max-w-2xl text-base text-muted-foreground">
            {tFaq("subtitle")}
          </p>
        </div>
      </section>

      <section className="bg-background py-16">
        <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
          <div className="space-y-4">
            {faqs.map((faq, idx) => {
              const isOpen = openIndex === idx;
              return (
                <Card
                  key={idx}
                  className="border border-border bg-card shadow-xs rounded-xl overflow-hidden"
                >
                  <button
                    onClick={() => toggleFaq(idx)}
                    className="flex w-full items-center justify-between p-6 text-left hover:bg-muted transition-colors cursor-pointer"
                  >
                    <span className="flex items-center gap-3 font-semibold text-foreground text-sm md:text-base">
                      <HelpCircle className="h-5 w-5 text-muted-foreground shrink-0" />
                      {faq.q}
                    </span>
                    <ChevronDown
                      className={`h-5 w-5 text-muted-foreground transition-transform duration-300 shrink-0 ${
                        isOpen ? "rotate-180" : ""
                      }`}
                    />
                  </button>

                  <AnimatePresence initial={false}>
                    {isOpen && (
                      <motion.div
                        initial={{ height: 0 }}
                        animate={{ height: "auto" }}
                        exit={{ height: 0 }}
                        transition={{ duration: 0.3 }}
                        className="overflow-hidden"
                      >
                        <div className="px-6 pb-6 pt-0 text-sm text-muted-foreground leading-relaxed border-t border-border">
                          <p className="pt-4">{faq.a}</p>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </Card>
              );
            })}
          </div>
        </div>
      </section>
    </PublicLayout>
  );
}

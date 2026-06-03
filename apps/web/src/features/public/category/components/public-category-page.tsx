"use client";

import React from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { PublicLayout } from "@/features/public/components/public-layout";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Factory,
  Layers,
  Building,
  HardHat,
  Tv,
  Car,
  ChevronRight,
} from "lucide-react";

interface PublicCategoryPageProps {
  locale: string;
}

export function PublicCategoryPage({ locale }: PublicCategoryPageProps) {
  const tCat = useTranslations("public.categories");

  const categories = [
    { id: "manufacturing", name: tCat("manufacturing"), icon: Factory, count: 0 },
    { id: "agriculture", name: tCat("agriculture"), icon: Layers, count: 0 },
    { id: "textile", name: tCat("textile"), icon: Layers, count: 0 },
    { id: "chemical", name: tCat("chemical"), icon: Layers, count: 0 },
    { id: "furniture", name: tCat("furniture"), icon: Building, count: 0 },
    { id: "construction", name: tCat("construction"), icon: HardHat, count: 0 },
    { id: "electronics", name: tCat("electronics"), icon: Tv, count: 0 },
    { id: "automotive", name: tCat("automotive"), icon: Car, count: 0 },
  ];

  return (
    <PublicLayout locale={locale}>
      <section className="bg-muted py-16 md:py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center">
          <h1 className="text-4xl font-bold tracking-tight text-foreground font-heading sm:text-5xl">
            {tCat("title")}
          </h1>
          <p className="mx-auto mt-4 max-w-2xl text-base text-muted-foreground">
            {tCat("subtitle")}
          </p>
        </div>
      </section>

      <section className="bg-background py-16">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
            {categories.map((cat) => {
              const IconComponent = cat.icon;
              return (
                <Card
                  key={cat.id}
                  className="group relative border border-border shadow-xs hover:shadow-md transition-all duration-300 rounded-xl overflow-hidden cursor-pointer"
                >
                  <CardContent className="p-6 space-y-4">
                    <div className="flex items-center gap-3">
                      <div className="p-3 bg-primary text-primary-foreground rounded-lg group-hover:scale-105 transition-transform">
                        <IconComponent className="h-5 w-5" />
                      </div>
                      <div>
                        <h3 className="font-semibold text-foreground text-sm">{cat.name}</h3>
                        <span className="text-xs text-muted-foreground font-medium">
                          {cat.count} {tCat("countSuffix")}
                        </span>
                      </div>
                    </div>

                    <div className="pt-2 flex justify-end">
                      <Button asChild size="sm" variant="ghost" className="text-xs font-semibold group-hover:text-primary transition-colors cursor-pointer p-0">
                        <Link href={`/demo/search?category=${cat.id}`} className="flex items-center gap-1">
                          {tCat("viewSuppliers")}
                          <ChevronRight className="h-4 w-4" />
                        </Link>
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>
      </section>
    </PublicLayout>
  );
}

"use client";

import React from "react";
import Image from "next/image";
import { Link } from "@/i18n/routing";
import { ChevronRight } from "lucide-react";
import type { DemoQuickAction } from "../types/demo.types";

interface QuickActionsBarProps {
  actions: DemoQuickAction[];
  isEn: boolean;
}

export function QuickActionsBar({ actions, isEn }: Readonly<QuickActionsBarProps>) {
  const getIconPath = (type: string) => {
    switch (type) {
      case "rfq":
        return "/images/icons/action-rfq.svg";
      case "cart":
        return "/images/icons/action-quick-order.svg";
      case "shield":
        return "/images/icons/action-verified.svg";
      case "calendar":
      default:
        return "/images/icons/action-payment-terms.svg";
    }
  };

  return (
    <section className="w-full">
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {actions.map((action) => {
          const title = isEn ? action.titleEn : action.titleId;
          const subtitle = isEn ? action.subtitleEn : action.subtitleId;

          return (
            <Link
              key={action.id}
              href={action.href}
              className="group flex cursor-pointer items-center justify-between rounded-xl border border-border bg-card p-3.5 shadow-2xs transition-all duration-200 hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-md"
            >
              <div className="flex items-center gap-3.5 min-w-0">
                {/* Minimalist Vector Illustration Only (No placeholder container box) */}
                <div className="relative h-11 w-11 shrink-0">
                  <Image
                    src={getIconPath(action.iconType)}
                    alt={title}
                    width={44}
                    height={44}
                    className="h-full w-full object-contain transition-transform duration-200 group-hover:scale-105"
                  />
                </div>

                {/* Text */}
                <div className="min-w-0 flex-1">
                  <h3 className="truncate text-xs sm:text-sm font-bold text-foreground group-hover:text-primary transition-colors">
                    {title}
                  </h3>
                  <p className="truncate text-[11px] text-muted-foreground mt-0.5">
                    {subtitle}
                  </p>
                </div>
              </div>

              {/* Right Chevron */}
              <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground/60 transition-transform group-hover:translate-x-0.5 group-hover:text-primary" />
            </Link>
          );
        })}
      </div>
    </section>
  );
}

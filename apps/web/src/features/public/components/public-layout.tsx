"use client";

import React from "react";
import { PublicNavbar } from "./public-navbar";
import { PublicFooter } from "./public-footer";

interface PublicLayoutProps {
  children: React.ReactNode;
  locale: string;
}

export function PublicLayout({ children, locale }: PublicLayoutProps) {
  return (
    <div className="min-h-screen bg-background text-foreground font-jost antialiased flex flex-col justify-between">
      <div>
        <PublicNavbar locale={locale} />
        <main>{children}</main>
      </div>
      <PublicFooter />
    </div>
  );
}



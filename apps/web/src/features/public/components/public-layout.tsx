"use client";

import React from "react";
import { PublicNavbar } from "./public-navbar";
import { PublicFooter } from "./public-footer";

interface PublicLayoutProps {
  children: React.ReactNode;
  locale: string;
  showFooter?: boolean;
  overlapNavbar?: boolean;
}

export function PublicLayout({
  children,
  locale,
  showFooter = true,
  overlapNavbar = false,
}: Readonly<PublicLayoutProps>) {
  return (
    <div className="min-h-screen bg-background text-foreground font-jost antialiased flex flex-col justify-between relative">
      <div className="w-full flex-1 flex flex-col">
        <div className={overlapNavbar ? "absolute top-0 left-0 right-0 z-50" : "w-full"}>
          <PublicNavbar locale={locale} />
        </div>
        <main className="flex-1 flex flex-col">{children}</main>
      </div>
      {showFooter && <PublicFooter />}
    </div>
  );
}



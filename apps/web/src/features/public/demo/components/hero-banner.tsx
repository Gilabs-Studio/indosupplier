"use client";

import React from "react";
import Image from "next/image";

export function HeroBanner() {
  return (
    <section className="relative w-full overflow-hidden rounded-xl shadow-xs transition-all duration-300">
      <div className="relative aspect-[894/192] w-full overflow-hidden rounded-xl">
        <Image
          src="/images/hero-banner-b2b.png"
          alt="Beli Langsung dari Supplier - Harga Grosir, Lebih Kompetitif"
          fill
          priority
          sizes="(max-width: 1280px) 100vw, 1280px"
          className="object-cover w-full h-full rounded-xl"
        />
      </div>
    </section>
  );
}

"use client";

import { useState, useEffect } from "react";

export function useDemoHome() {
  const [activeNewsTab, setActiveNewsTab] = useState<string>("Berita");
  const [showBackToTop, setShowBackToTop] = useState<boolean>(false);

  useEffect(() => {
    const handleScroll = () => {
      setShowBackToTop(window.scrollY > 400);
    };

    window.addEventListener("scroll", handleScroll);
    return () => {
      window.removeEventListener("scroll", handleScroll);
    };
  }, []);

  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  };

  return {
    activeNewsTab,
    setActiveNewsTab,
    showBackToTop,
    scrollToTop,
  };
}

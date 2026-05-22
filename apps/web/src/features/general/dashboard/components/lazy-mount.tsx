"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";

interface LazyMountProps {
  readonly children: ReactNode;
  readonly rootMargin?: string;
}

export function LazyMount({ children, rootMargin = "200px" }: LazyMountProps) {
  const ref = useRef<HTMLDivElement | null>(null);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    if (mounted) return;
    const el = ref.current;
    if (!el) return;
    const obs = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setMounted(true);
            obs.disconnect();
          }
        });
      },
      { root: null, rootMargin },
    );
    obs.observe(el);
    return () => obs.disconnect();
  }, [mounted, rootMargin]);

  return <div ref={ref}>{mounted ? children : <div className="min-h-40" />}</div>;
}

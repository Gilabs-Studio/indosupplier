"use client";

import { useEffect, useState, useRef } from "react";

interface UseTypedHeadlineOptions {
  speed?: number; // ms per character
  pause?: number; // ms pause at full string
}

/**
 * Simple typed-headline hook.
 * Cycles through `phrases` with a typing + pause + erase loop.
 */
export function useTypedHeadline(phrases: string[], options?: UseTypedHeadlineOptions) {
  const speed = options?.speed ?? 60;
  const pause = options?.pause ?? 1200;

  const [index, setIndex] = useState(0);
  const [display, setDisplay] = useState("");
  const [isDeleting, setIsDeleting] = useState(false);
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useEffect(() => {
    let timeout: number | undefined;
    const current = phrases[index % phrases.length] ?? "";

    if (!isDeleting && display === current) {
      // pause at full string, then start deleting
      timeout = window.setTimeout(() => {
        if (!mounted.current) return;
        setIsDeleting(true);
      }, pause);
    } else if (isDeleting && display === "") {
      // move to next phrase
      timeout = window.setTimeout(() => {
        if (!mounted.current) return;
        setIsDeleting(false);
        setIndex((i) => (i + 1) % phrases.length);
      }, 200);
    } else {
      // typing or deleting
      timeout = window.setTimeout(() => {
        if (!mounted.current) return;
        setDisplay((d) => {
          if (!isDeleting) return current.slice(0, d.length + 1);
          return current.slice(0, Math.max(0, d.length - 1));
        });
      }, isDeleting ? Math.max(30, speed / 2) : speed);
    }

    return () => {
      if (timeout) window.clearTimeout(timeout);
    };
  }, [display, isDeleting, index, phrases, speed, pause]);

  return display;
}

export default useTypedHeadline;

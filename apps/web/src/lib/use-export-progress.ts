"use client";

import { useCallback, useMemo, useState } from "react";

import { runAsyncExport, type RunAsyncExportOptions, type ExportJobStatus } from "@/lib/async-export";

type UseExportProgressResult = {
  isExporting: boolean;
  progress: number;
  runWithProgress: (options: RunAsyncExportOptions) => Promise<void>;
  label: (idleLabel: string, activeLabel?: string) => string;
};

export function useExportProgress(): UseExportProgressResult {
  const [isExporting, setIsExporting] = useState(false);
  const [progress, setProgress] = useState(0);

  const runWithProgress = useCallback(async (options: RunAsyncExportOptions) => {
    if (isExporting) {
      return;
    }

    setIsExporting(true);
    setProgress(0);

    const originalOnProgress = options.onProgress;

    try {
      await runAsyncExport({
        ...options,
        onProgress: (status: ExportJobStatus, nextProgress: number) => {
          if (status === "queued") {
            setProgress(0);
          } else {
            setProgress(Math.max(0, Math.min(100, nextProgress)));
          }
          originalOnProgress?.(status, nextProgress);
        },
      });
      setProgress(100);
    } finally {
      setIsExporting(false);
      window.setTimeout(() => setProgress(0), 500);
    }
  }, [isExporting]);

  const label = useCallback((idleLabel: string, activeLabel = "Exporting") => {
    if (!isExporting) {
      return idleLabel;
    }
    return `${activeLabel} ${progress}%`;
  }, [isExporting, progress]);

  return useMemo(() => ({
    isExporting,
    progress,
    runWithProgress,
    label,
  }), [isExporting, progress, runWithProgress, label]);
}
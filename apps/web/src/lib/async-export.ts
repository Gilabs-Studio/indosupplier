import { apiClient } from "@/lib/api-client";
import { toast } from "sonner";

export type ExportJobStatus = "queued" | "processing" | "completed" | "failed";

interface ExportJob {
  id: string;
  status: ExportJobStatus;
  progress?: number;
  file_name?: string;
  error?: string;
}

interface ApiEnvelope<T> {
  success: boolean;
  data: T;
}

export interface RunAsyncExportOptions {
  endpoint: string;
  params?: Record<string, unknown>;
  timeoutMs?: number;
  pollIntervalMs?: number;
  onProgress?: (status: ExportJobStatus, progress: number) => void;
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms);
  });
}

function resolveFileName(job: ExportJob, fallbackName: string): string {
  const fromJob = job.file_name?.trim();
  return fromJob && fromJob.length > 0 ? fromJob : fallbackName;
}

function progressLabel(status: ExportJobStatus, progress: number): string {
  if (status === "queued") {
    return "Export queued...";
  }
  if (status === "processing") {
    return `Export processing ${progress}%`;
  }
  return "Export processing...";
}

export async function runAsyncExport(options: RunAsyncExportOptions): Promise<void> {
  const timeoutMs = options.timeoutMs ?? 5 * 60 * 1000;
  const pollIntervalMs = options.pollIntervalMs ?? 1500;
  const progressToastId = toast.loading("Export queued...");

  const queued = await apiClient.get<ApiEnvelope<ExportJob>>(options.endpoint, {
    params: {
      ...(options.params ?? {}),
      async: true,
    },
  });

  const queuedJob = queued.data.data;
  const jobId = queuedJob?.id;
  if (!jobId) {
    toast.dismiss(progressToastId);
    throw new Error("Export job id not found");
  }

  toast.loading(progressLabel(queuedJob.status, queuedJob.progress ?? 0), {
    id: progressToastId,
  });
  options.onProgress?.(queuedJob.status, queuedJob.progress ?? 0);

  const startedAt = Date.now();
  let latest = queuedJob;

  while (Date.now() - startedAt < timeoutMs) {
    const statusResponse = await apiClient.get<ApiEnvelope<ExportJob>>(`/exports/jobs/${jobId}`);
    latest = statusResponse.data.data;

    toast.loading(progressLabel(latest.status, latest.progress ?? 0), {
      id: progressToastId,
    });
    options.onProgress?.(latest.status, latest.progress ?? 0);

    if (latest.status === "completed") {
      const fileResponse = await apiClient.get<Blob>(`/exports/jobs/${jobId}/download`, {
        responseType: "blob",
      });

      const fileName = resolveFileName(latest, `export_${jobId}.bin`);
      const blob = new Blob([fileResponse.data]);
      const objectUrl = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = objectUrl;
      link.setAttribute("download", fileName);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(objectUrl);
      toast.dismiss(progressToastId);
      return;
    }

    if (latest.status === "failed") {
      toast.dismiss(progressToastId);
      throw new Error(latest.error ?? "Export failed");
    }

    await sleep(pollIntervalMs);
  }

  toast.dismiss(progressToastId);
  throw new Error("Export timed out");
}

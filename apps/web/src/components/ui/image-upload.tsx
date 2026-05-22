"use client";

import * as React from "react";
import Image from "next/image";
import { Search, X, Loader2, ImagePlus, Link2 } from "lucide-react";
import { useDropzone, type FileRejection } from "react-dropzone";
import { toast } from "sonner";
import { cn, resolveImageUrl } from "@/lib/utils";
import apiClient from "@/lib/api-client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const DATA_IMAGE_PREFIX = "data:image/";

function inferUploadFolderFromPath(pathname?: string): string {
  if (!pathname) return "general";

  const segments = pathname.split("/").filter(Boolean);
  if (segments.length === 0) return "general";

  const localeSet = new Set(["en", "id"]);
  const normalized = localeSet.has(segments[0]) ? segments.slice(1) : segments;
  if (normalized.length === 0) return "general";

  const root = normalized[0].toLowerCase();
  const child = normalized[1]?.toLowerCase();

  if (child) {
    return `${root}/${child}`;
  }

  return root;
}

interface ImageUploadProps {
  value?: string | null;
  onChange: (url: string) => void;
  uploadFolder?: string;
  disabled?: boolean;
  className?: string;
  helperText?: string;
  enableUrlInput?: boolean;
  enableImageSearch?: boolean;
  searchEndpoint?: string;
  initialSearchQuery?: string;
  labels?: {
    dragActive?: string;
    dragInactive?: string;
    uploadSuccess?: string;
    uploadFailed?: string;
    invalidFile?: string;
    removeImage?: string;
    pasteImageUrl?: string;
    imageUrlPlaceholder?: string;
    applyUrl?: string;
    searchImage?: string;
    searchPlaceholder?: string;
    searchButton?: string;
    searchNoResults?: string;
    searchFailed?: string;
    selectFromResults?: string;
    attributionPrefix?: string;
    onUnsplash?: string;
    termsNotice?: string;
  };
}

interface UploadResponse {
  success: boolean;
  data: {
    url: string;
    filename: string;
    // other fields ignored
  };
}

interface ImageSearchResponse {
  success: boolean;
  data: Array<{
    id: string;
    url: string;
    thumb_url: string;
    alt: string;
    photographer_name: string;
    photographer_url: string;
    unsplash_photo_url: string;
    download_location: string;
    source: "unsplash";
  }>;
}

export function ImageUpload({
  value,
  onChange,
  uploadFolder,
  disabled,
  className,
  helperText = "SVG, PNG, JPG or GIF (max. 10MB)",
  enableUrlInput = false,
  enableImageSearch = false,
  searchEndpoint = "/api/image-search",
  initialSearchQuery,
  labels = {
    dragActive: "Drop image here",
    dragInactive: "Click or drag image",
    uploadSuccess: "Image uploaded successfully",
    uploadFailed: "Upload failed",
    invalidFile: "Invalid file",
    removeImage: "Remove image",
    pasteImageUrl: "Paste image URL",
    imageUrlPlaceholder: "https://images.unsplash.com/...",
    applyUrl: "Apply",
    searchImage: "Search from image API",
    searchPlaceholder: "Search image, e.g. mineral water bottle",
    searchButton: "Search",
    searchNoResults: "No images found",
    searchFailed: "Image search failed",
    selectFromResults: "Select",
    attributionPrefix: "Photo by",
    onUnsplash: "on Unsplash",
    termsNotice: "By selecting Unsplash photos, you agree to keep attribution and allow download tracking.",
  },
}: ImageUploadProps) {
  const [loading, setLoading] = React.useState(false);
  const [localPreviewObjectUrl, setLocalPreviewObjectUrl] = React.useState<string | null>(null);
  const [urlInput, setUrlInput] = React.useState("");
  const [searchQuery, setSearchQuery] = React.useState(initialSearchQuery ?? "");
  const [isSearching, setIsSearching] = React.useState(false);
  const [searchResults, setSearchResults] = React.useState<ImageSearchResponse["data"]>([]);
  const inferredFolder = React.useMemo(() => {
    if (typeof window === "undefined") {
      return "general";
    }
    return inferUploadFolderFromPath(window.location.pathname);
  }, []);

  const sanitizeImageSrc = React.useCallback((raw?: string | null): string | undefined => {
    if (!raw) return undefined;

    const value = raw.trim();
    if (!value) return undefined;

    if (value.startsWith("/")) {
      return value;
    }

    // Keep local blob previews generated in-browser (from dropped files).
    if (/^blob:/i.test(value)) {
      return value;
    }

    // Allow only image data URLs, disallow other data:* payloads.
    if (value.toLowerCase().startsWith(DATA_IMAGE_PREFIX)) {
      return value;
    }

    try {
      const parsed = new URL(value);
      if (parsed.protocol === "http:" || parsed.protocol === "https:") {
        return parsed.toString();
      }
      return undefined;
    } catch {
      return undefined;
    }
  }, []);

  const sanitizeExternalHref = React.useCallback((raw?: string | null): string | undefined => {
    if (!raw) return undefined;

    try {
      const parsed = new URL(raw.trim());
      if (parsed.protocol === "http:" || parsed.protocol === "https:") {
        return parsed.toString();
      }
      return undefined;
    } catch {
      return undefined;
    }
  }, []);

  const preview = React.useMemo(() => {
    if (localPreviewObjectUrl) return localPreviewObjectUrl;
    if (value) return resolveImageUrl(value) ?? null;
    return null;
  }, [localPreviewObjectUrl, value]);

  React.useEffect(() => {
    if (initialSearchQuery && initialSearchQuery !== searchQuery) {
      React.startTransition(() => {
        setSearchQuery(initialSearchQuery);
      });
    }
  }, [initialSearchQuery, searchQuery]);

  const onDrop = React.useCallback(
    async (acceptedFiles: File[], fileRejections: FileRejection[]) => {
      if (fileRejections.length > 0) {
        const error = fileRejections[0].errors[0];
        toast.error(labels.invalidFile ?? "Invalid file", {
          description: error.message,
        });
        return;
      }

      const file = acceptedFiles[0];
      if (!file) return;

      // Create preview immediately
      const objectUrl = URL.createObjectURL(file);
      setLocalPreviewObjectUrl(objectUrl);
      setLoading(true);

      const formData = new FormData();
      formData.append("file", file);

      try {
        const folder = uploadFolder?.trim() || inferredFolder;
        const endpoint = `/upload/image?folder=${encodeURIComponent(folder)}`;

        const res = await apiClient.post<UploadResponse>(endpoint, formData, {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        });

        if (res.data.success && res.data.data.url) {
          // Construct full URL if returned URL is relative
          // Assuming backend returns relative path like "/uploads/..." 
          // or fully qualified. If relative, prepend API_URL or serve statically.
          // The standard says "url": "/uploads/uuid.webp", so it's likely relative to domain root or static server.
          // For now, assume the URL is directly usable (or proxied).
          // But usually we need to prepend BE base URL if served by BE, or CDN.
          // Let's assume the URL is absolute or root-relative which works with <img src>.
          onChange(res.data.data.url);
          toast.success(labels.uploadSuccess ?? "Image uploaded successfully");
        } else {
          throw new Error(labels.uploadFailed ?? "Upload failed");
        }
      } catch (err) {
        setLocalPreviewObjectUrl(null);
        toast.error(labels.uploadFailed ?? "Upload failed", {
          description: "Something went wrong while uploading the image.",
        });
        console.error(err);
      } finally {
        setLoading(false);
      }
    },
    [inferredFolder, onChange, labels.invalidFile, labels.uploadFailed, labels.uploadSuccess, uploadFolder]
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: {
      "image/png": [],
      "image/jpeg": [],
      "image/jpg": [],
      "image/gif": [],
      "image/webp": [],
    },
    maxSize: 10 * 1024 * 1024, // 10MB
    maxFiles: 1,
    disabled: disabled || loading,
  });

  const handleRemove = (e: React.MouseEvent) => {
    e.stopPropagation();
    onChange("");
    setLocalPreviewObjectUrl(null);
  };

  const handleApplyUrl = React.useCallback(() => {
    const trimmed = urlInput.trim();
    if (!trimmed) {
      return;
    }

    let safeValue: string | null = null;

    if (trimmed.startsWith("/")) {
      safeValue = trimmed;
    } else {
      try {
        const parsed = new URL(trimmed);
        if (parsed.protocol === "http:" || parsed.protocol === "https:") {
          safeValue = parsed.toString();
        }
      } catch {
        safeValue = null;
      }
    }

    if (!safeValue) {
      toast.error(labels.invalidFile ?? "Invalid file", {
        description: "Use full http/https URL or a root-relative path (/...)",
      });
      return;
    }

    onChange(safeValue);
    setUrlInput("");
  }, [labels.invalidFile, onChange, urlInput]);

  const handleSearch = React.useCallback(async () => {
    const keyword = searchQuery.trim();
    if (keyword.length < 2) {
      return;
    }

    setIsSearching(true);
    try {
      const res = await fetch(
        `${searchEndpoint}?q=${encodeURIComponent(keyword)}&page=1&per_page=12`,
        { method: "GET" }
      );

      if (!res.ok) {
        throw new Error("SEARCH_FAILED");
      }

      const payload = (await res.json()) as ImageSearchResponse;
      setSearchResults(payload.data ?? []);
    } catch {
      setSearchResults([]);
      toast.error(labels.searchFailed ?? "Image search failed");
    } finally {
      setIsSearching(false);
    }
  }, [labels.searchFailed, searchEndpoint, searchQuery]);

  const triggerUnsplashDownload = React.useCallback(
    async (photoId: string, downloadLocation: string) => {
      try {
        await fetch("/api/image-search/download", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            photo_id: photoId,
            download_location: downloadLocation,
          }),
        });
      } catch {
        // Do not block user selection when provider event fails.
      }
    },
    []
  );

  const handleSelectFromResults = React.useCallback(
    (item: ImageSearchResponse["data"][number]) => {
      onChange(item.url);

      if (item.source === "unsplash" && item.download_location) {
        void triggerUnsplashDownload(item.id, item.download_location);
      }
    },
    [onChange, triggerUnsplashDownload]
  );

  const safePreviewSrc = React.useMemo(() => sanitizeImageSrc(preview), [preview, sanitizeImageSrc]);

  return (
    <div className={cn("space-y-2", className)}>
      <div
        {...getRootProps()}
        className={cn(
          "relative flex flex-col items-center justify-center w-full border-2 border-dashed rounded-lg cursor-pointer transition-colors bg-muted/5 xs:bg-background hover:bg-muted/10",
          isDragActive ? "border-primary bg-primary/5" : "border-muted-foreground/25",
          disabled && "opacity-50 cursor-not-allowed hover:bg-background",
          preview ? "p-0 aspect-square w-32 md:w-40 border-0 overflow-hidden" : "p-6 py-8"
        )}
      >
        <input {...getInputProps()} />

        {loading && (
          <div className="absolute inset-0 z-50 flex items-center justify-center bg-background/50 backdrop-blur-sm rounded-lg">
            <Loader2 className="w-6 h-6 animate-spin text-primary" />
          </div>
        )}

        {safePreviewSrc ? (
          <div className="relative w-full h-full group">
            <div className="relative w-full h-full">
              <Image
                src={safePreviewSrc}
                alt="Product preview"
                fill
                sizes="(max-width: 768px) 128px, 160px"
                className="rounded-lg object-cover"
              />
            </div>
            {!disabled && !loading && (
              <button
                type="button"
                className="absolute top-1 right-1 p-1 bg-destructive/90 text-destructive-foreground rounded-full opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer hover:bg-destructive"
                onClick={handleRemove}
                title={labels.removeImage ?? "Remove image"}
                aria-label={labels.removeImage ?? "Remove image"}
              >
                <X className="w-3 h-3" />
              </button>
            )}
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center text-center space-y-2">
            <div className="p-3 bg-muted rounded-full">
              <ImagePlus className="w-6 h-6 text-muted-foreground" />
            </div>
            <div className="space-y-1">
              <p className="text-sm font-medium text-muted-foreground">
                {isDragActive ? (labels.dragActive ?? "Drop image here") : (labels.dragInactive ?? "Click or drag image")}
              </p>
              {helperText && (
                <p className="text-xs text-muted-foreground/70">{helperText}</p>
              )}
            </div>
          </div>
        )}
      </div>

      {enableUrlInput && (
        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground flex items-center gap-1">
            <Link2 className="h-3.5 w-3.5" />
            {labels.pasteImageUrl ?? "Paste image URL"}
          </p>
          <div className="flex gap-2">
            <Input
              value={urlInput}
              onChange={(event) => setUrlInput(event.target.value)}
              placeholder={labels.imageUrlPlaceholder ?? "https://images.unsplash.com/..."}
              disabled={disabled || loading}
            />
            <Button
              type="button"
              variant="secondary"
              onClick={handleApplyUrl}
              disabled={disabled || loading || urlInput.trim().length === 0}
            >
              {labels.applyUrl ?? "Apply"}
            </Button>
          </div>
        </div>
      )}

      {enableImageSearch && (
        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground flex items-center gap-1">
            <Search className="h-3.5 w-3.5" />
            {labels.searchImage ?? "Search from image API"}
          </p>
          <div className="flex gap-2">
            <Input
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              placeholder={labels.searchPlaceholder ?? "Search image..."}
              disabled={disabled || isSearching}
            />
            <Button
              type="button"
              variant="secondary"
              onClick={handleSearch}
              disabled={disabled || isSearching || searchQuery.trim().length < 2}
            >
              {isSearching ? <Loader2 className="h-4 w-4 animate-spin" /> : (labels.searchButton ?? "Search")}
            </Button>
          </div>

          {searchResults.length > 0 && (
            <div className="grid grid-cols-3 gap-2">
              {searchResults.map((item) => (
                <div key={item.id} className="space-y-1">
                  <button
                    type="button"
                    className="group relative aspect-square w-full overflow-hidden rounded-md border border-border cursor-pointer"
                    onClick={() => handleSelectFromResults(item)}
                    title={labels.selectFromResults ?? "Select"}
                  >
                    <Image
                      src={sanitizeImageSrc(item.thumb_url) ?? ""}
                      alt={item.alt || "Search result"}
                      fill
                      sizes="(max-width: 768px) 33vw, 25vw"
                      className="object-cover"
                    />
                    <span className="absolute inset-x-0 bottom-0 bg-background/80 px-1 py-0.5 text-[10px] text-foreground opacity-0 transition-opacity group-hover:opacity-100">
                      {labels.selectFromResults ?? "Select"}
                    </span>
                  </button>
                  <p className="text-[10px] leading-tight text-muted-foreground">
                    {labels.attributionPrefix ?? "Photo by"}{" "}
                    <a
                      href={sanitizeExternalHref(item.photographer_url)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="underline"
                    >
                      {item.photographer_name}
                    </a>{" "}
                    <a
                      href={sanitizeExternalHref(item.unsplash_photo_url)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="underline"
                    >
                      {labels.onUnsplash ?? "on Unsplash"}
                    </a>
                  </p>
                </div>
              ))}
            </div>
          )}

          <p className="text-xs text-muted-foreground/80">
            {labels.termsNotice ?? "By selecting Unsplash photos, you agree to keep attribution and allow download tracking."}
          </p>

          {!isSearching && searchQuery.trim().length >= 2 && searchResults.length === 0 && (
            <p className="text-xs text-muted-foreground">{labels.searchNoResults ?? "No images found"}</p>
          )}
        </div>
      )}
    </div>
  );
}

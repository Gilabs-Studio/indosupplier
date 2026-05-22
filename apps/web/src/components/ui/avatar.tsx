import * as React from "react";
import { cn } from "@/lib/utils";

const AvatarContext = React.createContext<{
  imageLoaded: boolean;
  setImageLoaded: (loaded: boolean) => void;
}>({
  imageLoaded: false,
  setImageLoaded: () => {},
});

/**
 * Validates and sanitises a URL for use in an img src.
 * Prevents dangerous protocols like javascript: and data: non-image URLs.
 *
 * Critically, for remote URLs the function returns `parsed.href` — the URL
 * reconstructed by the browser's URL parser — rather than the raw input
 * string. This lets static-analysis tools (e.g. CodeQL js/xss-through-dom)
 * recognise the output as sanitised, because the value no longer flows
 * directly from user-controlled input to a sink.
 */
const sanitizeImageUrl = (url?: string): string | undefined => {
  if (!url || typeof url !== "string") return undefined;

  const trimmed = url.trim();
  if (!trimmed) return undefined;

  // Allow site-relative paths — no protocol injection possible.
  if (trimmed.startsWith("/")) return trimmed;

  // Allow image data URLs only.
  if (/^data:image\/[a-z0-9.+-]+;base64,[a-z0-9+/=\s]+$/i.test(trimmed)) {
    return trimmed;
  }

  // Allow browser-created blob URLs only (blob:<origin>/<uuid>).
  if (/^blob:https?:\/\/[^/]+\/[0-9a-f-]{36}$/i.test(trimmed)) {
    return trimmed;
  }

  // Allow standard remote image URLs over HTTP(S).
  // Return `parsed.href` (the browser-normalised form) rather than the raw
  // input so that the sanitised value is traceable by static-analysis tools.
  try {
    const parsed = new URL(trimmed);
    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
      return undefined;
    }
    return parsed.href; // ← reconstructed, not the raw user-controlled string
  } catch {
    return undefined;
  }
};

function Avatar({ className, ...props }: React.ComponentProps<"div">) {
  const [imageLoaded, setImageLoaded] = React.useState(false);

  const contextValue = React.useMemo(
    () => ({ imageLoaded, setImageLoaded }),
    [imageLoaded],
  );

  return (
    <AvatarContext.Provider value={contextValue}>
      <div
        className={cn(
          "relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full",
          className,
        )}
        {...props}
      />
    </AvatarContext.Provider>
  );
}

function AvatarImage({
  className,
  src,
  alt,
  ...props
}: React.ComponentProps<"img">) {
  const { setImageLoaded } = React.useContext(AvatarContext);

  // ── Derived-state pattern (React docs: "You Might Not Need an Effect") ──
  //
  // We store the last-seen `src` alongside `hasError` in a single state
  // object. When `src` changes we reset `hasError` synchronously *during
  // render* — no effect required, no cascading renders.
  //
  // This fixes the ESLint react-hooks/set-state-in-effect violation that
  // arose from calling setHasError() directly inside a useEffect body.
  const [{ prevSrc, hasError }, setErrorState] = React.useState<{
    prevSrc: typeof src;
    hasError: boolean;
  }>({ prevSrc: src, hasError: false });

  if (src !== prevSrc) {
    // Safe: we are only calling this component's own state setter during
    // render, which React explicitly supports for derived-state resets.
    setErrorState({ prevSrc: src, hasError: false });
  }

  // Notify the parent Avatar that the image is no longer loaded whenever
  // src changes. Kept in an effect because it calls an *external* (context)
  // setter — that is the correct use case for useEffect.
  React.useEffect(() => {
    setImageLoaded(false);
  }, [src, setImageLoaded]);

  // Validate and sanitise src — returns a protocol-safe URL or undefined.
  const safeSrc = React.useMemo(
    () => sanitizeImageUrl(src as string | undefined),
    [src],
  );

  if (hasError || !safeSrc) {
    return null;
  }

  return (
    /* eslint-disable-next-line @next/next/no-img-element */
    <img
      src={safeSrc}
      alt={alt ?? ""}
      className={cn("aspect-square h-full w-full object-cover", className)}
      onLoad={() => setImageLoaded(true)}
      onError={() => {
        setErrorState((prev) => ({ ...prev, hasError: true }));
        setImageLoaded(false);
      }}
      {...props}
    />
  );
}

function AvatarFallback({
  className,
  children,
  dataSeed,
  ...props
}: React.PropsWithChildren<
  React.ComponentProps<"div"> & { dataSeed?: string }
>) {
  const { imageLoaded } = React.useContext(AvatarContext);

  if (imageLoaded) {
    return null;
  }

  // Safely extract text from children for the seed, or default to "user".
  const getSeedText = (): string => {
    if (dataSeed) return dataSeed;
    if (typeof children === "string") return children.trim();
    if (typeof children === "number") return String(children);
    return "user";
  };

  const seedText = getSeedText() || "user";
  // encodeURIComponent prevents any special characters in seedText from
  // escaping the query-string context in the DiceBear URL.
  const seed = encodeURIComponent(seedText);
  const dicebearUrl = `https://api.dicebear.com/7.x/lorelei/svg?seed=${seed}`;

  return (
    <div
      className={cn(
        "absolute flex h-full w-full items-center justify-center rounded-full bg-muted overflow-hidden",
        className,
      )}
      {...props}
    >
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img
        src={dicebearUrl}
        alt={seedText}
        className="h-full w-full object-cover"
      />
    </div>
  );
}

export { Avatar, AvatarImage, AvatarFallback };
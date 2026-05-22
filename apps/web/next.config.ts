import type { NextConfig } from "next";
import type { RemotePattern } from "next/dist/shared/lib/image-config";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin("./src/i18n/request.ts");

const isDevelopment = process.env.NODE_ENV === "development";

const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL?.trim();
const r2PublicUrl =
  process.env.NEXT_PUBLIC_R2_PUBLIC_URL?.trim() ??
  process.env.R2_PUBLIC_URL?.trim() ??
  "https://pub-d746125bb1b6423491682404faec8132.r2.dev";
const apiProxyDestination = apiBaseUrl
  ? (() => {
      try {
        const parsed = new URL(apiBaseUrl);
        return `${parsed.origin}`;
      } catch {
        return null;
      }
    })()
  : null;

// Check if API URL is localhost (for development/testing builds)
const isApiLocalhost =
  apiBaseUrl &&
  (() => {
    try {
      const parsed = new URL(apiBaseUrl);
      return (
        parsed.hostname === "localhost" ||
        parsed.hostname === "127.0.0.1" ||
        parsed.hostname.startsWith("127.")
      );
    } catch {
      return false;
    }
  })();

const remoteImagePatterns: RemotePattern[] = [
  {
    protocol: "http",
    hostname: "localhost",
    pathname: "/uploads/**",
  },
  {
    protocol: "http",
    hostname: "127.0.0.1",
    pathname: "/uploads/**",
  },
  {
    protocol: "https",
    hostname: "source.unsplash.com",
    pathname: "/**",
  },
  {
    protocol: "https",
    hostname: "images.unsplash.com",
    pathname: "/**",
  },
  ...(apiBaseUrl
    ? (() => {
        try {
          const parsed = new URL(apiBaseUrl);
          return [
            {
              protocol: parsed.protocol.replace(":", "") as "http" | "https",
              hostname: parsed.hostname,
              ...(parsed.port ? { port: parsed.port } : {}),
              pathname: "/**",
            } satisfies RemotePattern,
          ];
        } catch {
          return [];
        }
      })()
    : []),
  ...(r2PublicUrl
    ? (() => {
        try {
          const parsed = new URL(r2PublicUrl);
          return [
            {
              protocol: parsed.protocol.replace(":", "") as "http" | "https",
              hostname: parsed.hostname,
              ...(parsed.port ? { port: parsed.port } : {}),
              pathname: "/**",
            } satisfies RemotePattern,
          ];
        } catch {
          return [];
        }
      })()
    : []),
];

// Bundle Analyzer - only load when ANALYZE=true
let withBundleAnalyzer = (config: NextConfig) => config;
if (process.env.ANALYZE === "true") {
  // eslint-disable-next-line @typescript-eslint/no-require-imports
  const bundleAnalyzer = require("@next/bundle-analyzer")({
    enabled: true,
  });
  withBundleAnalyzer = bundleAnalyzer;
}

const nextConfig: NextConfig = {
  async rewrites() {
    if (!apiProxyDestination) {
      return [];
    }

    return [
      {
        source: "/api/v1/:path*",
        destination: `${apiProxyDestination}/api/v1/:path*`,
      },
      {
        source: "/internal/:path*",
        destination: `${apiProxyDestination}/internal/:path*`,
      },
      {
        source: "/uploads/:path*",
        destination: `${apiProxyDestination}/uploads/:path*`,
      },
    ];
  },

  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          // HSTS (only in production)
          ...(isDevelopment
            ? []
            : [
                {
                  key: "Strict-Transport-Security",
                  value: "max-age=31536000; includeSubDomains; preload",
                },
              ]),

          // Resource hints for critical resources
          {
            key: "Link",
            value: [
              // Preconnect to Google Fonts
              '<https://fonts.googleapis.com>; rel=preconnect',
              '<https://fonts.gstatic.com>; rel=preconnect; crossorigin',
              // DNS prefetch for R2 (image CDN)
              '<https://pub-d746125bb1b6423491682404faec8132.r2.dev>; rel=dns-prefetch',
            ].join(", "),
          },

          // CSP (optimized for Next.js and React)
          // Note: 'unsafe-inline' and 'unsafe-eval' are required for Next.js hydration
          // and development mode. In production, consider using nonce-based CSP.
          {
            key: "Content-Security-Policy",
            value: [
              "default-src 'self'",
              isDevelopment || isApiLocalhost
                ? "img-src 'self' https: data: blob: http://localhost:* http://127.0.0.1:*"
                : "img-src 'self' https: data: blob:",
              "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
              "script-src-elem 'self' 'unsafe-inline'",
              "worker-src 'self' blob:",
              "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
              "font-src 'self' https://fonts.gstatic.com data:",
              // Allow localhost connections in development or when API URL is localhost
              isDevelopment || isApiLocalhost
                ? "connect-src 'self' https: wss: ws: http://localhost:* http://127.0.0.1:*"
                : "connect-src 'self' https: wss: ws:",
              // Allow map iframe sources used by Travel Planner and other map integrations
              "frame-src 'self' https://www.openstreetmap.org/ https://*.openstreetmap.org/ https://carto.com/",
              "frame-ancestors 'none'",
              "base-uri 'self'",
              "form-action 'self'",
              ...(isDevelopment ? [] : ["upgrade-insecure-requests"]),
            ].join("; "),
          },

          // Clickjacking
          { key: "X-Frame-Options", value: "DENY" },

          // MIME Sniffing
          { key: "X-Content-Type-Options", value: "nosniff" },

          // Referrer policy
          {
            key: "Referrer-Policy",
            value: "strict-origin-when-cross-origin",
          },

          // Feature permissions
          {
            key: "Permissions-Policy",
            value: [
              "camera=(self)",
              "microphone=()",
              "geolocation=(self)",
              "fullscreen=(self)",
              "payment=()",
              "usb=()",
            ].join(", "),
          },

          // Cross-Origin protections
          // Relaxed in development for Next.js HMR and WebSocket connections
          ...(isDevelopment
            ? []
            : [
                {
                  key: "Cross-Origin-Resource-Policy",
                  value: "same-origin",
                },
                {
                  key: "Cross-Origin-Opener-Policy",
                  value: "same-origin",
                },
              ]),

          // COEP disabled for compatibility
          // { key: "Cross-Origin-Embedder-Policy", value: "require-corp" },
        ],
      },

      // static asset caching
      {
        source:
          "/:path*.(ico|png|jpg|jpeg|svg|gif|webp|woff|woff2|ttf|eot)",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },

      // HTML pages — short-term cache with revalidation (stale-while-revalidate)
      {
        source: "/:path((?!_next).*)?",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=3600, s-maxage=3600, stale-while-revalidate=86400",
          },
        ],
      },
    ];
  },

  images: {
    remotePatterns: remoteImagePatterns,
    // Optimize image format and quality for better performance
    formats: ["image/avif", "image/webp"],
    // Reduce image quality slightly for better performance
    // while maintaining visual quality on landing page
    minimumCacheTTL: 60 * 60 * 24 * 365, // 1 year for hashed images
  },

  poweredByHeader: false,
};

export default withNextIntl(withBundleAnalyzer(nextConfig));
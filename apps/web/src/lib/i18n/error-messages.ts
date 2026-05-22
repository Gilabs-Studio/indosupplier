/**
 * Helper to get error messages based on current locale
 * This works outside React context (e.g., in axios interceptors)
 */

import { getLocalePreferenceFromBrowser } from "@/lib/locale-preference";

type Locale = "en" | "id";

interface ErrorMessages {
  network: {
    timeout: { title: string; description: string };
    connectionFailed: { title: string; description: string };
    serverUnreachable: { title: string; description: string };
    generic: { title: string; description: string };
  };
  backend: {
    validationError: { title: string; description: string };
    fieldError: { title: string; description: string };
    unauthorized: { title: string; description: string };
    forbidden: { title: string; description: string };
    notFound: { title: string; description: string };
    conflict: { title: string; description: string };
    emailExists: { title: string; description: string };
    resourceExists: { title: string; description: string };
    serverError: { title: string; description: string };
    serviceUnavailable: { title: string; description: string };
    rateLimit: { title: string; description: string };
    unexpectedError: { title: string; description: string };
    invalidFormat: { title: string; description: string };
    csrfError: { title: string; description: string };
  };
}

const cachedMessages: { en?: ErrorMessages; id?: ErrorMessages } = {};

/**
 * Get current locale from URL or default to 'en'
 */
function getCurrentLocale(): Locale {
  if (typeof window === "undefined") {
    return "en";
  }

  // Try to get locale from URL path (e.g., /en/..., /id/...)
  const pathname = window.location.pathname;
  const localeMatch = pathname.match(/^\/(en|id)(\/|$)/);
  if (localeMatch) {
    return localeMatch[1] as Locale;
  }

  const savedLocale = getLocalePreferenceFromBrowser();
  if (savedLocale) {
    return savedLocale;
  }

  // Default to English
  return "en";
}

/**
 * Load error messages for a specific locale
 */
async function loadErrorMessages(locale: Locale): Promise<ErrorMessages> {
  // Return cached if available
  if (cachedMessages[locale]) {
    return cachedMessages[locale]!;
  }

  try {
    const messages = await import(`./errors/${locale}.json`);
    cachedMessages[locale] = messages.default as ErrorMessages;
    return cachedMessages[locale]!;
  } catch (error) {
    console.error(`Failed to load error messages for locale ${locale}:`, error);
    // Fallback to English if locale not found
    if (locale !== "en") {
      return loadErrorMessages("en");
    }
    // If English also fails, return a minimal fallback
    return {
      network: {
        timeout: {
          title: "Connection Timeout",
          description: "The request took too long to complete.",
        },
        connectionFailed: {
          title: "Connection Failed",
          description: "Unable to connect to the server.",
        },
        serverUnreachable: {
          title: "Server Unreachable",
          description: "Cannot connect to the server.",
        },
        generic: {
          title: "Network Error",
          description: "A network error occurred.",
        },
      },
      backend: {
        validationError: {
          title: "Validation Error",
          description: "Please check your input.",
        },
        fieldError: {
          title: "Invalid Input",
          description: "There is an error in the {field} field.",
        },
        unauthorized: {
          title: "Session Expired",
          description: "Please log in again.",
        },
        forbidden: {
          title: "Access Denied",
          description: "You do not have permission.",
        },
        notFound: {
          title: "Not Found",
          description: "The requested information could not be found.",
        },
        conflict: {
          title: "Duplicate Entry",
          description: "This information already exists.",
        },
        emailExists: {
          title: "Email Already Registered",
          description: "The email is already registered.",
        },
        resourceExists: {
          title: "Duplicate Entry",
          description: "This value already exists.",
        },
        serverError: {
          title: "Server Error",
          description: "An error occurred on the server.",
        },
        serviceUnavailable: {
          title: "Service Unavailable",
          description: "The service is temporarily unavailable.",
        },
        rateLimit: {
          title: "Too Many Requests",
          description: "Please wait before trying again.",
        },
        unexpectedError: {
          title: "Unexpected Error",
          description: "An unexpected error occurred.",
        },
        invalidFormat: {
          title: "Invalid Response Format",
          description: "The server returned an unexpected response.",
        },
        csrfError: {
          title: "Session Expired",
          description: "Your session has expired. Please refresh the page and try again.",
        },
      },
    };
  }
}

/**
 * Get error message with optional parameter replacement
 */
function replaceParams(text: string, params: Record<string, string>): string {
  let result = text;
  for (const [key, value] of Object.entries(params)) {
    result = result.replace(new RegExp(`\\{${key}\\}`, "g"), value);
  }
  return result;
}

/**
 * Get error messages for current locale
 */
export async function getErrorMessages(): Promise<ErrorMessages> {
  const locale = getCurrentLocale();
  return loadErrorMessages(locale);
}

/**
 * Get error messages synchronously (uses cached or default)
 * This is for cases where we can't await (like in interceptors)
 */
export function getErrorMessagesSync(): ErrorMessages {
  const locale = getCurrentLocale();
  const cached = cachedMessages[locale];
  if (cached) {
    return cached;
  }

  // If not cached, load asynchronously and return default for now
  loadErrorMessages(locale).catch(console.error);

  // Return English defaults as fallback
  return {
    network: {
      timeout: {
        title: "Connection Timeout",
        description:
          "The request took too long to complete. Please check your internet connection and try again.",
      },
      connectionFailed: {
        title: "Connection Failed",
        description:
          "Unable to connect to the server. Please check your internet connection or contact support if the problem persists.",
      },
      serverUnreachable: {
        title: "Server Unreachable",
        description:
          "Cannot connect to the server. The server may be down or your internet connection is unstable. Please try again later.",
      },
      generic: {
        title: "Network Error",
        description:
          "A network error occurred. Please check your internet connection and try again.",
      },
    },
    backend: {
      validationError: {
        title: "Validation Error",
        description: "Please check the information you entered and try again.",
      },
      fieldError: {
        title: "Invalid Input",
        description: "There is an error in the {field} field: {message}",
      },
      unauthorized: {
        title: "Session Expired",
        description:
          "Your session has expired. Please log in again to continue.",
      },
      forbidden: {
        title: "Access Denied",
        description:
          "You do not have permission to perform this action. Please contact your administrator if you believe this is an error.",
      },
      notFound: {
        title: "Not Found",
        description:
          "The requested information could not be found. It may have been deleted or moved.",
      },
      conflict: {
        title: "Duplicate Entry",
        description:
          "This information already exists in the system. Please use different information.",
      },
      emailExists: {
        title: "Email Already Registered",
        description:
          'The email "{email}" is already registered. Please use a different email address.',
      },
      resourceExists: {
        title: "Duplicate Entry",
        description:
          'The {field} "{value}" already exists. Please use a different value.',
      },
      serverError: {
        title: "Server Error",
        description:
          "An error occurred on the server. Our team has been notified. Please try again later or contact support if the problem persists.",
      },
      serviceUnavailable: {
        title: "Service Unavailable",
        description:
          "The service is temporarily unavailable. Please try again in a few moments.",
      },
      rateLimit: {
        title: "Too Many Requests",
        description:
          "You have made too many requests. Please try again in {countdown}.",
      },
      unexpectedError: {
        title: "Unexpected Error",
        description:
          "An unexpected error occurred. Please try again or contact support if the problem persists.",
      },
      invalidFormat: {
        title: "Invalid Response Format",
        description:
          "The server returned an unexpected response format. Please contact support.",
      },
      csrfError: {
        title: "Session Expired",
        description:
          "Your session has expired. Please refresh the page and try again.",
      },
    },
  };
}

/**
 * Helper to format error message with parameters
 */
export function formatError(
  category: "network" | "backend",
  key: keyof ErrorMessages["network"] | keyof ErrorMessages["backend"],
  params?: Record<string, string>,
): { title: string; description: string } {
  const messages = getErrorMessagesSync();
  const message = messages[category][
    key as keyof (typeof messages)[typeof category]
  ] as {
    title: string;
    description: string;
  };

  if (params) {
    return {
      title: replaceParams(message.title, params),
      description: replaceParams(message.description, params),
    };
  }

  return message;
}

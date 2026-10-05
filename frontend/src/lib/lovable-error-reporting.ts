export function reportLovableError(error: unknown, context?: Record<string, unknown>) {
  if (typeof window !== "undefined" && import.meta.env?.DEV) {
    console.error("[LovableError]", error, context);
  }
}

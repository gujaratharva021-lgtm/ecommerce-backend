export function getErrorMessage(err: unknown, fallback = 'Something went wrong. Please try again.'): string {
  const anyErr = err as any
  return anyErr?.response?.data?.error ?? anyErr?.message ?? fallback
}

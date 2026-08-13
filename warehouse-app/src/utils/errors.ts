export function getErrorMessage(err: unknown, fallback = 'Something went wrong. Please try again.'): string {
  const anyErr = err as any

  // A server error message (400/403/404/409/500 etc, whatever the backend
  // put in { error: "..." }) is always the most specific thing we have.
  if (anyErr?.response?.data?.error) {
    return anyErr.response.data.error
  }

  // No response at all means the request never reached the server -
  // network failure or our own timeout. Distinguish those two, since
  // "check your connection" vs "try again" point the person in different
  // directions.
  if (anyErr?.code === 'ECONNABORTED') {
    return 'The request timed out. Please check your connection and try again.'
  }
  if (anyErr?.message === 'Network Error') {
    return 'Could not reach the server. Please check your connection and try again.'
  }

  return anyErr?.message ?? fallback
}

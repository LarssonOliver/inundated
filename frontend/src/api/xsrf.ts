const XSRF_COOKIE = "XSRF-TOKEN";

/**
 * Reads the anti-CSRF token the server sets in the `XSRF-TOKEN` cookie. The
 * server requires it echoed back in the `X-XSRF-TOKEN` header on every mutating
 * request. Returns an empty string when the cookie is not present yet.
 */
export function xsrfToken(): string {
  for (const pair of document.cookie.split(";")) {
    const eq = pair.indexOf("=");
    if (eq === -1) continue;
    if (pair.slice(0, eq).trim() === XSRF_COOKIE) {
      return decodeURIComponent(pair.slice(eq + 1).trim());
    }
  }
  return "";
}

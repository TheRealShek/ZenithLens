/**
 * API fetch wrapper. All API calls go through this function.
 * Returns { data, error } — never throws.
 */
export async function fetchAPI(endpoint, options = {}) {
  try {
    const res = await fetch(endpoint, options);
    const data = await res.json().catch(() => null);
    if (!res.ok) return { error: data?.error || res.statusText };
    return { data };
  } catch (e) {
    return { error: e.message };
  }
}

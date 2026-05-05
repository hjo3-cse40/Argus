export type Platform = {
  id: string;
  name: string;
  discord_webhook: string;
  created_at: string;
};

export type DestinationFilter = {
  id: string;
  platform_id: string;
  filter_type: "keyword_include" | "keyword_exclude";
  pattern: string;
  created_at: string;
};

export type DeliveryStatus = "queued" | "delivered" | "failed";

export type Delivery = {
  event_id: string;
  source: string;
  title: string;
  url: string;
  status: DeliveryStatus;
  subsource_id?: string;
  created_at: string;
  updated_at: string;
  retry_count: number;
  last_error?: string;
};

/**
 * All browser requests go to same-origin `/api/...`, which the Next.js route
 * `app/api/[...path]/route.ts` proxies to Go (`API_UPSTREAM`, default http://127.0.0.1:8080).
 * Do not use NEXT_PUBLIC_API_URL for fetch — it bypasses the proxy and breaks when CORS
 * or localhost/IPv6 differ.
 */
function apiUrl(path: string): string {
  const p = path.startsWith("/") ? path : `/${path}`;
  return p;
}

async function errorMessageFromResponse(res: Response, fallback: string): Promise<string> {
  const text = await res.text();
  if (!text) return `${fallback} (${res.status})`;
  try {
    const body = JSON.parse(text) as { error?: string; details?: string[] };
    if (body.details?.length) return `${body.error ?? "Error"}: ${body.details.join(", ")}`;
    if (body.error) return body.error;
  } catch {
    /* not JSON */
  }
  const snippet = text.replace(/\s+/g, " ").slice(0, 160);
  return snippet ? `${fallback} (${res.status}): ${snippet}` : `${fallback} (${res.status})`;
}

async function readJson<T>(res: Response, fallback: string): Promise<T> {
  const text = await res.text();
  if (!res.ok) {
    let msg = `${fallback} (${res.status})`;
    if (text) {
      try {
        const body = JSON.parse(text) as { error?: string; details?: string[] };
        if (body.details?.length) msg = `${body.error ?? "Error"}: ${body.details.join(", ")}`;
        else if (body.error) msg = body.error;
        else msg = `${fallback} (${res.status}): ${text.slice(0, 120)}`;
      } catch {
        msg = `${fallback} (${res.status}): ${text.slice(0, 120)}`;
      }
    }
    throw new Error(msg);
  }
  if (!text) throw new Error(`${fallback}: empty response`);
  try {
    return JSON.parse(text) as T;
  } catch {
    throw new Error(`${fallback}: invalid JSON — ${text.slice(0, 80)}`);
  }
}

async function doFetch(input: string, init?: RequestInit): Promise<Response> {
  try {
    return await fetch(input, init);
  } catch (e) {
    const msg = e instanceof Error ? e.message : "Network error";
    throw new Error(
      `${msg}. Check that Next is running and set API_UPSTREAM in frontend/.env.local if the Go API is not at http://127.0.0.1:8080.`
    );
  }
}

export type Subsource = {
  id: string;
  platform_id: string;
  platform_name: string;
  name: string;
  identifier: string;
  url?: string;
  created_at: string;
};

export type CreateSubsourcePayload = {
  name: string;
  identifier: string;
  url?: string;
};

export type UpdateSubsourcePayload = {
  name: string;
  identifier: string;
  url?: string;
};

export type CreatePlatformPayload = {
  name: string;
  discord_webhook: string;
  webhook_secret?: string;
};

export async function createPlatform(body: CreatePlatformPayload): Promise<Platform> {
  const res = await doFetch(apiUrl("/api/platforms"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  return readJson<Platform>(res, "Create platform failed");
}

export async function fetchPlatforms(): Promise<Platform[]> {
  const res = await doFetch(apiUrl("/api/platforms"));
  const data = await readJson<unknown>(res, "Load platforms failed");
  if (!Array.isArray(data)) {
    throw new Error("Load platforms failed: server did not return a JSON array");
  }
  return data as Platform[];
}

export type FetchDeliveriesOptions = {
  status?: DeliveryStatus;
  limit?: number;
  offset?: number;
  platformId?: string;
  subsourceId?: string;
};

export async function fetchDeliveries(
  options: FetchDeliveriesOptions = {}
): Promise<Delivery[]> {
  const params = new URLSearchParams();
  if (options.status) params.set("status", options.status);
  if (options.limit && options.limit > 0) params.set("limit", String(options.limit));
  if (typeof options.offset === "number" && options.offset >= 0) {
    params.set("offset", String(options.offset));
  }
  if (options.platformId) params.set("platform_id", options.platformId);
  if (options.subsourceId) params.set("subsource_id", options.subsourceId);

  const query = params.toString();
  const path = query ? `/api/deliveries?${query}` : "/api/deliveries";
  const res = await doFetch(apiUrl(path));
  const data = await readJson<unknown>(res, "Load deliveries failed");
  if (!Array.isArray(data)) {
    throw new Error("Load deliveries failed: server did not return a JSON array");
  }
  return data as Delivery[];
}

export async function fetchFilters(platformId: string): Promise<DestinationFilter[]> {
  const res = await doFetch(
    apiUrl(`/api/platforms/${encodeURIComponent(platformId)}/filters`)
  );
  const data = await readJson<unknown>(res, "Load filters failed");
  if (!Array.isArray(data)) {
    throw new Error("Load filters failed: server did not return a JSON array");
  }
  return data as DestinationFilter[];
}

export async function createFilter(
  platformId: string,
  filterType: "keyword_include" | "keyword_exclude",
  pattern: string
): Promise<DestinationFilter> {
  const res = await doFetch(
    apiUrl(`/api/platforms/${encodeURIComponent(platformId)}/filters`),
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ filter_type: filterType, pattern }),
    }
  );
  return readJson<DestinationFilter>(res, "Create filter failed");
}

export async function deleteFilter(filterId: string): Promise<void> {
  const res = await doFetch(apiUrl(`/api/filters/${encodeURIComponent(filterId)}`), {
    method: "DELETE",
  });
  if (res.status === 204) return;
  if (res.ok) return;
  throw new Error(await errorMessageFromResponse(res, "Delete filter failed"));
}

export async function fetchSubsources(platformId: string): Promise<Subsource[]> {
  const res = await doFetch(
    apiUrl(`/api/platforms/${encodeURIComponent(platformId)}/subsources`)
  );
  const data = await readJson<unknown>(res, "Load sub-channels failed");
  if (!Array.isArray(data)) {
    throw new Error("Load sub-channels failed: server did not return a JSON array");
  }
  return data as Subsource[];
}

export async function createSubsource(
  platformId: string,
  body: CreateSubsourcePayload
): Promise<Subsource> {
  const payload: Record<string, string> = {
    name: body.name.trim(),
    identifier: body.identifier.trim(),
  };
  const u = body.url?.trim();
  if (u) payload.url = u;

  const res = await doFetch(
    apiUrl(`/api/platforms/${encodeURIComponent(platformId)}/subsources`),
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }
  );
  return readJson<Subsource>(res, "Create sub-channel failed");
}

export async function updateSubsource(
  subsourceId: string,
  body: UpdateSubsourcePayload
): Promise<Subsource> {
  const payload: Record<string, string> = {
    name: body.name.trim(),
    identifier: body.identifier.trim(),
  };
  const u = body.url?.trim();
  if (u) payload.url = u;

  const res = await doFetch(apiUrl(`/api/subsources/${encodeURIComponent(subsourceId)}`), {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return readJson<Subsource>(res, "Update sub-channel failed");
}

export async function deleteSubsource(subsourceId: string): Promise<void> {
  const res = await doFetch(apiUrl(`/api/subsources/${encodeURIComponent(subsourceId)}`), {
    method: "DELETE",
  });
  if (res.status === 204) return;
  if (res.ok) return;
  throw new Error(await errorMessageFromResponse(res, "Delete sub-channel failed"));
}

export type LoginPayload = {
  email: string;
  password: string;
};

export type LoginResponse = {
  token: string;
};

export async function login(body: LoginPayload): Promise<LoginResponse> {
  const res = await doFetch(apiUrl("/api/login"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  return readJson<LoginResponse>(res, "Login failed");
}

export async function register(body: LoginPayload): Promise<void> {
  const res = await doFetch(apiUrl("/api/register"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (res.status === 204 || res.ok) return;
  throw new Error(await errorMessageFromResponse(res, "Registration failed"));
}

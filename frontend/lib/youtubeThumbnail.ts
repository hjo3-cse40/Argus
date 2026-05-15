const YT_ID = /^[a-zA-Z0-9_-]{11}$/;

function normalizeId(id: string | null | undefined): string | null {
  if (!id) return null;
  const trimmed = id.trim();
  if (!YT_ID.test(trimmed)) return null;
  return trimmed;
}

/**
 * Returns the canonical 11-char video id for common YouTube watch / shorts / live URLs.
 */
export function extractYouTubeVideoId(raw: string | undefined | null): string | null {
  if (!raw) return null;
  try {
    const u = new URL(raw);
    const host = u.hostname.replace(/^www\./, "");

    if (host === "youtu.be") {
      return normalizeId(u.pathname.split("/").filter(Boolean)[0]);
    }

    if (host === "m.youtube.com" || host === "youtube.com" || host.endsWith(".youtube.com")) {
      if (u.pathname === "/watch" || u.pathname.startsWith("/watch")) {
        const fromQuery = normalizeId(u.searchParams.get("v"));
        if (fromQuery) return fromQuery;
        const rest = u.pathname.replace(/^\/watch\/?/, "").split("/")[0];
        return normalizeId(rest || null);
      }
      const seg = u.pathname.split("/").filter(Boolean);
      const kind = seg[0];
      if (kind === "shorts" || kind === "live" || kind === "embed" || kind === "v") {
        return normalizeId(seg[1]);
      }
    }
  } catch {
    return null;
  }
  return null;
}

export function youtubeThumbnailUrl(videoId: string, quality: "mqdefault" | "hqdefault" = "mqdefault"): string {
  return `https://i.ytimg.com/vi/${encodeURIComponent(videoId)}/${quality}.jpg`;
}

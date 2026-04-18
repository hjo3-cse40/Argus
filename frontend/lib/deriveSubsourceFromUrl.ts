/**
 * Derive RSS / Argus subsource identifier from a user-pasted URL.
 * Must match how `backend/cmd/rss` builds RSSHub paths:
 * youtube → youtube/channel/{identifier}, reddit → reddit/subreddit/{identifier}, x → twitter/user/{identifier}
 */

export type DeriveResult =
  | { ok: true; identifier: string }
  | { ok: false; error: string };

function normalizeInput(raw: string): URL | null {
  const trimmed = raw.trim();
  if (!trimmed) return null;
  try {
    return new URL(trimmed.includes("://") ? trimmed : `https://${trimmed}`);
  } catch {
    return null;
  }
}

function deriveYouTube(url: URL): DeriveResult {
  const path = url.pathname;
  const channel = path.match(/\/channel\/([a-zA-Z0-9_-]+)/);
  if (channel) return { ok: true, identifier: channel[1] };

  const at = path.match(/\/@([^/]+)/);
  if (at) return { ok: true, identifier: decodeURIComponent(at[1]) };

  const c = path.match(/\/c\/([^/]+)/);
  if (c) return { ok: true, identifier: decodeURIComponent(c[1]) };

  const user = path.match(/\/user\/([^/]+)/);
  if (user) return { ok: true, identifier: decodeURIComponent(user[1]) };

  return {
    ok: false,
    error:
      "Could not read a YouTube channel from this URL. Try a link that contains /channel/UC… or /@handle.",
  };
}

function deriveReddit(url: URL): DeriveResult {
  const path = url.pathname;
  const sub = path.match(/\/r\/([^/]+)/);
  if (sub) return { ok: true, identifier: sub[1] };
  return {
    ok: false,
    error: "Could not read a subreddit from this URL. Use a link like https://reddit.com/r/name",
  };
}

function deriveX(url: URL): DeriveResult {
  const host = url.hostname.replace(/^www\./, "");
  if (host !== "x.com" && host !== "twitter.com") {
    return { ok: false, error: "Use an x.com or twitter.com profile URL." };
  }
  const parts = url.pathname.split("/").filter(Boolean);
  const skip = new Set([
    "intent",
    "i",
    "home",
    "search",
    "explore",
    "settings",
    "compose",
    "hashtag",
  ]);
  const first = parts[0];
  if (!first || skip.has(first.toLowerCase())) {
    return {
      ok: false,
      error: "Paste a profile URL like https://x.com/username or https://twitter.com/username",
    };
  }
  return { ok: true, identifier: first.replace(/^@/, "") };
}

/**
 * @param platformName lowercase platform name from API: `youtube` | `reddit` | `x`
 */
export function deriveSubsourceIdentifierFromUrl(
  platformName: string,
  rawUrl: string
): DeriveResult {
  const url = normalizeInput(rawUrl);
  if (!url) {
    return { ok: false, error: "Enter a valid URL." };
  }

  switch (platformName) {
    case "youtube":
      return deriveYouTube(url);
    case "reddit":
      return deriveReddit(url);
    case "x":
      return deriveX(url);
    default:
      return { ok: false, error: `Unknown platform: ${platformName}` };
  }
}

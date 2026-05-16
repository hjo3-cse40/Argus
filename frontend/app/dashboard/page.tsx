"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import Image from "next/image";
import { AppNav } from "@/components/AppNav";
import { AppTopNav } from "@/components/AppTopNav";
import { RequireAuth } from "@/components/RequireAuth";
import { fetchDeliveries, fetchPlatforms, type Delivery } from "@/lib/api";
import {
  type DashboardSortPreset,
  sortDashboardPosts,
} from "@/lib/notificationSort";
import { subsourceDisplayLine } from "@/lib/subsourceDisplay";
import { extractYouTubeVideoId, youtubeThumbnailUrl } from "@/lib/youtubeThumbnail";
import "../app-shell.css";

type DashboardPost = {
  id: string;
  title: string;
  url?: string;
  updatedAt?: string;
  subsourceName?: string;
  subsourceIdentifier?: string;
};

function formatRelativeTime(iso?: string): string {
  if (!iso) return "";
  const then = new Date(iso).getTime();
  const now = Date.now();
  const seconds = Math.max(1, Math.floor((now - then) / 1000));
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

function PlatformSection({
  title,
  posts,
}: {
  title: string;
  posts: DashboardPost[];
}) {
  const scrollRef = useRef<HTMLDivElement | null>(null);
  const isYoutube = title === "youtube";

  const scroll = (dir: "left" | "right") => {
    if (!scrollRef.current) return;
    scrollRef.current.scrollBy({
      left: dir === "left" ? -300 : 300,
      behavior: "smooth",
    });
  };

  const titleBlock = (post: DashboardPost) => {
    const subLine = subsourceDisplayLine(post.subsourceName, post.subsourceIdentifier);
    const inner =
      post.url ? (
        <a
          href={post.url}
          target="_blank"
          rel="noopener noreferrer"
          className="app-post-title-link"
        >
          {post.title}
        </a>
      ) : (
        <span>{post.title}</span>
      );
    return (
      <>
        <p style={{ margin: 0 }}>{inner}</p>
        {subLine ? <p className="app-subsource-line">{subLine}</p> : null}
      </>
    );
  };

  return (
    <div className="app-section-block">
      <div className="app-section-head">
        <h3>{title}</h3>
        <div className="app-icon-btns">
          <button
            type="button"
            className="app-icon-btn"
            onClick={() => scroll("left")}
            aria-label="Scroll left"
          >
            ←
          </button>
          <button
            type="button"
            className="app-icon-btn"
            onClick={() => scroll("right")}
            aria-label="Scroll right"
          >
            →
          </button>
        </div>
      </div>
      <div ref={scrollRef} className="app-row-scroll">
        {posts.length === 0 ? (
          <p className="app-muted">No delivered notifications yet.</p>
        ) : (
          posts.map((post) => {
            const ytId =
              isYoutube && post.url ? extractYouTubeVideoId(post.url) : null;
            const thumbSrc = ytId ? youtubeThumbnailUrl(ytId) : null;
            return (
              <div
                key={post.id}
                className={`app-post-card${isYoutube ? " app-post-card--youtube" : ""}`}
              >
                {thumbSrc && post.url ? (
                  <a
                    href={post.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="app-yt-thumb-link"
                    aria-label={`Video thumbnail: ${post.title}`}
                  >
                    <Image
                      src={thumbSrc}
                      alt=""
                      fill
                      className="app-yt-thumb"
                      sizes="280px"
                    />
                  </a>
                ) : null}
                {titleBlock(post)}
                {post.updatedAt ? (
                  <p className="app-muted">{formatRelativeTime(post.updatedAt)}</p>
                ) : null}
                {post.url ? (
                  <a
                    href={post.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="app-link-quiet"
                  >
                    Open
                  </a>
                ) : null}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}

export default function Dashboard() {
  const [deliveryGroups, setDeliveryGroups] = useState<{
    youtube: DashboardPost[];
    reddit: DashboardPost[];
    x: DashboardPost[];
  }>({ youtube: [], reddit: [], x: [] });
  const [dashboardSort, setDashboardSort] = useState<DashboardSortPreset>("updated_desc");
  const [deliveriesLoading, setDeliveriesLoading] = useState(true);
  const [deliveriesError, setDeliveriesError] = useState<string | null>(null);

  const [platformCount, setPlatformCount] = useState<number | null>(null);
  const [platformsError, setPlatformsError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const list = await fetchPlatforms();
        if (!cancelled) {
          setPlatformCount(list.length);
          setPlatformsError(null);
        }
      } catch (e) {
        if (!cancelled) {
          setPlatformCount(null);
          setPlatformsError(e instanceof Error ? e.message : "Failed to load platforms");
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setDeliveriesLoading(true);
      setDeliveriesError(null);
      try {
        const delivered = await fetchDeliveries({
          status: "delivered",
          limit: 100,
        });
        const next = {
          youtube: [],
          reddit: [],
          x: [],
        } as {
          youtube: DashboardPost[];
          reddit: DashboardPost[];
          x: DashboardPost[];
        };
        for (const d of delivered) {
          const source = d.source === "twitter" ? "x" : d.source;
          if (source !== "youtube" && source !== "reddit" && source !== "x") continue;
          const group = next[source];
          if (group.length >= 15) continue;
          if (group.some((item) => item.id === d.event_id)) continue;
          group.push({
            id: d.event_id,
            title: d.title || "(untitled)",
            url: d.url,
            updatedAt: d.updated_at,
            subsourceName: d.subsource_name,
            subsourceIdentifier: d.subsource_identifier,
          });
        }
        if (!cancelled) setDeliveryGroups(next);
      } catch (e) {
        if (!cancelled) {
          setDeliveriesError(e instanceof Error ? e.message : "Failed to load notifications");
          setDeliveryGroups({ youtube: [], reddit: [], x: [] });
        }
      } finally {
        if (!cancelled) setDeliveriesLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const es = new EventSource("/api/deliveries/stream");
    es.addEventListener("delivered", (evt) => {
      try {
        const incoming = JSON.parse((evt as MessageEvent).data) as Delivery;
        const source = incoming.source === "twitter" ? "x" : incoming.source;
        if (source !== "youtube" && source !== "reddit" && source !== "x") return;
        setDeliveryGroups((prev) => {
          const current = prev[source];
          if (current.some((item) => item.id === incoming.event_id)) return prev;
          const nextList = [
            {
              id: incoming.event_id,
              title: incoming.title || "(untitled)",
              url: incoming.url,
              updatedAt: incoming.updated_at,
              subsourceName: incoming.subsource_name,
              subsourceIdentifier: incoming.subsource_identifier,
            },
            ...current,
          ].slice(0, 15);
          return { ...prev, [source]: nextList };
        });
      } catch {
        // ignore malformed payloads
      }
    });
    return () => es.close();
  }, []);

  const deliveredTotal =
    deliveryGroups.youtube.length + deliveryGroups.reddit.length + deliveryGroups.x.length;

  const sortedYoutube = useMemo(
    () => sortDashboardPosts(deliveryGroups.youtube, dashboardSort),
    [deliveryGroups.youtube, dashboardSort]
  );
  const sortedReddit = useMemo(
    () => sortDashboardPosts(deliveryGroups.reddit, dashboardSort),
    [deliveryGroups.reddit, dashboardSort]
  );
  const sortedX = useMemo(
    () => sortDashboardPosts(deliveryGroups.x, dashboardSort),
    [deliveryGroups.x, dashboardSort]
  );

  return (
    <RequireAuth>
    <div className="app-shell">
      <AppTopNav />
      <div className="app-body">
        <aside className="app-sidebar">
          <AppNav />
          <p className="app-sidebar-hint">
            Tune keyword include/exclude rules on the <strong>Filters</strong> page.
          </p>
          <div className="app-sidebar-sort">
            <label className="app-label" htmlFor="dashboard-sort">
              Sort rows
            </label>
            <select
              id="dashboard-sort"
              className="app-select"
              value={dashboardSort}
              onChange={(e) => setDashboardSort(e.target.value as DashboardSortPreset)}
            >
              <option value="updated_desc">Time · Newest first</option>
              <option value="updated_asc">Time · Oldest first</option>
              <option value="title_asc">Title · A to Z</option>
              <option value="title_desc">Title · Z to A</option>
            </select>
          </div>
        </aside>

        <main className="app-main">
          <div className="app-main-header">
            <div>
              <p className="app-eyebrow">Overview</p>
              <h1 className="app-page-title">
                My <em>dashboard</em>
              </h1>
            </div>
            <p className="app-kicker">Welcome back</p>
          </div>

          {platformsError ? <p className="app-error">{platformsError}</p> : null}

          <div className="app-stat-grid">
            <div className="app-stat">
              <span className="app-stat-val">{deliveredTotal}</span>
              <div className="app-stat-label">Delivered shown</div>
            </div>
            <div className="app-stat">
              <span className="app-stat-val">
                {platformCount === null ? "…" : platformCount}
              </span>
              <div className="app-stat-label">Platforms</div>
            </div>
            <div className="app-stat">
              <span className="app-stat-val">{deliveriesLoading ? "…" : "live"}</span>
              <div className="app-stat-label">Notification feed</div>
            </div>
          </div>
          {deliveriesError ? <p className="app-error">{deliveriesError}</p> : null}

          <PlatformSection title="youtube" posts={sortedYoutube} />
          <PlatformSection title="reddit" posts={sortedReddit} />
          <PlatformSection title="twitter" posts={sortedX} />
        </main>
      </div>
    </div>
    </RequireAuth>
  );
}

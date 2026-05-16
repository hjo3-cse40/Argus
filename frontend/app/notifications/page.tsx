"use client";

import { useEffect, useMemo, useState } from "react";
import { AppNav } from "@/components/AppNav";
import { AppTopNav } from "@/components/AppTopNav";
import { RequireAuth } from "@/components/RequireAuth";
import { fetchDeliveries, type Delivery } from "@/lib/api";
import {
  type NotificationsSortPreset,
  notificationsPresetToQuery,
  sortDeliveries,
} from "@/lib/notificationSort";
import { subsourceDisplayLine } from "@/lib/subsourceDisplay";
import "../app-shell.css";

function formatRelativeTime(iso: string): string {
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

function formatShortDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function NotificationsPage() {
  const pageSize = 15;
  const countBatchSize = 100;
  const [notifications, setNotifications] = useState<Delivery[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [totalLoaded, setTotalLoaded] = useState(0);
  const [sortPreset, setSortPreset] = useState<NotificationsSortPreset>("updated_desc");
  const [latestMeta, setLatestMeta] = useState<{ source: string; at: string } | null>(null);

  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        let offset = 0;
        let total = 0;

        while (true) {
          const chunk = await fetchDeliveries({
            status: "delivered",
            limit: countBatchSize,
            offset,
          });
          total += chunk.length;
          if (chunk.length < countBatchSize) break;
          offset += countBatchSize;
        }

        const latest = await fetchDeliveries({
          status: "delivered",
          limit: 1,
          offset: 0,
          sort: "updated_at",
          order: "desc",
        });

        if (!cancelled) {
          setTotalLoaded(total);
          const u = latest[0]?.updated_at;
          const s = latest[0]?.source;
          if (u && s) setLatestMeta({ source: s, at: u });
        }
      } catch {
        // Keep page usable even if summary stats fail.
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  const sortQuery = useMemo(() => notificationsPresetToQuery(sortPreset), [sortPreset]);

  useEffect(() => {
    let cancelled = false;
    const offset = (page - 1) * pageSize;

    (async () => {
      setLoading(true);
      setError(null);
      try {
        const initial = await fetchDeliveries({
          status: "delivered",
          limit: pageSize,
          offset,
          sort: sortQuery.sort,
          order: sortQuery.order as "asc" | "desc",
        });
        if (!cancelled) {
          setNotifications(initial);
          setHasMore(initial.length === pageSize);
        }
      } catch (e) {
        if (!cancelled) {
          setNotifications([]);
          setHasMore(false);
          setError(e instanceof Error ? e.message : "Failed to load notifications");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [page, pageSize, sortQuery.order, sortQuery.sort]);

  useEffect(() => {
    const es = new EventSource("/api/deliveries/stream");

    const onDelivered = (evt: MessageEvent) => {
      try {
        const incoming = JSON.parse(evt.data) as Delivery;
        if (incoming.status !== "delivered") return;
        setTotalLoaded((prev) => prev + 1);
        const t = incoming.updated_at;
        const s = incoming.source || "unknown-source";
        if (t) {
          setLatestMeta((prev) => {
            if (!prev || new Date(t).getTime() > new Date(prev.at).getTime()) {
              return { source: s, at: t };
            }
            return prev;
          });
        }

        if (page !== 1) return;

        setNotifications((prev) => {
          if (prev.some((n) => n.event_id === incoming.event_id)) return prev;
          const merged = [incoming, ...prev];
          const sorted = sortDeliveries(merged, sortQuery.sort, sortQuery.order);
          return sorted.slice(0, pageSize);
        });
      } catch {
        // Ignore malformed stream payloads.
      }
    };

    es.addEventListener("delivered", onDelivered as EventListener);
    es.onerror = () => {
      setError((prev) => prev ?? "Realtime connection interrupted. Reconnecting...");
    };
    es.onopen = () => {
      setError((prev) =>
        prev === "Realtime connection interrupted. Reconnecting..." ? null : prev
      );
    };

    return () => {
      es.removeEventListener("delivered", onDelivered as EventListener);
      es.close();
    };
  }, [page, pageSize, sortQuery.order, sortQuery.sort]);

  const hasItems = notifications.length > 0;
  const stats = useMemo(
    () => ({
      total: totalLoaded,
      latest: latestMeta?.source ?? "—",
      lastTime: latestMeta?.at ? formatShortDateTime(latestMeta.at) : "—",
    }),
    [latestMeta, totalLoaded]
  );
  const canGoBack = page > 1;
  const canGoNext = hasMore && !loading;

  return (
    <RequireAuth>
    <div className="app-shell">
      <AppTopNav />
      <div className="app-body">
        <aside className="app-sidebar">
          <AppNav />
          <p className="app-sidebar-hint">
            New deliveries can appear on page 1 in real time. Sort applies across pages for the
            full delivered list.
          </p>
          <div className="app-sidebar-sort">
            <label className="app-label" htmlFor="notifications-sort">
              Sort list
            </label>
            <select
              id="notifications-sort"
              className="app-select"
              value={sortPreset}
              onChange={(e) => {
                setSortPreset(e.target.value as NotificationsSortPreset);
                setPage(1);
              }}
            >
              <option value="updated_desc">Time · Newest first</option>
              <option value="updated_asc">Time · Oldest first</option>
              <option value="title_asc">Title · A to Z</option>
              <option value="title_desc">Title · Z to A</option>
              <option value="source_asc">Source · A to Z</option>
              <option value="source_desc">Source · Z to A</option>
            </select>
          </div>
        </aside>

        <main className="app-main">
          <div className="app-main-header">
            <div>
              <p className="app-eyebrow">Realtime</p>
              <h1 className="app-page-title">
                Notification <em>inbox</em>
              </h1>
            </div>
            <p className="app-kicker">Delivered updates</p>
          </div>

          <div className="app-stat-grid">
            <div className="app-stat">
              <span className="app-stat-val">{stats.total}</span>
              <div className="app-stat-label">Loaded</div>
            </div>
            <div className="app-stat">
              <span className="app-stat-val">{stats.latest}</span>
              <div className="app-stat-label">Latest source</div>
            </div>
            <div className="app-stat">
              <span className="app-stat-val app-stat-val-compact">{stats.lastTime}</span>
              <div className="app-stat-label">Time of last notification</div>
            </div>
          </div>

          {loading ? <p className="app-muted">Loading notifications...</p> : null}
          {error ? <p className="app-error">{error}</p> : null}
          {!loading && !hasItems ? (
            <p className="app-muted">
              No delivered notifications yet. Publish an event to see updates in real time.
            </p>
          ) : null}

          {hasItems ? (
            <section className="app-notifications-list" aria-label="Delivered notifications">
              {notifications.map((item) => {
                const subLine = subsourceDisplayLine(
                  item.subsource_name,
                  item.subsource_identifier
                );
                return (
                <article key={item.event_id} className="app-notification-card">
                  <div className="app-notification-head">
                    <p className="app-notification-source">{item.source || "unknown-source"}</p>
                    <span className="app-notification-pill">delivered</span>
                  </div>
                  {subLine ? <p className="app-subsource-line">{subLine}</p> : null}
                  <h3 className="app-notification-title">
                    {item.url ? (
                      <a
                        href={item.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="app-post-title-link"
                      >
                        {item.title || "Untitled event"}
                      </a>
                    ) : (
                      item.title || "Untitled event"
                    )}
                  </h3>
                  {item.url ? (
                    <a
                      href={item.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="app-link-quiet"
                    >
                      Open source link
                    </a>
                  ) : null}
                  <p className="app-muted">
                    {formatRelativeTime(item.updated_at)} · {new Date(item.updated_at).toLocaleString()}
                  </p>
                </article>
                );
              })}
            </section>
          ) : null}
          <div className="app-pagination">
            <button
              type="button"
              className="app-btn app-btn-sm"
              disabled={!canGoBack || loading}
              onClick={() => setPage((prev) => Math.max(1, prev - 1))}
            >
              <span className="app-btn-label">Back</span>
            </button>
            <p className="app-pagination-indicator">Page {page}</p>
            <button
              type="button"
              className="app-btn app-btn-sm"
              disabled={!canGoNext}
              onClick={() => setPage((prev) => prev + 1)}
            >
              <span className="app-btn-label">Next</span>
            </button>
          </div>
        </main>
      </div>
    </div>
    </RequireAuth>
  );
}

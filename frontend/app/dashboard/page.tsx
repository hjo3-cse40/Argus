"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AppNav } from "@/components/AppNav";
import { AppTopNav } from "@/components/AppTopNav";
import {
  createFilter,
  deleteFilter,
  fetchFilters,
  fetchPlatforms,
  type DestinationFilter,
  type Platform,
} from "@/lib/api";
import "../app-shell.css";

function PlatformSection({
  title,
  posts,
}: {
  title: string;
  posts: { id: number; title: string }[];
}) {
  const scrollRef = useRef<HTMLDivElement | null>(null);

  const scroll = (dir: "left" | "right") => {
    if (!scrollRef.current) return;
    scrollRef.current.scrollBy({
      left: dir === "left" ? -300 : 300,
      behavior: "smooth",
    });
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
        {posts.map((post) => (
          <div key={post.id} className="app-post-card">
            {post.title}
          </div>
        ))}
      </div>
    </div>
  );
}

const demoPosts = {
  youtube: [
    { id: 1, title: "NBA Shorts - Best Dunks" },
    { id: 2, title: "Full Game Highlights Lakers vs Celtics" },
    { id: 3, title: "Top 10 Plays of the Week" },
  ],
  reddit: [
    { id: 4, title: "Reddit: Lakers discussion thread" },
    { id: 5, title: "Reddit: Best NBA plays debate" },
  ],
  twitter: [
    { id: 6, title: "LeBron tweet reaction" },
    { id: 7, title: "NBA trending hashtag" },
  ],
};

export default function Dashboard() {
  const [platforms, setPlatforms] = useState<Platform[]>([]);
  const [platformsLoading, setPlatformsLoading] = useState(true);
  const [platformsError, setPlatformsError] = useState<string | null>(null);

  const [selectedPlatformId, setSelectedPlatformId] = useState("");

  const [destinationFilters, setDestinationFilters] = useState<DestinationFilter[]>([]);
  const [filtersLoading, setFiltersLoading] = useState(false);
  const [filtersError, setFiltersError] = useState<string | null>(null);

  const [includeInput, setIncludeInput] = useState("");
  const [excludeInput, setExcludeInput] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);

  const reloadFilters = useCallback(async (platformId: string) => {
    if (!platformId) {
      setDestinationFilters([]);
      return;
    }
    setFiltersLoading(true);
    setFiltersError(null);
    try {
      const list = await fetchFilters(platformId);
      setDestinationFilters(list);
    } catch (e) {
      setDestinationFilters([]);
      setFiltersError(e instanceof Error ? e.message : "Failed to load filters");
    } finally {
      setFiltersLoading(false);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setPlatformsLoading(true);
      setPlatformsError(null);
      try {
        const list = await fetchPlatforms();
        if (cancelled) return;
        setPlatforms(list);
        setSelectedPlatformId((prev) => prev || list[0]?.id || "");
      } catch (e) {
        if (!cancelled) {
          setPlatforms([]);
          setPlatformsError(
            e instanceof Error ? e.message : "Failed to load platforms"
          );
        }
      } finally {
        if (!cancelled) setPlatformsLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!selectedPlatformId) {
      setDestinationFilters([]);
      return;
    }
    let cancelled = false;
    (async () => {
      setFiltersLoading(true);
      setFiltersError(null);
      try {
        const list = await fetchFilters(selectedPlatformId);
        if (!cancelled) setDestinationFilters(list);
      } catch (e) {
        if (!cancelled) {
          setDestinationFilters([]);
          setFiltersError(
            e instanceof Error ? e.message : "Failed to load filters"
          );
        }
      } finally {
        if (!cancelled) setFiltersLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [selectedPlatformId]);

  const includes = destinationFilters.filter((f) => f.filter_type === "keyword_include");
  const excludes = destinationFilters.filter((f) => f.filter_type === "keyword_exclude");

  const addInclude = async () => {
    const pattern = includeInput.trim();
    if (!pattern || !selectedPlatformId) return;
    setActionError(null);
    try {
      await createFilter(selectedPlatformId, "keyword_include", pattern);
      setIncludeInput("");
      await reloadFilters(selectedPlatformId);
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Could not add include filter");
    }
  };

  const addExclude = async () => {
    const pattern = excludeInput.trim();
    if (!pattern || !selectedPlatformId) return;
    setActionError(null);
    try {
      await createFilter(selectedPlatformId, "keyword_exclude", pattern);
      setExcludeInput("");
      await reloadFilters(selectedPlatformId);
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Could not add exclude filter");
    }
  };

  const removeFilterRow = async (id: string) => {
    setActionError(null);
    try {
      await deleteFilter(id);
      if (selectedPlatformId) await reloadFilters(selectedPlatformId);
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Could not delete filter");
    }
  };

  const demoTotal =
    demoPosts.youtube.length + demoPosts.reddit.length + demoPosts.twitter.length;

  const noPlatforms = !platformsLoading && platforms.length === 0 && !platformsError;

  return (
    <div className="app-shell">
      <AppTopNav />
      <div className="app-body">
        <aside className="app-sidebar">
          <AppNav />

          <div id="keyword-filters" className="app-filters-panel">
            <h2>Filters</h2>
            <p className="app-filters-help">
              Per platform. Include: need <strong>at least one</strong> match. Exclude: drop if
              any match (checked first).
            </p>

            {platformsError && <p className="app-error">{platformsError}</p>}
            {actionError && <p className="app-error">{actionError}</p>}

            {platformsLoading ? (
              <p className="app-muted">Loading platforms…</p>
            ) : noPlatforms ? (
              <p className="app-muted">
                No platforms yet. Add one on the Platforms page, then refresh.
              </p>
            ) : (
              <>
                <label className="app-label" htmlFor="filter-platform">
                  Platform
                </label>
                <select
                  id="filter-platform"
                  value={selectedPlatformId}
                  onChange={(e) => setSelectedPlatformId(e.target.value)}
                  className="app-select"
                >
                  {platforms.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name} ({p.id.slice(0, 8)}…)
                    </option>
                  ))}
                </select>

                {filtersLoading ? (
                  <p className="app-muted" style={{ marginTop: "1rem" }}>
                    Loading filters…
                  </p>
                ) : filtersError ? (
                  <p className="app-error" style={{ marginTop: "1rem" }}>
                    {filtersError}
                  </p>
                ) : (
                  <div className="app-filter-block">
                    <div className="app-filter-block">
                      <p className="app-label">Include</p>
                      <div className="app-form-row">
                        <input
                          value={includeInput}
                          onChange={(e) => setIncludeInput(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === "Enter") {
                              e.preventDefault();
                              void addInclude();
                            }
                          }}
                          disabled={!selectedPlatformId}
                          placeholder="Add keyword…"
                          className="app-input"
                        />
                        <button
                          type="button"
                          onClick={() => void addInclude()}
                          disabled={!selectedPlatformId}
                          className="app-btn app-btn-amber app-btn-icon"
                          aria-label="Add include"
                        >
                          <span className="app-btn-label">+</span>
                        </button>
                      </div>
                      {includes.length === 0 ? (
                        <p className="app-muted" style={{ marginTop: "0.35rem" }}>
                          None yet
                        </p>
                      ) : (
                        <div className="app-chip-row">
                          {includes.map((f) => (
                            <div key={f.id} className="app-chip">
                              <span title={f.pattern}>{f.pattern}</span>
                              <button
                                type="button"
                                onClick={() => void removeFilterRow(f.id)}
                                aria-label={`Remove ${f.pattern}`}
                              >
                                ×
                              </button>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>

                    <div className="app-filter-block">
                      <p className="app-label">Exclude</p>
                      <div className="app-form-row">
                        <input
                          value={excludeInput}
                          onChange={(e) => setExcludeInput(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === "Enter") {
                              e.preventDefault();
                              void addExclude();
                            }
                          }}
                          disabled={!selectedPlatformId}
                          placeholder="Add keyword…"
                          className="app-input"
                        />
                        <button
                          type="button"
                          onClick={() => void addExclude()}
                          disabled={!selectedPlatformId}
                          className="app-btn app-btn-amber app-btn-icon"
                          aria-label="Add exclude"
                        >
                          <span className="app-btn-label">+</span>
                        </button>
                      </div>
                      {excludes.length === 0 ? (
                        <p className="app-muted" style={{ marginTop: "0.35rem" }}>
                          None yet
                        </p>
                      ) : (
                        <div className="app-chip-row">
                          {excludes.map((f) => (
                            <div key={f.id} className="app-chip">
                              <span title={f.pattern}>{f.pattern}</span>
                              <button
                                type="button"
                                onClick={() => void removeFilterRow(f.id)}
                                aria-label={`Remove ${f.pattern}`}
                              >
                                ×
                              </button>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                )}
              </>
            )}
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

          <div className="app-stat-grid">
            <div className="app-stat">
              <span className="app-stat-val">{demoTotal}</span>
              <div className="app-stat-label">Demo posts</div>
            </div>
            <div className="app-stat">
              <span className="app-stat-val">
                {selectedPlatformId ? destinationFilters.length : 0}
              </span>
              <div className="app-stat-label">Active filters</div>
            </div>
            <div className="app-stat">
              <span className="app-stat-val">{demoTotal}</span>
              <div className="app-stat-label">Matches</div>
            </div>
          </div>

          <PlatformSection title="youtube" posts={demoPosts.youtube} />
          <PlatformSection title="reddit" posts={demoPosts.reddit} />
          <PlatformSection title="twitter" posts={demoPosts.twitter} />
        </main>
      </div>
    </div>
  );
}

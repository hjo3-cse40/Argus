"use client";

import { useCallback, useEffect, useState } from "react";
import {
  createFilter,
  deleteFilter,
  fetchFilters,
  fetchPlatforms,
  updatePlatform,
  type DestinationFilter,
  type FilterCombineMode,
  type Platform,
} from "@/lib/api";
import { pickDefaultPlatformId, PlatformKindButtons } from "@/components/PlatformKindButtons";

export function KeywordFiltersPanel() {
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
        setSelectedPlatformId((prev) =>
          prev && list.some((p) => p.id === prev) ? prev : pickDefaultPlatformId(list)
        );
      } catch (e) {
        if (!cancelled) {
          setPlatforms([]);
          setPlatformsError(e instanceof Error ? e.message : "Failed to load platforms");
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
          setFiltersError(e instanceof Error ? e.message : "Failed to load filters");
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

  const selectedPlatform = platforms.find((p) => p.id === selectedPlatformId);

  const savePlatformCombine = async (patch: {
    filter_include_combine?: FilterCombineMode;
    filter_exclude_combine?: FilterCombineMode;
  }) => {
    if (!selectedPlatform) return;
    setActionError(null);
    try {
      const updated = await updatePlatform(selectedPlatform.id, {
        discord_webhook: selectedPlatform.discord_webhook,
        filter_include_combine:
          patch.filter_include_combine ?? selectedPlatform.filter_include_combine ?? "any",
        filter_exclude_combine:
          patch.filter_exclude_combine ?? selectedPlatform.filter_exclude_combine ?? "any",
      });
      setPlatforms((prev) => prev.map((x) => (x.id === updated.id ? updated : x)));
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Could not update filter combine mode");
    }
  };

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

  const noPlatforms = !platformsLoading && platforms.length === 0 && !platformsError;

  return (
    <div className="app-filters-panel">
      <p className="app-filters-help">
        Per platform. Choose whether multiple keywords use <strong>any</strong> (OR) or{" "}
        <strong>all</strong> (AND). Excludes are checked first.
      </p>

      {platformsError && <p className="app-error">{platformsError}</p>}
      {actionError && <p className="app-error">{actionError}</p>}

      {platformsLoading ? (
        <p className="app-muted">Loading platforms…</p>
      ) : noPlatforms ? (
        <p className="app-muted">No platforms yet. Add one on the Platforms page, then refresh.</p>
      ) : (
        <>
          <p className="app-label" style={{ marginBottom: "0.25rem" }}>
            Platform
          </p>
          <PlatformKindButtons
            platforms={platforms}
            selectedPlatformId={selectedPlatformId}
            onSelect={setSelectedPlatformId}
            ariaLabel="Filter platform"
          />

          {selectedPlatform ? (
            <div className="app-filter-block" style={{ marginTop: "0.75rem" }}>
              <label className="app-label" htmlFor="filter-include-combine">
                Include keywords match
              </label>
              <select
                id="filter-include-combine"
                value={selectedPlatform.filter_include_combine ?? "any"}
                onChange={(e) =>
                  void savePlatformCombine({
                    filter_include_combine: e.target.value as FilterCombineMode,
                  })
                }
                className="app-select"
              >
                <option value="any">Any keyword (OR)</option>
                <option value="all">All keywords (AND)</option>
              </select>
              <label
                className="app-label"
                htmlFor="filter-exclude-combine"
                style={{ marginTop: "0.65rem" }}
              >
                Exclude keywords match
              </label>
              <select
                id="filter-exclude-combine"
                value={selectedPlatform.filter_exclude_combine ?? "any"}
                onChange={(e) =>
                  void savePlatformCombine({
                    filter_exclude_combine: e.target.value as FilterCombineMode,
                  })
                }
                className="app-select"
              >
                <option value="any">Any keyword blocks (OR)</option>
                <option value="all">All keywords must match to block (AND)</option>
              </select>
            </div>
          ) : null}

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
  );
}

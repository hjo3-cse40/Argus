"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { AppNav } from "@/components/AppNav";
import {
    createFilter,
    deleteFilter,
    fetchFilters,
    fetchPlatforms,
    type DestinationFilter,
    type Platform,
} from "@/lib/api";

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
        <div className="mb-10">
            <div className="flex justify-between items-center mb-3">
                <h3 className="text-lg font-semibold text-purple-400 capitalize">
                    {title}
                </h3>

                <div className="flex gap-2">
                    <button
                        type="button"
                        onClick={() => scroll("left")}
                        className="px-3 py-1 bg-purple-700/40 rounded hover:bg-purple-600"
                    >
                        ←
                    </button>
                    <button
                        type="button"
                        onClick={() => scroll("right")}
                        className="px-3 py-1 bg-purple-700/40 rounded hover:bg-purple-600"
                    >
                        →
                    </button>
                </div>
            </div>

            <div
                ref={scrollRef}
                className="flex gap-4 overflow-x-auto pb-2"
            >
                {posts.map((post) => (
                    <div
                        key={post.id}
                        className="min-w-[240px] bg-[#1e293b] p-4 rounded-lg 
                       hover:scale-[1.03] transition shadow"
                    >
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
        <div className="flex h-screen bg-gradient-to-br from-[#0f172a] to-[#1e1b4b] text-white">

            <div className="w-64 shrink-0 bg-gradient-to-b from-purple-700 to-purple-900 p-5 flex flex-col overflow-y-auto">
                <h1 className="text-2xl font-bold mb-6">Argus</h1>

                <AppNav />

                <div
                    id="keyword-filters"
                    className="mt-8 w-full max-w-[220px] min-w-0 border-t border-purple-500/30 pt-6 scroll-mt-4"
                >
                    <h2 className="text-sm text-purple-200 mb-2">Filters</h2>
                    <p className="text-xs text-purple-300 mb-3 leading-snug">
                        Per platform. Include: need <strong className="text-purple-100">at least one</strong> match.
                        Exclude: drop if any match (checked first).
                    </p>

                    {platformsError && (
                        <p className="text-xs text-red-300 mb-2">{platformsError}</p>
                    )}
                    {actionError && (
                        <p className="text-xs text-red-300 mb-2">{actionError}</p>
                    )}

                    {platformsLoading ? (
                        <p className="text-xs text-purple-200">Loading platforms…</p>
                    ) : noPlatforms ? (
                        <p className="text-xs text-purple-200">
                            No platforms yet. Add one on the Platforms page, then refresh.
                        </p>
                    ) : (
                        <>
                            <label className="block text-xs text-purple-200 mb-1">Platform</label>
                            <select
                                value={selectedPlatformId}
                                onChange={(e) => setSelectedPlatformId(e.target.value)}
                                className="w-full bg-purple-900/40 text-white text-sm rounded-md px-3 py-2 outline-none focus:ring-2 focus:ring-purple-400"
                            >
                                {platforms.map((p) => (
                                    <option key={p.id} value={p.id}>
                                        {p.name} ({p.id.slice(0, 8)}…)
                                    </option>
                                ))}
                            </select>

                            {filtersLoading ? (
                                <p className="text-xs text-purple-200 mt-3">Loading filters…</p>
                            ) : filtersError ? (
                                <p className="text-xs text-red-300 mt-3">{filtersError}</p>
                            ) : (
                                <div className="mt-4 space-y-4">
                                    <div>
                                        <p className="text-xs text-purple-200 mb-1">Include</p>
                                        <div className="flex gap-2">
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
                                                className="min-w-0 flex-1 bg-purple-900/40 text-white placeholder-purple-300 
                                                    px-3 py-2 rounded-md text-sm outline-none 
                                                    focus:ring-2 focus:ring-purple-400 disabled:opacity-50"
                                            />
                                            <button
                                                type="button"
                                                onClick={() => void addInclude()}
                                                disabled={!selectedPlatformId}
                                                className="shrink-0 bg-yellow-400 text-black font-bold px-3 py-2 rounded-md text-sm hover:bg-yellow-300 disabled:opacity-50"
                                            >
                                                +
                                            </button>
                                        </div>
                                        {includes.length === 0 ? (
                                            <p className="text-xs text-purple-300/70 mt-1">None yet</p>
                                        ) : (
                                            <div className="mt-3 flex flex-wrap gap-2 max-h-28 overflow-y-auto">
                                                {includes.map((f) => (
                                                    <div
                                                        key={f.id}
                                                        className="flex items-center gap-2 bg-purple-400/20 text-purple-200 px-3 py-1 rounded-full text-xs max-w-full"
                                                    >
                                                        <span className="truncate max-w-[120px]" title={f.pattern}>
                                                            {f.pattern}
                                                        </span>
                                                        <button
                                                            type="button"
                                                            onClick={() => void removeFilterRow(f.id)}
                                                            className="hover:text-red-400 shrink-0"
                                                            aria-label={`Remove ${f.pattern}`}
                                                        >
                                                            ×
                                                        </button>
                                                    </div>
                                                ))}
                                            </div>
                                        )}
                                    </div>

                                    <div>
                                        <p className="text-xs text-purple-200 mb-1">Exclude</p>
                                        <div className="flex gap-2">
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
                                                className="min-w-0 flex-1 bg-purple-900/40 text-white placeholder-purple-300 
                                                    px-3 py-2 rounded-md text-sm outline-none 
                                                    focus:ring-2 focus:ring-purple-400 disabled:opacity-50"
                                            />
                                            <button
                                                type="button"
                                                onClick={() => void addExclude()}
                                                disabled={!selectedPlatformId}
                                                className="shrink-0 bg-yellow-400 text-black font-bold px-3 py-2 rounded-md text-sm hover:bg-yellow-300 disabled:opacity-50"
                                            >
                                                +
                                            </button>
                                        </div>
                                        {excludes.length === 0 ? (
                                            <p className="text-xs text-purple-300/70 mt-1">None yet</p>
                                        ) : (
                                            <div className="mt-3 flex flex-wrap gap-2 max-h-28 overflow-y-auto">
                                                {excludes.map((f) => (
                                                    <div
                                                        key={f.id}
                                                        className="flex items-center gap-2 bg-purple-400/20 text-purple-200 px-3 py-1 rounded-full text-xs max-w-full"
                                                    >
                                                        <span className="truncate max-w-[120px]" title={f.pattern}>
                                                            {f.pattern}
                                                        </span>
                                                        <button
                                                            type="button"
                                                            onClick={() => void removeFilterRow(f.id)}
                                                            className="hover:text-red-400 shrink-0"
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
            </div>

            <div className="flex-1 p-8 overflow-y-auto min-w-0">
                <div className="flex justify-between mb-6">
                    <h1 className="text-3xl font-bold text-purple-300">My Dashboard</h1>
                    <span className="text-purple-300 text-sm">Welcome back</span>
                </div>

                <div className="grid grid-cols-3 gap-4 mb-8">
                    <div className="bg-white/10 p-4 rounded-lg">
                        <p className="text-sm text-gray-300">Total Posts</p>
                        <p className="text-xl font-bold">{demoTotal}</p>
                    </div>

                    <div className="bg-white/10 p-4 rounded-lg">
                        <p className="text-sm text-gray-300">Active Filters</p>
                        <p className="text-xl font-bold text-purple-400">
                            {selectedPlatformId ? destinationFilters.length : 0}
                        </p>
                    </div>

                    <div className="bg-white/10 p-4 rounded-lg">
                        <p className="text-sm text-gray-300">Matches</p>
                        <p className="text-xl font-bold">{demoTotal}</p>
                    </div>
                </div>

                <PlatformSection title="youtube" posts={demoPosts.youtube} />
                <PlatformSection title="reddit" posts={demoPosts.reddit} />
                <PlatformSection title="twitter" posts={demoPosts.twitter} />
            </div>
        </div>
    );
}

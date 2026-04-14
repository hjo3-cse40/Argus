"use client";

import { useState, useRef } from "react";

/* ===========================
   PLATFORM SECTION COMPONENT
=========================== */
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
            {/* Header */}
            <div className="flex justify-between items-center mb-3">
                <h3 className="text-lg font-semibold text-purple-400 capitalize">
                    {title}
                </h3>

                <div className="flex gap-2">
                    <button
                        onClick={() => scroll("left")}
                        className="px-3 py-1 bg-purple-700/40 rounded hover:bg-purple-600"
                    >
                        ←
                    </button>
                    <button
                        onClick={() => scroll("right")}
                        className="px-3 py-1 bg-purple-700/40 rounded hover:bg-purple-600"
                    >
                        →
                    </button>
                </div>
            </div>

            {/* Posts */}
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

/* MAIN DASHBOARD */
export default function Dashboard() {
    const [filters, setFilters] = useState<string[]>([]);
    const [input, setInput] = useState("");

    const [confirmClear, setConfirmClear] = useState(false);

    /* ===========================
       MOCK DATA
    =========================== */
    const data = {
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

    /* ===========================
       FILTER LOGIC
    =========================== */
    const applyFilters = (posts: { id: number; title: string }[]) => {
        if (filters.length === 0) return posts;

        return posts.filter((p) =>
            filters.every((f) =>
                p.title.toLowerCase().split(" ").includes(f.toLowerCase())
            )
        );
    };

    /* ===========================
       HANDLERS
    =========================== */
    const addFilter = () => {
        if (!input.trim()) return;
        setFilters([...filters, input.trim()]);
        setInput("");
    };

    const removeFilter = (index: number) => {
        setFilters(filters.filter((_, i) => i !== index));
    };

    const handleClearAll = () => {
        if (!confirmClear) {
            setConfirmClear(true);

            // auto-reset after 3 seconds 
            setTimeout(() => setConfirmClear(false), 3000);
            return;
        }

        // actually clear
        setFilters([]);
        setConfirmClear(false);
    };

    /* ===========================
       DERIVED DATA
    =========================== */
    const filteredData = {
        youtube: applyFilters(data.youtube),
        reddit: applyFilters(data.reddit),
        twitter: applyFilters(data.twitter),
    };

    const totalPosts =
        data.youtube.length + data.reddit.length + data.twitter.length;

    const totalMatches =
        filteredData.youtube.length +
        filteredData.reddit.length +
        filteredData.twitter.length;

    /* ===========================
       UI
    =========================== */
    return (
        <div className="flex h-screen bg-gradient-to-br from-[#0f172a] to-[#1e1b4b] text-white">

            {/* ================= SIDEBAR ================= */}
            <div className="w-64 bg-gradient-to-b from-purple-700 to-purple-900 p-5 flex flex-col">
                <h1 className="text-2xl font-bold mb-6">Argus</h1>

                {/* Nav */}
                <nav className="space-y-3 text-sm">
                    <div className="bg-purple-500/40 p-2 rounded">Dashboard</div>
                    <div className="text-purple-200">Platforms</div>
                    <div className="text-purple-200">Filters</div>
                    <div className="text-purple-200">Settings</div>
                </nav>

                {/* Filters UI */}
                <div className="mt-8 w-full max-w-[200px]">
                    <h3 className="text-sm text-purple-200 mb-2">Filters</h3>

                    {/* Input */}
                    <div className="flex gap-2">
                        <input
                            value={input}
                            onChange={(e) => setInput(e.target.value)}
                            onKeyDown={(e) => e.key === "Enter" && addFilter()}
                            placeholder="Add keyword..."
                            className="w-full bg-purple-900/40 text-white placeholder-purple-300 
                                        px-3 py-2 rounded-md text-sm outline-none 
                                        focus:ring-2 focus:ring-purple-400"
                        />

                        <button
                            onClick={addFilter}
                            className="bg-yellow-400 text-black font-bold px-3 py-2 rounded-md text-sm"
                        >
                            +
                        </button>
                    </div>

                    {/* CLEAR BUTTON  */}
                    {filters.length > 0 && (
                        <button
                            onClick={handleClearAll}
                            className={`w-full mt-2 text-sm font-medium py-2 rounded-md transition ${confirmClear
                                ? "bg-red-600 hover:bg-red-700 text-white"
                                : "bg-pink-600/80 hover:bg-pink-600 text-white"
                                }`}
                        >
                            {confirmClear ? "Clear" : "Clear All Filters"}
                        </button>
                    )}

                    {/* Filter chips */}
                    <div className="mt-3 flex flex-wrap gap-2 max-h-[120px] overflow-y-auto">
                        {filters.map((f, i) => (
                            <div
                                key={i}
                                className="flex items-center gap-2 
                                            bg-purple-400/20 text-purple-200 
                                            px-3 py-1 rounded-full text-xs"
                            >
                                <span className="truncate max-w-[120px]" title={f}>
                                    {f}
                                </span>

                                <button
                                    onClick={() => removeFilter(i)}
                                    className="hover:text-red-400"
                                >
                                    ✕
                                </button>
                            </div>
                        ))}
                    </div>
                </div>
            </div>

            {/* ================= MAIN ================= */}
            <div className="flex-1 p-8 overflow-y-auto">

                {/* Header */}
                <div className="flex justify-between mb-6">
                    <h1 className="text-3xl font-bold text-purple-300">
                        My Dashboard
                    </h1>
                    <span className="text-purple-300 text-sm">
                        Welcome back
                    </span>
                </div>

                {/* Stats */}
                <div className="grid grid-cols-3 gap-4 mb-8">
                    <div className="bg-white/10 p-4 rounded-lg">
                        <p className="text-sm text-gray-300">Total Posts</p>
                        <p className="text-xl font-bold">{totalPosts}</p>
                    </div>

                    <div className="bg-white/10 p-4 rounded-lg">
                        <p className="text-sm text-gray-300">Active Filters</p>
                        <p className="text-xl font-bold text-purple-400">
                            {filters.length}
                        </p>
                    </div>

                    <div className="bg-white/10 p-4 rounded-lg">
                        <p className="text-sm text-gray-300">Matches</p>
                        <p className="text-xl font-bold">{totalMatches}</p>
                    </div>
                </div>

                {/* Feed Sections */}
                <PlatformSection
                    title="youtube"
                    posts={filteredData.youtube}
                />
                <PlatformSection
                    title="reddit"
                    posts={filteredData.reddit}
                />
                <PlatformSection
                    title="twitter"
                    posts={filteredData.twitter}
                />
            </div>
        </div>
    );
}
"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { AppNav } from "@/components/AppNav";
import { deriveSubsourceIdentifierFromUrl } from "@/lib/deriveSubsourceFromUrl";
import {
    createPlatform,
    createSubsource,
    deleteSubsource,
    fetchPlatforms,
    fetchSubsources,
    updateSubsource,
    type CreatePlatformPayload,
    type Platform,
    type Subsource,
} from "@/lib/api";

const PLATFORM_NAMES = [
    { value: "youtube", label: "YouTube" },
    { value: "reddit", label: "Reddit" },
    { value: "x", label: "X (Twitter)" },
] as const;

export default function PlatformsPage() {
    const [platforms, setPlatforms] = useState<Platform[]>([]);
    const [loading, setLoading] = useState(true);
    const [listError, setListError] = useState<string | null>(null);

    const [name, setName] = useState<string>("youtube");
    const [webhook, setWebhook] = useState("");
    const [secret, setSecret] = useState("");
    const [submitting, setSubmitting] = useState(false);
    const [formError, setFormError] = useState<string | null>(null);

    const [subsPlatformId, setSubsPlatformId] = useState("");
    const [subsources, setSubsources] = useState<Subsource[]>([]);
    const [subsLoading, setSubsLoading] = useState(false);
    const [subsError, setSubsError] = useState<string | null>(null);
    const [subActionError, setSubActionError] = useState<string | null>(null);

    const [subName, setSubName] = useState("");
    const [subUrl, setSubUrl] = useState("");
    const [subSubmitting, setSubSubmitting] = useState(false);

    const [editingId, setEditingId] = useState<string | null>(null);
    const [editName, setEditName] = useState("");
    const [editUrl, setEditUrl] = useState("");
    const [editSaving, setEditSaving] = useState(false);

    const load = useCallback(async () => {
        setListError(null);
        setLoading(true);
        try {
            const list = await fetchPlatforms();
            setPlatforms(list);
        } catch (e) {
            setPlatforms([]);
            setListError(e instanceof Error ? e.message : "Failed to load platforms");
        } finally {
            setLoading(false);
        }
    }, []);

    const loadSubsources = useCallback(async (platformId: string) => {
        if (!platformId) {
            setSubsources([]);
            return;
        }
        setSubsLoading(true);
        setSubsError(null);
        try {
            const list = await fetchSubsources(platformId);
            setSubsources(list);
        } catch (e) {
            setSubsources([]);
            setSubsError(e instanceof Error ? e.message : "Failed to load sub-channels");
        } finally {
            setSubsLoading(false);
        }
    }, []);

    useEffect(() => {
        void load();
    }, [load]);

    useEffect(() => {
        if (platforms.length === 0) {
            setSubsPlatformId("");
            setSubsources([]);
            return;
        }
        setSubsPlatformId((prev) =>
            prev && platforms.some((p) => p.id === prev) ? prev : platforms[0].id
        );
    }, [platforms]);

    useEffect(() => {
        if (!subsPlatformId) {
            setSubsources([]);
            return;
        }
        void loadSubsources(subsPlatformId);
    }, [subsPlatformId, loadSubsources]);

    const selectedPlatform = platforms.find((p) => p.id === subsPlatformId);

    const addUrlDerived = useMemo(() => {
        if (!selectedPlatform?.name || !subUrl.trim()) return null;
        return deriveSubsourceIdentifierFromUrl(selectedPlatform.name, subUrl);
    }, [selectedPlatform?.name, subUrl]);

    const editUrlDerived = useMemo(() => {
        if (!editingId || !selectedPlatform?.name || !editUrl.trim()) return null;
        return deriveSubsourceIdentifierFromUrl(selectedPlatform.name, editUrl);
    }, [editingId, selectedPlatform?.name, editUrl]);

    const onSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setFormError(null);
        const payload: CreatePlatformPayload = {
            name,
            discord_webhook: webhook.trim(),
        };
        const s = secret.trim();
        if (s) payload.webhook_secret = s;

        setSubmitting(true);
        try {
            await createPlatform(payload);
            setWebhook("");
            setSecret("");
            await load();
        } catch (err) {
            setFormError(err instanceof Error ? err.message : "Could not create platform");
        } finally {
            setSubmitting(false);
        }
    };

    const onAddSubsource = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!subsPlatformId || !selectedPlatform) return;
        setSubActionError(null);
        const trimmedUrl = subUrl.trim();
        if (!trimmedUrl) {
            setSubActionError("Paste the channel, subreddit, or profile URL.");
            return;
        }
        const derived = deriveSubsourceIdentifierFromUrl(selectedPlatform.name, trimmedUrl);
        if (!derived.ok) {
            setSubActionError(derived.error);
            return;
        }
        setSubSubmitting(true);
        try {
            await createSubsource(subsPlatformId, {
                name: subName.trim() || derived.identifier,
                identifier: derived.identifier,
                url: trimmedUrl,
            });
            setSubName("");
            setSubUrl("");
            await loadSubsources(subsPlatformId);
        } catch (err) {
            setSubActionError(
                err instanceof Error ? err.message : "Could not create sub-channel"
            );
        } finally {
            setSubSubmitting(false);
        }
    };

    const startEdit = (s: Subsource) => {
        setEditingId(s.id);
        setEditName(s.name);
        setEditUrl(s.url ?? "");
        setSubActionError(null);
    };

    const cancelEdit = () => {
        setEditingId(null);
        setEditName("");
        setEditUrl("");
    };

    const saveEdit = async () => {
        if (!editingId || !selectedPlatform) return;
        setSubActionError(null);
        const trimmedUrl = editUrl.trim();
        if (!trimmedUrl) {
            setSubActionError("URL is required so the feed id can be detected.");
            return;
        }
        const derived = deriveSubsourceIdentifierFromUrl(selectedPlatform.name, trimmedUrl);
        if (!derived.ok) {
            setSubActionError(derived.error);
            return;
        }
        setEditSaving(true);
        try {
            await updateSubsource(editingId, {
                name: editName.trim() || derived.identifier,
                identifier: derived.identifier,
                url: trimmedUrl,
            });
            cancelEdit();
            if (subsPlatformId) await loadSubsources(subsPlatformId);
        } catch (err) {
            setSubActionError(
                err instanceof Error ? err.message : "Could not update sub-channel"
            );
        } finally {
            setEditSaving(false);
        }
    };

    const removeSubsource = async (id: string) => {
        setSubActionError(null);
        try {
            await deleteSubsource(id);
            if (subsPlatformId) await loadSubsources(subsPlatformId);
            if (editingId === id) cancelEdit();
        } catch (err) {
            setSubActionError(
                err instanceof Error ? err.message : "Could not delete sub-channel"
            );
        }
    };

    return (
        <div className="flex min-h-screen bg-gradient-to-br from-[#0f172a] to-[#1e1b4b] text-white">
            <div className="w-64 shrink-0 bg-gradient-to-b from-purple-700 to-purple-900 p-5 flex flex-col">
                <h1 className="text-2xl font-bold mb-6">Argus</h1>
                <AppNav />
                <p className="mt-8 text-xs text-purple-300 leading-snug">
                    Add Discord webhooks and sub-channels per platform, then set filters on the
                    Dashboard.
                </p>
            </div>

            <main className="flex-1 p-8 overflow-y-auto min-w-0">
                <h1 className="text-3xl font-bold text-purple-300 mb-2">Platforms</h1>
                <p className="text-sm text-purple-200 mb-8">
                    Names must be{" "}
                    <code className="text-purple-100">youtube</code>,{" "}
                    <code className="text-purple-100">reddit</code>, or{" "}
                    <code className="text-purple-100">x</code>.
                </p>

                <section className="mb-10 max-w-xl">
                    <h2 className="text-lg font-semibold text-purple-400 mb-3">Add platform</h2>
                    <form onSubmit={(e) => void onSubmit(e)} className="space-y-3">
                        <div>
                            <label className="block text-xs text-purple-200 mb-1">Name</label>
                            <select
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                className="w-full bg-purple-900/40 text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-purple-400"
                            >
                                {PLATFORM_NAMES.map((o) => (
                                    <option key={o.value} value={o.value}>
                                        {o.label}
                                    </option>
                                ))}
                            </select>
                        </div>
                        <div>
                            <label className="block text-xs text-purple-200 mb-1">
                                Discord webhook URL
                            </label>
                            <input
                                type="url"
                                required
                                value={webhook}
                                onChange={(e) => setWebhook(e.target.value)}
                                placeholder="https://discord.com/api/webhooks/..."
                                className="w-full bg-purple-900/40 text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-purple-400 placeholder:text-purple-300"
                            />
                        </div>
                        <div>
                            <label className="block text-xs text-purple-200 mb-1">
                                Webhook secret (optional)
                            </label>
                            <input
                                type="password"
                                value={secret}
                                onChange={(e) => setSecret(e.target.value)}
                                className="w-full bg-purple-900/40 text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-purple-400"
                            />
                        </div>
                        {formError && (
                            <p className="text-sm text-red-300">{formError}</p>
                        )}
                        <button
                            type="submit"
                            disabled={submitting}
                            className="bg-yellow-400 text-black font-bold px-4 py-2 rounded-md text-sm hover:bg-yellow-300 disabled:opacity-50"
                        >
                            {submitting ? "Creating…" : "Create platform"}
                        </button>
                    </form>
                </section>

                <section className="mb-10">
                    <h2 className="text-lg font-semibold text-purple-400 mb-3">Your platforms</h2>
                    {loading ? (
                        <p className="text-sm text-purple-200">Loading…</p>
                    ) : listError ? (
                        <p className="text-sm text-red-300">{listError}</p>
                    ) : platforms.length === 0 ? (
                        <p className="text-sm text-purple-200">No platforms yet.</p>
                    ) : (
                        <ul className="space-y-2 max-w-2xl">
                            {platforms.map((p) => (
                                <li
                                    key={p.id}
                                    className="bg-[#1e293b] rounded-lg p-4 text-sm shadow"
                                >
                                    <p className="font-semibold text-purple-400 capitalize">
                                        {p.name}
                                    </p>
                                    <p className="text-purple-200/80 text-xs mt-1 font-mono break-all">
                                        id: {p.id}
                                    </p>
                                    <p className="text-purple-200/80 text-xs mt-1 font-mono break-all">
                                        webhook: {p.discord_webhook}
                                    </p>
                                </li>
                            ))}
                        </ul>
                    )}
                </section>

                <section className="max-w-3xl border-t border-purple-500/30 pt-8">
                    <h2 className="text-lg font-semibold text-purple-400 mb-2">Sub-channels</h2>
                    <p className="text-xs text-purple-300 mb-4 leading-snug">
                        Paste the public URL for the channel, subreddit, or profile. We detect the
                        feed id automatically (same rules as the RSS worker: YouTube channel / @handle,
                        <code className="text-purple-100"> /r/name</code>, X/Twitter profile).
                    </p>

                    {platforms.length === 0 ? (
                        <p className="text-sm text-purple-200">Add a platform first.</p>
                    ) : (
                        <>
                            <label className="block text-xs text-purple-200 mb-1">Platform</label>
                            <select
                                value={subsPlatformId}
                                onChange={(e) => setSubsPlatformId(e.target.value)}
                                className="w-full max-w-md bg-purple-900/40 text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-purple-400 mb-4"
                            >
                                {platforms.map((p) => (
                                    <option key={p.id} value={p.id}>
                                        {p.name} ({p.id.slice(0, 8)}…)
                                    </option>
                                ))}
                            </select>

                            {subActionError && (
                                <p className="text-sm text-red-300 mb-3">{subActionError}</p>
                            )}

                            <h3 className="text-sm text-purple-200 mb-2">
                                Add sub-channel
                                {selectedPlatform ? (
                                    <span className="text-purple-300 font-normal">
                                        {" "}
                                        for <span className="capitalize">{selectedPlatform.name}</span>
                                    </span>
                                ) : null}
                            </h3>
                            <form
                                onSubmit={(e) => void onAddSubsource(e)}
                                className="grid gap-3 sm:grid-cols-2 mb-6"
                            >
                                <div className="sm:col-span-1">
                                    <label className="block text-xs text-purple-200 mb-1">
                                        Display name (optional)
                                    </label>
                                    <input
                                        value={subName}
                                        onChange={(e) => setSubName(e.target.value)}
                                        placeholder="Defaults from URL if empty"
                                        className="w-full bg-purple-900/40 text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-purple-400 placeholder:text-purple-300"
                                    />
                                </div>
                                <div className="sm:col-span-2">
                                    <label className="block text-xs text-purple-200 mb-1">
                                        Page URL
                                    </label>
                                    <input
                                        type="url"
                                        value={subUrl}
                                        onChange={(e) => setSubUrl(e.target.value)}
                                        placeholder="https://www.youtube.com/channel/… or /r/… or x.com/…"
                                        required
                                        className="w-full bg-purple-900/40 text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-purple-400 placeholder:text-purple-300"
                                    />
                                    {addUrlDerived ? (
                                        addUrlDerived.ok ? (
                                            <p className="text-xs text-purple-300 mt-1">
                                                Detected feed id:{" "}
                                                <code className="text-purple-100">{addUrlDerived.identifier}</code>
                                            </p>
                                        ) : (
                                            <p className="text-xs text-amber-200 mt-1">{addUrlDerived.error}</p>
                                        )
                                    ) : null}
                                </div>
                                <div className="sm:col-span-2">
                                    <button
                                        type="submit"
                                        disabled={subSubmitting || !subsPlatformId}
                                        className="bg-yellow-400 text-black font-bold px-4 py-2 rounded-md text-sm hover:bg-yellow-300 disabled:opacity-50"
                                    >
                                        {subSubmitting ? "Adding…" : "Add sub-channel"}
                                    </button>
                                </div>
                            </form>

                            {subsLoading ? (
                                <p className="text-sm text-purple-200">Loading sub-channels…</p>
                            ) : subsError ? (
                                <p className="text-sm text-red-300">{subsError}</p>
                            ) : subsources.length === 0 ? (
                                <p className="text-sm text-purple-200">No sub-channels yet.</p>
                            ) : (
                                <ul className="space-y-3">
                                    {subsources.map((s) => (
                                        <li
                                            key={s.id}
                                            className="bg-[#1e293b] rounded-lg p-4 text-sm shadow"
                                        >
                                            {editingId === s.id ? (
                                                <div className="space-y-2">
                                                    <label className="block text-xs text-purple-200">
                                                        Display name
                                                    </label>
                                                    <input
                                                        value={editName}
                                                        onChange={(e) => setEditName(e.target.value)}
                                                        className="w-full bg-purple-900/40 text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-purple-400"
                                                    />
                                                    <label className="block text-xs text-purple-200">
                                                        Page URL
                                                    </label>
                                                    <input
                                                        type="url"
                                                        value={editUrl}
                                                        onChange={(e) => setEditUrl(e.target.value)}
                                                        placeholder="https://…"
                                                        className="w-full bg-purple-900/40 text-white rounded-md px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-purple-400 placeholder:text-purple-300"
                                                    />
                                                    {editUrlDerived ? (
                                                        editUrlDerived.ok ? (
                                                            <p className="text-xs text-purple-300">
                                                                Will save feed id:{" "}
                                                                <code className="text-purple-100">
                                                                    {editUrlDerived.identifier}
                                                                </code>
                                                            </p>
                                                        ) : (
                                                            <p className="text-xs text-amber-200">{editUrlDerived.error}</p>
                                                        )
                                                    ) : null}
                                                    <div className="flex gap-2 pt-1">
                                                        <button
                                                            type="button"
                                                            onClick={() => void saveEdit()}
                                                            disabled={editSaving}
                                                            className="bg-yellow-400 text-black font-bold px-3 py-1.5 rounded-md text-xs hover:bg-yellow-300 disabled:opacity-50"
                                                        >
                                                            {editSaving ? "Saving…" : "Save"}
                                                        </button>
                                                        <button
                                                            type="button"
                                                            onClick={cancelEdit}
                                                            disabled={editSaving}
                                                            className="bg-purple-700/50 text-white px-3 py-1.5 rounded-md text-xs hover:bg-purple-600 disabled:opacity-50"
                                                        >
                                                            Cancel
                                                        </button>
                                                    </div>
                                                </div>
                                            ) : (
                                                <>
                                                    <div className="flex flex-wrap items-start justify-between gap-2">
                                                        <div>
                                                            <p className="font-semibold text-purple-400">
                                                                {s.name}
                                                            </p>
                                                            <p className="text-purple-200/80 text-xs mt-1 font-mono">
                                                                id: {s.id}
                                                            </p>
                                                            <p className="text-purple-200/80 text-xs mt-0.5">
                                                                identifier:{" "}
                                                                <span className="font-mono">
                                                                    {s.identifier}
                                                                </span>
                                                            </p>
                                                            {s.url ? (
                                                                <p className="text-purple-200/80 text-xs mt-0.5 font-mono break-all">
                                                                    url: {s.url}
                                                                </p>
                                                            ) : null}
                                                        </div>
                                                        <div className="flex gap-2 shrink-0">
                                                            <button
                                                                type="button"
                                                                onClick={() => startEdit(s)}
                                                                className="text-xs text-purple-200 hover:text-white underline"
                                                            >
                                                                Edit
                                                            </button>
                                                            <button
                                                                type="button"
                                                                onClick={() => void removeSubsource(s.id)}
                                                                className="text-xs text-red-300 hover:text-red-200"
                                                            >
                                                                Delete
                                                            </button>
                                                        </div>
                                                    </div>
                                                </>
                                            )}
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </>
                    )}
                </section>
            </main>
        </div>
    );
}

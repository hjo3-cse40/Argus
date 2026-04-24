"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { AppNav } from "@/components/AppNav";
import { AppTopNav } from "@/components/AppTopNav";
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
import "../app-shell.css";

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
    <div className="app-shell">
      <AppTopNav />
      <div className="app-body">
        <aside className="app-sidebar">
          <AppNav />
          <p className="app-sidebar-hint">
            Add Discord webhooks and sub-channels per platform, then set filters on the
            Dashboard.
          </p>
        </aside>

        <main className="app-main">
          <p className="app-eyebrow">Configuration</p>
          <h1 className="app-page-title">
            <em>Platforms</em>
          </h1>
          <p className="app-page-sub">
            Names must be <code className="app-code">youtube</code>,{" "}
            <code className="app-code">reddit</code>, or <code className="app-code">x</code>.
          </p>

          <section className="app-section-block app-max-w-form">
            <h2 className="app-section-title">Add platform</h2>
            <form onSubmit={(e) => void onSubmit(e)} className="app-form-stack">
              <div>
                <label className="app-label" htmlFor="platform-name">
                  Name
                </label>
                <select
                  id="platform-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="app-select"
                >
                  {PLATFORM_NAMES.map((o) => (
                    <option key={o.value} value={o.value}>
                      {o.label}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="app-label" htmlFor="webhook-url">
                  Discord webhook URL
                </label>
                <input
                  id="webhook-url"
                  type="url"
                  required
                  value={webhook}
                  onChange={(e) => setWebhook(e.target.value)}
                  placeholder="https://discord.com/api/webhooks/..."
                  className="app-input"
                />
              </div>
              <div>
                <label className="app-label" htmlFor="webhook-secret">
                  Webhook secret (optional)
                </label>
                <input
                  id="webhook-secret"
                  type="password"
                  value={secret}
                  onChange={(e) => setSecret(e.target.value)}
                  className="app-input"
                />
              </div>
              {formError && <p className="app-error">{formError}</p>}
              <button type="submit" disabled={submitting} className="app-btn app-btn-amber">
                <span className="app-btn-label">{submitting ? "Creating…" : "Create platform"}</span>
              </button>
            </form>
          </section>

          <section className="app-section-block app-max-w-wide">
            <h2 className="app-section-title">Your platforms</h2>
            {loading ? (
              <p className="app-muted">Loading…</p>
            ) : listError ? (
              <p className="app-error">{listError}</p>
            ) : platforms.length === 0 ? (
              <p className="app-muted">No platforms yet.</p>
            ) : (
              <ul className="app-card-list">
                {platforms.map((p) => (
                  <li key={p.id} className="app-card">
                    <p style={{ fontFamily: "'DM Serif Display', serif", margin: "0 0 0.35rem" }}>
                      <span style={{ textTransform: "capitalize" }}>{p.name}</span>
                    </p>
                    <p className="app-muted" style={{ margin: 0, fontSize: "0.75rem", wordBreak: "break-all" }}>
                      id: {p.id}
                    </p>
                    <p className="app-muted" style={{ margin: "0.35rem 0 0", fontSize: "0.75rem", wordBreak: "break-all" }}>
                      webhook: {p.discord_webhook}
                    </p>
                  </li>
                ))}
              </ul>
            )}
          </section>

          <hr className="app-divider" />

          <section className="app-max-w-wide">
            <h2 className="app-section-title">Sub-channels</h2>
            <p className="app-page-sub" style={{ marginBottom: "1.25rem" }}>
              Paste the public URL for the channel, subreddit, or profile. We detect the feed id
              automatically (same rules as the RSS worker: YouTube channel / @handle,{" "}
              <code className="app-code">/r/name</code>, X/Twitter profile).
            </p>

            {platforms.length === 0 ? (
              <p className="app-muted">Add a platform first.</p>
            ) : (
              <>
                <label className="app-label" htmlFor="subs-platform">
                  Platform
                </label>
                <select
                  id="subs-platform"
                  value={subsPlatformId}
                  onChange={(e) => setSubsPlatformId(e.target.value)}
                  className="app-select"
                  style={{ maxWidth: "28rem", marginBottom: "1rem" }}
                >
                  {platforms.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name} ({p.id.slice(0, 8)}…)
                    </option>
                  ))}
                </select>

                {subActionError && <p className="app-error">{subActionError}</p>}

                <h3
                  style={{
                    fontFamily: "'DM Serif Display', serif",
                    fontSize: "1rem",
                    margin: "0 0 0.75rem",
                    fontWeight: 400,
                  }}
                >
                  Add sub-channel
                  {selectedPlatform ? (
                    <span className="app-muted" style={{ fontFamily: "inherit", fontWeight: 400 }}>
                      {" "}
                      for <span style={{ textTransform: "capitalize" }}>{selectedPlatform.name}</span>
                    </span>
                  ) : null}
                </h3>
                <form
                  onSubmit={(e) => void onAddSubsource(e)}
                  className="app-form-stack"
                  style={{ marginBottom: "1.5rem", maxWidth: "40rem" }}
                >
                  <div>
                    <label className="app-label" htmlFor="sub-name">
                      Display name (optional)
                    </label>
                    <input
                      id="sub-name"
                      value={subName}
                      onChange={(e) => setSubName(e.target.value)}
                      placeholder="Defaults from URL if empty"
                      className="app-input"
                    />
                  </div>
                  <div>
                    <label className="app-label" htmlFor="sub-url">
                      Page URL
                    </label>
                    <input
                      id="sub-url"
                      type="url"
                      value={subUrl}
                      onChange={(e) => setSubUrl(e.target.value)}
                      placeholder="https://www.youtube.com/channel/… or /r/… or x.com/…"
                      required
                      className="app-input"
                    />
                    {addUrlDerived ? (
                      addUrlDerived.ok ? (
                        <p className="app-muted" style={{ marginTop: "0.35rem" }}>
                          Detected feed id:{" "}
                          <code className="app-code">{addUrlDerived.identifier}</code>
                        </p>
                      ) : (
                        <p className="app-warn" style={{ marginTop: "0.35rem" }}>
                          {addUrlDerived.error}
                        </p>
                      )
                    ) : null}
                  </div>
                  <button
                    type="submit"
                    disabled={subSubmitting || !subsPlatformId}
                    className="app-btn app-btn-amber app-btn-sm"
                    style={{ alignSelf: "flex-start" }}
                  >
                    <span className="app-btn-label">
                      {subSubmitting ? "Adding…" : "Add sub-channel"}
                    </span>
                  </button>
                </form>

                {subsLoading ? (
                  <p className="app-muted">Loading sub-channels…</p>
                ) : subsError ? (
                  <p className="app-error">{subsError}</p>
                ) : subsources.length === 0 ? (
                  <p className="app-muted">No sub-channels yet.</p>
                ) : (
                  <ul className="app-card-list">
                    {subsources.map((s) => (
                      <li key={s.id} className="app-card">
                        {editingId === s.id ? (
                          <div className="app-form-stack">
                            <div>
                              <label className="app-label" htmlFor={`edit-name-${s.id}`}>
                                Display name
                              </label>
                              <input
                                id={`edit-name-${s.id}`}
                                value={editName}
                                onChange={(e) => setEditName(e.target.value)}
                                className="app-input"
                              />
                            </div>
                            <div>
                              <label className="app-label" htmlFor={`edit-url-${s.id}`}>
                                Page URL
                              </label>
                              <input
                                id={`edit-url-${s.id}`}
                                type="url"
                                value={editUrl}
                                onChange={(e) => setEditUrl(e.target.value)}
                                placeholder="https://…"
                                className="app-input"
                              />
                              {editUrlDerived ? (
                                editUrlDerived.ok ? (
                                  <p className="app-muted" style={{ marginTop: "0.35rem" }}>
                                    Will save feed id:{" "}
                                    <code className="app-code">{editUrlDerived.identifier}</code>
                                  </p>
                                ) : (
                                  <p className="app-warn" style={{ marginTop: "0.35rem" }}>
                                    {editUrlDerived.error}
                                  </p>
                                )
                              ) : null}
                            </div>
                            <div style={{ display: "flex", gap: "0.5rem" }}>
                              <button
                                type="button"
                                onClick={() => void saveEdit()}
                                disabled={editSaving}
                                className="app-btn app-btn-amber app-btn-sm"
                              >
                                <span className="app-btn-label">
                                  {editSaving ? "Saving…" : "Save"}
                                </span>
                              </button>
                              <button
                                type="button"
                                onClick={cancelEdit}
                                disabled={editSaving}
                                className="app-btn app-btn-sm"
                              >
                                <span className="app-btn-label">Cancel</span>
                              </button>
                            </div>
                          </div>
                        ) : (
                          <div
                            style={{
                              display: "flex",
                              flexWrap: "wrap",
                              alignItems: "flex-start",
                              justifyContent: "space-between",
                              gap: "0.75rem",
                            }}
                          >
                            <div>
                              <p
                                style={{
                                  fontFamily: "'DM Serif Display', serif",
                                  margin: "0 0 0.35rem",
                                }}
                              >
                                {s.name}
                              </p>
                              <p className="app-muted" style={{ margin: 0, fontSize: "0.75rem" }}>
                                id: {s.id}
                              </p>
                              <p className="app-muted" style={{ margin: "0.25rem 0 0", fontSize: "0.75rem" }}>
                                identifier: <span style={{ fontFamily: "monospace" }}>{s.identifier}</span>
                              </p>
                              {s.url ? (
                                <p
                                  className="app-muted"
                                  style={{
                                    margin: "0.25rem 0 0",
                                    fontSize: "0.75rem",
                                    wordBreak: "break-all",
                                  }}
                                >
                                  url: {s.url}
                                </p>
                              ) : null}
                            </div>
                            <div style={{ display: "flex", gap: "0.75rem", flexShrink: 0 }}>
                              <button
                                type="button"
                                onClick={() => startEdit(s)}
                                className="app-link-quiet"
                              >
                                Edit
                              </button>
                              <button
                                type="button"
                                onClick={() => void removeSubsource(s.id)}
                                className="app-link-danger"
                              >
                                Delete
                              </button>
                            </div>
                          </div>
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
    </div>
  );
}

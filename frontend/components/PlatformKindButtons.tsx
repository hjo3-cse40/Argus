"use client";

import type { Platform } from "@/lib/api";

const KINDS = [
  { name: "youtube" as const, label: "YouTube" },
  { name: "reddit" as const, label: "Reddit" },
  { name: "x" as const, label: "X (Twitter)" },
];

export function pickDefaultPlatformId(platforms: Platform[]): string {
  for (const k of KINDS) {
    const p = platforms.find((x) => x.name === k.name);
    if (p) return p.id;
  }
  return platforms[0]?.id ?? "";
}

/** Pick a platform by fixed YouTube / Reddit / X slots; disabled if not created yet. */
export function PlatformKindButtons({
  platforms,
  selectedPlatformId,
  onSelect,
  ariaLabel = "Platform",
}: {
  platforms: Platform[];
  selectedPlatformId: string;
  onSelect: (platformId: string) => void;
  ariaLabel?: string;
}) {
  return (
    <div className="app-platform-kind-row" role="group" aria-label={ariaLabel}>
      {KINDS.map(({ name, label }) => {
        const p = platforms.find((x) => x.name === name);
        const active = Boolean(p && p.id === selectedPlatformId);
        return (
          <button
            key={name}
            type="button"
            disabled={!p}
            title={p ? undefined : `Add ${label} on the Platforms page first`}
            aria-pressed={active}
            className={`app-platform-kind-btn${active ? " app-platform-kind-btn-active" : ""}`}
            onClick={() => {
              if (p) onSelect(p.id);
            }}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}

const CREATE_OPTIONS = [
  { value: "youtube", label: "YouTube" },
  { value: "reddit", label: "Reddit" },
  { value: "x", label: "X (Twitter)" },
] as const;

/** Toggle `name` state when adding a new platform (youtube | reddit | x). */
export function PlatformNameButtons({
  value,
  onChange,
}: {
  value: string;
  onChange: (name: string) => void;
}) {
  return (
    <div className="app-platform-kind-row" role="group" aria-label="Platform name">
      {CREATE_OPTIONS.map(({ value: v, label }) => {
        const active = value === v;
        return (
          <button
            key={v}
            type="button"
            aria-pressed={active}
            className={`app-platform-kind-btn${active ? " app-platform-kind-btn-active" : ""}`}
            onClick={() => onChange(v)}
          >
            {label}
          </button>
        );
      })}
    </div>
  );
}

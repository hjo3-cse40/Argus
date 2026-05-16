import type { Delivery } from "./api";

/** Dashboard carousel: time + title only (each row is one source). */
export type DashboardSortPreset = "updated_desc" | "updated_asc" | "title_asc" | "title_desc";

/** Notifications inbox: time, title, and source. */
export type NotificationsSortPreset =
  | "updated_desc"
  | "updated_asc"
  | "title_asc"
  | "title_desc"
  | "source_asc"
  | "source_desc";

export function notificationsPresetToQuery(preset: NotificationsSortPreset): {
  sort: string;
  order: string;
} {
  switch (preset) {
    case "updated_desc":
      return { sort: "updated_at", order: "desc" };
    case "updated_asc":
      return { sort: "updated_at", order: "asc" };
    case "title_asc":
      return { sort: "title", order: "asc" };
    case "title_desc":
      return { sort: "title", order: "desc" };
    case "source_asc":
      return { sort: "source", order: "asc" };
    case "source_desc":
      return { sort: "source", order: "desc" };
  }
}

function timeMs(iso?: string): number {
  if (!iso) return 0;
  const t = new Date(iso).getTime();
  return Number.isFinite(t) ? t : 0;
}

export function sortDashboardPosts<T extends { title: string; updatedAt?: string }>(
  posts: T[],
  preset: DashboardSortPreset
): T[] {
  const copy = [...posts];
  switch (preset) {
    case "updated_desc":
      copy.sort((a, b) => timeMs(b.updatedAt) - timeMs(a.updatedAt));
      break;
    case "updated_asc":
      copy.sort((a, b) => timeMs(a.updatedAt) - timeMs(b.updatedAt));
      break;
    case "title_asc":
      copy.sort((a, b) =>
        a.title.localeCompare(b.title, undefined, { sensitivity: "base" })
      );
      break;
    case "title_desc":
      copy.sort((a, b) =>
        b.title.localeCompare(a.title, undefined, { sensitivity: "base" })
      );
      break;
  }
  return copy;
}

/** Matches server-side sort in `deliveries.go` (for SSE merges on page 1). */
export function sortDeliveries(list: Delivery[], sort: string, order: string): Delivery[] {
  const copy = [...list];
  const field = sort.trim().toLowerCase();
  const allowed = new Set(["created_at", "updated_at", "title", "source"]);
  const f = allowed.has(field) ? field : "created_at";
  const o = order.trim().toLowerCase();
  let asc: boolean;
  if (o === "asc") {
    asc = true;
  } else if (o === "desc") {
    asc = false;
  } else {
    asc = f === "title" || f === "source";
  }

  copy.sort((a, b) => {
    switch (f) {
      case "updated_at": {
        const ta = new Date(a.updated_at).getTime();
        const tb = new Date(b.updated_at).getTime();
        const cmp = ta - tb;
        return asc ? cmp : -cmp;
      }
      case "title": {
        const cmp = a.title.localeCompare(b.title, undefined, { sensitivity: "base" });
        return asc ? cmp : -cmp;
      }
      case "source": {
        const cmp = a.source.localeCompare(b.source, undefined, { sensitivity: "base" });
        return asc ? cmp : -cmp;
      }
      default: {
        const ta = new Date(a.created_at).getTime();
        const tb = new Date(b.created_at).getTime();
        const cmp = ta - tb;
        return asc ? cmp : -cmp;
      }
    }
  });
  return copy;
}

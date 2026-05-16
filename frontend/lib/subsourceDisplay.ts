/** Human-readable sub-channel line for deliveries (from API subsource fields). */
export function subsourceDisplayLine(
  name?: string | null,
  identifier?: string | null
): string | null {
  const n = name?.trim() ?? "";
  const i = identifier?.trim() ?? "";
  if (n && i && n.toLowerCase() !== i.toLowerCase()) {
    return `${n} · ${i}`;
  }
  if (n) return n;
  if (i) return i;
  return null;
}

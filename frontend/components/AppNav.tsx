"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useSyncExternalStore } from "react";

function subscribeLocationHash(onChange: () => void) {
  window.addEventListener("hashchange", onChange);
  window.addEventListener("popstate", onChange);
  return () => {
    window.removeEventListener("hashchange", onChange);
    window.removeEventListener("popstate", onChange);
  };
}

function getLocationHashSnapshot() {
  return window.location.hash;
}

function getLocationHashServerSnapshot() {
  return "";
}

const links: { href: string; label: string; matchHash?: string }[] = [
  { href: "/notifications", label: "Notifications" },
  { href: "/dashboard", label: "Dashboard", matchHash: "" },
  { href: "/platforms", label: "Platforms" },
  { href: "/dashboard#keyword-filters", label: "Filters", matchHash: "#keyword-filters" },
];

export function AppNav() {
  const pathname = usePathname();
  const hash = useSyncExternalStore(
    subscribeLocationHash,
    getLocationHashSnapshot,
    getLocationHashServerSnapshot
  );

  return (
    <nav className="app-nav-vertical" aria-label="App sections">
      {links.map(({ href, label, matchHash }) => {
        const pathOnly = href.split("#")[0] ?? href;
        let active = pathname === pathOnly;
        if (active && matchHash !== undefined) {
          active = hash === matchHash;
        }
        return (
          <Link
            key={href}
            href={href}
            className={active ? "app-nav-link app-nav-link-active" : "app-nav-link"}
          >
            {label}
          </Link>
        );
      })}
      <div className="app-nav-muted">Settings</div>
    </nav>
  );
}

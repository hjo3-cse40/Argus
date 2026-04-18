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
    { href: "/dashboard", label: "Dashboard", matchHash: "" },
    { href: "/platforms", label: "Platforms" },
    { href: "/dashboard#keyword-filters", label: "Filters", matchHash: "#keyword-filters" },
];

function itemClass(active: boolean) {
    return [
        "block rounded p-2 text-sm transition-colors",
        active
            ? "bg-purple-500/40 text-white"
            : "text-purple-200 hover:bg-purple-500/20 hover:text-white",
    ].join(" ");
}

export function AppNav() {
    const pathname = usePathname();
    const hash = useSyncExternalStore(
        subscribeLocationHash,
        getLocationHashSnapshot,
        getLocationHashServerSnapshot
    );

    return (
        <nav className="space-y-3 text-sm">
            {links.map(({ href, label, matchHash }) => {
                const pathOnly = href.split("#")[0] ?? href;
                let active = pathname === pathOnly;
                if (active && matchHash !== undefined) {
                    active = hash === matchHash;
                }
                return (
                    <Link key={href} href={href} className={itemClass(active)}>
                        {label}
                    </Link>
                );
            })}
            <div className="text-purple-200 p-2 text-sm">Settings</div>
        </nav>
    );
}

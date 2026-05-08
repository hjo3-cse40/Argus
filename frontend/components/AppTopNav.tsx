"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { fetchCurrentUser, logout, type CurrentUser } from "@/lib/api";

export function AppTopNav() {
  const [user, setUser] = useState<CurrentUser | null | undefined>(undefined);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const u = await fetchCurrentUser();
        if (!cancelled) setUser(u);
      } catch {
        if (!cancelled) setUser(null);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const onLogout = async () => {
    try {
      await logout();
    } finally {
      window.location.href = "/login";
    }
  };

  return (
    <nav className="app-topnav" aria-label="Primary">
      <Link href="/landing" className="app-topnav-logo">
        Arg<span>u</span>s
      </Link>
      <div className="app-topnav-center">
        <Link href="/landing" className="app-topnav-link">
          Home
        </Link>
        <Link href="/dashboard" className="app-topnav-link">
          Dashboard
        </Link>
        <Link href="/platforms" className="app-topnav-link">
          Platforms
        </Link>
        <Link href="/notifications" className="app-topnav-link">
          Notifications
        </Link>
      </div>
      <div className="app-topnav-actions">
        {user === undefined ? (
          <span className="app-topnav-link opacity-60">…</span>
        ) : user ? (
          <>
            <span className="app-topnav-link max-w-[220px] truncate opacity-90" title={user.email}>
              {user.email}
            </span>
            <button type="button" className="app-topnav-link" onClick={onLogout}>
              Log out
            </button>
          </>
        ) : (
          <>
            <Link href="/login" className="app-topnav-link">
              Log in
            </Link>
            <Link href="/register" className="app-topnav-cta">
              <span>Get started</span>
            </Link>
          </>
        )}
      </div>
    </nav>
  );
}

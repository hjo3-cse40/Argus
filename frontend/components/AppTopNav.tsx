"use client";

import Link from "next/link";

export function AppTopNav() {
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
      </div>
      <div className="app-topnav-actions">
        <Link href="/login" className="app-topnav-link">
          Log in
        </Link>
        <Link href="/register" className="app-topnav-cta">
          <span>Get started</span>
        </Link>
      </div>
    </nav>
  );
}

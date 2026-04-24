"use client";

import Link from "next/link";

export function AppTopNav() {
  return (
    <nav className="app-topnav" aria-label="Primary">
      <Link href="/landing" className="app-topnav-logo">
        Arg<span>u</span>s
      </Link>
      <div className="app-topnav-links">
        <Link href="/landing" className="app-topnav-link">
          Home
        </Link>
        <Link href="/dashboard" className="app-topnav-link">
          Dashboard
        </Link>
        <Link href="/platforms" className="app-topnav-link">
          Platforms
        </Link>
        <Link href="/register" className="app-topnav-link">
          Register
        </Link>
        <Link href="/login" className="app-topnav-cta">
          <span>Sign in →</span>
        </Link>
      </div>
    </nav>
  );
}

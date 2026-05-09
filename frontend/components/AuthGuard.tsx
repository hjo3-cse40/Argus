"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";

function getCookie(name: string): string | null {
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : null;
}

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    let token = localStorage.getItem("token");
    if (!token) {
      token = getCookie("token");
      if (token) {
        localStorage.setItem("token", token);
      } else {
        router.replace("/login");
        return;
      }
    }
    document.cookie = `token=${token}; path=/; max-age=${7 * 24 * 60 * 60}; SameSite=Lax`;
    setReady(true);
  }, [router]);

  if (!ready) return null;
  return <>{children}</>;
}

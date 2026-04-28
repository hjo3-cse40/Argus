"use client";

import { useEffect, useRef } from "react";
import { usePathname } from "next/navigation";
import { useToast } from "@/components/ToastProvider";
import type { Delivery } from "@/lib/api";

function shouldShowToast(pathname: string): boolean {
  return pathname === "/dashboard" || pathname === "/platforms" || pathname === "/notifications";
}

export function GlobalDeliveryToasts() {
  const pathname = usePathname();
  const { showToast } = useToast();
  const recentEventIDs = useRef<Set<string>>(new Set());

  useEffect(() => {
    const es = new EventSource("/api/deliveries/stream");

    const onDelivered = (evt: MessageEvent) => {
      if (!shouldShowToast(pathname)) return;

      try {
        const incoming = JSON.parse(evt.data) as Delivery;
        if (incoming.status !== "delivered") return;
        if (recentEventIDs.current.has(incoming.event_id)) return;

        recentEventIDs.current.add(incoming.event_id);
        if (recentEventIDs.current.size > 60) {
          const first = recentEventIDs.current.values().next().value;
          if (first) recentEventIDs.current.delete(first);
        }

        showToast({
          title: incoming.title || "New notification delivered",
          message: incoming.source || "Argus",
          variant: "success",
        });
      } catch {
        // Ignore malformed stream payloads.
      }
    };

    es.addEventListener("delivered", onDelivered as EventListener);

    return () => {
      es.removeEventListener("delivered", onDelivered as EventListener);
      es.close();
    };
  }, [pathname, showToast]);

  return null;
}

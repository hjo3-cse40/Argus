"use client";

import { KeywordFiltersPanel } from "@/components/KeywordFiltersPanel";
import { AppNav } from "@/components/AppNav";
import { AppTopNav } from "@/components/AppTopNav";
import { RequireAuth } from "@/components/RequireAuth";
import "../app-shell.css";

export default function FiltersPage() {
  return (
    <RequireAuth>
      <div className="app-shell">
        <AppTopNav />
        <div className="app-body">
          <aside className="app-sidebar">
            <AppNav />
            <p className="app-sidebar-hint">
              Keyword rules apply per Discord destination platform. Sub-channels inherit delivery
              when events pass these filters.
            </p>
          </aside>

          <main className="app-main">
            <p className="app-eyebrow">Routing</p>
            <h1 className="app-page-title">
              <em>Filters</em>
            </h1>
            <p className="app-page-sub">
              Include and exclude patterns match text in the event title and description.
            </p>

            <section className="app-section-block app-max-w-form">
              <KeywordFiltersPanel />
            </section>
          </main>
        </div>
      </div>
    </RequireAuth>
  );
}

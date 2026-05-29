import "../app/app-shell.css";

export default function LoadingSkeleton() {
  return (
    <div className="app-shell">
      <nav className="app-topnav" aria-label="Primary">
        <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
          <span style={{ fontFamily: "'DM Serif Display', serif", fontSize: "1.5rem", opacity: 0.2 }}>
            Arg<span style={{ opacity: 0.3 }}>u</span>s
          </span>
        </div>
        <div className="app-topnav-center">
          <span style={{ opacity: 0.15 }}>Dashboard</span>
          <span style={{ opacity: 0.15 }}>Platforms</span>
          <span style={{ opacity: 0.15 }}>Notifications</span>
        </div>
        <div className="app-topnav-actions">
          <span style={{ opacity: 0.15 }}>user@example.com</span>
        </div>
      </nav>

      <div className="app-body">
        <aside className="app-sidebar">
          <nav className="app-nav" aria-label="Sidebar">
            <span style={{ display: "block", opacity: 0.1, height: 36, marginBottom: "0.25rem", background: "var(--line)", borderRadius: 4 }} />
            <span style={{ display: "block", opacity: 0.1, height: 36, marginBottom: "0.25rem", background: "var(--line)", borderRadius: 4 }} />
            <span style={{ display: "block", opacity: 0.1, height: 36, marginBottom: "0.25rem", background: "var(--line)", borderRadius: 4 }} />
            <span style={{ display: "block", opacity: 0.1, height: 36, marginBottom: "0.25rem", background: "var(--line)", borderRadius: 4 }} />
          </nav>
          <span style={{ display: "block", opacity: 0.1, height: 40, width: "90%", marginTop: "2rem", background: "var(--line)", borderRadius: 4 }} />
        </aside>

        <main className="app-main">
          <span style={{ display: "block", opacity: 0.1, width: 80, height: 12, marginBottom: "0.75rem", background: "var(--line)", borderRadius: 4 }} />
          <span style={{ display: "block", opacity: 0.1, width: 280, height: 32, marginBottom: "1rem", background: "var(--line)", borderRadius: 4 }} />
          <span style={{ display: "block", opacity: 0.1, width: 200, height: 14, marginBottom: "2rem", background: "var(--line)", borderRadius: 4 }} />
          <span style={{ display: "block", opacity: 0.08, width: "100%", height: 200, background: "var(--line)", borderRadius: 4 }} />
        </main>
      </div>
    </div>
  );
}

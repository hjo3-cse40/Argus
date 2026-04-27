"use client";

import Link from "next/link";
import "./landing.css";

export default function Landing() {
  return (
    <div className="landing-root">
      <nav className="nav">
        <Link href="/landing" className="nav-logo">
          Arg<span>u</span>s
        </Link>
        <div className="nav-center">
          <Link href="/landing" className="nav-link">Home</Link>
          <Link href="/dashboard" className="nav-link">Dashboard</Link>
          <Link href="/platforms" className="nav-link">Platforms</Link>
        </div>
        <div className="nav-actions">
          <Link href="/login" className="nav-link">Log in</Link>
          <Link href="/register" className="nav-cta">Get started</Link>
        </div>
      </nav>

      <main>
        <section className="hero">
          <div className="hero-left">
            <p className="hero-eyebrow">Notification intelligence</p>
            <h1>
              Always<br />
              <em>watching.</em>
              <br />
              Never missing.
            </h1>
            <p className="hero-desc">
              Argus aggregates signals from YouTube, Reddit, and X — routing them to
              Discord the moment they happen. One system. Full visibility.
            </p>
            <div className="hero-actions">
              <Link href="/register" className="btn-primary">
                <span>Get started →</span>
              </Link>
              <Link href="/login" className="btn-ghost">Log in</Link>
            </div>
          </div>

          <div className="hero-right">
            <div className="eye-graphic">
              <div className="eye-ring"></div>
              <div className="eye-ring"></div>
              <div className="eye-ring"></div>
              <div className="eye-core">
                <div className="eye-pupil"></div>
              </div>
            </div>

            <div className="stats-row">
              <div className="stat">
                <span className="stat-val">3</span>
                <div className="stat-label">Platforms</div>
              </div>
              <div className="stat">
                <span className="stat-val">∞</span>
                <div className="stat-label">Subsources</div>
              </div>
              <div className="stat">
                <span className="stat-val">0ms</span>
                <div className="stat-label">Delay</div>
              </div>
            </div>
          </div>
        </section>

        <section className="features" id="features">
          <div className="feature">
            <div className="feature-num">01 / Aggregate</div>
            <h3>Hierarchical Source Management</h3>
            <p>
              Organise platforms and subsources with precision. YouTube channels,
              subreddits, X accounts — all in one place.
            </p>
          </div>
          <div className="feature">
            <div className="feature-num">02 / Route</div>
            <h3>Discord Webhook Delivery</h3>
            <p>
              Each platform routes to its own webhook. Events land in the right
              channel, every time, with full status tracking.
            </p>
          </div>
          <div className="feature">
            <div className="feature-num">03 / Audit</div>
            <h3>Notification History</h3>
            <p>
              Full delivery log with source, destination, and status. Filter by
              success, failure, or pending — nothing disappears.
            </p>
          </div>
        </section>

        <section className="use-cases" id="use-cases">
          <p className="use-cases-kicker">Use cases</p>
          <h2>What you can do with Argus</h2>
          <div className="use-cases-grid">
            <article className="use-case-card">
              <div className="use-case-meta">📡 Monitoring</div>
              <h3>Creator monitoring</h3>
              <p>
                Track channels and accounts from multiple platforms and get alerts
                as soon as new content appears.
              </p>
            </article>
            <article className="use-case-card">
              <div className="use-case-meta">🧭 Strategy</div>
              <h3>Competitor watch</h3>
              <p>
                Follow competitor activity in one dashboard and route critical
                updates directly to your team&apos;s Discord channels.
              </p>
            </article>
            <article className="use-case-card">
              <div className="use-case-meta">🚨 Alerts</div>
              <h3>Incident response</h3>
              <p>
                Monitor keywords and high-priority sources to detect incidents
                early and respond before they escalate.
              </p>
            </article>
            <article className="use-case-card">
              <div className="use-case-meta">💬 Community</div>
              <h3>Community ops</h3>
              <p>
                Keep moderators informed with real-time notifications from Reddit,
                X, and YouTube without manual checking.
              </p>
            </article>
            <article className="use-case-card">
              <div className="use-case-meta">📈 Growth</div>
              <h3>Campaign tracking</h3>
              <p>
                Watch campaign mentions across platforms and verify what was
                delivered, when, and to which destination.
              </p>
            </article>
            <article className="use-case-card">
              <div className="use-case-meta">🗂️ Insights</div>
              <h3>Audit and reporting</h3>
              <p>
                Review delivery history by source and status to understand signal
                quality and improve routing rules over time.
              </p>
            </article>
          </div>
        </section>

        <div className="cta-band">
          <h2>
            Start watching
            <br />
            your <em>sources.</em>
          </h2>
          <Link href="/register" className="btn-amber">
            Get started →
          </Link>
        </div>
      </main>

      <footer>
        <span>© 2024 Argus</span>
        <Link href="/login">Log in</Link>
      </footer>
    </div>
  );
}
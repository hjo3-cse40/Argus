"use client";

import { useState } from "react";
import { login } from "@/lib/api";
import Link from "next/link";
import "./login.css";

export default function Login() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [bannerError, setBannerError] = useState("");

  const validate = () => {
    const newErrors: Record<string, string> = {};

    if (!email || !/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = "Please enter a valid email address.";
    }

    if (!password) {
      newErrors.password = "Password is required.";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBannerError("");

    if (!validate()) return;

    setLoading(true);

    try {
      await login({ email, password });
      window.location.href = "/dashboard";
    } catch (err) {
      setLoading(false);
      setBannerError(err instanceof Error ? err.message : "Invalid email or password.");
    }
  };

  return (
    <div className="login-root">
      {/* LEFT PANEL */}
      <div className="panel-left">
        <Link href="/landing" className="panel-logo">
          Arg<span>u</span>s
        </Link>

        <div className="panel-center">
          <div className="eye-mini">
            <div className="eye-mini-ring"></div>
            <div className="eye-mini-ring"></div>
            <div className="eye-mini-core"></div>
          </div>
          <div className="panel-tagline">
            Your sources,<br />
            <em>always visible.</em>
          </div>
          <div className="panel-sub">
            Hierarchical notification aggregation
          </div>
        </div>

        <div className="panel-footer">© 2024 Argus</div>
      </div>

      {/* RIGHT PANEL */}
      <div className="panel-right">
        <div className="form-header">
          <div className="form-eyebrow">Authentication</div>
          <h1 className="form-title">
            Sign <em>in</em>
          </h1>
        </div>

        {bannerError && (
          <div className="error-banner visible">
            <div className="error-icon">✕</div>
            <div className="error-text">{bannerError}</div>
          </div>
        )}

        <form onSubmit={handleSubmit} noValidate>
          <div className="field">
            <label htmlFor="email">Email address</label>
            <input
              type="email"
              id="email"
              name="email"
              placeholder="you@example.com"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className={errors.email ? "invalid" : ""}
            />
            {errors.email && (
              <div className="field-error visible">{errors.email}</div>
            )}
          </div>

          <div className="field">
            <div className="field-row">
              <label htmlFor="password">Password</label>
            </div>
            <input
              type="password"
              id="password"
              name="password"
              placeholder="••••••••"
              autoComplete="current-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className={errors.password ? "invalid" : ""}
            />
            {errors.password && (
              <div className="field-error visible">{errors.password}</div>
            )}
          </div>

          <button
            type="submit"
            className={`btn-submit ${loading ? "loading" : ""}`}
            disabled={loading}
          >
            {!loading && <span className="btn-label">Sign in →</span>}
            {loading && <div className="spinner"></div>}
          </button>
        </form>

        <div className="form-footer">
          <Link href="/landing">← Back to home</Link>
        </div>
      </div>
    </div>
  );
}
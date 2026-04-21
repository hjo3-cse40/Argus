"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import "./register.css";

export default function Register() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [errorBanner, setErrorBanner] = useState("");
  const [successBanner, setSuccessBanner] = useState(false);
  const [strength, setStrength] = useState(0);

  const validate = () => {
    const newErrors: Record<string, string> = {};
    const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    if (!email) {
      newErrors.email = "Email address is required.";
    } else if (!emailRe.test(email)) {
      newErrors.email = "Please enter a valid email address.";
    }

    if (!password) {
      newErrors.password = "Password is required.";
    } else if (password.length < 8) {
      newErrors.password = "Password must be at least 8 characters.";
    }

    if (!confirm) {
      newErrors.confirm = "Please confirm your password.";
    } else if (confirm !== password) {
      newErrors.confirm = "Passwords do not match.";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const scorePassword = (p: string) => {
    let score = 0;
    if (p.length >= 8) score++;
    if (p.length >= 12) score++;
    if (/[A-Z]/.test(p) && /[a-z]/.test(p)) score++;
    if (/[0-9]/.test(p)) score++;
    if (/[^A-Za-z0-9]/.test(p)) score++;
    return Math.min(4, score);
  };

  useEffect(() => {
    setStrength(scorePassword(password));
  }, [password]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorBanner("");

    if (!validate()) return;

    setLoading(true);

    try {
      const res = await fetch("/api/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (res.ok) {
        setSuccessBanner(true);
        setTimeout(() => {
          window.location.href = "/login";
        }, 1800);
        return;
      }

      if (res.status === 409) {
        setErrors({ email: "This email is already registered." });
        return;
      }

      if (res.status === 400) {
        const data = await res.json().catch(() => ({}));
        const detail = data.details?.[0] || data.error || "Validation failed.";
        setErrorBanner(detail);
        return;
      }

      setErrorBanner("Server error. Please try again in a moment.");
    } catch {
      setErrorBanner("Cannot reach the server. Check your connection.");
    } finally {
      setLoading(false);
    }
  };

  const strengthLabels = ["", "Weak", "Fair", "Good", "Strong"];

  return (
    <div className="register-root">
      <div className="panel-left">
        <Link href="/" className="panel-logo">
          Arg<span>u</span>s
        </Link>

        <div className="panel-center">
          <div className="eye-mini">
            <div className="eye-mini-ring"></div>
            <div className="eye-mini-ring"></div>
            <div className="eye-mini-core"></div>
          </div>
          <div className="panel-tagline">
            Start <em>watching</em>
            <br />
            your sources.
          </div>
          <div className="panel-sub">Everything in one place</div>
          <ul className="panel-checklist">
            <li>YouTube, Reddit &amp; X aggregation</li>
            <li>Discord webhook delivery</li>
            <li>Full notification history</li>
            <li>Hierarchical source management</li>
          </ul>
        </div>

        <div className="panel-footer">© 2024 Argus</div>
      </div>

      <div className="panel-right">
        <div className="form-header">
          <div className="form-eyebrow">New account</div>
          <h1 className="form-title">
            Create <em>account</em>
          </h1>
        </div>

        {errorBanner && (
          <div className="banner banner-error visible">
            <div className="banner-icon-error">✕</div>
            <div className="banner-text-error">{errorBanner}</div>
          </div>
        )}

        {successBanner && (
          <div className="banner banner-success visible">
            <div className="banner-icon-success">✓</div>
            <div className="banner-text-success">
              Account created! Redirecting to login…
            </div>
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
              onChange={(e) => {
                setEmail(e.target.value);
                if (errors.email) setErrors({ ...errors, email: "" });
              }}
              className={errors.email ? "invalid" : email && !errors.email ? "valid" : ""}
            />
            {errors.email && (
              <div className="field-error visible">{errors.email}</div>
            )}
          </div>

          <div className="field">
            <label htmlFor="password">Password</label>
            <input
              type="password"
              id="password"
              name="password"
              placeholder="••••••••"
              autoComplete="new-password"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                if (errors.password) setErrors({ ...errors, password: "" });
              }}
              className={errors.password ? "invalid" : password && !errors.password ? "valid" : ""}
            />
            <div className={`strength-bar ${password ? "visible" : ""}`}>
              <div className={`strength-seg ${strength >= 1 ? `filled-${strength}` : ""}`}></div>
              <div className={`strength-seg ${strength >= 2 ? `filled-${strength}` : ""}`}></div>
              <div className={`strength-seg ${strength >= 3 ? `filled-${strength}` : ""}`}></div>
              <div className={`strength-seg ${strength >= 4 ? `filled-${strength}` : ""}`}></div>
            </div>
            <div className={`strength-label ${password ? "visible" : ""}`}>
              {strengthLabels[strength]}
            </div>
            {errors.password && (
              <div className="field-error visible">{errors.password}</div>
            )}
          </div>

          <div className="field">
            <label htmlFor="confirm">Confirm password</label>
            <input
              type="password"
              id="confirm"
              name="confirm"
              placeholder="••••••••"
              autoComplete="new-password"
              value={confirm}
              onChange={(e) => {
                setConfirm(e.target.value);
                if (errors.confirm) setErrors({ ...errors, confirm: "" });
              }}
              className={errors.confirm ? "invalid" : confirm && !errors.confirm && password === confirm ? "valid" : ""}
            />
            {errors.confirm && (
              <div className="field-error visible">{errors.confirm}</div>
            )}
          </div>

          <button
            type="submit"
            className={`btn-submit ${loading ? "loading" : ""}`}
            disabled={loading}
          >
            {!loading && <span className="btn-label">Create account →</span>}
            {loading && <div className="spinner"></div>}
          </button>
        </form>

        <div className="form-footer">
          Already have an account? <Link href="/login">Sign in</Link>
          &nbsp;·&nbsp; <Link href="/">Home</Link>
        </div>
      </div>
    </div>
  );
}
"use client";

import { createContext, useCallback, useContext, useMemo, useState } from "react";

type ToastVariant = "info" | "success" | "error";

type ToastInput = {
  title: string;
  message?: string;
  variant?: ToastVariant;
  durationMs?: number;
};

type Toast = ToastInput & {
  id: string;
  variant: ToastVariant;
};

type ToastContextValue = {
  showToast: (toast: ToastInput) => void;
};

const ToastContext = createContext<ToastContextValue | null>(null);

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const dismissToast = useCallback((id: string) => {
    setToasts((current) => current.filter((t) => t.id !== id));
  }, []);

  const showToast = useCallback(
    ({ title, message, variant = "info", durationMs = 4200 }: ToastInput) => {
      const id = `${Date.now()}-${Math.random().toString(16).slice(2, 8)}`;
      setToasts((current) => [...current, { id, title, message, variant }]);
      window.setTimeout(() => dismissToast(id), durationMs);
    },
    [dismissToast]
  );

  const value = useMemo(() => ({ showToast }), [showToast]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="app-toast-region" aria-live="polite" aria-atomic="true">
        {toasts.map((toast) => (
          <article key={toast.id} className={`app-toast app-toast-${toast.variant}`}>
            <div className="app-toast-body">
              <p className="app-toast-title">{toast.title}</p>
              {toast.message ? <p className="app-toast-message">{toast.message}</p> : null}
            </div>
            <button
              type="button"
              className="app-toast-close"
              aria-label="Close notification"
              onClick={() => dismissToast(toast.id)}
            >
              ×
            </button>
          </article>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within ToastProvider");
  }
  return context;
}

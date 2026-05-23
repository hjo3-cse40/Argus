"use client";

import { useState } from "react";
import { SunMedium, MoonStar } from "lucide-react";

type ThemeToggleProps = {
    initialDark?: boolean;
};

export default function ThemeToggle({
    initialDark = false,
}: ThemeToggleProps) {
    const [dark, setDark] = useState(initialDark);

    function toggleTheme() {
        const next = !dark;
        setDark(next);
        document.documentElement.classList.toggle("dark", next);
        localStorage.setItem("theme", next ? "dark" : "light");
        document.cookie = `theme=${next ? "dark" : "light"
            }; path=/; max-age=31536000`;
    }

    return (
        <button
            onClick={toggleTheme}
            className={`theme-toggle ${dark ? "active" : ""}`}
            aria-label="Toggle theme"
        >
            <span className="toggle-thumb">
                <span className="toggle-icon">
                    {dark ? <MoonStar size={14} /> : <SunMedium size={14} />}
                </span>
            </span>
        </button>
    );
}
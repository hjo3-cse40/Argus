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

        if (next) {
            document.documentElement.classList.add("dark");
            localStorage.setItem("theme", "dark");
            document.cookie = "theme=dark; path=/; max-age=31536000";
        } else {
            document.documentElement.classList.remove("dark");
            localStorage.setItem("theme", "light");
            document.cookie = "theme=light; path=/; max-age=31536000";
        }
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
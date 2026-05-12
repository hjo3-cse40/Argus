import ThemeToggle from "@/components/ThemeToggle";
import { AppNav } from "@/components/AppNav";
import { AppTopNav } from "@/components/AppTopNav";
import { cookies } from "next/headers";
import "../app-shell.css";

export default async function SettingsPage() {
    const cookieStore = await cookies();
    const isDark = cookieStore.get("theme")?.value === "dark";
    return (
        <div className="app-shell">
            <AppTopNav />

            <div className="app-body">
                <aside className="app-sidebar">
                    <AppNav />

                    <div className="app-sidebar-hint">
                        Change your display preferences for Argus.
                    </div>
                </aside>

                <main className="app-main">
                    <p className="app-eyebrow">Settings</p>

                    <h1 className="app-page-title">
                        Appearance
                    </h1>

                    <section className="app-max-w-form space-y-2">
                        <h2 className="app-section-title">Theme</h2>

                        <div className="theme-setting-row">
                            <ThemeToggle initialDark={isDark} />
                        </div>
                    </section>
                </main>
            </div>
        </div>
    );
}
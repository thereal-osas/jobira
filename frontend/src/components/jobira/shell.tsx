import { Link, useRouterState } from "@tanstack/react-router";
import { ArrowRight, BriefcaseBusiness, LayoutDashboard, Menu, Moon, Plus, Sun, X } from "lucide-react";
import { useState, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { Logo } from "./logo";
import { useTheme } from "./theme";
import { cn } from "@/lib/utils";

const nav = [
  { label: "Find Work", to: "/jobs" },
  { label: "Find Professionals", to: "/professionals" },
  { label: "For Business", to: "/dashboard" },
  { label: "How It Works", to: "/", hash: "how-it-works" },
  { label: "Pricing", to: "/pricing" },
] as const;

export function ThemeButton() {
  const { theme, toggleTheme } = useTheme();
  return (
    <Button variant="ghost" size="icon" onClick={toggleTheme} aria-label={`Switch to ${theme === "light" ? "dark" : "light"} mode`} title="Change appearance">
      {theme === "light" ? <Moon /> : <Sun />}
    </Button>
  );
}

export function Header() {
  const [open, setOpen] = useState(false);
  const path = useRouterState({ select: (state) => state.location.pathname });
  return (
    <header className="sticky top-0 z-50 border-b border-border bg-background/95 backdrop-blur-md">
      <div className="page-wrap flex h-[76px] items-center justify-between">
        <Logo />
        <nav className="hidden items-center gap-7 lg:flex" aria-label="Main navigation">
          {nav.map((item) => <Link key={item.label} to={item.to} {...("hash" in item ? { hash: item.hash } : {})} className={cn("nav-link", path === item.to && item.to !== "/" && "text-foreground")}>{item.label}</Link>)}
        </nav>
        <div className="hidden items-center gap-2 lg:flex">
          <ThemeButton />
          <Button variant="ghost" asChild><Link to="/client">Client area</Link></Button>
          <Button asChild><Link to="/coming-soon">Sign Up <ArrowRight /></Link></Button>
        </div>
        <div className="flex items-center gap-1 lg:hidden">
          <ThemeButton />
          <Button variant="ghost" size="icon" onClick={() => setOpen(!open)} aria-label="Toggle navigation">{open ? <X /> : <Menu />}</Button>
        </div>
      </div>
      {open && <nav className="border-t border-border bg-background px-5 py-4 lg:hidden" aria-label="Mobile navigation">
        {nav.map((item) => <Link key={item.label} to={item.to} {...("hash" in item ? { hash: item.hash } : {})} onClick={() => setOpen(false)} className="block border-b border-border py-3 text-sm font-semibold">{item.label}</Link>)}
        <Button className="mt-4 w-full" asChild><Link to="/coming-soon">Create account</Link></Button>
      </nav>}
    </header>
  );
}

export function Footer() {
  return (
    <footer className="border-t border-border bg-background py-10">
      <div className="page-wrap flex flex-col gap-7 md:flex-row md:items-end md:justify-between">
        <div><Logo /><p className="mt-4 max-w-sm text-sm text-muted-foreground">A fairer way for people, professionals and businesses to create lasting opportunities.</p></div>
        <div className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-muted-foreground">
          <Link to="/jobs">Find work</Link><Link to="/professionals">Professionals</Link><Link to="/pricing">Pricing</Link><Link to="/client">Client area</Link><Link to="/admin">Admin</Link><Link to="/coming-soon">Contact</Link>
        </div>
        <p className="text-xs text-muted-foreground">© 2026 Jobira</p>
      </div>
    </footer>
  );
}

export function PublicShell({ children }: { children: ReactNode }) {
  return <div className="min-h-screen bg-background text-foreground"><Header /><main>{children}</main><Footer /></div>;
}

const dashboardNav = [
  { label: "Overview", to: "/dashboard", icon: LayoutDashboard },
  { label: "My jobs", to: "/jobs", icon: BriefcaseBusiness },
  { label: "Post a job", to: "/post-job", icon: Plus },
] as const;

export function DashboardShell({ children, title, description }: { children: ReactNode; title: string; description: string }) {
  const path = useRouterState({ select: (state) => state.location.pathname });
  return <div className="min-h-screen bg-background text-foreground md:grid md:grid-cols-[236px_1fr]">
    <aside className="border-b border-border bg-sidebar px-5 py-5 md:min-h-screen md:border-b-0 md:border-r">
      <Logo compact />
      <nav className="mt-8 flex gap-2 overflow-x-auto md:flex-col">
        {dashboardNav.map((item) => <Link key={item.label} to={item.to} className={cn("flex shrink-0 items-center gap-3 rounded-md px-3 py-2.5 text-sm font-semibold text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground", path === item.to && "bg-sidebar-accent text-foreground")}><item.icon className="h-4 w-4" />{item.label}</Link>)}
      </nav>
      <div className="mt-6 hidden border-t border-border pt-5 md:block"><ThemeButton /></div>
    </aside>
    <main className="min-w-0">
      <div className="flex items-center justify-between border-b border-border px-5 py-5 sm:px-8"><div><p className="section-kicker">Employer workspace</p><h1 className="mt-1 font-display text-2xl font-semibold">{title}</h1><p className="mt-1 text-sm text-muted-foreground">{description}</p></div><div className="md:hidden"><ThemeButton /></div></div>
      <div className="p-5 sm:p-8">{children}</div>
    </main>
  </div>;
}
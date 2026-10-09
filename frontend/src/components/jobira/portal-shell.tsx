import { Link, useRouterState } from "@tanstack/react-router";
import type { LucideIcon } from "lucide-react";
import { Bell, ChevronDown, Menu, Search, X } from "lucide-react";
import { useState, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { Logo } from "./logo";
import { ThemeButton } from "./shell";
import { cn } from "@/lib/utils";

export type PortalNavItem = { label: string; to: string; icon: LucideIcon };

export function PortalShell({ children, navigation, name, role, initials, context }: { children: ReactNode; navigation: PortalNavItem[]; name: string; role: string; initials: string; context: string }) {
  const [mobileOpen, setMobileOpen] = useState(false);
  const path = useRouterState({ select: (state) => state.location.pathname });
  return <div className="min-h-screen bg-muted/45 text-foreground lg:grid lg:grid-cols-[232px_1fr]">
    <aside className={cn("fixed inset-y-0 left-0 z-50 flex w-[232px] flex-col border-r border-sidebar-border bg-sidebar p-4 transition-transform lg:sticky lg:top-0 lg:h-screen lg:translate-x-0", mobileOpen ? "translate-x-0" : "-translate-x-full")}>
      <div className="flex items-center justify-between"><Logo compact /><Button className="lg:hidden" variant="ghost" size="icon" onClick={() => setMobileOpen(false)} aria-label="Close navigation"><X /></Button></div>
      <p className="mt-7 px-3 text-[0.62rem] font-extrabold uppercase tracking-[0.16em] text-muted-foreground">{context}</p>
      <nav className="mt-3 space-y-1" aria-label={`${context} navigation`}>{navigation.map((item, index) => {
        const firstExactMatch = navigation.findIndex((candidate) => candidate.to === path) === index;
        const active = firstExactMatch || (item.to !== "/admin" && item.to !== "/client" && path.startsWith(item.to));
        return <Link key={`${item.label}-${item.to}`} to={item.to} onClick={() => setMobileOpen(false)} className={cn("flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-semibold text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground", active && "bg-sidebar-accent text-sidebar-accent-foreground")}><item.icon className="h-4 w-4" />{item.label}</Link>;
      })}</nav>
      <div className="mt-auto rounded-md border border-sidebar-border bg-sidebar-accent/55 p-4"><p className="font-display text-sm font-semibold">A fairer future, together.</p><p className="mt-2 text-xs leading-5 text-muted-foreground">Connecting people, creating opportunities, building brighter communities.</p></div>
    </aside>
    {mobileOpen && <button className="fixed inset-0 z-40 bg-foreground/20 lg:hidden" onClick={() => setMobileOpen(false)} aria-label="Close navigation overlay" />}
    <div className="min-w-0">
      <header className="sticky top-0 z-30 flex h-[72px] items-center gap-3 border-b border-border bg-background/95 px-4 backdrop-blur-md sm:px-6">
        <Button className="lg:hidden" variant="ghost" size="icon" onClick={() => setMobileOpen(true)} aria-label="Open navigation"><Menu /></Button>
        <label className="relative hidden min-w-0 max-w-md flex-1 md:block"><span className="sr-only">Search</span><Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" /><input className="h-10 w-full rounded-md border border-input bg-card pl-10 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" placeholder="Search Jobira..." /></label>
        <div className="ml-auto flex items-center gap-1 sm:gap-2"><ThemeButton /><Button variant="ghost" size="icon" aria-label="Notifications" className="relative"><Bell /><span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-destructive" /></Button><div className="ml-1 flex items-center gap-2 border-l border-border pl-3"><div className="flex h-9 w-9 items-center justify-center rounded-full bg-secondary text-xs font-bold">{initials}</div><div className="hidden sm:block"><p className="text-xs font-bold">{name}</p><p className="text-[0.68rem] text-muted-foreground">{role}</p></div><ChevronDown className="hidden h-4 w-4 text-muted-foreground sm:block" /></div></div>
      </header>
      <main className="p-4 sm:p-6">{children}</main>
    </div>
  </div>;
}

export function PortalTitle({ title, description, action }: { title: string; description: string; action?: ReactNode }) {
  return <div className="mb-5 flex flex-col justify-between gap-4 sm:flex-row sm:items-end"><div><h1 className="font-display text-2xl font-semibold sm:text-3xl">{title}</h1><p className="mt-1 text-sm text-muted-foreground">{description}</p></div>{action}</div>;
}

export function MetricGrid({ items }: { items: { label: string; value: string; note: string; icon: LucideIcon; tone?: "blue" | "green" | "orange" | "violet" }[] }) {
  return <section className="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">{items.map((item) => <article key={item.label} className="dashboard-card p-5"><div className="flex items-center gap-3"><span className={cn("flex h-9 w-9 items-center justify-center rounded-full", item.tone === "blue" && "bg-accent/10 text-accent", item.tone === "orange" && "bg-warning/15 text-warning", item.tone === "violet" && "bg-violet/10 text-violet", (!item.tone || item.tone === "green") && "bg-soft-green text-brand-leaf")}><item.icon className="h-4 w-4" /></span><p className="text-xs font-bold">{item.label}</p></div><div className="mt-4 flex items-end gap-3"><strong className="font-display text-3xl">{item.value}</strong><span className="pb-1 text-xs font-bold text-brand-leaf">{item.note}</span></div></article>)}</section>;
}

export function MiniChart({ bars = false }: { bars?: boolean }) {
  const values = [30, 42, 28, 52, 48, 66, 54, 72, 61, 78, 70, 88];
  return <div className="mt-6 flex h-40 items-end gap-2 border-b border-border">{values.map((value, index) => <div key={`${value}-${index}`} className="flex h-full flex-1 items-end gap-0.5"><span className={cn("chart-bar block w-full bg-accent/80", bars && "w-1/2")} style={{ height: `${value}%`, animationDelay: `${index * 55}ms` }} />{bars && <span className="chart-bar block w-1/2 bg-primary/75" style={{ height: `${Math.max(20, value - 17)}%`, animationDelay: `${index * 55 + 40}ms` }} />}</div>)}</div>;
}

const lineSeries = [
  { label: "Posted", color: "var(--accent)", values: [24, 36, 31, 49, 44, 61, 57, 72, 68, 81, 76, 88] },
  { label: "Completed", color: "var(--brand-leaf)", values: [15, 22, 26, 31, 37, 35, 49, 52, 58, 62, 70, 74] },
  { label: "Cancelled", color: "var(--warning)", values: [8, 11, 7, 13, 10, 15, 12, 18, 14, 17, 13, 16] },
];

export function MultiSeriesLineChart() {
  const points = (values: number[]) => values.map((value, index) => `${24 + index * 42},${132 - value}`).join(" ");
  return <div className="mt-5">
    <div className="flex flex-wrap gap-4 text-xs">{lineSeries.map((series) => <span key={series.label} className="flex items-center gap-2"><i className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: series.color }} />{series.label}</span>)}</div>
    <div className="mt-3 overflow-hidden"><svg viewBox="0 0 510 150" className="h-44 w-full" role="img" aria-label="Posted, completed and cancelled jobs over twelve weeks"><g className="text-border" stroke="currentColor" strokeWidth="1">{[32, 62, 92, 122].map((y) => <line key={y} x1="20" x2="496" y1={y} y2={y} />)}</g>{lineSeries.map((series) => <g key={series.label}><polyline points={points(series.values)} fill="none" stroke={series.color} strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" className="chart-line" />{series.values.map((value, index) => <circle key={`${series.label}-${index}`} cx={24 + index * 42} cy={132 - value} r="2.5" fill={series.color} />)}</g>)}</svg></div>
    <div className="grid grid-cols-6 text-[0.62rem] text-muted-foreground sm:grid-cols-12">{["W1", "W2", "W3", "W4", "W5", "W6", "W7", "W8", "W9", "W10", "W11", "W12"].map((week, index) => <span key={week} className={index % 2 ? "hidden sm:block" : "block"}>{week}</span>)}</div>
  </div>;
}
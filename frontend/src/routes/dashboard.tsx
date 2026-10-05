import { createFileRoute, Link } from "@tanstack/react-router";
import { ArrowUpRight, BriefcaseBusiness, CalendarClock, Eye, MoreHorizontal, Plus, Users } from "lucide-react";
import { DashboardShell } from "@/components/jobira/shell";
import { Button } from "@/components/ui/button";

const metrics = [
  { label: "Active jobs", value: "6", note: "+2 this month", icon: BriefcaseBusiness },
  { label: "New applicants", value: "48", note: "+18 this week", icon: Users },
  { label: "Profile views", value: "1,284", note: "+12.4%", icon: Eye },
  { label: "Interviews", value: "9", note: "Next: tomorrow", icon: CalendarClock },
];

const applicants = [
  ["AM", "Amara Mensah", "Office Cleaning Professional", "4.9", "2 hours ago"],
  ["JL", "Jordan Lewis", "Facilities Team Lead", "5.0", "Yesterday"],
  ["SO", "Sofia Oliveira", "Housekeeping Specialist", "4.8", "Yesterday"],
];

export const Route = createFileRoute("/dashboard")({
  head: () => ({ meta: [{ title: "Employer Dashboard | Jobira" }, { name: "description", content: "Manage Jobira jobs, candidates and interviews." }, { property: "og:title", content: "Employer Dashboard | Jobira" }, { property: "og:description", content: "Manage Jobira jobs, candidates and interviews." }, { property: "og:type", content: "website" }, { name: "twitter:card", content: "summary_large_image" }] }),
  component: DashboardPage,
});

function DashboardPage() {
  return <DashboardShell title="Good evening, Jamie." description="Here’s what’s happening with your opportunities.">
    <div className="flex justify-end"><Button asChild><Link to="/post-job"><Plus />Post a job</Link></Button></div>
    <section className="mt-5 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">{metrics.map((metric) => <article key={metric.label} className="border border-border bg-card p-5"><div className="flex items-center justify-between"><p className="text-xs font-bold uppercase tracking-wider text-muted-foreground">{metric.label}</p><metric.icon className="h-4 w-4 text-brand-leaf" /></div><p className="mt-5 font-display text-3xl font-semibold">{metric.value}</p><p className="mt-2 text-xs text-muted-foreground">{metric.note}</p></article>)}</section>
    <div className="mt-6 grid gap-6 xl:grid-cols-[1.35fr_0.65fr]">
      <section className="border border-border bg-card"><div className="flex items-center justify-between border-b border-border p-5"><div><p className="section-kicker">Recent activity</p><h2 className="mt-1 font-display text-lg font-semibold">New applicants</h2></div><Button variant="ghost" size="sm">View all <ArrowUpRight /></Button></div><div>{applicants.map(([initials, name, role, rating, time]) => <div key={name} className="grid grid-cols-[auto_1fr_auto] items-center gap-3 border-b border-border p-4 last:border-b-0"><div className="flex h-9 w-9 items-center justify-center rounded-full bg-soft-green text-xs font-bold text-secondary-foreground">{initials}</div><div className="min-w-0"><p className="truncate text-sm font-semibold">{name}</p><p className="truncate text-xs text-muted-foreground">{role} · ★ {rating}</p></div><div className="flex items-center gap-2"><span className="hidden text-xs text-muted-foreground sm:block">{time}</span><Button variant="ghost" size="icon" aria-label={`Actions for ${name}`}><MoreHorizontal /></Button></div></div>)}</div></section>
      <section className="border border-border bg-card p-5"><p className="section-kicker">Hiring pulse</p><h2 className="mt-2 font-display text-lg font-semibold">Applications this week</h2><div className="mt-8 flex h-44 items-end gap-3 border-b border-border">{[38, 55, 43, 76, 92, 62, 84].map((height, index) => <div key={index} className="group relative flex-1"><div className="w-full bg-primary/75 transition-colors group-hover:bg-primary" style={{ height: `${height * 1.5}px` }} /></div>)}</div><div className="mt-3 flex justify-between text-[0.62rem] uppercase tracking-wider text-muted-foreground"><span>Mon</span><span>Tue</span><span>Wed</span><span>Thu</span><span>Fri</span><span>Sat</span><span>Sun</span></div></section>
    </div>
    <section className="mt-6 border border-border bg-card"><div className="flex items-center justify-between border-b border-border p-5"><div><p className="section-kicker">Your listings</p><h2 className="mt-1 font-display text-lg font-semibold">Active jobs</h2></div><Button variant="outline" size="sm" asChild><Link to="/post-job"><Plus />New listing</Link></Button></div><div className="overflow-x-auto"><table className="w-full min-w-[660px] text-left text-sm"><thead className="bg-muted text-xs uppercase tracking-wider text-muted-foreground"><tr><th className="px-5 py-3">Role</th><th className="px-5 py-3">Status</th><th className="px-5 py-3">Applicants</th><th className="px-5 py-3">Closes</th></tr></thead><tbody>{[["Office Cleaning Professional", "Live", "24", "22 Sep"], ["Facilities Team Lead", "Live", "11", "29 Sep"], ["Weekend Housekeeper", "Review", "13", "18 Sep"]].map((row) => <tr key={row[0]} className="border-t border-border"><td className="px-5 py-4 font-semibold">{row[0]}</td><td className="px-5 py-4"><span className="rounded-sm bg-secondary px-2 py-1 text-xs font-semibold">{row[1]}</span></td><td className="px-5 py-4">{row[2]}</td><td className="px-5 py-4 text-muted-foreground">{row[3]}</td></tr>)}</tbody></table></div></section>
  </DashboardShell>;
}
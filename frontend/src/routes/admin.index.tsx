import { createFileRoute, Link } from "@tanstack/react-router";
import { BriefcaseBusiness, CheckCircle2, ClipboardCheck, ShieldCheck, UserRound, Users } from "lucide-react";
import { MetricGrid, MiniChart, MultiSeriesLineChart, PortalTitle } from "@/components/jobira/portal-shell";
import { Button } from "@/components/ui/button";

const recentJobs = [
  ["#231", "End of tenancy clean", "London", "End of tenancy", "Open"],
  ["#230", "Regular home clean", "Stratford", "Domestic", "Open"],
  ["#229", "Airbnb turnover", "Canary Wharf", "Airbnb", "Open"],
  ["#228", "Office cleaning", "Hackney", "Commercial", "Open"],
];
const recentUsers = [["SM", "Sarah M.", "Cleaner"], ["DK", "Daniel K.", "Client"], ["MT", "Maria T.", "Cleaner"], ["JR", "James R.", "Cleaner"]];

export const Route = createFileRoute("/admin/")({
  head: () => ({ meta: [{ title: "Admin Dashboard | Jobira" }, { name: "description", content: "Monitor Jobira platform users, jobs and activity." }, { property: "og:title", content: "Admin Dashboard | Jobira" }, { property: "og:description", content: "Monitor Jobira platform users, jobs and activity." }, { property: "og:type", content: "website" }, { name: "twitter:card", content: "summary_large_image" }] }),
  component: AdminOverview,
});

function AdminOverview() {
  return <>
    <PortalTitle title="Admin Dashboard" description="Overview of platform activity, users, jobs and performance." />
    <MetricGrid items={[
      { label: "Total users", value: "1,248", note: "↑ 12%", icon: UserRound, tone: "blue" },
      { label: "Cleaners", value: "892", note: "↑ 14%", icon: Users },
      { label: "Clients", value: "356", note: "↑ 8%", icon: Users, tone: "violet" },
      { label: "Active jobs", value: "187", note: "↑ 21%", icon: ClipboardCheck, tone: "blue" },
      { label: "Bookings", value: "624", note: "↑ 18%", icon: CheckCircle2 },
    ]} />
    <section className="mt-4 grid gap-4 xl:grid-cols-2">
      <article className="dashboard-card p-5"><div className="flex justify-between"><div><h2 className="font-display text-lg font-semibold">Job activity</h2><p className="text-xs text-muted-foreground">Posted, completed and cancelled jobs</p></div><span className="text-xs font-semibold">Last 12 weeks</span></div><MultiSeriesLineChart /></article>
      <article className="dashboard-card p-5"><div className="flex justify-between"><div><h2 className="font-display text-lg font-semibold">User growth</h2><p className="text-xs text-muted-foreground">New signups by account type</p></div><span className="text-xs font-semibold">Last 30 days</span></div><MiniChart bars /></article>
    </section>
    <section className="mt-4 grid gap-4 xl:grid-cols-[1.15fr_0.8fr_0.85fr]">
      <article className="dashboard-card overflow-hidden"><div className="flex items-center justify-between border-b border-border p-4"><h2 className="font-display font-semibold">Recent jobs</h2><Button variant="link" size="sm" asChild><Link to="/admin/jobs">View all</Link></Button></div><div className="overflow-x-auto"><table className="w-full min-w-[520px] text-left text-xs"><thead className="bg-muted text-muted-foreground"><tr><th className="px-4 py-3">#</th><th>Title</th><th>Location</th><th>Type</th><th>Status</th></tr></thead><tbody>{recentJobs.map((row) => <tr key={row[0]} className="border-t border-border"><td className="px-4 py-3 font-semibold">{row[0]}</td><td>{row[1]}</td><td>{row[2]}</td><td>{row[3]}</td><td><span className="rounded-sm bg-soft-green px-2 py-1 text-brand-leaf">{row[4]}</span></td></tr>)}</tbody></table></div></article>
      <article className="dashboard-card"><div className="flex items-center justify-between border-b border-border p-4"><h2 className="font-display font-semibold">Recent users</h2><Button variant="link" size="sm" asChild><Link to="/admin/users">View all</Link></Button></div>{recentUsers.map(([initials, name, role]) => <div key={name} className="flex items-center gap-3 border-b border-border p-3 last:border-b-0"><span className="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-[0.65rem] font-bold">{initials}</span><div><p className="text-xs font-semibold">{name}</p><p className="text-[0.65rem] text-muted-foreground">{role}</p></div></div>)}</article>
      <article className="dashboard-card"><div className="border-b border-border p-4"><h2 className="font-display font-semibold">Recent activity</h2></div>{[[BriefcaseBusiness, "New job posted", "5 mins ago"], [UserRound, "New user registered", "18 mins ago"], [CheckCircle2, "Booking completed", "1 hour ago"], [ShieldCheck, "User verified", "3 hours ago"]].map(([Icon, label, time]) => { const ActivityIcon = Icon as typeof BriefcaseBusiness; return <div key={String(label)} className="flex items-center gap-3 border-b border-border p-3 last:border-b-0"><span className="flex h-8 w-8 items-center justify-center rounded-full bg-secondary"><ActivityIcon className="h-4 w-4 text-brand-leaf" /></span><div className="min-w-0"><p className="truncate text-xs font-semibold">{String(label)}</p><p className="text-[0.65rem] text-muted-foreground">{String(time)}</p></div></div>; })}</article>
    </section>
  </>;
}
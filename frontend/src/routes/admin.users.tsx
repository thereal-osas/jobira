import { createFileRoute } from "@tanstack/react-router";
import { Download, MoreHorizontal, Plus, Search, ShieldCheck, UserRound, Users } from "lucide-react";
import { useState } from "react";
import { MetricGrid, PortalTitle } from "@/components/jobira/portal-shell";
import { Button } from "@/components/ui/button";

type UserRow = [string, string, string, string, string, string, string];
const users: UserRow[] = [
  ["SM", "Sarah M.", "sarah.m@example.com", "Cleaner", "Verified", "East London", "Active"],
  ["DK", "Daniel K.", "daniel.k@example.com", "Client", "Verified", "Stratford", "Active"],
  ["MT", "Maria T.", "maria.t@example.com", "Cleaner", "Pending", "North London", "Active"],
  ["JR", "James R.", "james.r@example.com", "Cleaner", "Verified", "West London", "Active"],
  ["AP", "Aisha P.", "aisha.p@example.com", "Client", "Verified", "East London", "Active"],
  ["LW", "Liam W.", "liam.w@example.com", "Cleaner", "Pending", "Hackney", "Active"],
  ["RT", "Ryan T.", "ryan.t@example.com", "Admin", "Verified", "London", "Active"],
  ["MJ", "Michael J.", "michael.j@example.com", "Client", "Verified", "Bromley", "Suspended"],
];

export const Route = createFileRoute("/admin/users")({
  head: () => ({ meta: [{ title: "User Management | Jobira" }, { name: "description", content: "Review and manage Jobira platform users." }, { property: "og:title", content: "User Management | Jobira" }, { property: "og:description", content: "Review and manage Jobira platform users." }, { property: "og:type", content: "website" }, { name: "twitter:card", content: "summary_large_image" }] }), component: UsersPage,
});

function UsersPage() {
  const [query, setQuery] = useState("");
  const [active, setActive] = useState<string | null>(null);
  const filtered = users.filter((user) => `${user[1]} ${user[2]} ${user[3]}`.toLowerCase().includes(query.toLowerCase()));
  return <><PortalTitle title="Users" description="Manage all users on the platform. View, edit, verify and take action." />
    <MetricGrid items={[{ label: "Total users", value: "1,248", note: "↑ 12%", icon: UserRound, tone: "blue" }, { label: "Cleaners", value: "892", note: "↑ 14%", icon: Users }, { label: "Clients", value: "356", note: "↑ 8%", icon: Users, tone: "violet" }, { label: "Admins", value: "12", note: "— 0%", icon: ShieldCheck, tone: "blue" }, { label: "Pending verification", value: "28", note: "↑ 33%", icon: ShieldCheck, tone: "orange" }]} />
    <section className="dashboard-card mt-4 p-3"><div className="grid gap-3 md:grid-cols-[1.5fr_repeat(3,1fr)_auto]"><label className="relative"><Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" /><input value={query} onChange={(event) => setQuery(event.target.value)} className="h-10 w-full rounded-md border border-input bg-background pl-9 pr-3 text-sm" placeholder="Name, email or ID..." /></label>{["All roles", "All statuses", "All locations"].map((label) => <select key={label} className="h-10 rounded-md border border-input bg-background px-3 text-sm"><option>{label}</option></select>)}<Button variant="outline" onClick={() => setQuery("")}>Reset</Button></div></section>
    <section className="dashboard-card mt-4"><div className="flex items-center justify-between border-b border-border p-4"><h2 className="font-display font-semibold">Users ({filtered.length})</h2><div className="flex gap-2"><Button variant="outline" size="sm"><Download />Export</Button><Button size="sm"><Plus />Add user</Button></div></div><div className="overflow-x-auto"><table className="w-full min-w-[900px] text-left text-xs"><thead className="bg-muted text-muted-foreground"><tr><th className="px-4 py-3">User</th><th>Email</th><th>Role</th><th>Verification</th><th>Location</th><th>Status</th><th className="text-right pr-4">Actions</th></tr></thead><tbody>{filtered.map((user) => <tr key={user[2]} className="border-t border-border"><td className="px-4 py-3"><div className="flex items-center gap-2"><span className="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-[0.65rem] font-bold">{user[0]}</span><strong>{user[1]}</strong></div></td><td>{user[2]}</td><td><span className="rounded-sm bg-secondary px-2 py-1">{user[3]}</span></td><td className={user[4] === "Pending" ? "text-warning" : "text-brand-leaf"}>{user[4]}</td><td>{user[5]}</td><td><span className={user[6] === "Active" ? "text-brand-leaf" : "text-destructive"}>{user[6]}</span></td><td className="relative pr-4 text-right"><Button variant="ghost" size="icon" onClick={() => setActive(active === user[2] ? null : user[2])} aria-label={`Actions for ${user[1]}`}><MoreHorizontal /></Button>{active === user[2] && <div className="absolute right-12 top-2 z-20 w-40 rounded-md border border-border bg-popover p-1 text-left shadow-lg">{["View profile", "Edit user", "Change role", "Verify user", "Suspend user", "Delete user"].map((action) => <Button key={action} variant="ghost" size="sm" className="w-full justify-start" onClick={() => setActive(null)}>{action}</Button>)}</div>}</td></tr>)}</tbody></table></div><div className="flex items-center justify-between border-t border-border p-4 text-xs text-muted-foreground"><span>Showing {filtered.length} users</span><div className="flex gap-1"><Button size="sm">1</Button><Button size="sm" variant="ghost">2</Button><Button size="sm" variant="ghost">3</Button></div></div></section>
  </>;
}
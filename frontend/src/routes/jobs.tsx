import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Clock, MapPin, SlidersHorizontal } from "lucide-react";
import { PublicShell } from "@/components/jobira/shell";
import { JobCard, PageIntro, SearchBar } from "@/components/jobira/shared";
import { Button } from "@/components/ui/button";

const jobs = [
  { title: "Office Cleaning Professional", company: "North & Co. Workspaces", location: "Manchester", pay: "£16–£19 / hour", type: "Part-time", description: "Join a trusted facilities team looking after two premium city-centre offices. Consistent evening hours and all supplies provided." },
  { title: "Housekeeping Specialist", company: "The Harrington Collection", location: "London", pay: "£145 / day", type: "Contract", description: "Support a small portfolio of serviced apartments with flexible weekday scheduling and repeat bookings." },
  { title: "End of Tenancy Cleaner", company: "MoveWell Property", location: "Birmingham", pay: "£180 / job", type: "Flexible", description: "Experienced detail-focused cleaner wanted for regular end-of-tenancy appointments across central Birmingham." },
  { title: "Domestic Cleaning Partner", company: "Kindred Homes", location: "Leeds", pay: "£15 / hour", type: "Part-time", description: "Build a consistent local schedule with friendly clients and simple weekly bookings." },
];

export const Route = createFileRoute("/jobs")({
  head: () => ({ meta: [{ title: "Find Work | Jobira" }, { name: "description", content: "Browse trusted jobs and flexible work opportunities on Jobira." }, { property: "og:title", content: "Find Work | Jobira" }, { property: "og:description", content: "Browse trusted jobs and flexible work opportunities on Jobira." }, { property: "og:type", content: "website" }, { name: "twitter:card", content: "summary_large_image" }] }),
  component: JobsPage,
});

function JobsPage() {
  const [selected, setSelected] = useState(0);
  const job = jobs[selected] ?? jobs[0];
  return <PublicShell><section className="border-b border-border bg-secondary py-16"><div className="page-wrap"><PageIntro kicker="Find work" title="Good work starts with a fair opportunity." copy="Explore flexible, trusted roles from businesses that value what you do." /><SearchBar placeholder="Job title, skill or company" buttonText="Find jobs" /></div></section><section className="py-10"><div className="page-wrap"><div className="mb-6 flex items-center justify-between"><p className="text-sm font-semibold">128 opportunities near you</p><Button variant="outline" size="sm"><SlidersHorizontal /> Filters</Button></div><div className="grid gap-6 lg:grid-cols-[minmax(0,0.9fr)_minmax(360px,1.1fr)]"><div className="space-y-3">{jobs.map((item, index) => <JobCard key={item.title} {...item} selected={selected === index} onClick={() => setSelected(index)} />)}</div>{job && <aside className="h-fit rounded-md border border-border bg-card p-6 lg:sticky lg:top-24"><p className="section-kicker">Featured opportunity</p><h2 className="mt-4 font-display text-2xl font-semibold">{job.title}</h2><p className="mt-2 text-muted-foreground">{job.company}</p><div className="mt-6 flex flex-wrap gap-5 border-y border-border py-4 text-sm"><span className="flex items-center gap-2"><MapPin className="h-4 w-4 text-brand-leaf" />{job.location}</span><span className="flex items-center gap-2"><Clock className="h-4 w-4 text-brand-leaf" />{job.type}</span><strong>{job.pay}</strong></div><h3 className="mt-6 font-display font-semibold">About the opportunity</h3><p className="mt-3 text-sm leading-7 text-muted-foreground">{job.description}</p><h3 className="mt-6 font-display font-semibold">What you’ll need</h3><ul className="mt-3 space-y-2 text-sm text-muted-foreground"><li>• A reliable, professional approach</li><li>• Recent relevant experience</li><li>• Eligibility to work in the UK</li></ul><Button className="mt-7 w-full">Apply for this role</Button></aside>}</div></div></section></PublicShell>;
}
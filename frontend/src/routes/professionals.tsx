import { createFileRoute } from "@tanstack/react-router";
import { PublicShell } from "@/components/jobira/shell";
import { PageIntro, ProfessionalCard, SearchBar, EmptyCallout } from "@/components/jobira/shared";

const people = [
  { initials: "AM", name: "Amara Mensah", role: "Commercial Cleaning Specialist", location: "Manchester", rating: "4.9", skills: ["Offices", "Deep cleaning", "8 years"] },
  { initials: "SO", name: "Sofia Oliveira", role: "Housekeeping Professional", location: "London", rating: "4.8", skills: ["Hospitality", "Turnovers", "Eco-friendly"] },
  { initials: "JL", name: "Jordan Lewis", role: "Facilities Team Lead", location: "Birmingham", rating: "5.0", skills: ["Team management", "Commercial", "COSHH"] },
  { initials: "RN", name: "Rina Nair", role: "Domestic Cleaning Partner", location: "Leeds", rating: "4.9", skills: ["Homes", "Weekly service", "References"] },
  { initials: "DK", name: "Daniel King", role: "End of Tenancy Specialist", location: "Bristol", rating: "4.7", skills: ["Move-outs", "Carpets", "Inventory"] },
  { initials: "FB", name: "Fatima Bello", role: "Hospitality Housekeeper", location: "Liverpool", rating: "4.9", skills: ["Hotels", "Linen", "Fast turnaround"] },
];

export const Route = createFileRoute("/professionals")({
  head: () => ({ meta: [{ title: "Find Professionals | Jobira" }, { name: "description", content: "Discover trusted, experienced professionals ready to work." }, { property: "og:title", content: "Find Professionals | Jobira" }, { property: "og:description", content: "Discover trusted, experienced professionals ready to work." }, { property: "og:type", content: "website" }, { name: "twitter:card", content: "summary_large_image" }] }), component: ProfessionalsPage,
});

function ProfessionalsPage() { return <PublicShell><section className="border-b border-border bg-secondary py-16"><div className="page-wrap"><PageIntro kicker="Find professionals" title="Trusted people, ready to help." copy="Discover experienced professionals with clear skills, real reviews and availability that suits your business." /><SearchBar placeholder="Skill, service or location" buttonText="Search people" /></div></section><section className="section-pad"><div className="page-wrap"><div className="mb-8 flex items-end justify-between"><div><p className="section-kicker">Recommended near you</p><h2 className="mt-2 font-display text-2xl font-semibold">Professionals with proven reputations</h2></div><p className="hidden text-sm text-muted-foreground sm:block">Showing 6 of 2,480</p></div><div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">{people.map((person) => <ProfessionalCard key={person.name} {...person} />)}</div></div></section><EmptyCallout title="Need to hire a whole team?" copy="Post your requirements once and connect with suitable professionals." to="/post-job" action="Post a job" /></PublicShell>; }
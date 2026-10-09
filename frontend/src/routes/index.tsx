import { createFileRoute, Link } from "@tanstack/react-router";
import { ArrowDown, ArrowRight, CheckCircle2, Search, UserPlus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { PublicShell } from "@/components/jobira/shell";
import heroImage from "@/assets/jobira-hero.jpg";
import pricingImage from "@/assets/jobira-pricing.jpg";
import type { LucideIcon } from "lucide-react";

export const Route = createFileRoute("/")({
  head: () => ({ meta: [
    { title: "Jobira | Work Meets Opportunity" },
    { name: "description", content: "Find meaningful work and trusted professionals with Jobira." },
    { property: "og:title", content: "Jobira | Work Meets Opportunity" },
    { property: "og:description", content: "Find meaningful work and trusted professionals with Jobira." },
    { property: "og:type", content: "website" },
    { name: "twitter:card", content: "summary_large_image" },
  ] }),
  component: Index,
});

// IMPORTANT: Replace this placeholder. See ./README.md for routing conventions.
function Index() {
  return (
    <PublicShell>
      <section className="relative min-h-[calc(100vh-76px)] overflow-hidden">
        <img src={heroImage} width={1600} height={1000} alt="A calm, sunlit interior opening onto green trees" className="absolute inset-0 h-full w-full object-cover object-[64%_center] dark:brightness-[.38]" />
        <div className="absolute inset-0 bg-[linear-gradient(90deg,var(--background)_0%,color-mix(in_oklab,var(--background)_92%,transparent)_34%,transparent_70%)]" />
        <div className="page-wrap relative z-10 flex min-h-[calc(100vh-76px)] flex-col justify-center py-20">
          <div className="reveal max-w-2xl"><p className="section-kicker">A better way forward</p><h1 className="mt-6 font-display text-5xl font-medium leading-[1.02] sm:text-6xl lg:text-[5rem]">Work meets<br /><em className="font-normal text-brand-leaf">opportunity.</em></h1><p className="mt-7 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg">Jobira connects people, businesses and professionals to create real opportunities that last.</p><div className="mt-9 flex flex-wrap gap-3"><Button size="lg" asChild><Link to="/jobs">Find Work <ArrowRight /></Link></Button><Button size="lg" variant="outline" asChild><Link to="/professionals">Find a Professional</Link></Button></div></div>
          <a href="#how-it-works" className="absolute bottom-8 left-5 flex items-center gap-2 text-[0.68rem] font-bold uppercase tracking-[0.16em] text-muted-foreground sm:left-[max(1.25rem,calc((100vw-1240px)/2))]">Explore Jobira <ArrowDown className="h-3.5 w-3.5" /></a>
        </div>
      </section>
      <section id="how-it-works" className="section-pad bg-background"><div className="page-wrap"><p className="section-kicker">How Jobira Works</p><div className="mt-5 grid gap-8 lg:grid-cols-[1fr_1fr]"><h2 className="max-w-lg font-display text-4xl font-medium leading-tight sm:text-5xl">Real connections.<br />Three simple steps.</h2><p className="max-w-xl self-end text-base leading-7 text-muted-foreground">We take the friction out of finding work and hiring trusted people. Clear profiles, transparent reputations, better matches.</p></div><div className="mt-16 grid border-t border-border md:grid-cols-3">{([
        ["01", "Create an account", "Tell us what you do or who you need. A clear profile helps the right people find you.", UserPlus],
        ["02", "Find the right people", "Search trusted opportunities or discover skilled professionals near you.", Search],
        ["03", "Make it happen", "Connect directly, agree the details and build a reputation through real work.", CheckCircle2],
      ] as [string, string, string, LucideIcon][]).map(([number, title, copy, Icon], index) => <article key={number} className={`py-8 md:px-8 ${index > 0 ? "md:border-l md:border-border" : "md:pl-0"}`}><div className="flex items-center justify-between"><span className="font-display text-sm text-brand-leaf">{number}</span><Icon className="h-5 w-5 text-muted-foreground" /></div><h3 className="mt-14 font-display text-xl font-semibold">{title}</h3><p className="mt-3 text-sm leading-6 text-muted-foreground">{copy}</p></article>)}</div></div></section>
      <section className="relative min-h-[520px] overflow-hidden"><img src={pricingImage} loading="lazy" width={1600} height={720} alt="Sunlight across a sage green architectural wall" className="absolute inset-0 h-full w-full object-cover dark:brightness-[.42]" /><div className="absolute inset-0 bg-[linear-gradient(90deg,color-mix(in_oklab,var(--background)_88%,transparent),transparent_75%)]" /><div className="page-wrap relative flex min-h-[520px] items-center"><div className="max-w-lg"><p className="section-kicker">Simple, transparent pricing</p><h2 className="mt-5 font-display text-4xl font-medium leading-tight sm:text-5xl">Start free.<br />Grow with confidence.</h2><p className="mt-5 text-muted-foreground">Straightforward plans for people and companies ready to move forward.</p><Button className="mt-7" variant="outline" asChild><Link to="/pricing">View Pricing <ArrowRight /></Link></Button></div></div></section>
    </PublicShell>
  );
}

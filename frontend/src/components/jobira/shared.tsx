import { ArrowRight, Check, MapPin, Star } from "lucide-react";
import { Link } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export function PageIntro({ kicker, title, copy, centered = false }: { kicker: string; title: string; copy: string; centered?: boolean }) {
  return <div className={cn("max-w-3xl", centered && "mx-auto text-center")}><p className="section-kicker">{kicker}</p><h1 className="mt-5 font-display text-4xl font-medium leading-[1.08] sm:text-5xl lg:text-6xl">{title}</h1><p className={cn("mt-5 max-w-2xl text-base leading-7 text-muted-foreground sm:text-lg", centered && "mx-auto")}>{copy}</p></div>;
}

export function SearchBar({ placeholder, buttonText = "Search" }: { placeholder: string; buttonText?: string }) {
  return <form className="mt-8 flex max-w-2xl flex-col gap-2 sm:flex-row" onSubmit={(event) => event.preventDefault()}><label className="sr-only" htmlFor="site-search">Search</label><input id="site-search" className="h-12 min-w-0 flex-1 rounded-md border border-input bg-card px-4 text-sm outline-none ring-ring focus:ring-2" placeholder={placeholder} /><Button size="lg">{buttonText}<ArrowRight /></Button></form>;
}

export function JobCard({ title, company, location, pay, selected, onClick }: { title: string; company: string; location: string; pay: string; selected?: boolean; onClick?: () => void }) {
  return <button onClick={onClick} className={cn("w-full rounded-md border bg-card p-5 text-left transition-colors hover:border-primary", selected ? "border-primary" : "border-border")}><div className="flex items-start justify-between gap-4"><div><h3 className="font-display text-base font-semibold">{title}</h3><p className="mt-1 text-sm text-muted-foreground">{company}</p></div><span className="rounded-sm bg-secondary px-2 py-1 text-[0.67rem] font-bold uppercase tracking-wider text-secondary-foreground">New</span></div><div className="mt-5 flex flex-wrap gap-4 text-xs text-muted-foreground"><span className="flex items-center gap-1.5"><MapPin className="h-3.5 w-3.5" />{location}</span><strong className="text-foreground">{pay}</strong></div></button>;
}

export function ProfessionalCard({ initials, name, role, location, rating, skills }: { initials: string; name: string; role: string; location: string; rating: string; skills: string[] }) {
  return <article className="rounded-md border border-border bg-card p-5"><div className="flex gap-4"><div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-soft-green font-display font-semibold text-secondary-foreground">{initials}</div><div><h2 className="font-display font-semibold">{name}</h2><p className="text-sm text-muted-foreground">{role}</p><div className="mt-2 flex items-center gap-3 text-xs text-muted-foreground"><span>{location}</span><span className="flex items-center gap-1 font-semibold text-foreground"><Star className="h-3.5 w-3.5 fill-current text-brand-leaf" />{rating}</span></div></div></div><div className="mt-5 flex flex-wrap gap-2">{skills.map((skill) => <span key={skill} className="rounded-sm border border-border px-2 py-1 text-xs text-muted-foreground">{skill}</span>)}</div><Button variant="outline" className="mt-5 w-full">View profile</Button></article>;
}

export function PricingCard({ name, price, description, features, popular }: { name: string; price: string; description: string; features: string[]; popular?: boolean | undefined }) {
  return <article className={cn("relative flex min-h-[460px] flex-col border bg-card p-6", popular ? "border-primary" : "border-border")}>
    {popular && <span className="absolute right-0 top-0 bg-primary px-3 py-1.5 text-[0.65rem] font-extrabold uppercase tracking-wider text-primary-foreground">Most Popular</span>}
    <p className="section-kicker">{name}</p><div className="mt-6 flex items-end gap-1"><span className="font-display text-4xl font-semibold">{price}</span><span className="pb-1 text-sm text-muted-foreground">/month</span></div><p className="mt-4 min-h-12 text-sm leading-6 text-muted-foreground">{description}</p><div className="editorial-rule my-6" /><ul className="space-y-3">{features.map((item) => <li key={item} className="flex gap-2.5 text-sm"><Check className="mt-0.5 h-4 w-4 shrink-0 text-brand-leaf" />{item}</li>)}</ul><Button className="mt-auto" variant={popular ? "default" : "outline"}>{name === "Free" ? "Start free" : `Choose ${name}`}</Button>
  </article>;
}

export function EmptyCallout({ title, copy, to, action }: { title: string; copy: string; to: "/jobs" | "/post-job" | "/professionals"; action: string }) {
  return <section className="border-y border-border bg-secondary py-16"><div className="page-wrap flex flex-col justify-between gap-6 md:flex-row md:items-center"><div><h2 className="font-display text-2xl font-semibold">{title}</h2><p className="mt-2 text-sm text-muted-foreground">{copy}</p></div><Button asChild><Link to={to}>{action}<ArrowRight /></Link></Button></div></section>;
}
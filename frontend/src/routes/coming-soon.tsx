import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { ArrowRight, Check, ChevronDown, Mail } from "lucide-react";
import { Logo } from "@/components/jobira/logo";
import { ThemeButton } from "@/components/jobira/shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

export const Route = createFileRoute("/coming-soon")({
  head: () => ({ meta: [{ title: "Coming Soon | Jobira" }, { name: "description", content: "A fairer way to find work and grow your future. Join the Jobira waitlist and be the first to know when we launch." }, { property: "og:title", content: "Coming Soon | Jobira" }, { property: "og:description", content: "A fairer way to find work and grow your future. Join the Jobira waitlist." }, { property: "og:type", content: "website" }, { name: "twitter:card", content: "summary_large_image" }] }), component: ComingSoonPage,
});

function CornerArcs() {
  return (
    <svg aria-hidden="true" viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice" className="pointer-events-none absolute inset-0 h-full w-full">
      {/* top-left leaf arc */}
      <path d="M0 240 Q0 60 260 0 L0 0 Z" fill="url(#arc-green)" opacity="0.5" />
      <path d="M0 260 Q120 200 210 40" fill="none" stroke="url(#line-blue)" strokeWidth="1.2" opacity="0.6" />
      {/* bottom-right leaf arc */}
      <path d="M1440 620 Q1440 860 1120 900 L1440 900 Z" fill="url(#arc-blue)" opacity="0.5" />
      <path d="M1440 640 Q1320 720 1180 900" fill="none" stroke="url(#line-green)" strokeWidth="1.2" opacity="0.6" />
      <defs>
        <linearGradient id="arc-green" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stopColor="var(--brand-leaf)" /><stop offset="1" stopColor="var(--brand-blue)" stopOpacity="0.35" /></linearGradient>
        <linearGradient id="arc-blue" x1="1" y1="1" x2="0" y2="0"><stop offset="0" stopColor="var(--brand-blue)" /><stop offset="1" stopColor="var(--brand-leaf)" stopOpacity="0.4" /></linearGradient>
        <linearGradient id="line-blue" x1="0" y1="0" x2="1" y2="0"><stop offset="0" stopColor="var(--brand-blue)" /><stop offset="1" stopColor="var(--brand-leaf)" /></linearGradient>
        <linearGradient id="line-green" x1="1" y1="0" x2="0" y2="1"><stop offset="0" stopColor="var(--brand-leaf)" /><stop offset="1" stopColor="var(--brand-blue)" /></linearGradient>
      </defs>
    </svg>
  );
}

function ComingSoonPage() {
  const [email, setEmail] = useState("");
  const [joined, setJoined] = useState(false);
  return (
    <main className="relative min-h-screen overflow-hidden bg-background">
      <CornerArcs />
      <header className="page-wrap relative z-10 flex h-24 items-center justify-between">
        <Logo />
        <ThemeButton />
      </header>
      <section className="relative z-10 flex min-h-[calc(100vh-96px)] flex-col items-center justify-center px-5 pb-20 text-center">
        <p className="text-xs font-semibold uppercase tracking-[0.35em] text-muted-foreground">Coming soon</p>
        <h1 className="mt-6 font-display text-5xl font-semibold leading-[1.06] tracking-tight sm:text-6xl lg:text-7xl">
          Real opportunities
          <br />
          <span className="bg-[linear-gradient(90deg,var(--brand-leaf),var(--brand-blue))] bg-clip-text text-transparent">for real people.</span>
        </h1>
        <p className="mt-7 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg">
          A fairer way to find work and grow your future.
          <br />
          Be the first to know when we launch.
        </p>
        {joined ? (
          <div className="mt-9 flex w-full max-w-xl items-center justify-center gap-3 rounded-lg border border-primary bg-secondary p-4 text-sm font-semibold">
            <Check className="h-5 w-5 text-brand-leaf" />
            You’re on the list. We’ll be in touch.
          </div>
        ) : (
          <form className="mt-9 flex w-full max-w-xl items-center gap-2 rounded-lg border border-border bg-card p-1.5 shadow-sm" onSubmit={(event) => { event.preventDefault(); if (email.includes("@")) setJoined(true); }}>
            <Mail className="ml-3 h-5 w-5 shrink-0 text-muted-foreground" aria-hidden="true" />
            <label className="sr-only" htmlFor="waitlist-email">Email address</label>
            <Input id="waitlist-email" type="email" required value={email} onChange={(event) => setEmail(event.target.value)} placeholder="Enter your email" className="h-11 flex-1 border-0 bg-transparent shadow-none focus-visible:ring-0" />
            <Button type="submit" size="lg" className="shrink-0">Join Waitlist <ArrowRight /></Button>
          </form>
        )}
        <p className="mt-4 text-xs text-muted-foreground">No spam. Just important updates.</p>
        <div className="mt-14 flex flex-col items-center gap-5">
          <span className="h-10 w-px bg-border" aria-hidden="true" />
          <span className="text-[0.65rem] font-bold uppercase tracking-[0.3em] text-muted-foreground">More opportunities ahead</span>
          <ChevronDown className="h-4 w-4 animate-bounce text-muted-foreground" aria-hidden="true" />
        </div>
      </section>
    </main>
  );
}

import { Link } from "@tanstack/react-router";
import { cn } from "@/lib/utils";
import logoDark from "@/assets/jobira_logo_dark.png";
import logoWhite from "@/assets/jobira_logo_white.png";

interface LogoProps {
  className?: string;
  compact?: boolean;
  to?: string;
}

export function Logo({ className, compact = false, to = "/" }: LogoProps) {
  const content = (
    <>
      {/* Light theme logo (Dark text) */}
      <img
        src={logoDark}
        alt="Jobira Logo"
        className={cn(
          "w-auto object-contain dark:hidden",
          compact ? "h-8" : "h-10",
          className
        )}
      />
      {/* Dark theme logo (White text) */}
      <img
        src={logoWhite}
        alt="Jobira Logo"
        className={cn(
          "hidden w-auto object-contain dark:block",
          compact ? "h-8" : "h-10",
          className
        )}
      />
    </>
  );

  if (!to) {
    return <span className="inline-flex items-center">{content}</span>;
  }

  return (
    <Link to={to} className="inline-flex items-center" aria-label="Jobira home">
      {content}
    </Link>
  );
}
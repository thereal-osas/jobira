import { createFileRoute, Outlet } from "@tanstack/react-router";
import { BriefcaseBusiness, CalendarDays, CreditCard, Heart, Home, MapPin, PlusCircle, Settings, Star } from "lucide-react";
import { PortalShell, type PortalNavItem } from "@/components/jobira/portal-shell";

const clientNavigation: PortalNavItem[] = [
  { label: "Home", to: "/client", icon: Home },
  { label: "Post a job", to: "/client/post-job", icon: PlusCircle },
  { label: "My jobs", to: "/client/jobs", icon: BriefcaseBusiness },
  { label: "My bookings", to: "/client/bookings", icon: CalendarDays },
  { label: "Favourites", to: "/client/favourites", icon: Heart },
  { label: "Payments", to: "/client/payments", icon: CreditCard },
  { label: "Reviews", to: "/client/reviews", icon: Star },
  { label: "Saved locations", to: "/client/locations", icon: MapPin },
  { label: "Settings", to: "/client/settings", icon: Settings },
];

export const Route = createFileRoute("/client")({ component: ClientLayout });

function ClientLayout() {
  return <PortalShell navigation={clientNavigation} name="Sarah M." role="Client" initials="SC" context="Client workspace"><Outlet /></PortalShell>;
}
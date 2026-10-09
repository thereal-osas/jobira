import { createFileRoute, Outlet } from "@tanstack/react-router";
import { BarChart3, BriefcaseBusiness, CalendarDays, ClipboardList, CreditCard, LayoutDashboard, MessageCircle, Settings, ShieldCheck, Star, Users } from "lucide-react";
import { PortalShell, type PortalNavItem } from "@/components/jobira/portal-shell";

const adminNavigation: PortalNavItem[] = [
  { label: "Dashboard", to: "/admin", icon: LayoutDashboard },
  { label: "Users", to: "/admin/users", icon: Users },
  { label: "Jobs", to: "/admin/jobs", icon: BriefcaseBusiness },
  { label: "Applications", to: "/admin/applications", icon: ClipboardList },
  { label: "Bookings", to: "/admin/bookings", icon: CalendarDays },
  { label: "Reputation", to: "/admin/reputation", icon: Star },
  { label: "Payments", to: "/admin/payments", icon: CreditCard },
  { label: "Verification", to: "/admin/verification", icon: ShieldCheck },
  { label: "Reports", to: "/admin/reports", icon: BarChart3 },
  { label: "Messages", to: "/admin/messages", icon: MessageCircle },
  { label: "Settings", to: "/admin/settings", icon: Settings },
];

export const Route = createFileRoute("/admin")({ component: AdminLayout });

function AdminLayout() {
  return <PortalShell navigation={adminNavigation} name="Admin User" role="Administrator" initials="AU" context="Platform control"><Outlet /></PortalShell>;
}
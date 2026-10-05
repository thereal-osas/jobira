import { createFileRoute } from "@tanstack/react-router";
import { AdminJobsPage } from "@/components/jobira/admin-extra-pages";

export const Route = createFileRoute("/admin/jobs")({ head: () => ({ meta: [{ title: "Jobs | Jobira Admin" }, { name: "description", content: "Manage every job listing across Jobira." }, { property: "og:title", content: "Jobs | Jobira Admin" }, { property: "og:description", content: "Manage every job listing across Jobira." }, { property: "og:type", content: "website" }, { name: "twitter:card", content: "summary_large_image" }] }), component: AdminJobsPage });
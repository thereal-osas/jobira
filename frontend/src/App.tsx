import type { QueryClient } from "@tanstack/react-query";
import { QueryClientProvider } from "@tanstack/react-query";
import { Outlet } from "@tanstack/react-router";

import { ThemeProvider } from "@/components/jobira/theme";
import { ScrollMotion } from "@/components/jobira/scroll-motion";

export interface AppProps {
  queryClient?: QueryClient;
}

export default function App({ queryClient }: AppProps) {
  const content = (
    <ThemeProvider>
      <ScrollMotion />
      <Outlet />
    </ThemeProvider>
  );

  if (queryClient) {
    return (
      <QueryClientProvider client={queryClient}>
        {content}
      </QueryClientProvider>
    );
  }

  return content;
}
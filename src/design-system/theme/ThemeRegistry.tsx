"use client";

import { AppRouterCacheProvider } from "@mui/material-nextjs/v16-appRouter";
import { ThemeProvider } from "@mui/material/styles";
import CssBaseline from "@mui/material/CssBaseline";
import { createAppTheme } from "@/design-system/theme/theme";
import type { SiteSettings } from "@/features/settings/types";

export default function ThemeRegistry({
  children,
  settings,
}: {
  children: React.ReactNode;
  settings: SiteSettings;
}) {
  const theme = createAppTheme(settings);

  return (
    <AppRouterCacheProvider options={{ enableCssLayer: true }}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </AppRouterCacheProvider>
  );
}

"use client";

import { createTheme, darken, getContrastRatio, lighten } from "@mui/material/styles";
import type { SiteSettings } from "@/features/settings/types";

export function createAppTheme(settings: SiteSettings) {
  const secondary = settings.secondary_color || "#1D4E89";
  const contrastText = getContrastRatio(secondary, "#fff") >= 4.5 ? "#fff" : "#211D19";
  const linkColor = getContrastRatio(secondary, "#F8F9FA") >= 4.5
    ? secondary
    : darken(secondary, 0.25);

  return createTheme({
    cssVariables: true,
    palette: {
      mode: "light",
      primary: {
        main: "#211D19",
        light: "#4D4741",
        dark: "#0F0D0B",
        contrastText: "#F8F9FA",
      },
      secondary: {
        main: secondary,
        light: lighten(secondary, 0.18),
        dark: darken(secondary, 0.2),
        contrastText,
      },
      background: {
        default: "#F8F9FA",
        paper: "#FBF9F5",
      },
      text: {
        primary: "#211D19",
        secondary: "#655E55",
      },
    },
    typography: {
      fontFamily: [
        "var(--font-geist-sans)",
        "-apple-system",
        "BlinkMacSystemFont",
        '"Segoe UI"',
        "Roboto",
        '"Helvetica Neue"',
        "Arial",
        "sans-serif",
      ].join(","),
      body1: { lineHeight: 1.65 },
      h1: { fontWeight: 700, letterSpacing: "-0.035em" },
      h2: { fontWeight: 700, letterSpacing: "-0.03em" },
      h3: { fontWeight: 700, letterSpacing: "-0.025em" },
      h4: { fontWeight: 700, letterSpacing: "-0.02em" },
      h5: { fontWeight: 600 },
      h6: { fontWeight: 600 },
    },
    shape: { borderRadius: 18 },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: { backgroundColor: "#F8F9FA" },
          a: {
            color: linkColor,
            transition: "color 160ms ease",
            "&:hover": { color: secondary },
            "&:focus-visible": { outline: `3px solid ${secondary}`, outlineOffset: 3 },
          },
          "::selection": { backgroundColor: secondary, color: contrastText },
        },
      },
      MuiButton: {
        defaultProps: { disableElevation: true },
        styleOverrides: {
          root: {
            minHeight: 44,
            borderRadius: 999,
            textTransform: "none",
            fontWeight: 600,
            "&:focus-visible": { outline: `3px solid ${secondary}`, outlineOffset: 2 },
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: { backgroundColor: "#FBF9F5", border: "1px solid rgba(80, 65, 45, 0.12)", boxShadow: "none" },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: { backgroundImage: "none" },
        },
      },
      MuiTextField: {
        defaultProps: { variant: "outlined", fullWidth: true },
      },
      MuiTab: {
        styleOverrides: {
          root: {
            minHeight: 48,
            textTransform: "none",
            fontWeight: 600,
            "&:focus-visible": { outline: `3px solid ${secondary}`, outlineOffset: -3 },
          },
        },
      },
    },
  });
}

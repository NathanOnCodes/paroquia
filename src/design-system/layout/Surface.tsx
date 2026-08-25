import Box from "@mui/material/Box";

export default function Surface({
  children,
  component = "section",
  sx,
}: {
  children: React.ReactNode;
  component?: "section" | "div" | "article";
  sx?: object;
}) {
  return (
    <Box component={component} sx={{ backgroundColor: "background.paper", border: 1, borderColor: "divider", borderRadius: 4, p: { xs: 2, sm: 3 }, ...sx }}>
      {children}
    </Box>
  );
}

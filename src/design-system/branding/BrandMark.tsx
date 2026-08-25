import Box from "@mui/material/Box";

export default function BrandMark() {
  return (
    <Box component="svg" viewBox="0 0 64 64" aria-hidden="true" sx={{ width: 28, height: 28, flexShrink: 0, color: "secondary.main" }}>
      <path
        fill="currentColor"
        d="M29 7h6v8h8v6h-8v8h-6v-8h-8v-6h8V7Zm-3 26h12v24H26V33Zm-17 24h46v4H9v-4Zm5-4V31l18-14 18 14v22h-7V35H21v18h-7Z"
      />
    </Box>
  );
}

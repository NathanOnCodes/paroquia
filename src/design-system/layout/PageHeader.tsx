import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

export default function PageHeader({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: React.ReactNode;
}) {
  return (
    <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 2, flexWrap: "wrap", mb: 4 }}>
      <Box sx={{ maxWidth: 720 }}>
        <Typography variant="h4" component="h1" sx={{ fontSize: { xs: "1.5rem", sm: "2.125rem" } }}>{title}</Typography>
        {description && <Typography color="text.secondary" sx={{ mt: 1 }}>{description}</Typography>}
      </Box>
      {action}
    </Box>
  );
}

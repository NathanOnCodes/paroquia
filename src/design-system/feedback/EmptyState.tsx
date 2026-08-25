import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";

export default function EmptyState({
  title,
  description,
}: {
  title: string;
  description?: string;
}) {
  return (
    <Box sx={{ py: 5, textAlign: "center" }}>
      <Typography variant="h6">{title}</Typography>
      {description && <Typography color="text.secondary" sx={{ mt: 0.5 }}>{description}</Typography>}
    </Box>
  );
}

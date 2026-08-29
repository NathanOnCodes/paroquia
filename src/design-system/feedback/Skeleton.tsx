import Box from "@mui/material/Box";
import MuiSkeleton from "@mui/material/Skeleton";
import Paper from "@mui/material/Paper";
import Grid from "@mui/material/Grid";

export function CardSkeleton({ count = 1 }: { count?: number }) {
  return (
    <Grid container spacing={2}>
      {Array.from({ length: count }).map((_, i) => (
        <Grid size={{ xs: 12, sm: 6, md: 4 }} key={i}>
          <Paper sx={{ p: 2 }}>
            <MuiSkeleton variant="rectangular" height={140} sx={{ borderRadius: 1, mb: 1 }} />
            <MuiSkeleton variant="text" width="70%" />
            <MuiSkeleton variant="text" width="40%" />
          </Paper>
        </Grid>
      ))}
    </Grid>
  );
}

export function TableSkeleton({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <Paper sx={{ overflowX: "auto" }}>
      <Box sx={{ minWidth: 400 }}>
        <Box sx={{ display: "flex", p: 2, gap: 2 }}>
          {Array.from({ length: cols }).map((_, i) => (
            <MuiSkeleton key={i} variant="text" sx={{ flex: i === 0 ? 2 : 1 }} />
          ))}
        </Box>
        {Array.from({ length: rows }).map((_, row) => (
          <Box key={row} sx={{ display: "flex", p: 2, gap: 2, borderTop: 1, borderColor: "divider" }}>
            {Array.from({ length: cols }).map((_, col) => (
              <MuiSkeleton key={col} variant="text" sx={{ flex: col === 0 ? 2 : 1 }} />
            ))}
          </Box>
        ))}
      </Box>
    </Paper>
  );
}

export function FormSkeleton({ fields = 4 }: { fields?: number }) {
  return (
    <Paper sx={{ p: 3 }}>
      <Box sx={{ display: "flex", flexDirection: "column", gap: 2.5 }}>
        {Array.from({ length: fields }).map((_, i) => (
          <Box key={i}>
            <MuiSkeleton variant="text" width={80} sx={{ mb: 0.5 }} />
            <MuiSkeleton variant="rectangular" height={56} sx={{ borderRadius: 1 }} />
          </Box>
        ))}
        <MuiSkeleton variant="rectangular" height={44} sx={{ borderRadius: 999, mt: 1 }} />
      </Box>
    </Paper>
  );
}

export function PageSkeleton() {
  return (
    <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
      <MuiSkeleton variant="text" width={200} height={40} />
      <MuiSkeleton variant="text" width={300} />
      <CardSkeleton count={3} />
    </Box>
  );
}

"use client";

import { useState } from "react";
import Box from "@mui/material/Box";

export default function EventImage({ src, alt }: { src: string; alt: string }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <Box
      component="img"
      src={src}
      alt={alt}
      onClick={() => setExpanded(!expanded)}
      sx={{
        display: "block",
        width: expanded ? "100%" : { xs: "85%", sm: "70%" },
        mx: "auto",
        borderRadius: 2,
        mb: 2,
        cursor: "pointer",
        transition: "width 0.2s ease",
      }}
    />
  );
}

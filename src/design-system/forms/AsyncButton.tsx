"use client";

import Button, { type ButtonProps } from "@mui/material/Button";
import CircularProgress from "@mui/material/CircularProgress";

export default function AsyncButton({ loading, children, ...props }: ButtonProps & { loading?: boolean }) {
  return (
    <Button {...props} disabled={loading || props.disabled}>
      {loading ? <CircularProgress size={20} color="inherit" /> : children}
    </Button>
  );
}

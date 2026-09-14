"use client";

import type { ReactNode } from "react";

interface ActionFormProps {
  action: (formData: FormData) => Promise<unknown>;
  children: ReactNode;
}

export default function ActionForm({ action, children }: ActionFormProps) {
  return (
    <form
      action={async (formData) => {
        await action(formData);
      }}
    >
      {children}
    </form>
  );
}

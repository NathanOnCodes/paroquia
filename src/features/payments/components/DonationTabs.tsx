"use client";

import { useState } from "react";
import Box from "@mui/material/Box";
import Tab from "@mui/material/Tab";
import Tabs from "@mui/material/Tabs";

function tabProps(index: number) {
  return { id: `doacao-tab-${index}`, "aria-controls": `doacao-tabpanel-${index}` };
}

export default function DonationTabs({
  pixPanel,
  recurringPanel,
}: {
  pixPanel: React.ReactNode;
  recurringPanel: React.ReactNode;
}) {
  const [value, setValue] = useState(0);

  return (
    <Box>
      <Tabs
        value={value}
        onChange={(_, nextValue) => setValue(nextValue)}
        centered
        variant="fullWidth"
        aria-label="Tipos de doação"
        sx={{ maxWidth: 560, mx: "auto", mb: 3 }}
      >
        <Tab label="Doação" {...tabProps(0)} />
        <Tab label="Doação recorrente" {...tabProps(1)} />
      </Tabs>
      <Box role="tabpanel" hidden={value !== 0} id="doacao-tabpanel-0" aria-labelledby="doacao-tab-0" tabIndex={0}>
        {value === 0 && pixPanel}
      </Box>
      <Box role="tabpanel" hidden={value !== 1} id="doacao-tabpanel-1" aria-labelledby="doacao-tab-1" tabIndex={0}>
        {value === 1 && recurringPanel}
      </Box>
    </Box>
  );
}

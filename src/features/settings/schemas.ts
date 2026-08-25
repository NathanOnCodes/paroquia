import { z } from "zod";

const optionalText = (max: number) =>
  z.string().trim().max(max).optional().default("");

export const siteSettingsSchema = z.object({
  site_name: z.string().trim().min(2, "Informe o nome da paróquia").max(120),
  site_title: z.string().trim().min(2, "Informe o título do site").max(120),
  city: optionalText(120),
  description: z.string().trim().min(2).max(240),
  secondary_color: z.string().trim().regex(/^#[0-9A-Fa-f]{6}$/, "Use uma cor hexadecimal válida"),
  pix_key: optionalText(254),
  pix_receiver_name: optionalText(25),
  pix_receiver_city: optionalText(15),
  home_title: optionalText(180),
  home_description: optionalText(400),
});

export type SiteSettingsInput = z.infer<typeof siteSettingsSchema>;

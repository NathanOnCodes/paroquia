export interface SiteSettings {
  id: number;
  site_name: string;
  site_title: string;
  city: string;
  description: string;
  secondary_color: string;
  pix_key: string | null;
  pix_receiver_name: string | null;
  pix_receiver_city: string | null;
  home_title: string | null;
  home_description: string | null;
  created_at: string;
  updated_at: string;
}

export const DEFAULT_SITE_SETTINGS: SiteSettings = {
  id: 1,
  site_name: "Paróquia",
  site_title: "Paróquia",
  city: "",
  description: "Comunidade, fé e acolhimento.",
  secondary_color: "#1D4E89",
  pix_key: null,
  pix_receiver_name: null,
  pix_receiver_city: null,
  home_title: null,
  home_description: null,
  created_at: "",
  updated_at: "",
};

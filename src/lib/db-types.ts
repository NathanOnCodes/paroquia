export type Role = "admin" | "assistente";

export interface Profile {
  id: string;
  full_name: string;
  email: string;
  role: Role;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Community {
  id: string;
  name: string;
  description: string | null;
  city: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Donor {
  id: string;
  full_name: string;
  email: string;
  phone: string | null;
  city: string | null;
  neighborhood: string | null;
  community_id: string | null;
  community_name: string | null;
  privacy_consent_at: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export type RecurringStatus =
  | "active"
  | "past_due"
  | "cancel_at_period_end"
  | "canceled"
  | "incomplete";

export interface RecurringDonation {
  id: string;
  donor_id: string;
  community_id: string | null;
  amount_cents: number;
  currency: string;
  interval: "month";
  stripe_customer_id: string;
  stripe_subscription_id: string;
  status: RecurringStatus;
  started_at: string;
  cancel_at_period_end: boolean;
  cancellation_requested_at: string | null;
  canceled_at: string | null;
  current_period_end: string | null;
  created_at: string;
  updated_at: string;
}

export type PaymentStatus =
  | "pending"
  | "paid"
  | "failed"
  | "canceled"
  | "refunded";

export type PaymentType = "one_time" | "recurring";

export interface ContributionPayment {
  id: string;
  donor_id: string;
  community_id: string | null;
  community_name: string | null;
  recurring_donation_id: string | null;
  amount_cents: number;
  currency: string;
  payment_type: PaymentType;
  status: PaymentStatus;
  stripe_payment_intent_id: string | null;
  stripe_invoice_id: string | null;
  stripe_subscription_id: string | null;
  stripe_event_id: string | null;
  paid_at: string | null;
  refunded_at: string | null;
  created_at: string;
  updated_at: string;
}

export type EventStatus = "draft" | "published" | "archived";

export interface ChurchEvent {
  id: string;
  title: string;
  slug: string;
  description: string | null;
  image_path: string | null;
  starts_at: string;
  ends_at: string | null;
  location: string | null;
  status: EventStatus;
  created_by: string;
  updated_by: string | null;
  published_at: string | null;
  created_at: string;
  updated_at: string;
}

export type ContactStatus = "new" | "in_progress" | "resolved" | "archived";

export interface ContactMessage {
  id: string;
  name: string;
  email: string;
  phone: string | null;
  message: string;
  privacy_consent_at: string;
  ip_hash: string | null;
  status: ContactStatus;
  assigned_to: string | null;
  created_at: string;
  updated_at: string;
}

export type FinancialPeriodStatus = "draft" | "review" | "published" | "archived";

export interface FinancialPeriod {
  id: string;
  name: string;
  start_date: string;
  end_date: string;
  status: FinancialPeriodStatus;
  created_by: string;
  reviewed_by: string | null;
  published_by: string | null;
  published_at: string | null;
  created_at: string;
  updated_at: string;
}

export type FinancialEntryType = "income" | "expense";
export type FinancialEntrySource = "online_contribution" | "manual" | "other";

export interface FinancialEntry {
  id: string;
  financial_period_id: string;
  type: FinancialEntryType;
  category: string;
  description: string;
  amount_cents: number;
  entry_date: string;
  source: FinancialEntrySource;
  created_by: string;
  updated_by: string | null;
  created_at: string;
  updated_at: string;
}

export interface TransparencyDocument {
  id: string;
  financial_period_id: string;
  file_path: string;
  original_file_name: string;
  mime_type: string;
  file_size: number;
  created_by: string;
  created_at: string;
}

export interface StripeWebhookEvent {
  id: string;
  stripe_event_id: string;
  event_type: string;
  payload_hash: string;
  processing_status: "processed" | "failed" | "skipped";
  error_message: string | null;
  created_at: string;
}

export interface AuditLog {
  id: string;
  actor_id: string;
  action: string;
  entity_type: string;
  entity_id: string | null;
  metadata: Record<string, unknown> | null;
  created_at: string;
}

export interface RecurringMagicToken {
  id: string;
  token_hash: string;
  email: string;
  expires_at: string;
  consumed_at: string | null;
  created_at: string;
}

export interface ContactRateLimit {
  id: string;
  ip_hash: string;
  created_at: string;
}
import { Resend } from "resend";
import { env } from "@/lib/env";

const resend = new Resend(env.CHAVE_API_RESEND);

export interface EmailRecipient {
  to: string;
  subject: string;
  html: string;
  text?: string;
}

export async function sendEmail({
  to,
  subject,
  html,
  text,
}: EmailRecipient): Promise<{ ok: boolean; id?: string; error?: string }> {
  const { data, error } = await resend.emails.send({
    from: env.EMAIL_REMETENTE,
    to: [to],
    subject,
    html,
    text: text ?? stripHtml(html),
  });

  if (error) {
    return { ok: false, error: error.message };
  }
  return { ok: true, id: data?.id };
}

function stripHtml(html: string): string {
  return html
    .replace(/<[^>]+>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

import { createStaticPix, hasError } from "pix-utils";
import type { SiteSettings } from "@/features/settings/types";

export async function generateOpenPixQr(settings: SiteSettings): Promise<{
  brCode: string;
  qrCodeImage: string;
  receiverName: string;
  receiverCity: string;
} | null> {
  if (!settings.pix_key || !settings.pix_receiver_name || !settings.pix_receiver_city) {
    return null;
  }

  const pix = createStaticPix({
    pixKey: settings.pix_key,
    merchantName: settings.pix_receiver_name,
    merchantCity: settings.pix_receiver_city,
    transactionAmount: 0,
    isTransactionUnique: false,
  });

  if (hasError(pix)) {
    return null;
  }

  return {
    brCode: pix.toBRCode(),
    qrCodeImage: await pix.toImage(),
    receiverName: settings.pix_receiver_name,
    receiverCity: settings.pix_receiver_city,
  };
}

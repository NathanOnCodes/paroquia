import { redirect } from "next/navigation";
import AdminShell from "@/components/admin/AdminShell";
import { getAdminUser } from "@/lib/auth/authorization";
import { getSiteSettings } from "@/features/settings/queries";

export const dynamic = "force-dynamic";

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const authUser = await getAdminUser();
  const user = authUser?.profile;

  if (!authUser || !user) {
    redirect("/login");
  }

  const settings = await getSiteSettings();

  return (
    <AdminShell
      user={{
        name: user.full_name,
        email: user.email,
        role: user.role,
        siteName: settings.site_name,
      }}
    >
      {children}
    </AdminShell>
  );
}

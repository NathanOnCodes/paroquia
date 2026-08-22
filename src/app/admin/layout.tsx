import { redirect } from "next/navigation";
import AdminShell from "@/components/admin/AdminShell";
import { getAdminUser } from "@/lib/auth/authorization";

export const dynamic = "force-dynamic";

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const authUser = await getAdminUser();
  const user = authUser?.profile;

  if (!authUser || !user) {
    redirect("/login");
  }

  return (
    <AdminShell
      user={{
        name: user.full_name,
        email: user.email,
        role: user.role,
      }}
    >
      {children}
    </AdminShell>
  );
}
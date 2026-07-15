import { AuthLayout } from "@/components/account/auth-layout";
import { ResetPasswordForm } from "@/components/account/password-form";

export default async function ResetPasswordPage({ searchParams }: { searchParams: Promise<{ token?: string }> }) {
  const params = await searchParams;
  return <AuthLayout><ResetPasswordForm token={params.token ?? ""} /></AuthLayout>;
}

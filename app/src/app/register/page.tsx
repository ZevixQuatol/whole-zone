import { AuthLayout } from "@/components/account/auth-layout";
import { RegisterForm } from "@/components/account/register-form";

export default async function RegisterPage({ searchParams }: { searchParams: Promise<{ invite?: string }> }) {
  const params = await searchParams;
  return <AuthLayout><RegisterForm initialInvite={params.invite} /></AuthLayout>;
}

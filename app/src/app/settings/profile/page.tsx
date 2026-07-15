import { AccountShell } from "@/components/account/account-shell";
import { ProfileForm } from "@/components/account/profile-form";

export default function ProfilePage() { return <AccountShell><h1 className="page-title">公开资料</h1><p className="page-lead">这些资料会显示在你的新时代名片上。</p><ProfileForm /></AccountShell>; }

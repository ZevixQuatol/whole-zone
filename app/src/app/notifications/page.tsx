import { AccountShell } from "@/components/account/account-shell";
import { NotificationList } from "@/components/account/notification-list";

export default function NotificationsPage() { return <AccountShell><h1 className="page-title">系统通知</h1><p className="page-lead">这里只展示权重处理、评价结果和平台操作通知。</p><NotificationList /></AccountShell>; }

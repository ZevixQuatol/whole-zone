import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "寰域",
  description: "以寰点连接内容、域与人的沟通平台",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="zh-CN">
      <body>{children}</body>
    </html>
  );
}

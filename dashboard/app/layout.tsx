import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import { LayoutShell } from "@/components/layout-shell";
import { DemoBanner } from "@/components/demo-banner";
import "./globals.css";

export const metadata: Metadata = {
  title: "VMMonitor",
  description: "Monitor and manage your Oracle Cloud VMs",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${GeistSans.variable} ${GeistMono.variable} dark`}>
      <body className="bg-background text-text-primary font-sans antialiased h-screen flex flex-col">
        <DemoBanner />
        <LayoutShell>{children}</LayoutShell>
      </body>
    </html>
  );
}

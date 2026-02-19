import type { Metadata } from "next";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";
import { Sidebar } from "@/components/sidebar";
import "./globals.css";

export const metadata: Metadata = {
  title: "VMMonitor",
  description: "Monitor and manage your Oracle Cloud VMs",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${GeistSans.variable} ${GeistMono.variable} dark`}>
      <body className="bg-background text-text-primary font-sans antialiased h-screen flex flex-col">
        <header className="h-14 border-b border-border flex items-center px-4 shrink-0">
          <span className="text-sm font-semibold text-text-primary">VMMonitor</span>
        </header>
        <div className="flex flex-1 overflow-hidden">
          <Sidebar />
          <main className="flex-1 overflow-y-auto px-6 py-6">
            <div className="max-w-6xl">
              {children}
            </div>
          </main>
        </div>
      </body>
    </html>
  );
}

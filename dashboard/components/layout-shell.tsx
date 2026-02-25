"use client";

import { useState, useEffect } from "react";
import { usePathname } from "next/navigation";
import { Menu } from "lucide-react";
import { Sidebar } from "./sidebar";

export function LayoutShell({ children }: { children: React.ReactNode }) {
  const [navOpen, setNavOpen] = useState(false);
  const pathname = usePathname();

  // Close mobile nav on navigation
  useEffect(() => {
    setNavOpen(false);
  }, [pathname]);

  return (
    <>
      <header className="h-14 border-b border-border flex items-center px-4 shrink-0 gap-3">
        <button
          className="md:hidden p-1 -ml-1 text-text-muted hover:text-text-primary transition-colors"
          onClick={() => setNavOpen(true)}
          aria-label="Open navigation"
        >
          <Menu size={20} />
        </button>
        <span className="text-sm font-semibold text-text-primary">VMMonitor</span>
      </header>

      <div className="flex flex-1 overflow-hidden">
        {/* Desktop sidebar — always visible */}
        <div className="hidden md:flex">
          <Sidebar />
        </div>

        {/* Mobile sidebar — overlay drawer */}
        {navOpen && (
          <>
            <div
              className="fixed inset-0 z-40 bg-black/50 md:hidden"
              onClick={() => setNavOpen(false)}
            />
            <div className="fixed inset-y-0 left-0 z-50 flex bg-background md:hidden">
              <Sidebar />
            </div>
          </>
        )}

        <main className="flex-1 overflow-y-auto px-4 sm:px-6 py-6">
          <div className="max-w-6xl">
            {children}
          </div>
        </main>
      </div>
    </>
  );
}

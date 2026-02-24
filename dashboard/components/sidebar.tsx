"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutGrid, Monitor, Plus } from "lucide-react";

const navItems = [
  { href: "/", label: "Apps", icon: LayoutGrid },
  { href: "/vms", label: "VMs", icon: Monitor },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-56 border-r border-border flex flex-col gap-1 p-3 shrink-0">
      {navItems.map(({ href, label, icon: Icon }) => {
        const active = href === "/" ? pathname === "/" : pathname.startsWith(href);
        return active ? (
          <Link
            key={href}
            href={href}
            className="flex items-center gap-3 px-3 py-2 rounded-md text-sm text-text-primary bg-surface-raised border-l-2 border-accent pl-[10px]"
          >
            <Icon size={16} className="text-accent" />
            {label}
          </Link>
        ) : (
          <Link
            key={href}
            href={href}
            className="flex items-center gap-3 px-3 py-2 rounded-md text-sm text-text-muted hover:text-text-primary hover:bg-surface-raised transition-colors"
          >
            <Icon size={16} />
            {label}
          </Link>
        );
      })}

      <div className="mt-2 pt-2 border-t border-border">
        <Link
          href="/apps/new"
          className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors ${
            pathname === "/apps/new"
              ? "text-text-primary bg-surface-raised border-l-2 border-accent pl-[10px]"
              : "text-text-muted hover:text-text-primary hover:bg-surface-raised"
          }`}
        >
          <Plus size={16} className={pathname === "/apps/new" ? "text-accent" : ""} />
          Register App
        </Link>
      </div>
    </aside>
  );
}

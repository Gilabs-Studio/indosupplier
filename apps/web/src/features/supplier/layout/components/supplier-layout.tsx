"use client";

import React from "react";
import { Link } from "@/i18n/routing";
import { useSupplierLayout } from "../hooks/use-supplier-layout";
import {
  Loader2,
  Search,
  Bell,
  User,
  Receipt,
  LogOut,
  AlertCircle,
  Menu,
  ChevronLeft,
  ChevronRight
} from "lucide-react";
import { toast } from "sonner";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getDicebearUrl } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogClose,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  useSidebar
} from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";

interface SupplierLayoutProps {
  children: React.ReactNode;
}

function SupplierSidebar() {
  const { menuGroups, pathname } = useSupplierLayout();

  return (
    <Sidebar collapsible="icon" className="h-screen shrink-0 border-none bg-card">
      <SidebarHeader className="h-16 flex flex-row items-center justify-start px-6 overflow-hidden shrink-0">
        <span className="font-extrabold text-foreground tracking-tight text-lg select-none leading-none">
          supplier
        </span>
      </SidebarHeader>
      <SidebarContent className="p-3 space-y-4">
        {menuGroups.map((group) => (
          <SidebarGroup key={group.title} className="p-0">
            <SidebarGroupLabel className="px-3 text-[10px] font-bold text-muted-foreground uppercase tracking-wider mb-2 mt-4 first:mt-0">
              {group.title}
            </SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu className="gap-1">
                {group.items.map((item) => {
                  const isActive = pathname === item.url || (item.url !== "/supplier/dashboard" && pathname.startsWith(item.url));
                  const Icon = item.icon;

                  return (
                    <SidebarMenuItem key={item.name}>
                      <SidebarMenuButton
                        asChild
                        isActive={isActive}
                        className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-300 group cursor-pointer hover:-translate-y-0.5 active:translate-y-0 ${
                          isActive
                            ? "text-primary bg-primary/8 font-semibold shadow-xs shadow-primary/10"
                            : "text-muted-foreground hover:text-foreground hover:bg-muted/40"
                        }`}
                      >
                        <Link href={item.url} className="flex items-center gap-3 w-full">
                          <Icon className={`h-4.5 w-4.5 transition-transform duration-300 group-hover:scale-105 ${isActive ? "text-primary" : ""}`} />
                          <span className="text-sm select-none truncate">
                            {item.name}
                          </span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  );
                })}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
    </Sidebar>
  );
}

function MobileMenuButton() {
  const { toggleSidebar } = useSidebar();
  return (
    <Button
      variant="ghost"
      size="icon"
      className="cursor-pointer md:hidden shrink-0"
      onClick={toggleSidebar}
      aria-label="Open dashboard menu"
    >
      <Menu className="h-5 w-5" />
    </Button>
  );
}

export default function SupplierLayoutComponent({ children }: SupplierLayoutProps) {
  const {
    t,
    locale,
    pathname,
    user,
    mounted,
    isAuthorizing,
    showLogoutConfirm,
    setShowLogoutConfirm,
    isVerified,
    menuGroups,
    handleLogout,
  } = useSupplierLayout();

  if (!mounted || isAuthorizing) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-background text-foreground">
        <Loader2 className="h-10 w-10 animate-spin text-primary mb-4" />
        <p className="text-sm font-semibold tracking-wide text-muted-foreground">
          {t("loading")}
        </p>
      </div>
    );
  }

  const activeItem = menuGroups
    .flatMap((group) => group.items)
    .find((item) => pathname === item.url || (item.url !== "/supplier/dashboard" && pathname.startsWith(item.url)));
  const activeTitle = activeItem ? activeItem.name : t("menu.dashboard");

  return (
    <TooltipProvider>
      <SidebarProvider defaultOpen={true}>
        <div className="h-screen w-screen overflow-hidden bg-background flex relative" data-slot="supplier-root-container">
          {/* Sidebar */}
          <SupplierSidebar />

          {/* Main Workspace */}
          <div className="flex-1 flex flex-col min-w-0 h-screen overflow-hidden relative">
            {/* Header */}
            <header className="shrink-0 z-20 bg-background/95 backdrop-blur h-16 w-full flex items-center justify-between px-4 md:px-6">
              <div className="flex flex-1 items-center justify-between min-w-0 gap-4">
                <div className="flex min-w-0 flex-1 items-center gap-3">
                  {/* Mobile toggle button */}
                  <MobileMenuButton />

                  {/* Navigation Arrows < > */}
                  <div className="hidden md:flex items-center gap-1 shrink-0 text-muted-foreground mr-1">
                    <button
                      onClick={() => window.history.back()}
                      className="p-1 hover:bg-muted rounded-md cursor-pointer transition-colors"
                      aria-label="Go back"
                    >
                      <ChevronLeft className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => window.history.forward()}
                      className="p-1 hover:bg-muted rounded-md cursor-pointer transition-colors"
                      aria-label="Go forward"
                    >
                      <ChevronRight className="h-4 w-4" />
                    </button>
                  </div>

                  {/* Breadcrumbs */}
                  <div className="min-w-0 flex items-center text-xs font-semibold text-muted-foreground gap-1.5 select-none">
                    <span className="truncate">{locale === "id" ? "Portal Supplier" : "Supplier Portal"}</span>
                    <span>/</span>
                    <span className="text-foreground font-bold truncate">
                      {activeTitle}
                    </span>
                  </div>
                </div>

                {/* Quick actions search bar */}
                <div className="relative hidden md:block w-80">
                  <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground/80" />
                  <input
                    type="text"
                    placeholder={t("searchPlaceholder")}
                    className="h-8.5 bg-muted/40 border border-border/80 focus-visible:ring-1 focus-visible:ring-primary pl-9 pr-4 rounded-lg text-xs font-medium w-full shadow-none cursor-pointer outline-none transition-all duration-300"
                  />
                </div>

                {/* User actions / notifications */}
                <div className="flex items-center gap-4 shrink-0">
                  <button
                    onClick={() => toast.info(t("notificationAlert"))}
                    className="relative p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted/40 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 cursor-pointer"
                    aria-label="Notifications"
                  >
                    <Bell className="h-4.5 w-4.5" />
                    <span className="absolute top-1.5 right-1.5 flex h-1.5 w-1.5 rounded-full bg-destructive" />
                  </button>

                  <div className="h-6 w-px bg-border/80" />

                  {/* Profile Dropdown */}
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <button className="flex items-center gap-2 cursor-pointer hover:bg-muted/40 p-1.5 rounded-lg transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 select-none outline-none">
                        <Avatar className="h-8 w-8 border border-border">
                          <AvatarImage src={getDicebearUrl(user?.email || "supplier", "lorelei")} alt={user?.name} />
                          <AvatarFallback className="bg-primary/10 text-primary font-semibold text-xs">
                            {user?.name?.slice(0, 2).toUpperCase() || "SP"}
                          </AvatarFallback>
                        </Avatar>
                        <div className="hidden sm:flex flex-col text-left">
                          <span className="text-xs font-semibold text-foreground leading-none">{user?.name || "PT Nusantara Supplier"}</span>
                          <span className="text-[10px] text-success font-semibold flex items-center gap-1 mt-1 leading-none">
                            <span className="h-1 w-1 rounded-full bg-success inline-block" />
                            {t("online")}
                          </span>
                        </div>
                      </button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-56 mt-1 rounded-lg">
                      <DropdownMenuLabel className="font-semibold text-xs text-muted-foreground uppercase tracking-wider px-3 py-2">
                        {locale === "id" ? "Portal Supplier" : "Supplier Portal"}
                      </DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem asChild>
                        <Link href="/supplier/profile" className="flex items-center gap-2 px-3 py-2 cursor-pointer w-full text-sm">
                          <User className="h-4 w-4" />
                          <span>{t("menu.profile")}</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem asChild>
                        <Link href="/supplier/subscription" className="flex items-center gap-2 px-3 py-2 cursor-pointer w-full text-sm">
                          <Receipt className="h-4 w-4" />
                          <span>{t("menu.subscription")}</span>
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        onClick={() => setShowLogoutConfirm(true)}
                        className="flex items-center gap-2 px-3 py-2 text-destructive focus:text-destructive cursor-pointer"
                      >
                        <LogOut className="h-4 w-4" />
                        <span>{t("signOut")}</span>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            </header>

            {/* Main content body with verification warning */}
            <main className="flex-1 min-h-0 bg-muted/10 overflow-y-auto p-6 md:p-8">
              <div className="max-w-6xl mx-auto space-y-6">
                {!isVerified && (
                  <div className="bg-card border border-border/80 border-l-4 border-l-primary p-4.5 rounded-lg flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 select-none shadow-xs transition-all duration-300 hover:shadow-md">
                    <div className="flex items-center gap-3.5 text-left">
                      <div className="h-9 w-9 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                        <AlertCircle className="h-5 w-5" />
                      </div>
                      <div className="flex flex-col">
                        <span className="font-extrabold text-[10px] uppercase tracking-wider text-primary">
                          {t("warningBannerBadge")}
                        </span>
                        <span className="text-xs text-muted-foreground font-medium mt-1 leading-relaxed">
                          {t("warningBannerMessage")}
                        </span>
                      </div>
                    </div>
                    <Link
                      href="/supplier/verification"
                      className="border border-primary text-primary hover:bg-primary hover:text-primary-foreground px-4 py-2 rounded-lg text-xs font-bold transition-all duration-300 cursor-pointer shrink-0 self-start sm:self-auto hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/20 flex items-center gap-1.5"
                    >
                      {t("warningBannerButton")}
                    </Link>
                  </div>
                )}
                {children}
              </div>
            </main>
          </div>
        </div>

        {/* Logout Dialog */}
        <Dialog open={showLogoutConfirm} onOpenChange={setShowLogoutConfirm}>
          <DialogContent size="sm">
            <DialogHeader>
              <DialogTitle className="font-heading">{t("signOut")}</DialogTitle>
              <DialogDescription>{t("signOutConfirm")}</DialogDescription>
            </DialogHeader>
            <DialogFooter className="gap-2 sm:gap-0">
              <DialogClose asChild>
                <Button variant="outline" className="cursor-pointer">
                  Cancel
                </Button>
              </DialogClose>
              <Button
                variant="destructive"
                className="cursor-pointer"
                onClick={handleLogout}
              >
                {t("signOut")}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SidebarProvider>
    </TooltipProvider>
  );
}

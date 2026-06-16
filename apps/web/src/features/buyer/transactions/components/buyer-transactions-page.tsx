"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/routing";
import { useBuyerTransactions } from "@/features/buyer/transactions/hooks/useBuyerTransactions";
import { BuyerLayout } from "../../components/buyer-layout";
import { TransactionItem } from "@/features/buyer/transactions/types/transaction.types";
import { formatCurrency } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Search, Loader2, Calendar, FileText, ShoppingBag, Eye, MessageSquare } from "lucide-react";
import { cn } from "@/lib/utils";

export function BuyerTransactionsPage() {
  const t = useTranslations("buyer.transactions");
  const [activeTab, setActiveTab] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [currentPage, setCurrentPage] = useState<number>(1);
  const itemsPerPage = 5;

  const { data, isLoading } = useBuyerTransactions({
    page: currentPage,
    per_page: itemsPerPage,
    status: activeTab,
  });

  const statusTabs = [
    { id: "all", label: t("tabAll") },
    { id: "pending", label: t("tabUnpaid") },
    { id: "processing", label: t("tabProcessing") },
    { id: "shipped", label: t("tabShipped") },
    { id: "completed", label: t("tabCompleted") },
    { id: "cancelled", label: t("tabCancelled") },
  ];

  // Client-side search filter simulation for search input
  const filteredItems = data?.items.filter((item: TransactionItem) =>
    item.product_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    item.po_number.toLowerCase().includes(searchQuery.toLowerCase()) ||
    item.supplier_name.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];

  const handleTabChange = (tabId: string) => {
    setActiveTab(tabId);
    setCurrentPage(1);
  };

  const getStatusColorClass = (status: string) => {
    switch (status) {
      case "completed":
        return "bg-success/10 text-success border-success/20";
      case "processing":
        return "bg-warning/10 text-warning border-warning/20";
      case "shipped":
        return "bg-cyan/10 text-cyan border-cyan/20";
      case "cancelled":
        return "bg-destructive/10 text-destructive border-destructive/20";
      default:
        return "bg-primary/10 text-primary border-primary/20";
    }
  };

  const renderContent = () => {
    if (isLoading) {
      return (
        <div className="flex flex-col items-center justify-center py-20 gap-3">
          <Loader2 className="h-8 w-8 text-primary animate-spin" />
          <p className="text-sm text-muted-foreground">Memuat daftar transaksi...</p>
        </div>
      );
    }

    if (filteredItems.length === 0) {
      return (
        <div className="bg-card rounded-xl border border-border p-12 text-center shadow-xs flex flex-col items-center justify-center gap-4">
          <div className="bg-muted p-4 rounded-full text-muted-foreground">
            <ShoppingBag className="h-8 w-8" />
          </div>
          <div className="space-y-1">
            <h3 className="font-bold text-base text-foreground">{t("empty")}</h3>
            <p className="text-xs text-muted-foreground">Coba cari produk atau supplier lain.</p>
          </div>
        </div>
      );
    }

    return (
      <div className="space-y-4">
        {filteredItems.map((tx: TransactionItem) => (
          <div
            key={tx.id}
            className="bg-card rounded-xl border border-border p-5 shadow-xs space-y-4 hover:shadow-md transition-all duration-300"
          >
            {/* Card Header (Date, PO number, Status Badge) */}
            <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border pb-3">
              <div className="flex items-center gap-3 text-xs text-muted-foreground">
                <span className="flex items-center gap-1">
                  <Calendar className="h-3.5 w-3.5" />
                  {new Date(tx.created_at).toLocaleDateString("id-ID", {
                    year: "numeric",
                    month: "short",
                    day: "numeric",
                  })}
                </span>
                <span className="h-3 w-px bg-border" />
                <span className="flex items-center gap-1 font-semibold text-foreground">
                  <FileText className="h-3.5 w-3.5" />
                  {tx.po_number}
                </span>
              </div>

              <div className="flex items-center gap-2">
                <span
                  className={cn(
                    "px-2.5 py-0.5 rounded-md text-[11px] font-semibold uppercase tracking-wider border",
                    getStatusColorClass(tx.status)
                  )}
                >
                  {tx.status}
                </span>
                <span
                  className={cn(
                    "px-2.5 py-0.5 rounded-md text-[11px] font-semibold uppercase tracking-wider border",
                    tx.payment_status === "paid"
                      ? "bg-success/10 text-success border-success/20"
                      : "bg-warning/10 text-warning border-warning/20"
                  )}
                >
                  {tx.payment_status}
                </span>
              </div>
            </div>

            {/* Card Body (Supplier info, Product detail) */}
            <div className="flex items-start gap-4">
              <div className="bg-primary/10 h-10 w-10 rounded-lg flex items-center justify-center shrink-0">
                <ShoppingBag className="h-5 w-5 text-primary" />
              </div>
              <div className="min-w-0 flex-1 space-y-1">
                <h4 className="font-extrabold text-sm text-foreground hover:text-primary transition-colors cursor-pointer">
                  {tx.product_name}
                </h4>
                <p className="text-xs text-muted-foreground font-semibold">
                  {t("supplier")}: <span className="text-foreground">{tx.supplier_name}</span>
                </p>
                <p className="text-xs text-muted-foreground font-medium">
                  {tx.quantity_value} {tx.quantity_unit} x {formatCurrency(tx.price_per_unit)}
                </p>
              </div>
              <div className="text-right space-y-0.5 shrink-0">
                <p className="text-xs text-muted-foreground font-semibold">{t("total")}</p>
                <p className="text-base font-extrabold text-foreground">{formatCurrency(tx.total_amount)}</p>
              </div>
            </div>

            {/* Card Footer (Actions) */}
            <div className="flex items-center justify-between gap-4 pt-1">
              <p className="text-xs text-muted-foreground truncate max-w-[50%]">
                {tx.notes && `${t("notes")}: "${tx.notes}"`}
              </p>

              <div className="flex items-center gap-2">
                <Link href={`/chat`}>
                  <Button variant="outline" size="sm" className="h-8.5 text-xs font-semibold cursor-pointer">
                    <MessageSquare className="mr-1.5 h-3.5 w-3.5" />
                    {t("chatBtn")}
                  </Button>
                </Link>

                <Link href={`/transactions/${tx.id}`}>
                  <Button variant="outline" size="sm" className="h-8.5 text-xs font-semibold cursor-pointer">
                    <Eye className="mr-1.5 h-3.5 w-3.5" />
                    {t("detailBtn")}
                  </Button>
                </Link>

                {tx.status === "completed" && (
                  <Link href={`/reviews`}>
                    <Button size="sm" className="h-8.5 text-xs font-semibold cursor-pointer bg-primary text-primary-foreground hover:-translate-y-0.5 active:translate-y-0">
                      {t("reviewBtn")}
                    </Button>
                  </Link>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  };

  return (
    <BuyerLayout>
      <div className="space-y-6">
        {/* Title Header */}
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">{t("title")}</h1>
          <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
        </div>

        {/* Filter Tabs & Search */}
        <div className="space-y-4">
          {/* Navigation Tabs (Tokopedia-style horizontal tabs) */}
          <div className="flex items-center gap-1.5 border-b border-border overflow-x-auto pb-1 scrollbar-none">
            {statusTabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => handleTabChange(tab.id)}
                className={cn(
                  "px-4 py-2 text-sm font-medium transition-all duration-200 whitespace-nowrap cursor-pointer hover:text-primary hover:-translate-y-0.5 active:translate-y-0",
                  activeTab === tab.id
                    ? "text-primary font-semibold"
                    : "text-muted-foreground"
                )}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* Search Input */}
          <div className="relative">
            <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4.5 w-4.5 text-muted-foreground" />
            <Input
              type="text"
              placeholder={t("searchPlaceholder")}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10 h-10.5 rounded-lg border-border focus-visible:ring-primary focus-visible:border-primary text-sm font-medium bg-card"
            />
          </div>
        </div>

        {/* Transactions List Content */}
        {renderContent()}

        {/* Pagination Controls */}
        {data && data.total > itemsPerPage && (
          <div className="flex items-center justify-between border-t border-border pt-4">
            <p className="text-xs text-muted-foreground font-semibold">
              Showing <span className="font-bold">{(currentPage - 1) * itemsPerPage + 1}</span> to{" "}
              <span className="font-bold">
                {Math.min(currentPage * itemsPerPage, data.total)}
              </span>{" "}
              of <span className="font-bold">{data.total}</span> entries
            </p>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage === 1}
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                className="h-8 text-xs font-semibold cursor-pointer"
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={currentPage * itemsPerPage >= data.total}
                onClick={() => setCurrentPage((p) => p + 1)}
                className="h-8 text-xs font-semibold cursor-pointer"
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </div>
    </BuyerLayout>
  );
}


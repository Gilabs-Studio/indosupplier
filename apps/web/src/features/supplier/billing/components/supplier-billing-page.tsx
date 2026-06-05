"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";
import { Landmark, Download, FileText, CheckCircle2, AlertCircle } from "lucide-react";

export function SupplierBillingPage() {
  const t = useTranslations("supplier.billing");

  const [invoices, setInvoices] = useState([
    { id: "INV-2026-004", desc: "Gold Membership Plan Subscription (Annual)", amount: "Rp 12.000.000", status: "paid", date: "2026-06-01" },
    { id: "INV-2026-003", desc: "Premium Search Result Slot Auction #098", amount: "Rp 3.100.000", status: "paid", date: "2026-05-15" },
    { id: "INV-2026-002", desc: "Advertising Clicks Charge (April 2026)", amount: "Rp 540.000", status: "paid", date: "2026-05-01" },
    { id: "INV-2026-005", desc: "Advertising Clicks Charge (May 2026)", amount: "Rp 750.000", status: "unpaid", date: "2026-06-05" },
  ]);

  const handlePay = (id: string) => {
    toast.promise(
      new Promise((resolve) => setTimeout(resolve, 1000)),
      {
        loading: "Processing payment...",
        success: () => {
          setInvoices(invoices.map(inv => inv.id === id ? { ...inv, status: "paid" } : inv));
          return "Payment successful! Thank you.";
        },
        error: "Payment failed",
      }
    );
  };

  const handleDownload = (id: string) => {
    toast.info(`Downloading invoice PDF for ${id}...`);
  };

  return (
    <div className="space-y-6 text-left">
      {/* Header */}
      <div className="border-b border-border/80 pb-6">
        <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
          {t("title")}
        </h1>
        <p className="text-sm text-muted-foreground">
          {t("subtitle")}
        </p>
      </div>

      {/* Overview stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Outstanding Invoice Card */}
        <Card className="border border-border shadow-xs rounded-xl bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs uppercase tracking-wider text-muted-foreground">{t("outstanding")}</CardTitle>
          </CardHeader>
          <CardContent className="flex items-center justify-between">
            <div className="space-y-1">
              <span className="text-2xl font-extrabold text-foreground">Rp 750.000</span>
              <p className="text-xs text-muted-foreground">1 invoice pending payment</p>
            </div>
            <div className="h-10 w-10 bg-warning/10 text-warning border border-border rounded-lg flex items-center justify-center">
              <AlertCircle className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>

        {/* Link Bank Payout */}
        <Card className="border border-border shadow-xs rounded-xl bg-card md:col-span-2">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs uppercase tracking-wider text-muted-foreground">{t("payoutBank")}</CardTitle>
          </CardHeader>
          <CardContent className="flex items-center justify-between flex-wrap gap-4">
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 bg-primary/10 text-primary border border-border rounded-lg flex items-center justify-center shrink-0">
                <Landmark className="h-5 w-5" />
              </div>
              <div>
                <p className="text-sm font-bold text-foreground">Bank Central Asia (BCA)</p>
                <p className="text-xs text-muted-foreground">Account Number: ****8920 (PT Nusantara Supplier Utama)</p>
              </div>
            </div>
            <Button variant="outline" size="sm" className="text-xs cursor-pointer border-border font-semibold">
              Manage Bank Account
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Ledger Table */}
      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card">
        <CardHeader className="border-b border-border bg-muted/20 py-4 px-6">
          <CardTitle className="text-sm font-bold text-foreground">{t("paymentHistory")}</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-b border-border bg-muted/10">
                  <TableHead className="font-bold text-foreground py-3 px-6">{t("invoiceNo")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("description")}</TableHead>
                  <TableHead className="font-bold text-foreground">{t("amount")}</TableHead>
                  <TableHead className="font-bold text-foreground">Date</TableHead>
                  <TableHead className="font-bold text-foreground">Status</TableHead>
                  <TableHead className="font-bold text-foreground text-right px-6">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {invoices.map((inv) => (
                  <TableRow key={inv.id} className="hover:bg-muted/10 border-b border-border transition-colors">
                    <TableCell className="py-4 px-6 font-bold text-muted-foreground">{inv.id}</TableCell>
                    <TableCell className="py-4 font-semibold text-foreground flex items-center gap-2">
                      <FileText className="h-4.5 w-4.5 text-muted-foreground shrink-0" />
                      <span>{inv.desc}</span>
                    </TableCell>
                    <TableCell className="py-4 font-extrabold text-foreground">{inv.amount}</TableCell>
                    <TableCell className="py-4 text-xs font-semibold text-muted-foreground">{inv.date}</TableCell>
                    <TableCell className="py-4">
                      {inv.status === "paid" ? (
                        <Badge className="bg-success/15 text-success border border-success/30 font-bold flex items-center gap-1 max-w-fit">
                          <CheckCircle2 className="h-3 w-3" /> Paid
                        </Badge>
                      ) : (
                        <Badge className="bg-warning/15 text-warning border border-warning/30 font-bold flex items-center gap-1 max-w-fit">
                          <AlertCircle className="h-3 w-3" /> Unpaid
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell className="py-4 px-6 text-right space-x-1.5">
                      {inv.status === "unpaid" && (
                        <Button onClick={() => handlePay(inv.id)} size="sm" className="text-xs h-8 cursor-pointer font-semibold">
                          {t("actionPay")}
                        </Button>
                      )}
                      <Button onClick={() => handleDownload(inv.id)} variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-primary cursor-pointer hover:bg-primary/5 border border-border">
                        <Download className="h-3.5 w-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

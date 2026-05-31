"use client";

import React, { useEffect, useState, useCallback } from "react";
import { useSysadminStore } from "@/features/sysadmin/stores/use-sysadmin-store";
import { waitingListService } from "@/features/waiting-list/services/waiting-list-service";
import type { WaitingListEntry } from "@/features/waiting-list/types";
import { toast } from "sonner";
import {
  LogOut,
  Users,
  Trash2,
  RefreshCw,
  Inbox,
  UserCheck,
  Mail,
  Building,
  Phone,
  ChevronLeft,
  ChevronRight,
  Loader2,
  SlidersHorizontal
} from "lucide-react";
import { format } from "date-fns";

import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ThemeToggleButton } from "@/components/ui/theme-toggle";

export default function SysadminDashboard() {
  const { admin, logout } = useSysadminStore();
  const [entries, setEntries] = useState<WaitingListEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [statusFilter, setStatusFilter] = useState<string>("");
  const [isLoading, setIsLoading] = useState(true);

  // Statistics state
  const [stats, setStats] = useState({
    total: 0,
    suppliers: 0,
    buyers: 0,
    pending: 0,
  });

  const fetchEntries = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await waitingListService.list({
        page,
        limit,
        status: statusFilter || undefined,
      });
      setEntries(data.items);
      setTotal(data.total);

      // Compute statistics globally by loading a larger batch or calculating local stats
      const allData = await waitingListService.list({ page: 1, limit: 1000 });
      const supplierCount = allData.items.filter(i => i.company_type === "supplier").length;
      const buyerCount = allData.items.filter(i => i.company_type === "buyer").length;
      const pendingCount = allData.items.filter(i => i.status === "pending").length;

      setStats({
        total: allData.total,
        suppliers: supplierCount,
        buyers: buyerCount,
        pending: pendingCount,
      });
    } catch (error) {
      toast.error("Failed to load waiting list entries.");
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, statusFilter]);

  useEffect(() => {
    fetchEntries();
  }, [fetchEntries]);

  const handleUpdateStatus = async (id: string, newStatus: string) => {
    try {
      await waitingListService.updateStatus(id, newStatus);
      toast.success(`Entry marked as ${newStatus}`);
      fetchEntries();
    } catch (error) {
      toast.error("Failed to update status.");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Are you sure you want to delete this waitlist entry?")) return;
    try {
      await waitingListService.delete(id);
      toast.success("Entry deleted successfully.");
      fetchEntries();
    } catch (error) {
      toast.error("Failed to delete entry.");
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "approved":
        return <Badge variant="success">Approved</Badge>;
      case "contacted":
        return <Badge variant="info">Contacted</Badge>;
      case "rejected":
        return <Badge variant="destructive">Rejected</Badge>;
      default:
        return <Badge variant="warning">Pending</Badge>;
    }
  };

  const getCompanyTypeBadge = (type: string) => {
    switch (type) {
      case "supplier":
        return (
          <Badge className="bg-purple/10 text-purple border border-purple/20">
            Supplier
          </Badge>
        );
      case "buyer":
        return (
          <Badge className="bg-cyan/10 text-cyan border border-cyan/20">
            Buyer
          </Badge>
        );
      default:
        return <Badge variant="secondary">Other</Badge>;
    }
  };

  const totalPages = Math.ceil(total / limit) || 1;

  return (
    <div className="min-h-screen bg-background text-foreground transition-colors duration-350">
      <main className="max-w-7xl mx-auto px-6 py-10 space-y-8">
        {/* Simplified Dashboard Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border pb-6">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">
              Waiting List
            </h1>
            <p className="text-muted-foreground text-xs mt-1">
              System Admin Portal
            </p>
          </div>

          <div className="flex items-center gap-3 self-start sm:self-center">
            <Button
              onClick={fetchEntries}
              variant="outline"
              size="sm"
              className="gap-1.5"
            >
              <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
              Refresh
            </Button>

            <ThemeToggleButton className="border border-border bg-background hover:bg-muted" />

            <div className="h-6 w-px bg-border hidden sm:block" />

            <div className="text-right hidden sm:block">
              <span className="text-xs font-semibold block">{admin?.name}</span>
              <span className="text-[10px] text-muted-foreground block capitalize leading-none mt-0.5">
                {admin?.role.replace("_", " ")}
              </span>
            </div>

            <Button
              variant="destructive"
              size="sm"
              onClick={logout}
              className="gap-1.5"
            >
              <LogOut className="h-3.5 w-3.5" />
              Sign Out
            </Button>
          </div>
        </div>

        {/* Statistics Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {/* Card 1: Total */}
          <Card className="hover:shadow-md transition-all duration-300">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Total Registrations
              </CardTitle>
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <Users className="h-5 w-5" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-extrabold">{stats.total}</div>
              <p className="text-xs text-muted-foreground mt-1">Overall waitlist signups</p>
            </CardContent>
          </Card>

          {/* Card 2: Suppliers */}
          <Card className="hover:shadow-md transition-all duration-300">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Suppliers
              </CardTitle>
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-purple/10 text-purple border border-purple/20">
                <Building className="h-5 w-5" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-extrabold">{stats.suppliers}</div>
              <p className="text-xs text-muted-foreground mt-1">Early supplier access</p>
            </CardContent>
          </Card>

          {/* Card 3: Buyers */}
          <Card className="hover:shadow-md transition-all duration-300">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Buyers
              </CardTitle>
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-cyan/10 text-cyan border border-cyan/20">
                <UserCheck className="h-5 w-5" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-extrabold">{stats.buyers}</div>
              <p className="text-xs text-muted-foreground mt-1">Early buyer access</p>
            </CardContent>
          </Card>

          {/* Card 4: Pending */}
          <Card className="hover:shadow-md transition-all duration-300">
            <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
              <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                Pending Review
              </CardTitle>
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-warning/10 text-warning border border-warning/20">
                <Inbox className="h-5 w-5" />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-extrabold">{stats.pending}</div>
              <p className="text-xs text-muted-foreground mt-1">Requires admin approval</p>
            </CardContent>
          </Card>
        </div>

        {/* Filter and Content Card */}
        <Card className="border border-border/80 shadow-sm overflow-hidden">
          {/* Filtering Header */}
          <div className="p-5 border-b border-border flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-muted/30">
            <div className="flex items-center gap-2">
              <SlidersHorizontal className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm font-semibold">Filter Registrants</span>
            </div>

            <div className="flex items-center gap-3">
              <Select
                value={statusFilter || "all"}
                onValueChange={(val) => {
                  setStatusFilter(val === "all" ? "" : val);
                  setPage(1);
                }}
              >
                <SelectTrigger className="w-[160px] bg-background">
                  <SelectValue placeholder="All Statuses" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="pending">Pending</SelectItem>
                  <SelectItem value="contacted">Contacted</SelectItem>
                  <SelectItem value="approved">Approved</SelectItem>
                  <SelectItem value="rejected">Rejected</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Table Content */}
          {isLoading ? (
            <div className="py-24 text-center">
              <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
              <span className="text-sm font-semibold text-muted-foreground">Fetching records...</span>
            </div>
          ) : entries.length === 0 ? (
            <div className="py-24 text-center space-y-2">
              <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
              <h3 className="font-bold text-lg text-foreground">No registrations found</h3>
              <p className="text-muted-foreground text-sm max-w-sm mx-auto">
                Try changing your status filter or wait for new waitlist submissions.
              </p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="pl-6">Company / Name</TableHead>
                  <TableHead>Contact</TableHead>
                  <TableHead>Business Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Signup Date</TableHead>
                  <TableHead className="pr-6 text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {entries.map((entry) => (
                  <TableRow key={entry.id} className="hover:bg-muted/30 transition-colors">
                    {/* Name / Company */}
                    <TableCell className="pl-6 py-4">
                      <div>
                        <span className="font-bold text-foreground block">{entry.company_name}</span>
                        <span className="text-xs text-muted-foreground block mt-0.5">{entry.name}</span>
                      </div>
                    </TableCell>

                    {/* Contact */}
                    <TableCell className="py-4">
                      <div className="space-y-1">
                        <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
                          <Mail className="h-3.5 w-3.5 text-muted-foreground/60" />
                          {entry.email}
                        </span>
                        {entry.phone && (
                          <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <Phone className="h-3.5 w-3.5 text-muted-foreground/60" />
                            {entry.phone}
                          </span>
                        )}
                      </div>
                    </TableCell>

                    {/* Business Type */}
                    <TableCell className="py-4">
                      {getCompanyTypeBadge(entry.company_type)}
                    </TableCell>

                    {/* Status */}
                    <TableCell className="py-4">
                      {getStatusBadge(entry.status)}
                    </TableCell>

                    {/* Date */}
                    <TableCell className="py-4 text-xs text-muted-foreground">
                      {format(new Date(entry.created_at), "MMM d, yyyy HH:mm")}
                    </TableCell>

                    {/* Actions */}
                    <TableCell className="pr-6 py-4 text-right">
                      <div className="inline-flex items-center justify-end gap-3">
                        <Select
                          value={entry.status}
                          onValueChange={(val) => handleUpdateStatus(entry.id, val)}
                        >
                          <SelectTrigger className="w-[120px] h-8 text-xs bg-background">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="pending">Pending</SelectItem>
                            <SelectItem value="contacted">Contacted</SelectItem>
                            <SelectItem value="approved">Approved</SelectItem>
                            <SelectItem value="rejected">Rejected</SelectItem>
                          </SelectContent>
                        </Select>

                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(entry.id)}
                          className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                          title="Delete entry"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}

          {/* Pagination Footer */}
          {entries.length > 0 && (
            <div className="p-4 border-t border-border flex items-center justify-between bg-muted/20">
              <span className="text-xs text-muted-foreground font-medium">
                Showing page <span className="font-semibold text-foreground">{page}</span> of{" "}
                <span className="font-semibold text-foreground">{totalPages}</span> ({total} entries)
              </span>

              <div className="inline-flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="h-8 w-8 p-0"
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="h-8 w-8 p-0"
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </Card>
      </main>
    </div>
  );
}

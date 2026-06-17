"use client";

import React, { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Edit2, FileText, Loader2, Plus, RefreshCw, Search, Trash2 } from "lucide-react";
import { sysadminContentService } from "../services";
import type { ContentArticle, ContentArticlePayload, ContentStatus, ContentType } from "../types";
import { contentTypeLabels } from "@/features/content/utils/content-format";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Field, FieldLabel } from "@/components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { DeleteDialog } from "@/components/ui/delete-dialog";

const contentTypes: ContentType[] = ["news", "feature", "tips", "editorial_review", "video"];
const statuses: ContentStatus[] = ["draft", "published", "archived"];

const emptyPayload: ContentArticlePayload = {
  type: "news",
  locale: "id",
  title: "",
  slug: "",
  excerpt: "",
  body: "",
  authorName: "",
  imageUrl: "",
  videoUrl: "",
  duration: "",
  viewCount: 0,
  status: "draft",
  isFeatured: false,
  sortOrder: 0,
};

export default function ContentManagement() {
  const t = useTranslations("sysadminContent");
  const queryClient = useQueryClient();
  const [search, setSearch] = useState("");
  const [selectedType, setSelectedType] = useState<ContentType | "all">("all");
  const [selectedStatus, setSelectedStatus] = useState<ContentStatus | "all">("all");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editing, setEditing] = useState<ContentArticle | null>(null);
  const [form, setForm] = useState<ContentArticlePayload>(emptyPayload);
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const params = useMemo(
    () => ({
      search,
      type: selectedType === "all" ? undefined : selectedType,
      status: selectedStatus,
      page: 1,
      perPage: 50,
    }),
    [search, selectedStatus, selectedType],
  );

  const contentQuery = useQuery({
    queryKey: ["sysadmin-content", params],
    queryFn: () => sysadminContentService.listAdmin(params),
  });

  const saveMutation = useMutation({
    mutationFn: (payload: ContentArticlePayload) => {
      const cleaned = cleanPayload(payload);
      return editing
        ? sysadminContentService.update(editing.id, cleaned)
        : sysadminContentService.create(cleaned);
    },
    onSuccess: () => {
      toast.success(t("successSave"));
      setIsDialogOpen(false);
      setEditing(null);
      setForm(emptyPayload);
      queryClient.invalidateQueries({ queryKey: ["sysadmin-content"] });
    },
    onError: () => toast.error(t("errorSave")),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => sysadminContentService.delete(id),
    onSuccess: () => {
      toast.success(t("successDelete"));
      queryClient.invalidateQueries({ queryKey: ["sysadmin-content"] });
    },
    onError: () => toast.error(t("errorDelete")),
  });

  const rows = contentQuery.data?.items || [];

  const openCreate = () => {
    setEditing(null);
    setForm(emptyPayload);
    setIsDialogOpen(true);
  };

  const openEdit = (item: ContentArticle) => {
    setEditing(item);
    setForm({
      type: item.type,
      locale: item.locale,
      title: item.title,
      slug: item.slug,
      excerpt: item.excerpt,
      body: item.body,
      authorName: item.authorName,
      imageUrl: item.imageUrl,
      videoUrl: item.videoUrl,
      duration: item.duration,
      viewCount: item.viewCount,
      supplierProfileId: item.supplierProfileId,
      supplierProductId: item.supplierProductId,
      status: item.status,
      isFeatured: item.isFeatured,
      sortOrder: item.sortOrder,
    });
    setIsDialogOpen(true);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="mt-0.5 text-xs text-muted-foreground">{t("subtitle")}</p>
        </div>
        <div className="flex gap-2">
          <Button onClick={openCreate} className="cursor-pointer gap-2 font-bold">
            <Plus className="h-4 w-4" />
            {t("newContent")}
          </Button>
          <Button
            onClick={() => contentQuery.refetch()}
            variant="outline"
            className="cursor-pointer gap-2"
          >
            <RefreshCw className={`h-4 w-4 ${contentQuery.isFetching ? "animate-spin" : ""}`} />
            {t("refresh")}
          </Button>
        </div>
      </div>

      <Card className="flex flex-col gap-4 border border-border bg-card p-4 md:flex-row md:items-center md:justify-between">
        <div className="relative w-full md:w-96">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={t("searchPlaceholder")}
            className="pl-9"
          />
        </div>
        <div className="flex flex-wrap gap-2">
          <Select value={selectedType} onValueChange={(value) => setSelectedType(value as ContentType | "all")}>
            <SelectTrigger className="w-44 cursor-pointer">
              <SelectValue placeholder={t("type")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              {contentTypes.map((type) => (
                <SelectItem key={type} value={type} className="cursor-pointer">
                  {contentTypeLabels[type].id}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select value={selectedStatus} onValueChange={(value) => setSelectedStatus(value as ContentStatus | "all")}>
            <SelectTrigger className="w-40 cursor-pointer">
              <SelectValue placeholder={t("status")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              {statuses.map((status) => (
                <SelectItem key={status} value={status} className="cursor-pointer">
                  {status}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </Card>

      <Card className="overflow-hidden border border-border bg-card">
        {contentQuery.isLoading ? (
          <div className="py-24 text-center">
            <Loader2 className="mx-auto mb-3 h-8 w-8 animate-spin text-primary" />
            <span className="text-sm font-semibold text-muted-foreground">Loading...</span>
          </div>
        ) : rows.length === 0 ? (
          <div className="py-24 text-center">
            <FileText className="mx-auto mb-3 h-10 w-10 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">{t("empty")}</p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="bg-muted/20">
                <TableHead className="pl-6">{t("titleColumn")}</TableHead>
                <TableHead>{t("type")}</TableHead>
                <TableHead>{t("locale")}</TableHead>
                <TableHead>{t("status")}</TableHead>
                <TableHead>{t("featured")}</TableHead>
                <TableHead>{t("updated")}</TableHead>
                <TableHead className="pr-6 text-right">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="pl-6">
                    <div className="max-w-md">
                      <p className="truncate font-bold text-foreground">{item.title}</p>
                      <p className="truncate text-xs text-muted-foreground">{item.slug}</p>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{contentTypeLabels[item.type].id}</Badge>
                  </TableCell>
                  <TableCell className="uppercase">{item.locale}</TableCell>
                  <TableCell>
                    <Badge variant={item.status === "published" ? "default" : "outline"}>{item.status}</Badge>
                  </TableCell>
                  <TableCell>{item.isFeatured ? "Ya" : "-"}</TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {new Intl.DateTimeFormat("id-ID", { day: "2-digit", month: "short", year: "numeric" }).format(new Date(item.updatedAt))}
                  </TableCell>
                  <TableCell className="pr-6 text-right">
                    <div className="flex justify-end gap-1">
                      <Button variant="ghost" size="icon" onClick={() => openEdit(item)} className="h-8 w-8 cursor-pointer">
                        <Edit2 className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => setDeleteId(item.id)} className="h-8 w-8 cursor-pointer text-destructive">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </Card>

      <Dialog open={isDialogOpen} onOpenChange={(open) => {
        setIsDialogOpen(open);
        if (!open) {
          setEditing(null);
          setForm(emptyPayload);
        }
      }}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("editTitle") : t("createTitle")}</DialogTitle>
          </DialogHeader>
          <form
            className="space-y-4"
            onSubmit={(event) => {
              event.preventDefault();
              saveMutation.mutate(form);
            }}
          >
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Field>
                <FieldLabel>{t("type")}</FieldLabel>
                <Select value={form.type} onValueChange={(value) => setForm((prev) => ({ ...prev, type: value as ContentType }))}>
                  <SelectTrigger className="cursor-pointer"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {contentTypes.map((type) => <SelectItem key={type} value={type} className="cursor-pointer">{contentTypeLabels[type].id}</SelectItem>)}
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>{t("status")}</FieldLabel>
                <Select value={form.status} onValueChange={(value) => setForm((prev) => ({ ...prev, status: value as ContentStatus }))}>
                  <SelectTrigger className="cursor-pointer"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    {statuses.map((status) => <SelectItem key={status} value={status} className="cursor-pointer">{status}</SelectItem>)}
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>{t("locale")}</FieldLabel>
                <Select value={form.locale} onValueChange={(value) => setForm((prev) => ({ ...prev, locale: value as ContentArticlePayload["locale"] }))}>
                  <SelectTrigger className="cursor-pointer"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="id" className="cursor-pointer">ID</SelectItem>
                    <SelectItem value="en" className="cursor-pointer">EN</SelectItem>
                    <SelectItem value="both" className="cursor-pointer">ID + EN</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
              <Field>
                <FieldLabel>{t("sortOrderLabel")}</FieldLabel>
                <Input type="number" value={form.sortOrder} onChange={(event) => setForm((prev) => ({ ...prev, sortOrder: Number(event.target.value) }))} />
              </Field>
            </div>

            <Field>
              <FieldLabel>{t("titleLabel")}</FieldLabel>
              <Input value={form.title} onChange={(event) => setForm((prev) => ({ ...prev, title: event.target.value }))} required />
            </Field>
            <Field>
              <FieldLabel>{t("slugLabel")}</FieldLabel>
              <Input value={form.slug || ""} onChange={(event) => setForm((prev) => ({ ...prev, slug: event.target.value }))} />
            </Field>
            <Field>
              <FieldLabel>{t("excerptLabel")}</FieldLabel>
              <Textarea value={form.excerpt || ""} onChange={(event) => setForm((prev) => ({ ...prev, excerpt: event.target.value }))} rows={3} />
            </Field>
            <Field>
              <FieldLabel>{t("bodyLabel")}</FieldLabel>
              <Textarea value={form.body || ""} onChange={(event) => setForm((prev) => ({ ...prev, body: event.target.value }))} rows={5} />
            </Field>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Field>
                <FieldLabel>{t("authorLabel")}</FieldLabel>
                <Input value={form.authorName || ""} onChange={(event) => setForm((prev) => ({ ...prev, authorName: event.target.value }))} />
              </Field>
              <Field>
                <FieldLabel>{t("imageUrlLabel")}</FieldLabel>
                <Input value={form.imageUrl || ""} onChange={(event) => setForm((prev) => ({ ...prev, imageUrl: event.target.value }))} />
              </Field>
              <Field>
                <FieldLabel>{t("videoUrlLabel")}</FieldLabel>
                <Input value={form.videoUrl || ""} onChange={(event) => setForm((prev) => ({ ...prev, videoUrl: event.target.value }))} />
              </Field>
              <Field>
                <FieldLabel>{t("durationLabel")}</FieldLabel>
                <Input value={form.duration || ""} onChange={(event) => setForm((prev) => ({ ...prev, duration: event.target.value }))} placeholder="05:42" />
              </Field>
              <Field>
                <FieldLabel>{t("viewCountLabel")}</FieldLabel>
                <Input type="number" value={form.viewCount || 0} onChange={(event) => setForm((prev) => ({ ...prev, viewCount: Number(event.target.value) }))} />
              </Field>
              <label className="flex cursor-pointer items-center gap-2 pt-7 text-sm font-semibold text-foreground">
                <input
                  type="checkbox"
                  checked={form.isFeatured}
                  onChange={(event) => setForm((prev) => ({ ...prev, isFeatured: event.target.checked }))}
                  className="h-4 w-4 cursor-pointer"
                />
                {t("featured")}
              </label>
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => {
                setIsDialogOpen(false);
                setEditing(null);
                setForm(emptyPayload);
              }} className="cursor-pointer">
                {t("cancel")}
              </Button>
              <Button type="submit" disabled={saveMutation.isPending} className="cursor-pointer">
                {saveMutation.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
                {t("save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <DeleteDialog
        open={!!deleteId}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title={t("deleteTitle")}
        description={t("deleteDescription")}
        itemName="content"
        isLoading={deleteMutation.isPending}
        onConfirm={async () => {
          if (deleteId) await deleteMutation.mutateAsync(deleteId);
        }}
      />
    </div>
  );
}

function cleanPayload(payload: ContentArticlePayload): ContentArticlePayload {
  return {
    ...payload,
    slug: payload.slug?.trim() || undefined,
    excerpt: payload.excerpt?.trim() || undefined,
    body: payload.body?.trim() || undefined,
    authorName: payload.authorName?.trim() || undefined,
    imageUrl: payload.imageUrl?.trim() || undefined,
    videoUrl: payload.videoUrl?.trim() || undefined,
    duration: payload.duration?.trim() || undefined,
    supplierProfileId: payload.supplierProfileId?.trim() || undefined,
    supplierProductId: payload.supplierProductId?.trim() || undefined,
  };
}

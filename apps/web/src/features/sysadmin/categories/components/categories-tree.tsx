"use client";

import React, { useEffect, useState, useCallback } from "react";
import { categoryService } from "@/features/sysadmin/categories/services";
import type { Category } from "@/features/sysadmin/categories/types";
import { toast } from "sonner";
import { useTranslations } from "next-intl";
import {
  Trash2,
  RefreshCw,
  Inbox,
  Plus,
  FolderTree,
  Edit2,
  Check,
  X,
  Loader2
} from "lucide-react";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function CategoriesTree() {
  const t = useTranslations("sysadminCategories");
  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  
  // Creation modal state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newName, setNewName] = useState("");
  const [newSlug, setNewSlug] = useState("");
  const [newParent, setNewParent] = useState<string | null>(null);
  const [newSortOrder, setNewSortOrder] = useState(1);

  // Edit sort order inline state
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editOrderVal, setEditOrderVal] = useState<number>(1);

  const fetchCategories = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await categoryService.list();
      setCategories(data);
    } catch {
      toast.error(t("errorLoad"));
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchCategories();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchCategories]);

  const handleToggleActive = async (id: string, currentVal: boolean) => {
    try {
      await categoryService.update(id, { active: !currentVal });
      toast.success(t("successStatus"));
      fetchCategories();
    } catch {
      toast.error(t("errorStatus"));
    }
  };

  const handleSaveOrder = async (id: string) => {
    try {
      await categoryService.update(id, { sortOrder: editOrderVal });
      toast.success(t("successOrder"));
      setEditingId(null);
      fetchCategories();
    } catch {
      toast.error(t("errorOrder"));
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("confirmDelete"))) return;
    try {
      await categoryService.delete(id);
      toast.success(t("successDelete"));
      fetchCategories();
    } catch {
      toast.error(t("errorDelete"));
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim() || !newSlug.trim()) {
      toast.error(t("fieldsRequired"));
      return;
    }
    try {
      await categoryService.create({
        name: newName,
        slug: newSlug,
        parentName: newParent || null,
        active: true,
        sortOrder: Number(newSortOrder)
      });
      toast.success(t("successCreate"));
      setIsCreateOpen(false);
      setNewName("");
      setNewSlug("");
      setNewParent(null);
      fetchCategories();
    } catch {
      toast.error(t("errorCreate"));
    }
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div className="text-left">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">{t("title")}</h1>
          <p className="text-xs text-muted-foreground mt-0.5">{t("subtitle")}</p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            onClick={() => setIsCreateOpen(true)}
            className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold text-xs uppercase tracking-wider py-5 px-4 rounded-lg flex items-center gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md cursor-pointer"
          >
            <Plus className="h-4 w-4" />
            {t("newCategory")}
          </Button>
          <Button
            onClick={fetchCategories}
            variant="outline"
            size="sm"
            className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
            {t("refresh")}
          </Button>
        </div>
      </div>

      {/* Content Card */}
      <Card className="border border-border/80 shadow-sm overflow-hidden bg-card">
        {/* Table Content */}
        {isLoading ? (
          <div className="py-24 text-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
            <span className="text-sm font-semibold text-muted-foreground">{t("fetching")}</span>
          </div>
        ) : categories.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">{t("noCategories")}</h3>
            <p className="text-muted-foreground text-sm max-w-sm mx-auto">
              {t("noCategoriesDesc")}
            </p>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("categoryName")}</TableHead>
                <TableHead className="font-semibold">{t("slug")}</TableHead>
                <TableHead className="font-semibold">{t("parentCategory")}</TableHead>
                <TableHead className="font-semibold">{t("sortOrder")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {categories.map((cat) => (
                <TableRow key={cat.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  {/* Category Name */}
                  <TableCell className="pl-6 py-4">
                    <div className="text-left flex items-center gap-2">
                      <FolderTree className="h-4 w-4 text-primary shrink-0" />
                      <span className="font-bold text-foreground">{cat.name}</span>
                    </div>
                  </TableCell>

                  {/* Slug */}
                  <TableCell className="py-4 text-left font-mono text-xs text-muted-foreground">
                    {cat.slug}
                  </TableCell>

                  {/* Parent */}
                  <TableCell className="py-4 text-left font-medium text-foreground">
                    {cat.parentName ? (
                      <Badge className="bg-slate-50 text-slate-600 border border-slate-200 dark:bg-muted dark:text-muted-foreground">
                        {cat.parentName}
                      </Badge>
                    ) : (
                      <span className="text-xs text-muted-foreground italic">{t("mainCategory")}</span>
                    )}
                  </TableCell>

                  {/* Sort Order */}
                  <TableCell className="py-4 text-left">
                    {editingId === cat.id ? (
                      <div className="flex items-center gap-1.5">
                        <Input
                          type="number"
                          value={editOrderVal}
                          onChange={(e) => setEditOrderVal(Number(e.target.value))}
                          className="w-16 h-7 text-xs px-2 border-border"
                          min={1}
                        />
                        <button
                          onClick={() => handleSaveOrder(cat.id)}
                          className="h-7 w-7 flex items-center justify-center bg-primary text-primary-foreground rounded hover:bg-primary/90 cursor-pointer"
                        >
                          <Check className="h-3.5 w-3.5" />
                        </button>
                        <button
                          onClick={() => setEditingId(null)}
                          className="h-7 w-7 flex items-center justify-center bg-secondary text-secondary-foreground rounded hover:bg-secondary/80 cursor-pointer"
                        >
                          <X className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    ) : (
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-foreground text-sm">{cat.sortOrder}</span>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => {
                            setEditingId(cat.id);
                            setEditOrderVal(cat.sortOrder);
                          }}
                          className="h-6 w-6 text-muted-foreground hover:text-primary cursor-pointer"
                        >
                          <Edit2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    )}
                  </TableCell>

                  {/* Active status */}
                  <TableCell className="py-4 text-left">
                    <button
                      onClick={() => handleToggleActive(cat.id, cat.active)}
                      className={`px-3 py-1 text-xs font-semibold rounded-full border cursor-pointer select-none transition-all ${
                        cat.active
                          ? "bg-success/15 text-success border border-success/30"
                          : "bg-muted text-muted-foreground border-border"
                      }`}
                    >
                      {cat.active ? t("active") : t("inactive")}
                    </button>
                  </TableCell>

                  {/* Actions */}
                  <TableCell className="pr-6 py-4 text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(cat.id)}
                      className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 cursor-pointer"
                      title={t("delete")}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </Card>

      {/* Create Category Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent className="max-w-md bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-primary font-bold mb-2">
              <FolderTree className="h-5 w-5" />
              {t("newCategory")}
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {t("dialogTitle")}
            </DialogTitle>
            <DialogDescription className="text-xs text-muted-foreground">
              {t("dialogSubtitle")}
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handleCreate} className="space-y-4 my-2 text-left">
            <FieldGroup className="space-y-3">
              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("categoryName")}</FieldLabel>
                <Input
                  type="text"
                  placeholder="e.g. Furnitur Jati, Rempah Kering"
                  value={newName}
                  onChange={(e) => {
                    setNewName(e.target.value);
                    setNewSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, "-"));
                  }}
                  className="bg-background border-border"
                />
              </Field>

              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("slug")}</FieldLabel>
                <Input
                  type="text"
                  placeholder="category-slug-auto"
                  value={newSlug}
                  onChange={(e) => setNewSlug(e.target.value)}
                  className="bg-background border-border"
                />
              </Field>

              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("parentCategory")} ({t("active")})</FieldLabel>
                <Select
                  value={newParent || "none"}
                  onValueChange={(val) => setNewParent(val === "none" ? null : val)}
                >
                  <SelectTrigger className="bg-background border-border cursor-pointer">
                    <SelectValue placeholder={t("parentCategory")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none" className="cursor-pointer">{t("parentNone")}</SelectItem>
                    {categories.filter(c => !c.parentName).map((c) => (
                      <SelectItem key={c.id} value={c.name} className="cursor-pointer">{c.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </Field>

              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("sortOrder")}</FieldLabel>
                <Input
                  type="number"
                  value={newSortOrder}
                  onChange={(e) => setNewSortOrder(Number(e.target.value))}
                  min={1}
                  className="bg-background border-border"
                />
              </Field>
            </FieldGroup>

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)} className="border-border cursor-pointer text-foreground">
                {t("batal")}
              </Button>
              <Button type="submit" className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold cursor-pointer">
                {t("simpan")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

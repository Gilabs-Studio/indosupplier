"use client";

import React, { useEffect, useState, useCallback } from "react";
import { faqService } from "@/features/sysadmin/faq/services";
import type { FAQArticle } from "@/features/sysadmin/faq/types";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  Trash2,
  RefreshCw,
  Inbox,
  Plus,
  BookOpen,
  Edit2,
  Eye,
  Search,
  Filter,
  Loader2,
  HelpCircle,
  Globe
} from "lucide-react";

import { Card } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Field, FieldLabel, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export default function FAQManagement() {
  const t = useTranslations("sysadminFaq");
  const [articles, setArticles] = useState<FAQArticle[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  
  // Search & Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedTopic, setSelectedTopic] = useState<string>("all");
  const [selectedLang, setSelectedLang] = useState<string>("all");

  // Create/Edit Dialog state
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingArticle, setEditingArticle] = useState<FAQArticle | null>(null);
  
  // Form values
  const [title, setTitle] = useState("");
  const [slug, setSlug] = useState("");
  const [topic, setTopic] = useState("");
  const [language, setLanguage] = useState<"id" | "en" | "both">("id");
  const [content, setContent] = useState("");
  const [active, setActive] = useState(true);

  // Preview Dialog state
  const [isPreviewOpen, setIsPreviewOpen] = useState(false);
  const [previewArticle, setPreviewArticle] = useState<FAQArticle | null>(null);

  const fetchArticles = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await faqService.list();
      setArticles(data);
    } catch {
      toast.error(t("subtitle")); // Fallback or direct toast
    } finally {
      setIsLoading(false);
    }
  }, [t]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchArticles();
    }, 0);
    return () => clearTimeout(timer);
  }, [fetchArticles]);

  const handleOpenCreate = () => {
    setEditingArticle(null);
    setTitle("");
    setSlug("");
    setTopic("");
    setLanguage("id");
    setContent("");
    setActive(true);
    setIsDialogOpen(true);
  };

  const handleOpenEdit = (article: FAQArticle) => {
    setEditingArticle(article);
    setTitle(article.title);
    setSlug(article.slug);
    setTopic(article.topic);
    setLanguage(article.language);
    setContent(article.content);
    setActive(article.active);
    setIsDialogOpen(true);
  };

  const handleOpenPreview = (article: FAQArticle) => {
    setPreviewArticle(article);
    setIsPreviewOpen(true);
  };

  const handleToggleActive = async (id: string, currentVal: boolean) => {
    try {
      await faqService.update(id, { active: !currentVal });
      toast.success(t("successSave"));
      fetchArticles();
    } catch {
      toast.error("Error");
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm(t("delete") + "?")) return;
    try {
      await faqService.delete(id);
      toast.success(t("successDelete"));
      fetchArticles();
    } catch {
      toast.error("Error");
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim() || !slug.trim() || !topic.trim() || !content.trim()) {
      toast.error("Fields required");
      return;
    }

    try {
      if (editingArticle) {
        await faqService.update(editingArticle.id, {
          title,
          slug,
          topic,
          language,
          content,
          active
        });
      } else {
        await faqService.create({
          title,
          slug,
          topic,
          language,
          content,
          active
        });
      }
      toast.success(t("successSave"));
      setIsDialogOpen(false);
      fetchArticles();
    } catch {
      toast.error("Error");
    }
  };

  // Filter logic
  const filteredArticles = articles.filter(a => {
    const matchesSearch = a.title.toLowerCase().includes(searchQuery.toLowerCase()) || 
                          a.content.toLowerCase().includes(searchQuery.toLowerCase()) ||
                          a.topic.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesTopic = selectedTopic === "all" || a.topic === selectedTopic;
    const matchesLang = selectedLang === "all" || a.language === selectedLang || a.language === "both";
    return matchesSearch && matchesTopic && matchesLang;
  });

  // Extract unique topics for filter select
  const uniqueTopics = Array.from(new Set(articles.map(a => a.topic)));

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
            onClick={handleOpenCreate}
            className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold text-xs uppercase tracking-wider py-5 px-4 rounded-lg flex items-center gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md cursor-pointer"
          >
            <Plus className="h-4 w-4" />
            {t("newArticle")}
          </Button>
          <Button
            onClick={fetchArticles}
            variant="outline"
            size="sm"
            className="gap-1.5 hover:-translate-y-0.5 active:translate-y-0 transition-all duration-300 hover:shadow-md border-border cursor-pointer"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
            {t("refresh")}
          </Button>
        </div>
      </div>

      {/* Filters Bar */}
      <Card className="p-4 border border-border bg-card flex flex-col md:flex-row gap-4 items-center justify-between">
        <div className="relative w-full md:w-80">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t("searchPlaceholder")}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 bg-background border-border text-sm"
          />
        </div>
        <div className="flex flex-wrap w-full md:w-auto items-center gap-3">
          {/* Topic filter */}
          <div className="flex items-center gap-2">
            <Filter className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">{t("topic")}:</span>
          </div>
          <Select value={selectedTopic} onValueChange={setSelectedTopic}>
            <SelectTrigger className="w-[150px] bg-background border-border text-xs h-9 cursor-pointer">
              <SelectValue placeholder={t("all")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              {uniqueTopics.map((topic) => (
                <SelectItem key={topic} value={topic} className="cursor-pointer">{topic}</SelectItem>
              ))}
            </SelectContent>
          </Select>

          {/* Language filter */}
          <div className="flex items-center gap-2">
            <Globe className="h-3.5 w-3.5 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">{t("language")}:</span>
          </div>
          <Select value={selectedLang} onValueChange={setSelectedLang}>
            <SelectTrigger className="w-[130px] bg-background border-border text-xs h-9 cursor-pointer">
              <SelectValue placeholder={t("all")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="cursor-pointer">{t("all")}</SelectItem>
              <SelectItem value="id" className="cursor-pointer">Indonesia (ID)</SelectItem>
              <SelectItem value="en" className="cursor-pointer">English (EN)</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </Card>

      {/* Content Card */}
      <Card className="border border-border/80 shadow-sm overflow-hidden bg-card">
        {isLoading ? (
          <div className="py-24 text-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
            <span className="text-sm font-semibold text-muted-foreground">Loading...</span>
          </div>
        ) : filteredArticles.length === 0 ? (
          <div className="py-24 text-center space-y-2">
            <Inbox className="h-12 w-12 text-muted-foreground/30 mx-auto" />
            <h3 className="font-bold text-lg text-foreground">Empty</h3>
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow className="border-b border-border bg-muted/10">
                <TableHead className="pl-6 font-semibold">{t("articleTitle")}</TableHead>
                <TableHead className="font-semibold">{t("slug")}</TableHead>
                <TableHead className="font-semibold text-center">{t("language")}</TableHead>
                <TableHead className="font-semibold">{t("status")}</TableHead>
                <TableHead className="pr-6 text-right font-semibold">{t("actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredArticles.map((art) => (
                <TableRow key={art.id} className="hover:bg-muted/10 border-b border-border transition-colors duration-200">
                  <TableCell className="pl-6 py-4">
                    <div className="text-left space-y-1">
                      <div className="font-bold text-foreground hover:text-primary transition-colors flex items-center gap-1.5">
                        <HelpCircle className="h-4 w-4 text-primary shrink-0" />
                        <span>{art.title}</span>
                      </div>
                      <div className="text-[10px]">
                        <span className="bg-muted text-muted-foreground font-bold px-2 py-0.5 rounded">
                          {art.topic}
                        </span>
                      </div>
                    </div>
                  </TableCell>

                  <TableCell className="py-4 text-left font-mono text-xs text-muted-foreground">
                    {art.slug}
                  </TableCell>

                  <TableCell className="py-4 text-center">
                    <Badge className="uppercase text-[10px] px-2 py-0.5 bg-secondary text-secondary-foreground">
                      {art.language === "both" ? "ID + EN" : art.language}
                    </Badge>
                  </TableCell>

                  <TableCell className="py-4 text-left">
                    <button
                      onClick={() => handleToggleActive(art.id, art.active)}
                      className={`px-3 py-1 text-xs font-semibold rounded-full border cursor-pointer select-none transition-all ${
                        art.active
                          ? "bg-success/15 text-success border-success/30"
                          : "bg-muted text-muted-foreground border-border"
                      }`}
                    >
                      {art.active ? t("active") : t("draft")}
                    </button>
                  </TableCell>

                  <TableCell className="pr-6 py-4 text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleOpenPreview(art)}
                        className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-muted cursor-pointer"
                        title={t("preview")}
                      >
                        <Eye className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleOpenEdit(art)}
                        className="h-8 w-8 text-muted-foreground hover:text-primary hover:bg-muted cursor-pointer"
                        title={t("edit")}
                      >
                        <Edit2 className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(art.id)}
                        className="h-8 w-8 text-muted-foreground hover:text-destructive hover:bg-destructive/10 cursor-pointer"
                        title={t("delete")}
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
      </Card>

      {/* Create / Edit Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="max-w-xl bg-card border-border text-foreground">
          <DialogHeader className="text-left">
            <div className="flex items-center gap-2 text-primary font-bold mb-1">
              <BookOpen className="h-5 w-5" />
              {editingArticle ? t("edit") : t("newArticle")}
            </div>
            <DialogTitle className="text-lg font-bold text-foreground">
              {t("title")}
            </DialogTitle>
          </DialogHeader>

          <form onSubmit={handleSubmit} className="space-y-4 my-2 text-left">
            <FieldGroup className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("articleTitle")}</FieldLabel>
                  <Input
                    type="text"
                    placeholder="Title"
                    value={title}
                    onChange={(e) => {
                      setTitle(e.target.value);
                      if (!editingArticle) {
                        setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, "-"));
                      }
                    }}
                    className="bg-background border-border"
                  />
                </Field>

                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("slug")}</FieldLabel>
                  <Input
                    type="text"
                    placeholder="slug"
                    value={slug}
                    onChange={(e) => setSlug(e.target.value)}
                    className="bg-background border-border text-xs font-mono"
                  />
                </Field>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("topic")}</FieldLabel>
                  <Input
                    type="text"
                    placeholder="Topic"
                    value={topic}
                    onChange={(e) => setTopic(e.target.value)}
                    className="bg-background border-border"
                  />
                </Field>

                <Field>
                  <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("language")}</FieldLabel>
                  <Select
                    value={language}
                    onValueChange={(val: "id" | "en" | "both") => setLanguage(val)}
                  >
                    <SelectTrigger className="bg-background border-border cursor-pointer">
                      <SelectValue placeholder={t("language")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="id" className="cursor-pointer">Indonesia (ID Only)</SelectItem>
                      <SelectItem value="en" className="cursor-pointer">English (EN Only)</SelectItem>
                      <SelectItem value="both" className="cursor-pointer">ID + EN</SelectItem>
                    </SelectContent>
                  </Select>
                </Field>
              </div>

              <Field>
                <FieldLabel className="text-xs font-semibold uppercase text-muted-foreground">{t("content")}</FieldLabel>
                <Textarea
                  placeholder="Instruksi..."
                  value={content}
                  onChange={(e) => setContent(e.target.value)}
                  className="bg-background border-border min-h-[150px]"
                />
              </Field>

              <div className="flex items-center gap-2 pt-1">
                <input
                  type="checkbox"
                  id="active-checkbox"
                  checked={active}
                  onChange={(e) => setActive(e.target.checked)}
                  className="h-4 w-4 text-primary focus:ring-primary border-border rounded cursor-pointer"
                />
                <label htmlFor="active-checkbox" className="text-sm font-semibold select-none text-foreground cursor-pointer">
                  {t("publish")}
                </label>
              </div>
            </FieldGroup>

            <DialogFooter className="pt-2">
              <Button type="button" variant="outline" onClick={() => setIsDialogOpen(false)} className="border-border cursor-pointer text-foreground">
                {t("cancel")}
              </Button>
              <Button type="submit" className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold cursor-pointer hover:-translate-y-0.5 active:translate-y-0 transition-transform">
                {t("saveArticle")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Preview Dialog */}
      <Dialog open={isPreviewOpen} onOpenChange={setIsPreviewOpen}>
        <DialogContent className="max-w-xl bg-card border-border text-foreground">
          <DialogHeader className="text-left border-b border-border pb-3">
            <div className="text-[10px] uppercase font-bold text-primary tracking-wider">
              {t("preview")}
            </div>
            <DialogTitle className="text-xl font-bold text-foreground">
              {previewArticle?.title}
            </DialogTitle>
            <div className="flex items-center gap-2 mt-1">
              <Badge className="bg-muted text-muted-foreground border-none font-bold text-[10px] px-2">
                {previewArticle?.topic}
              </Badge>
              <span className="text-xs text-muted-foreground">
                {t("language")}: <span className="font-semibold uppercase">{previewArticle?.language}</span>
              </span>
            </div>
          </DialogHeader>

          <div className="my-4 py-2 text-left space-y-4 max-h-[300px] overflow-y-auto">
            <div className="text-sm leading-relaxed text-foreground whitespace-pre-line font-normal">
              {previewArticle?.content}
            </div>
          </div>

          <DialogFooter className="border-t border-border pt-3">
            <Button
              type="button"
              onClick={() => setIsPreviewOpen(false)}
              className="bg-primary hover:bg-primary/90 text-primary-foreground font-bold cursor-pointer"
            >
              {t("close")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

"use client";

import React, { useState } from "react";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { Star, MessageCircle } from "lucide-react";

export function SupplierReviewsList() {
  const t = useTranslations("supplier.reviews");

  const [reviews, setReviews] = useState([
    { id: "REV-01", buyer: "PT Metalindo Utama", rating: 5, comment: "Bahan garnet sand berkualitas tinggi, pengiriman cepat, packing rapi menggunakan jumbo bag. Sangat direkomendasikan!", date: "2026-06-02", reply: "Terima kasih atas ulasannya! Kami selalu berupaya menjaga kualitas produk dan layanan logistik terbaik." },
    { id: "REV-02", buyer: "CV Borneo Abadi", rating: 4, comment: "Kualitas bentonite clay bagus, mengembang sempurna untuk drilling mud. Sayang respons logistik pelabuhan agak lambat sedikit.", date: "2026-05-28", reply: "" },
    { id: "REV-03", buyer: "PT Jakarta Chemical", rating: 5, comment: "Mineral quartz sangat murni, 325 mesh sesuai spesifikasi. Komunikasi sales sangat responsif.", date: "2026-05-15", reply: "" },
  ]);

  const [activeReviewId, setActiveReviewId] = useState<string | null>(null);
  const [replyText, setReplyText] = useState("");

  const handleReplySubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!replyText.trim()) return;

    setReviews(reviews.map(rev => rev.id === activeReviewId ? { ...rev, reply: replyText } : rev));
    toast.success(t("replySuccess"));
    setActiveReviewId(null);
    setReplyText("");
  };

  const getStars = (count: number) => {
    return Array.from({ length: 5 }).map((_, idx) => (
      <Star
        key={idx}
        className={`h-4 w-4 ${
          idx < count ? "fill-amber-400 stroke-amber-400" : "text-muted border-muted"
        }`}
      />
    ));
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

      {/* Overview Card */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card className="border border-border shadow-xs rounded-xl bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs uppercase tracking-wider text-muted-foreground">{t("totalReviews")}</CardTitle>
          </CardHeader>
          <CardContent className="flex items-center justify-between">
            <span className="text-2xl font-extrabold text-foreground">3 Reviews</span>
            <div className="h-10 w-10 bg-primary/10 text-primary border border-border rounded-lg flex items-center justify-center">
              <MessageCircle className="h-5 w-5" />
            </div>
          </CardContent>
        </Card>

        <Card className="border border-border shadow-xs rounded-xl bg-card">
          <CardHeader className="pb-2">
            <CardTitle className="text-xs uppercase tracking-wider text-muted-foreground">{t("avgRating")}</CardTitle>
          </CardHeader>
          <CardContent className="flex items-center gap-3">
            <span className="text-2xl font-extrabold text-foreground">4.7</span>
            <div className="flex gap-0.5">{getStars(5)}</div>
          </CardContent>
        </Card>
      </div>

      {/* Reviews List */}
      <div className="space-y-4">
        {reviews.map((rev) => (
          <Card key={rev.id} className="border border-border shadow-xs rounded-xl overflow-hidden bg-card p-6 space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-border pb-3">
              <div className="space-y-1">
                <h4 className="text-sm font-bold text-foreground">{rev.buyer}</h4>
                <div className="flex items-center gap-2 flex-wrap">
                  <div className="flex gap-0.5">{getStars(rev.rating)}</div>
                  <span className="text-[10px] text-muted-foreground font-semibold">{rev.date}</span>
                </div>
              </div>
              <Badge variant="secondary" className="font-bold text-[10px] uppercase max-w-fit">{rev.id}</Badge>
            </div>

            <p className="text-sm text-foreground leading-relaxed font-semibold bg-muted/10 p-3 rounded-lg border border-border">
              {rev.comment}
            </p>

            {/* Replies section */}
            {rev.reply ? (
              <div className="bg-primary/5 border border-primary/10 p-4 rounded-xl space-y-1 ml-4">
                <span className="text-[10px] font-bold text-primary uppercase tracking-wider block">Your Response:</span>
                <p className="text-xs text-foreground font-semibold leading-relaxed">
                  {rev.reply}
                </p>
              </div>
            ) : (
              <div className="flex justify-end">
                <Button onClick={() => setActiveReviewId(rev.id)} size="sm" className="text-xs font-semibold h-8 cursor-pointer">
                  {t("btnReply")}
                </Button>
              </div>
            )}
          </Card>
        ))}
      </div>

      {/* Reply Modal */}
      {activeReviewId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs">
          <Card className="max-w-md w-full border border-border bg-card shadow-2xl rounded-2xl overflow-hidden animate-in fade-in zoom-in duration-200">
            <CardHeader className="border-b border-border">
              <CardTitle className="text-base font-bold font-heading">{t("replyModalTitle")}</CardTitle>
              <CardDescription className="text-xs">Post a reply to the buyer&apos;s review.</CardDescription>
            </CardHeader>
            <form onSubmit={handleReplySubmit}>
              <CardContent className="p-6">
                <textarea
                  required
                  placeholder={t("replyPlaceholder")}
                  value={replyText}
                  onChange={(e) => setReplyText(e.target.value)}
                  rows={4}
                  className="w-full px-3 py-2 bg-card border border-border text-sm rounded-lg focus:outline-none focus:border-primary transition-all text-left text-foreground font-semibold resize-none"
                />
              </CardContent>
              <div className="p-4 border-t border-border bg-muted/10 flex justify-end gap-2">
                <Button variant="outline" onClick={() => setActiveReviewId(null)} className="text-xs h-9 cursor-pointer border-border">
                  Cancel
                </Button>
                <Button type="submit" className="text-xs h-9 bg-primary text-primary-foreground hover:bg-primary/95 cursor-pointer font-semibold">
                  {t("btnSubmitReply")}
                </Button>
              </div>
            </form>
          </Card>
        </div>
      )}
    </div>
  );
}

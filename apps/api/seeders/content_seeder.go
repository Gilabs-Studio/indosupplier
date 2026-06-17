package seeders

import (
	"fmt"

	"github.com/gilabs/indosupplier/api/internal/content/data/models"
	"github.com/gilabs/indosupplier/api/internal/core/apptime"
	"github.com/gilabs/indosupplier/api/internal/core/infrastructure/database"
)

func SeedContent() error {
	var count int64
	if err := database.DB.Model(&models.ContentArticle{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := apptime.Now()
	articles := []models.ContentArticle{
		{
			Type:        "news",
			Locale:      "id",
			Title:       "Kelompok Kopi Gayo Menambah Kapasitas Sortasi untuk Buyer Ekspor",
			Slug:        "kelompok-kopi-gayo-menambah-kapasitas-sortasi-buyer-ekspor",
			Excerpt:     "Produsen kopi Arabika Gayo meningkatkan proses sortasi dan traceability untuk memenuhi permintaan buyer Asia Timur.",
			Body:        "Peningkatan kapasitas sortasi membantu supplier menjaga konsistensi grade, kadar air, dan dokumentasi asal barang.",
			AuthorName:  "Redaksi IndoSupplier",
			ImageURL:    "https://images.unsplash.com/photo-1447933601403-0c6688de566e?auto=format&fit=crop&w=900&q=80",
			Status:      "published",
			IsFeatured:  true,
			SortOrder:   1,
			PublishedAt: &now,
		},
		{
			Type:        "feature",
			Locale:      "id",
			Title:       "Di Balik Audit Fasilitas Baja: Cara Buyer Membaca Kapasitas Pabrik",
			Slug:        "di-balik-audit-fasilitas-baja-cara-buyer-membaca-kapasitas-pabrik",
			Excerpt:     "Panduan ringkas membaca bukti kapasitas, sertifikasi, dan kesiapan pengiriman dari supplier manufaktur.",
			Body:        "Buyer B2B perlu melihat konsistensi lini produksi, dokumen kualitas, dan respons supplier sebelum mengirim RFQ besar.",
			AuthorName:  "Tim Trust IndoSupplier",
			ImageURL:    "https://images.unsplash.com/photo-1504917595217-d4dc5ebe6122?auto=format&fit=crop&w=900&q=80",
			Status:      "published",
			IsFeatured:  true,
			SortOrder:   2,
			PublishedAt: &now,
		},
		{
			Type:        "tips",
			Locale:      "id",
			Title:       "Checklist RFQ Pertama untuk Produk Bahan Baku",
			Slug:        "checklist-rfq-pertama-untuk-produk-bahan-baku",
			Excerpt:     "Spesifikasi teknis, target harga, toleransi mutu, dan jadwal pengiriman perlu disiapkan sebelum menghubungi supplier.",
			Body:        "RFQ yang lengkap mempercepat balasan supplier dan mengurangi risiko salah spesifikasi pada tahap penawaran.",
			AuthorName:  "Procurement Desk",
			ImageURL:    "https://images.unsplash.com/photo-1454165804606-c3d57bc86b40?auto=format&fit=crop&w=900&q=80",
			Status:      "published",
			SortOrder:   3,
			PublishedAt: &now,
		},
		{
			Type:        "editorial_review",
			Locale:      "id",
			Title:       "Review Redaksi: Konsistensi MOQ Supplier Tekstil untuk Brand Lokal",
			Slug:        "review-redaksi-konsistensi-moq-supplier-tekstil-brand-lokal",
			Excerpt:     "Kami membandingkan beberapa profil supplier tekstil dari sisi MOQ, respons, dan kelengkapan produk.",
			Body:        "MOQ yang realistis dan data kapasitas yang transparan menjadi sinyal penting untuk buyer yang sedang validasi pemasok.",
			AuthorName:  "B2B Editorial",
			ImageURL:    "https://images.unsplash.com/photo-1541099649105-f69ad21f3246?auto=format&fit=crop&w=900&q=80",
			Status:      "published",
			SortOrder:   4,
			PublishedAt: &now,
		},
		{
			Type:        "video",
			Locale:      "id",
			Title:       "Sortasi Manual Biji Kopi Arabika Gayo Grade Ekspor",
			Slug:        "sortasi-manual-biji-kopi-arabika-gayo-grade-ekspor",
			Excerpt:     "Tur singkat proses sortasi dan pengepakan biji kopi Arabika Gayo untuk buyer grosir.",
			AuthorName:  "IndoSupplier Video",
			ImageURL:    "https://images.unsplash.com/photo-1509042239860-f550ce710b93?auto=format&fit=crop&w=900&q=80",
			VideoURL:    "https://example.com/videos/kopi-gayo-sortasi",
			Duration:    "06:12",
			ViewCount:   1240,
			Status:      "published",
			IsFeatured:  true,
			SortOrder:   1,
			PublishedAt: &now,
		},
		{
			Type:        "video",
			Locale:      "id",
			Title:       "Tur Lini Produksi Baja dan Pemeriksaan Kualitas Material",
			Slug:        "tur-lini-produksi-baja-pemeriksaan-kualitas-material",
			Excerpt:     "Cuplikan fasilitas produksi dan proses inspeksi material untuk buyer konstruksi.",
			AuthorName:  "IndoSupplier Video",
			ImageURL:    "https://images.unsplash.com/photo-1581094288338-2314dddb7ecc?auto=format&fit=crop&w=900&q=80",
			VideoURL:    "https://example.com/videos/audit-baja",
			Duration:    "05:48",
			ViewCount:   890,
			Status:      "published",
			SortOrder:   2,
			PublishedAt: &now,
		},
	}

	for i := range articles {
		if err := database.DB.Create(&articles[i]).Error; err != nil {
			return err
		}
	}

	fmt.Printf("seeded %d content articles\n", len(articles))
	return nil
}

package mapper

import (
	"strings"

	buyerModels "github.com/gilabs/indosupplier/api/internal/buyer/data/models"
	"github.com/gilabs/indosupplier/api/internal/buyer/domain/dto"
	supplierModels "github.com/gilabs/indosupplier/api/internal/supplier/data/models"
)

func resolveProductPhoto(url, prodName, catName string) string {
	url = strings.TrimSpace(url)
	if url != "" {
		if strings.HasSuffix(url, ".png") && strings.HasPrefix(url, "/images/") {
			return strings.TrimSuffix(url, ".png") + ".webp"
		}
		return url
	}
	combined := strings.ToLower(prodName + " " + catName)
	switch {
	case strings.Contains(combined, "steel") || strings.Contains(combined, "baja") || strings.Contains(combined, "rebar") || strings.Contains(combined, "plate") || strings.Contains(combined, "denim") || strings.Contains(combined, "yarn"):
		return "/images/categories/cat-bahan-baku.webp"
	case strings.Contains(combined, "coffee") || strings.Contains(combined, "kopi") || strings.Contains(combined, "sugar") || strings.Contains(combined, "makanan"):
		return "/images/categories/cat-makanan-minuman.webp"
	case strings.Contains(combined, "sand") || strings.Contains(combined, "powder") || strings.Contains(combined, "bentonite"):
		return "/images/products/prod-mineral-powder.webp"
	case strings.Contains(combined, "masker") || strings.Contains(combined, "medis"):
		return "/images/products/prod-masker.webp"
	case strings.Contains(combined, "helm") || strings.Contains(combined, "safety"):
		return "/images/products/prod-helmet.webp"
	case strings.Contains(combined, "laptop") || strings.Contains(combined, "elektronik"):
		return "/images/products/prod-laptop.webp"
	case strings.Contains(combined, "pompa") || strings.Contains(combined, "pump") || strings.Contains(combined, "mesin"):
		return "/images/products/prod-water-pump.webp"
	case strings.Contains(combined, "kursi") || strings.Contains(combined, "chair"):
		return "/images/products/prod-office-chair.webp"
	case strings.Contains(combined, "karton") || strings.Contains(combined, "box"):
		return "/images/products/prod-carton-boxes.webp"
	default:
		return "/images/products/prod-hvs.webp"
	}
}

func ToBookmarkResponse(
	b *buyerModels.Bookmark,
	supplier *supplierModels.SupplierProfile,
	categoryName string,
	keyProducts []string,
	product *supplierModels.SupplierProduct,
	establishedYear int,
	slug string,
) dto.BookmarkResponse {
	var bookmarkType string
	var prodName, prodMinOrder, prodImage string
	var prodPrice float64

	if b.SupplierProductID != nil && *b.SupplierProductID != "" {
		bookmarkType = "product"
		if product != nil {
			prodName = product.Name
			prodPrice = product.StartingPrice
			prodMinOrder = product.MOQ
			rawPhoto := ""
			if len(product.Photos) > 0 {
				rawPhoto = product.Photos[0].FileURL
			}
			prodImage = resolveProductPhoto(rawPhoto, prodName, categoryName)
		}
	} else {
		bookmarkType = "supplier"
	}

	supplierID := ""
	companyName := ""
	location := ""
	businessType := ""
	var rating float64
	var reviewCount int
	isVerified := false

	if supplier != nil {
		supplierID = supplier.ID
		companyName = supplier.CompanyName
		location = supplier.CityID
		businessType = supplier.CompanyType
		rating = supplier.StarRating
		reviewCount = supplier.ReviewCount
		isVerified = supplier.VerificationLevel >= 2
	}

	return dto.BookmarkResponse{
		ID:                b.ID,
		SupplierProfileID: supplierID,
		SupplierProductID: b.SupplierProductID,
		Type:              bookmarkType,
		SupplierSlug:      slug,
		CompanyName:       companyName,
		Category:          categoryName,
		Location:          location,
		BusinessType:      businessType,
		EstablishedYear:   establishedYear,
		Rating:            rating,
		ReviewCount:       reviewCount,
		IsVerified:        isVerified,
		KeyProducts:       keyProducts,
		ProductName:       prodName,
		ProductPrice:      prodPrice,
		ProductMinOrder:   prodMinOrder,
		ProductImage:      prodImage,
		CreatedAt:         b.CreatedAt,
	}
}

func ToToggleBookmarkResponse(bookmarked bool, action string, bookmark *dto.BookmarkResponse) *dto.ToggleBookmarkResponse {
	return &dto.ToggleBookmarkResponse{
		Bookmarked: bookmarked,
		Action:     action,
		Bookmark:   bookmark,
	}
}

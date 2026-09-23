/**
 * Utility to reliably map category slugs, names, and icon URLs to valid local static webp assets.
 */

const CATEGORY_IMAGE_MAP: Record<string, string> = {
  // Electronics
  "electronics-it": "/images/categories/cat-elektronik.webp",
  "electronics": "/images/categories/cat-elektronik.webp",
  "elektronik": "/images/categories/cat-elektronik.webp",
  "elektronik-it": "/images/categories/cat-elektronik.webp",

  // Raw Materials / Steel
  "steel-metal": "/images/categories/cat-bahan-baku.webp",
  "bahan-baku": "/images/categories/cat-bahan-baku.webp",
  "manufacturing": "/images/categories/cat-bahan-baku.webp",
  "industrial-minerals": "/images/categories/cat-bahan-baku.webp",

  // Safety & Hygiene
  "safety-k3": "/images/categories/cat-kebersihan-k3.webp",
  "kebersihan-k3": "/images/categories/cat-kebersihan-k3.webp",
  "chemical": "/images/categories/cat-kebersihan-k3.webp",

  // Office & Stationery
  "kantor-atk": "/images/categories/cat-kantor-atk.webp",
  "office-stationery": "/images/categories/cat-kantor-atk.webp",
  "office": "/images/categories/cat-kantor-atk.webp",

  // Furniture
  "furniture": "/images/categories/cat-furniture.webp",
  "furnitur": "/images/categories/cat-furniture.webp",

  // Packaging
  "packaging": "/images/categories/cat-packaging.webp",
  "kemasan": "/images/categories/cat-packaging.webp",

  // Machinery & Industrial
  "machinery-industrial": "/images/categories/cat-mesin-industrial.webp",
  "mesin-industrial": "/images/categories/cat-mesin-industrial.webp",
  "machinery": "/images/categories/cat-mesin-industrial.webp",

  // Agriculture & Food
  "agricultural-products": "/images/categories/cat-makanan-minuman.webp",
  "agriculture": "/images/categories/cat-makanan-minuman.webp",
  "komoditas-tani": "/images/categories/cat-makanan-minuman.webp",
  "makanan-minuman": "/images/categories/cat-makanan-minuman.webp",
  "textiles-fabrics": "/images/categories/cat-makanan-minuman.webp",
  "textile": "/images/categories/cat-makanan-minuman.webp",
};

export function getCategoryThumbnail(slug?: string, iconUrl?: string, name?: string): string {
  // If iconUrl is provided and is a valid local webp path, prefer it
  if (iconUrl && iconUrl.startsWith("/images/") && !iconUrl.includes("cat-undefined")) {
    return iconUrl.replace(/\.png$/, ".webp");
  }

  const normalizedSlug = (slug || "").toLowerCase().trim();
  if (CATEGORY_IMAGE_MAP[normalizedSlug]) {
    return CATEGORY_IMAGE_MAP[normalizedSlug];
  }

  // Fallback checking by name / keyword
  const normalizedName = (name || "").toLowerCase();
  if (normalizedName.includes("elektronik") || normalizedName.includes("it") || normalizedName.includes("komputer") || normalizedName.includes("laptop")) {
    return "/images/categories/cat-elektronik.webp";
  }
  if (normalizedName.includes("kantor") || normalizedName.includes("atk") || normalizedName.includes("stationery")) {
    return "/images/categories/cat-kantor-atk.webp";
  }
  if (normalizedName.includes("k3") || normalizedName.includes("kebersihan") || normalizedName.includes("safety")) {
    return "/images/categories/cat-kebersihan-k3.webp";
  }
  if (normalizedName.includes("mesin") || normalizedName.includes("industrial") || normalizedName.includes("machinery")) {
    return "/images/categories/cat-mesin-industrial.webp";
  }
  if (normalizedName.includes("furniture") || normalizedName.includes("furnitur") || normalizedName.includes("mebel")) {
    return "/images/categories/cat-furniture.webp";
  }
  if (normalizedName.includes("pack") || normalizedName.includes("kemasan") || normalizedName.includes("kardus")) {
    return "/images/categories/cat-packaging.webp";
  }
  if (normalizedName.includes("tani") || normalizedName.includes("kopi") || normalizedName.includes("agri") || normalizedName.includes("makanan")) {
    return "/images/categories/cat-makanan-minuman.webp";
  }
  if (normalizedName.includes("baja") || normalizedName.includes("besi") || normalizedName.includes("steel") || normalizedName.includes("bahan")) {
    return "/images/categories/cat-bahan-baku.webp";
  }

  return "/images/categories/cat-bahan-baku.webp";
}

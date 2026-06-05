# Standar Pengambangan Frontend (Web) — GIMS Platform

Dokumen ini mendefinisikan standar arsitektur, panduan gaya kode, styling, i18n, serta praktik terbaik (best practices) untuk pengembangan frontend aplikasi GIMS menggunakan Next.js 16 (App Router) dan React 19.

---

## 1. Tech Stack Utama
*   **Framework:** Next.js 16.0.3 (App Router & Server Components)
*   **Library UI:** React 19.2.0 & shadcn/ui
*   **Styling:** Tailwind CSS v4 (menggunakan skema tema `@theme` CSS-native)
*   **State Management:**
    *   **Server State:** TanStack Query (React Query) v5 (via Axios client)
    *   **Client/UI State:** Zustand
*   **Forms & Validation:** React Hook Form & Zod
*   **Animasi:** Framer Motion
*   **Internationalization:** `next-intl` (dengan localized routing)

---

## 2. Struktur Folder & Arsitektur Fitur
Aplikasi web ini menggunakan pendekatan **Feature-Based Architecture**. Setiap modul atau domain fungsional diletakkan di dalam folder `src/features/<feature>/`.

### 📂 Anatomi Modul Fitur
Setiap fitur harus mengikuti struktur direktori berikut secara konsisten:

```text
src/features/<feature>/
├── types/             # Deklarasi tipe data (*.d.ts atau *.types.ts)
├── schemas/           # Validasi Zod schema (*.schema.ts)
├── services/          # Call API (fungsi fetch/axios, *.service.ts)
├── hooks/             # Custom React hooks (logika bisnis & query, use*.ts)
├── stores/            # Zustand stores untuk UI state saja (use*Store.ts)
├── components/        # Komponen UI presentasional (UI-only, no business logic)
├── utils/             # Helper fungsi spesifik fitur
└── i18n/              # Berisi file kamus lokalisasi (en.ts, id.ts)
```

> [!IMPORTANT]
> **Pemisahan Logika Bisnis (Component Logic Separation):**
> *   **Komponen UI (`components/`)** hanya boleh berisi kode visual, handling markup, styling, dan visual states.
> *   **Semua logika bisnis**, *fetching* data, *mutation*, koordinasi state, dan efek samping (side effects) **WAJIB** diekstrak ke custom hooks (`hooks/`).

*Contoh Pemisahan Logika:*
```tsx
// ✅ BENAR - Logika bisnis dan fetch didelegasikan ke hook useLogin
export default function LoginForm() {
  const { handleLogin, isLoading } = useLogin();
  return (
    <form onSubmit={handleSubmit(handleLogin)}>
      {/* UI Elements */}
      <Button disabled={isLoading}>Submit</Button>
    </form>
  );
}
```

---

## 3. Standar Styling & Tailwind CSS v4

GIMS menggunakan **Tailwind CSS v4** dengan integrasi CSS-native theme. Pengaturan tema global disimpan di `/app/globals.css`.

### 🎨 Sistem Warna & Tema (Theme Tokens)
*   Hindari penulisan nilai warna secara langsung (arbitrary values) seperti `bg-[#6366f1]` atau `text-[#e53e3e]`.
*   Gunakan variabel warna tema yang sudah didefinisikan di CSS `@theme`:
    *   `bg-primary`, `text-primary-foreground` (Warna utama - Electric Blue `hsl(221 83% 53%)`)
    *   `bg-card`, `text-card-foreground`
    *   `bg-muted`, `text-muted-foreground`
    *   `border-border` / `border-input`
    *   `text-destructive`, `bg-destructive`
    *   Semantic colors: `success` (`--color-success`), `warning` (`--color-warning`), `purple`, `cyan`, `rose`.

### 📐 Aturan Sudut (Border Radius Restriction)
GIMS menerapkan pembatasan radius sudut (rounded corners) secara global untuk menjaga konsistensi UI. Di dalam `globals.css` terdapat rule:
```css
.rounded-full,
.rounded-2xl,
.rounded-3xl,
.rounded-xl {
  border-radius: var(--radius) !important; /* Cap ke 0.5rem (rounded-lg) */
}
```
*   **Aturan:** Semua tombol, input, card, dan kontainer secara visual akan mentok pada radius `var(--radius)` (`0.5rem` / `rounded-lg`). Jangan memaksa menggunakan lengkungan super bulat atau oval kecuali untuk komponen ikon bulat yang spesifik.

### 🖱️ Elemen Interaktif (`cursor-pointer`)
Semua elemen yang dapat diklik atau merespons interaksi pengguna **WAJIB** memiliki kelas `cursor-pointer`.
*   Gunakan `cursor-pointer` pada tag `<Link>`, tombol, checkbox kustom, baris tabel interaktif, atau div berkemampuan klik.

### ✨ Animasi Mikro (Micro-animations)
Gunakan interaksi yang dinamis dan berkelas pada elemen tombol atau navigasi:
*   Tambahkan transisi halus (`transition-all duration-300`).
*   Tambahkan efek hover bergeser sedikit ke atas dan efek klik (active) kembali ke posisi semula:
    `hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg hover:shadow-primary/30`.

---

## 4. Standar Form & Input UI

Untuk menjaga konsistensi form input di seluruh aplikasi, gunakan pustaka `@/components/ui/field` untuk pembungkus input (wrapper).

### 📝 Komponen Field Semantik
Gunakan hierarki field berikut ketika membuat form:
```tsx
import { Field, FieldLabel, FieldDescription, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

// Penggunaan di dalam form:
<FieldGroup className="space-y-4">
  <Field className="space-y-2">
    <FieldLabel htmlFor="email">{t("emailLabel")}</FieldLabel>
    <Input
      id="email"
      type="email"
      placeholder={t("emailPlaceholder")}
      {...register("email")}
    />
    <FieldDescription>{t("emailDesc")}</FieldDescription>
    {errors.email && <FieldError>{errors.email.message}</FieldError>}
  </Field>
</FieldGroup>
```

### 🧠 Otomatisasi Input (`input.tsx`)
Komponen `Input` bawaan (`@/components/ui/input.tsx`) telah diprogram dengan standar:
1.  **Auto-select:** Teks akan otomatis terblok/terseleksi saat input mendapatkan fokus (`onFocus`), kecuali untuk input tipe date, time, color, dan file.
2.  **Filter Numeric Input:** Untuk input tipe `number`, komponen otomatis mencegah input karakter non-angka kecuali tombol navigasi, minus di awal, dan satu buah titik desimal.

### 🗑️ Dialog Konfirmasi Hapus (Delete Confirmation Dialog)
Untuk menjaga konsistensi UI, **DILARANG** menggunakan native `confirm()` browser untuk aksi penghapusan.
- **Wajib menggunakan** komponen `DeleteDialog` dari `@/components/ui/delete-dialog.tsx`.
- Contoh penggunaan:
  ```tsx
  import { DeleteDialog } from "@/components/ui/delete-dialog";
  
  const [deleteId, setDeleteId] = useState<string | null>(null);
  
  <DeleteDialog
    open={!!deleteId}
    onOpenChange={(open) => !open && setDeleteId(null)}
    onConfirm={async () => {
      if (deleteId) {
        await handleDelete(deleteId);
      }
    }}
    itemName="product"
  />
  ```

---

## 5. Lokalisasi & Navigasi Rute (i18n)

Aplikasi GIMS menggunakan localization bawaan dengan dukungan Bahasa Indonesia (`id`) dan Inggris (`en`).

### 🗺️ Aturan Navigasi & Router (Kritis)
Next.js standard router tidak mendukung dynamic locale prefixing secara bawaan. Oleh karena itu:
*   **JANGAN PERNAH** mengimpor `Link`, `redirect`, `usePathname`, atau `useRouter` dari `next/link` atau `next/navigation`.
*   **WAJIB** impor dari modul routing lokalisasi lokal:
    ```tsx
    // ✅ BENAR
    import { Link, useRouter, usePathname, redirect } from "@/i18n/routing";
    
    // ❌ SALAH (akan merusak locale path /id/ atau /en/)
    import Link from "next/link";
    import { useRouter } from "next/navigation";
    ```

### 🗂️ Pembuatan Kamus i18n
*   Kamus global disimpan di `src/i18n/messages/{en,id}.json`.
*   Kamus lokalisasi spesifik fitur diletakkan di `src/features/<feature>/i18n/{en,id}.ts` berupa export objek Javascript.
*   **Wajib Registrasi:** Gabungkan dictionary fitur baru Anda ke dalam file konfigurasi sentral `src/i18n/request.ts` di dalam objek `messages`:
    ```typescript
    // src/i18n/request.ts
    import { yourFeatureEn } from "@/features/your-feature/i18n/en";
    import { yourFeatureId } from "@/features/your-feature/i18n/id";

    const messages = {
      en: {
        ...globalEnMessages,
        yourFeatureKey: yourFeatureEn,
      },
      id: {
        ...globalIdMessages,
        yourFeatureKey: yourFeatureId,
      }
    };
    ```

---

## 6. Pengelolaan State & Pemanggilan API

### 🌐 Data Server (TanStack Query)
*   Gunakan query/mutation hooks untuk semua interaksi database/API eksternal.
*   Tangani status visual dengan benar: `isLoading`, `isError`, `isPending`, serta status kosong (empty state).
*   Gunakan `Axios` client yang ada di `@/lib/api-client.ts` untuk memanggil API dengan auto-refresh JWT token dan auto-redirect jika terdeteksi status `401`.

### 🖥️ Data Client (Zustand & Context)
*   Gunakan Zustand store (`src/features/<feature>/stores/`) hanya untuk UI state global seperti pembukaan modal, sidebar state, preferensi filter, atau cache sementara non-DB.
*   Gunakan context lokal jika state hanya dibagikan dalam radius satu halaman/fitur kecil.

### ⚡ Optimalisasi Kinerja Dashboard
1.  **React Server Components (RSC):** Jadikan halaman sebagai Server Component secara default. Gunakan Client Component (`"use client"`) hanya pada fragmen interaktif terkecil.
2.  **Route-level `loading.tsx`:** Tambahkan file `loading.tsx` di setiap folder rute halaman Next.js untuk memperlihatkan animasi transisi halaman menggunakan `PageMotion` + Skeleton.
3.  **Lazy Load Tabs:** Gunakan pemuatan lambat (lazy-load) pada konten tab agar tab yang tidak aktif tidak membebani proses render awal.

---

## 7. Keamanan & Kepatuhan Kode

*   **Type Safety:** **JANGAN PERNAH** menggunakan tipe data `any`. Jika tipe tidak diketahui, gunakan `unknown` dan buat tipe guard / tipe assertion yang aman.
*   **Clean Up UseEffects:** Pastikan effect handlers membersihkan (cleanup) timer `setInterval`/`setTimeout` atau event listener untuk mencegah kebocoran memori (memory leak).
*   **Routing Guard:** Selalu lindungi rute sensitif dengan `AuthGuard` atau `PermissionGuard` yang bersandar pada validasi backend (jangan percaya state localStorage secara mutlak).

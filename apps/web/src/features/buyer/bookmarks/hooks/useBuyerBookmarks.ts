import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { bookmarksService } from "../services/bookmarks.service";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { toast } from "sonner";

export function useBuyerBookmarks() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();

  const bookmarksQuery = useQuery({
    queryKey: ["buyer-bookmarks"],
    queryFn: () => bookmarksService.getBookmarks(),
    enabled: isAuthenticated,
  });

  const addBookmarkMutation = useMutation({
    mutationFn: ({ supplierProfileId, supplierProductId }: { supplierProfileId: string; supplierProductId?: string }) =>
      bookmarksService.addBookmark(supplierProfileId, supplierProductId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-bookmarks"] });
      toast.success("Produk berhasil disimpan!");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal menyimpan ke bookmark.");
    },
  });

  const deleteBookmarkMutation = useMutation({
    mutationFn: (id: string) => bookmarksService.removeBookmark(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-bookmarks"] });
      toast.success("Bookmark berhasil dihapus!");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Gagal menghapus bookmark.");
    },
  });

  return {
    bookmarks: bookmarksQuery.data || [],
    isLoading: bookmarksQuery.isLoading && isAuthenticated,
    isError: bookmarksQuery.isError,
    addBookmark: addBookmarkMutation.mutate,
    isAdding: addBookmarkMutation.isPending,
    deleteBookmark: deleteBookmarkMutation.mutate,
    isDeleting: deleteBookmarkMutation.isPending,
  };
}

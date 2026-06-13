import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { bookmarksService } from "../services/bookmarks.service";
import { toast } from "sonner";

export function useBuyerBookmarks() {
  const queryClient = useQueryClient();

  const bookmarksQuery = useQuery({
    queryKey: ["buyer-bookmarks"],
    queryFn: () => bookmarksService.getBookmarks(),
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
    isLoading: bookmarksQuery.isLoading,
    isError: bookmarksQuery.isError,
    deleteBookmark: deleteBookmarkMutation.mutate,
    isDeleting: deleteBookmarkMutation.isPending,
  };
}

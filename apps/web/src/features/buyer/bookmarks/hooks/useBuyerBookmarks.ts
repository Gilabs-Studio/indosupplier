import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { bookmarksService } from "../services/bookmarks.service";
import { useAuthStore } from "@/features/auth/stores/use-auth-store";
import { useTranslations } from "next-intl";
import { toast } from "sonner";

export function useBuyerBookmarks() {
  const queryClient = useQueryClient();
  const { isAuthenticated } = useAuthStore();
  const t = useTranslations("buyer.bookmarks");

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
      toast.success(t("toastSaveSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastSaveError"));
    },
  });

  const deleteBookmarkMutation = useMutation({
    mutationFn: (id: string) => bookmarksService.removeBookmark(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["buyer-bookmarks"] });
      toast.success(t("toastDeleteSuccess"));
    },
    onError: (error) => {
      console.error(error);
      toast.error(t("toastDeleteError"));
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

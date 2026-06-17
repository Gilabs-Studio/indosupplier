import { useQuery } from "@tanstack/react-query";
import { contentService } from "../services/content.service";
import type { ContentListParams } from "../types/content.types";

export function usePublicContent(params: ContentListParams) {
  return useQuery({
    queryKey: ["public-content", params],
    queryFn: () => contentService.listPublic(params),
    staleTime: 5 * 60 * 1000,
  });
}

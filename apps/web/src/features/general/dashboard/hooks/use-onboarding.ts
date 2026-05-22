import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { onboardingService } from "../services/onboarding-service";
import type { OnboardingState } from "../services/onboarding-service";

const ONBOARDING_KEY = ["onboarding"] as const;

export function useOnboardingState(options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ONBOARDING_KEY,
    queryFn: () => onboardingService.getState(),
    // Keep data fresh enough for onboarding UX without over-fetching.
    staleTime: 5_000,
    refetchOnMount: true,
    refetchOnWindowFocus: false,
    refetchOnReconnect: true,
    enabled: options?.enabled ?? true,
  });
}

function useOnboardingOptimisticMutation<TVariables>(params: {
  mutationFn: (variables: TVariables) => Promise<OnboardingState>;
  optimisticUpdater: (
    current: OnboardingState,
    variables: TVariables,
  ) => OnboardingState;
}) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: params.mutationFn,
    onMutate: async (variables) => {
      await queryClient.cancelQueries({ queryKey: ONBOARDING_KEY });

      const previous = queryClient.getQueryData<OnboardingState>(ONBOARDING_KEY);

      queryClient.setQueryData<OnboardingState>(ONBOARDING_KEY, (current) => {
        if (!current) return current;
        return params.optimisticUpdater(current, variables);
      });

      return { previous };
    },
    onError: (_error, _variables, context) => {
      if (context?.previous) {
        queryClient.setQueryData(ONBOARDING_KEY, context.previous);
      }
    },
    onSuccess: (data) => {
      queryClient.setQueryData(ONBOARDING_KEY, data);
    },
  });
}

export function useSetBusinessType() {
  return useOnboardingOptimisticMutation<string>({
    mutationFn: (businessType) => onboardingService.setBusinessType(businessType),
    optimisticUpdater: (current, businessType) => ({
      ...current,
      business_type: businessType,
    }),
  });
}

export function useMarkOnboardingComplete() {
  return useOnboardingOptimisticMutation<void>({
    mutationFn: () => onboardingService.markComplete(),
    optimisticUpdater: (current) => ({
      ...current,
      completed: true,
    }),
  });
}

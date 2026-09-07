"use client";

import React from "react";
import { useDemoPopularCategories } from "../../hooks/use-demo-home";
import { PopularCategoriesSection } from "../popular-categories-section";
import { PopularCategoriesSkeleton } from "../skeletons/popular-categories-skeleton";

interface PopularCategoriesSectionWrapperProps {
  locale: string;
}

export function PopularCategoriesSectionWrapper({
  locale,
}: Readonly<PopularCategoriesSectionWrapperProps>) {
  const { popularCategories, isLoading, isEn } = useDemoPopularCategories(locale);

  if (isLoading) {
    return <PopularCategoriesSkeleton />;
  }

  return <PopularCategoriesSection categories={popularCategories} isEn={isEn} />;
}

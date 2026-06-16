export const formatReviewDate = (dateStr: string): string => {
  if (!dateStr) return "";
  // Return YYYY-MM-DD
  return dateStr.substring(0, 10);
};

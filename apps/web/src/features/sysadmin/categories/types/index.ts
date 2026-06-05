export interface Category {
  id: string;
  name: string;
  slug: string;
  parentName: string | null;
  active: boolean;
  sortOrder: number;
}

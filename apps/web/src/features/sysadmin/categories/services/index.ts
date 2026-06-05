import type { Category } from "../types";

let mockCategories: Category[] = [
  {
    id: "CAT-001",
    name: "Food & Ingredients",
    slug: "food-ingredients",
    parentName: null,
    active: true,
    sortOrder: 1
  },
  {
    id: "CAT-002",
    name: "Spices & Herbs",
    slug: "spices-herbs",
    parentName: "Food & Ingredients",
    active: true,
    sortOrder: 1
  },
  {
    id: "CAT-003",
    name: "Wooden Furniture",
    slug: "wooden-furniture",
    parentName: null,
    active: true,
    sortOrder: 2
  },
  {
    id: "CAT-004",
    name: "Batik & Apparel",
    slug: "batik-apparel",
    parentName: null,
    active: false,
    sortOrder: 3
  }
];

export const categoryService = {
  async list(): Promise<Category[]> {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([...mockCategories].sort((a, b) => a.sortOrder - b.sortOrder));
      }, 100);
    });
  },

  async update(id: string, updates: Partial<Category>): Promise<Category> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        const cat = mockCategories.find(c => c.id === id);
        if (!cat) {
          reject(new Error("Category not found"));
          return;
        }
        Object.assign(cat, updates);
        resolve({ ...cat });
      }, 100);
    });
  },

  async create(data: Omit<Category, "id">): Promise<Category> {
    return new Promise((resolve) => {
      setTimeout(() => {
        const newCat: Category = {
          ...data,
          id: `CAT-00${mockCategories.length + 1}`
        };
        mockCategories.push(newCat);
        resolve(newCat);
      }, 100);
    });
  },

  async delete(id: string): Promise<void> {
    return new Promise((resolve) => {
      setTimeout(() => {
        mockCategories = mockCategories.filter(c => c.id !== id);
        resolve();
      }, 100);
    });
  }
};

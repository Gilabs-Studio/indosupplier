import { z } from "zod";

export const supplierProfileSchema = z.object({
  companyName: z.string().min(2, "Company name must be at least 2 characters"),
  businessType: z.string().min(2, "Business type must be at least 2 characters"),
  established: z.string().min(4, "Established year is invalid"),
  employees: z.string().min(1, "Employee count is required"),
  email: z.string().email("Invalid email address"),
  phone: z.string().min(8, "Phone number must be at least 8 characters"),
  website: z.string().url("Invalid website URL"),
  taxId: z.string().min(5, "NPWP / Tax ID is required"),
  nib: z.string().min(5, "NIB number is required"),
  overview: z.string().min(10, "Overview must be at least 10 characters"),
  location: z.string().min(5, "Location/Address is required"),
});

export type SupplierProfileFormValues = z.infer<typeof supplierProfileSchema>;

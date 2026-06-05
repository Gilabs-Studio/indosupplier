"use client";

import React, { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Link } from "@/i18n/routing";
import { Edit2, Plus, Search, Trash2 } from "lucide-react";
import { toast } from "sonner";

export function SupplierProductsList() {
  const [search, setSearch] = useState("");
  const [products, setProducts] = useState([
    { id: "PROD-01", name: "Garnet Sand Mesh 80", category: "Industrial Minerals", moq: "20 Ton", price: "Rp 3.500.000 / Ton", capacity: "500 Ton / Month" },
    { id: "PROD-02", name: "Bentonite Clay Powder", category: "Industrial Minerals", moq: "10 Ton", price: "Rp 4.200.000 / Ton", capacity: "300 Ton / Month" },
    { id: "PROD-03", name: "Quartz Powder 325 Mesh", category: "Industrial Minerals", moq: "50 Ton", price: "Rp 2.800.000 / Ton", capacity: "1,000 Ton / Month" },
  ]);

  const handleDelete = (id: string) => {
    if (confirm("Are you sure you want to delete this product?")) {
      setProducts(products.filter(p => p.id !== id));
      toast.success("Product deleted successfully!");
    }
  };

  const filtered = products.filter(p => p.name.toLowerCase().includes(search.toLowerCase()));

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border/80 pb-6 text-left">
        <div className="space-y-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground font-heading">
            Products & Catalog
          </h1>
          <p className="text-sm text-muted-foreground">
            Manage your listed items, pricing terms, and wholesale minimum order quantities.
          </p>
        </div>
        <Button asChild className="cursor-pointer bg-primary text-primary-foreground hover:bg-primary/95 transition-all duration-300 hover:-translate-y-0.5 active:translate-y-0 hover:shadow-lg">
          <Link href="/supplier/profile/products/create">
            <Plus className="mr-2 h-4 w-4" /> Add Product
          </Link>
        </Button>
      </div>

      {/* Filtering Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="relative max-w-xs w-full text-left">
          <Search className="absolute left-3 top-2.5 h-4.5 w-4.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search products..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-4 py-1.5 bg-card border border-border text-sm rounded-lg outline-hidden focus:border-primary transition-all text-left"
          />
        </div>
      </div>

      {/* Catalog Table */}
      <Card className="border border-border shadow-xs rounded-xl overflow-hidden bg-card text-left">
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-b border-border bg-muted/10">
                  <TableHead className="font-bold text-foreground py-3 px-6">Product ID</TableHead>
                  <TableHead className="font-bold text-foreground">Name</TableHead>
                  <TableHead className="font-bold text-foreground">MOQ</TableHead>
                  <TableHead className="font-bold text-foreground">Price</TableHead>
                  <TableHead className="font-bold text-foreground">Supply Capacity</TableHead>
                  <TableHead className="font-bold text-foreground text-right px-6">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((prod) => (
                  <TableRow key={prod.id} className="hover:bg-muted/10 border-b border-border transition-colors">
                    <TableCell className="py-4 px-6 font-bold text-muted-foreground">{prod.id}</TableCell>
                    <TableCell className="py-4 font-semibold text-foreground">
                      <span>{prod.name}</span>
                      <p className="text-[11px] font-normal text-muted-foreground mt-0.5">{prod.category}</p>
                    </TableCell>
                    <TableCell className="py-4">
                      <Badge variant="outline" className="bg-primary/5 text-primary border-primary/20">{prod.moq}</Badge>
                    </TableCell>
                    <TableCell className="py-4 font-semibold">{prod.price}</TableCell>
                    <TableCell className="py-4 text-xs text-muted-foreground font-semibold">{prod.capacity}</TableCell>
                    <TableCell className="py-4 px-6 text-right space-x-2">
                      <Button asChild variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-primary cursor-pointer hover:bg-primary/5 border border-border">
                        <Link href={`/supplier/profile/products/${prod.id}/edit`}>
                          <Edit2 className="h-3.5 w-3.5" />
                        </Link>
                      </Button>
                      <Button onClick={() => handleDelete(prod.id)} variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground hover:text-destructive cursor-pointer hover:bg-destructive/10 border border-border">
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

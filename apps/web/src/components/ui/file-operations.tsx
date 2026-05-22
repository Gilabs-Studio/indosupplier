"use client";

import { useState, useRef } from "react";
import { Upload, Download, FileDown, MoreVertical } from "lucide-react";
import { Button } from "./button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./dialog";
import { Input } from "./input";
import { Label } from "./label";
import { cn } from "@/lib/utils";

interface FileOperationsProps {
  readonly onImport: (file: File) => void | Promise<void>;
  readonly onExport: () => void | Promise<void>;
  readonly onDownloadTemplate: () => void | Promise<void>;
  readonly hasImportPermission?: boolean;
  readonly hasExportPermission?: boolean;
  readonly className?: string;
  readonly importLoading?: boolean;
  readonly exportLoading?: boolean;
  readonly templateLoading?: boolean;
}

export function FileOperations({
  onImport,
  onExport,
  onDownloadTemplate,
  hasImportPermission = true,
  hasExportPermission = true,
  className,
  importLoading = false,
  exportLoading = false,
  templateLoading = false,
}: FileOperationsProps) {
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      // Validate file type
      const validTypes = [
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "application/vnd.ms-excel",
        "text/csv",
      ];
      if (
        !validTypes.includes(file.type) &&
        !file.name.endsWith(".xlsx") &&
        !file.name.endsWith(".xls") &&
        !file.name.endsWith(".csv")
      ) {
        alert("Please select a valid Excel file (.xlsx, .xls, or .csv)");
        return;
      }
      setSelectedFile(file);
    }
  };

  const handleImport = async () => {
    if (selectedFile) {
      await onImport(selectedFile);
      setIsImportDialogOpen(false);
      setSelectedFile(null);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  const hasAnyPermission = hasImportPermission || hasExportPermission;

  if (!hasAnyPermission) {
    return null;
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm" className={cn("shrink-0", className)}>
            <MoreVertical className="h-4 w-4 shrink-0" />
            <span className="hidden xl:inline ml-2">File Operations</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-56">
          {hasImportPermission && (
            <>
              <DropdownMenuItem
                onClick={() => setIsImportDialogOpen(true)}
                disabled={importLoading}
              >
                <Download className="mr-2 h-4 w-4" />
                {importLoading ? "Importing..." : "Import"}
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={onDownloadTemplate}
                disabled={templateLoading}
              >
                <FileDown className="mr-2 h-4 w-4" />
                {templateLoading ? "Downloading..." : "Download Template"}
              </DropdownMenuItem>
            </>
          )}
          {hasImportPermission && hasExportPermission && <DropdownMenuSeparator />}
          {hasExportPermission && (
            <DropdownMenuItem onClick={onExport} disabled={exportLoading}>
              <Upload className="mr-2 h-4 w-4" />
              {exportLoading ? "Exporting..." : "Export"}
            </DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Import Dialog */}
      <Dialog open={isImportDialogOpen} onOpenChange={setIsImportDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Import Data</DialogTitle>
            <DialogDescription>
              Select an Excel file (.xlsx, .xls, or .csv) to import data.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="file-input">Select File</Label>
              <Input
                id="file-input"
                ref={fileInputRef}
                type="file"
                accept=".xlsx,.xls,.csv"
                onChange={handleFileSelect}
                disabled={importLoading}
              />
              {selectedFile && (
                <p className="text-sm text-muted-foreground">
                  Selected: {selectedFile.name}
                </p>
              )}
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsImportDialogOpen(false);
                setSelectedFile(null);
                if (fileInputRef.current) {
                  fileInputRef.current.value = "";
                }
              }}
              disabled={importLoading}
            >
              Cancel
            </Button>
            <Button onClick={handleImport} disabled={!selectedFile || importLoading}>
              {importLoading ? "Importing..." : "Import"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

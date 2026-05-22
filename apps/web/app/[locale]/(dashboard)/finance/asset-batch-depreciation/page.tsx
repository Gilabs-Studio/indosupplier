"use client";

import { PageMotion } from "@/components/motion";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { useBatchDepreciation } from "@/features/finance/fixed-assets/hooks/use-batch-depreciation";
import { BatchDepreciationPreview } from "@/features/finance/fixed-assets/components/batch-depreciation-preview";
import { BatchDepreciationResults } from "@/features/finance/fixed-assets/components/batch-depreciation-results";
import { useState } from "react";

export default function AssetBatchDepreciationPage() {
  const [month, setMonth] = useState<string>("");
  const [year, setYear] = useState<string>("");
  const [activeTab, setActiveTab] = useState<"preview" | "results">("preview");
  const { preview, run, previewData, runData, isLoadingPreview, isRunning } = useBatchDepreciation();

  const getValidatedPeriod = (): { month: number; year: number } | null => {
    const parsedMonth = Number(month);
    const parsedYear = Number(year);
    if (!Number.isInteger(parsedMonth) || parsedMonth < 1 || parsedMonth > 12) {
      return null;
    }
    if (!Number.isInteger(parsedYear) || parsedYear < 2000 || parsedYear > 2100) {
      return null;
    }
    return { month: parsedMonth, year: parsedYear };
  };

  const handlePreview = async () => {
    const period = getValidatedPeriod();
    if (!period) return;
    await preview.mutateAsync(period);
  };

  const handleRun = async () => {
    const period = getValidatedPeriod();
    if (!period || !previewData) return;
    const result = await run.mutateAsync(period);
    if (result?.success) {
      setActiveTab("results");
    }
  };

  return (
    <PermissionGuard requiredPermission="asset.depreciate">
      <PageMotion>
        <div className="p-6 space-y-6">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Batch Depreciation</h1>
            <p className="text-muted-foreground">Run or preview depreciation for all assets in a given period.</p>
          </div>

          <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as "preview" | "results")}>
            <TabsList>
              <TabsTrigger value="preview">Preview</TabsTrigger>
              <TabsTrigger value="results" disabled={!runData}>Results</TabsTrigger>
            </TabsList>

            <TabsContent value="preview" className="space-y-4">
              <BatchDepreciationPreview
                month={month}
                year={year}
                onMonthChange={setMonth}
                onYearChange={setYear}
                onPreview={handlePreview}
                onRun={handleRun}
                previewData={previewData}
                isLoadingPreview={isLoadingPreview}
                isRunning={isRunning}
              />
            </TabsContent>

            <TabsContent value="results" className="space-y-4">
              {runData && <BatchDepreciationResults data={runData} />}
            </TabsContent>
          </Tabs>
        </div>
      </PageMotion>
    </PermissionGuard>
  );
}

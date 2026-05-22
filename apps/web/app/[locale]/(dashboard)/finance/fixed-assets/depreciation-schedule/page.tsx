'use client';

import { useMemo, useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useTranslations } from 'next-intl';
import { AlertCircle, CalendarDays, CheckCircle2, PlayCircle, RefreshCcw, TrendingUp } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { PermissionGuard } from '@/features/auth/components/permission-guard';
import { fixedAssetsService } from '@/features/finance/fixed-assets/services/fixed-assets-service';
import { FINANCE_PERMISSIONS } from '@/features/finance/permissions';
import type { BatchDepreciationRunResponse, DepreciationHistoryResponse, DepreciationScheduleResponse, RunDepreciationRequest } from '@/features/finance/fixed-assets/types';

function formatCurrency(value: number, currency = 'IDR') {
  const safeValue = Number.isFinite(value) ? value : 0;
  return safeValue.toLocaleString('id-ID', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  });
}

function currentPeriod() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
}

export default function DepreciationSchedulePage() {
  const t = useTranslations('financeFixedAssets.depreciationSchedule');
  const tCommon = useTranslations('common');
  
  const [period, setPeriod] = useState(currentPeriod());
  const [assetId, setAssetId] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [locationId, setLocationId] = useState('');
  const [runConfirmOpen, setRunConfirmOpen] = useState(false);
  const [approveConfirmOpen, setApproveConfirmOpen] = useState(false);
  const [schedule, setSchedule] = useState<DepreciationScheduleResponse | null>(null);
  const [history, setHistory] = useState<DepreciationHistoryResponse | null>(null);
  const [runResult, setRunResult] = useState<BatchDepreciationRunResponse | null>(null);

  const requestPayload = useMemo<RunDepreciationRequest>(() => ({
    period,
    asset_id: assetId.trim() || undefined,
    category_id: categoryId.trim() || undefined,
    location_id: locationId.trim() || undefined,
  }), [assetId, categoryId, locationId, period]);

  const previewMutation = useMutation({
    mutationFn: () => fixedAssetsService.getDepreciationSchedule(requestPayload),
    onSuccess: (response) => {
      setSchedule(response.data);
    },
  });

  const historyMutation = useMutation({
    mutationFn: () => fixedAssetsService.getDepreciationHistory(period),
    onSuccess: (response) => {
      setHistory(response.data);
    },
  });

  const runMutation = useMutation({
    mutationFn: () => fixedAssetsService.runDepreciation(requestPayload),
    onSuccess: (response) => {
      setRunResult(response.data);
      setRunConfirmOpen(false);
      void previewMutation.mutateAsync();
      void historyMutation.mutateAsync();
    },
  });

  const approveMutation = useMutation({
    mutationFn: () => fixedAssetsService.approveDepreciation(requestPayload),
    onSuccess: (response) => {
      setRunResult(response.data);
      setApproveConfirmOpen(false);
      void previewMutation.mutateAsync();
      void historyMutation.mutateAsync();
    },
  });

  const totalItems = schedule?.total_assets ?? 0;
  const postedCount = schedule?.posted ?? 0;
  const pendingCount = schedule?.pending ?? 0;
  const highlightedCount = schedule?.items.filter((item) => item.highlighted).length ?? 0;
  const generatedPendingCount = runResult?.items?.filter((item) => item.status === 'pending').length ?? 0;

  return (
    <PermissionGuard requiredPermission={FINANCE_PERMISSIONS.ASSET.DEPRECIATE}>
      <div className="mx-auto max-w-7xl space-y-6 p-4 md:p-8">
      <div className="relative overflow-hidden rounded-3xl border border-slate-200 bg-gradient-to-br from-slate-950 via-slate-900 to-slate-800 p-8 text-white shadow-2xl">
        <div className="absolute inset-0 opacity-20">
          <div className="absolute left-0 top-0 h-40 w-40 rounded-full bg-cyan-400 blur-3xl" />
          <div className="absolute right-0 bottom-0 h-52 w-52 rounded-full bg-amber-400 blur-3xl" />
        </div>
        <div className="relative flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div className="max-w-3xl space-y-3">
            <Badge className="bg-white/10 text-white hover:bg-white/10">{t('title')}</Badge>
            <h1 className="text-3xl font-semibold tracking-tight md:text-5xl">{t('title')}</h1>
            <p className="max-w-2xl text-sm text-slate-300 md:text-base">
              {t('description')}
            </p>
          </div>
          <div className="grid gap-3 text-sm sm:grid-cols-2 xl:grid-cols-4">
            <Stat label={t('stats.assets')} value={String(totalItems)} />
            <Stat label={t('stats.posted')} value={String(postedCount)} />
            <Stat label={t('stats.pending')} value={String(pendingCount)} />
            <Stat label={t('stats.highlighted')} value={String(highlightedCount)} />
          </div>
        </div>
      </div>

      <div className="grid gap-6 xl:grid-cols-[380px_1fr]">
        <Card className="border-slate-200 shadow-sm">
          <CardHeader>
            <CardTitle>{t('controls.title')}</CardTitle>
            <CardDescription>{t('controls.description')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="period">{t('controls.period')}</Label>
              <Input id="period" type="month" value={period} onChange={(event) => setPeriod(event.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="asset-id">{t('controls.asset')}</Label>
              <Input id="asset-id" value={assetId} onChange={(event) => setAssetId(event.target.value)} placeholder={t('controls.assetPlaceholder')} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="category-id">{t('controls.category')}</Label>
              <Input id="category-id" value={categoryId} onChange={(event) => setCategoryId(event.target.value)} placeholder={t('controls.categoryPlaceholder')} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="location-id">{t('controls.location')}</Label>
              <Input id="location-id" value={locationId} onChange={(event) => setLocationId(event.target.value)} placeholder={t('controls.locationPlaceholder')} />
            </div>

            <div className="flex flex-wrap gap-2 pt-2">
              <Button onClick={() => previewMutation.mutate()} disabled={previewMutation.isPending} className="gap-2">
                <CalendarDays className="h-4 w-4" />
                {t('buttons.preview')}
              </Button>
              <Button variant="outline" onClick={() => historyMutation.mutate()} disabled={historyMutation.isPending || !period} className="gap-2">
                <RefreshCcw className="h-4 w-4" />
                {t('buttons.history')}
              </Button>
              <Button
                variant="secondary"
                onClick={() => setRunConfirmOpen(true)}
                disabled={!schedule || runMutation.isPending}
                className="gap-2"
              >
                <PlayCircle className="h-4 w-4" />
                {t('buttons.generatePending')}
              </Button>
              <Button
                onClick={() => setApproveConfirmOpen(true)}
                disabled={!schedule || pendingCount === 0 || approveMutation.isPending}
                className="gap-2"
              >
                <CheckCircle2 className="h-4 w-4" />
                {t('buttons.approvePosting')}
              </Button>
            </div>

            <Alert className="border-amber-200 bg-amber-50 text-amber-900">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {t('alerts.info')}
              </AlertDescription>
            </Alert>
          </CardContent>
        </Card>

        <div className="space-y-6">
          {schedule && (
            <Card className="border-slate-200 shadow-sm">
              <CardHeader>
                <CardTitle>{t('preview.title')}</CardTitle>
                <CardDescription>
                  {t('preview.description', { period: schedule.period, posted: schedule.posted, pending: schedule.pending })}
                </CardDescription>
              </CardHeader>
              <CardContent className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('preview.columns.code')}</TableHead>
                      <TableHead>{t('preview.columns.name')}</TableHead>
                      <TableHead>{t('preview.columns.method')}</TableHead>
                      <TableHead className="text-right">{t('preview.columns.acquisition')}</TableHead>
                      <TableHead className="text-right">{t('preview.columns.accumulated')}</TableHead>
                      <TableHead className="text-right">{t('preview.columns.nbv')}</TableHead>
                      <TableHead className="text-right">{t('preview.columns.amount')}</TableHead>
                      <TableHead>{t('preview.columns.status')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {schedule.items.map((item) => (
                      <TableRow key={item.asset_id} className={item.highlighted ? 'bg-amber-50/70' : ''}>
                        <TableCell className="font-medium">{item.asset_code}</TableCell>
                        <TableCell>
                          <div className="space-y-1">
                            <div>{item.asset_name}</div>
                            <div className="text-xs text-muted-foreground">{item.category_name ?? item.category_id}</div>
                          </div>
                        </TableCell>
                        <TableCell>{item.method}</TableCell>
                        <TableCell className="text-right">{formatCurrency(item.acquisition_cost)}</TableCell>
                        <TableCell className="text-right">{formatCurrency(item.accumulated_depreciation)}</TableCell>
                        <TableCell className="text-right">{formatCurrency(item.net_book_value)}</TableCell>
                        <TableCell className="text-right">{formatCurrency(item.depreciation_amount)}</TableCell>
                        <TableCell>
                          <Badge variant={item.posted ? 'default' : 'secondary'}>
                            {item.posted ? t('statuses.posted') : item.highlighted ? t('statuses.notRun') : item.status}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {history && (
            <Card className="border-slate-200 shadow-sm">
              <CardHeader>
                <CardTitle>{t('history.title')}</CardTitle>
                <CardDescription>
                  {t('history.description', { total: history.total, period: history.period })}
                </CardDescription>
              </CardHeader>
              <CardContent className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>{t('history.columns.asset')}</TableHead>
                      <TableHead className="text-right">{t('history.columns.amount')}</TableHead>
                      <TableHead>{t('history.columns.status')}</TableHead>
                      <TableHead>{t('history.columns.journal')}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {history.items.map((item) => (
                      <TableRow key={item.depreciation_id}>
                        <TableCell>
                          <div className="space-y-1">
                            <div className="font-medium">{item.asset_code}</div>
                            <div className="text-xs text-muted-foreground">{item.asset_name}</div>
                          </div>
                        </TableCell>
                        <TableCell className="text-right">{formatCurrency(item.amount)}</TableCell>
                        <TableCell>
                          <Badge variant={item.posted ? 'default' : 'secondary'}>{item.posted ? t('statuses.posted') : t('statuses.pending')}</Badge>
                        </TableCell>
                        <TableCell className="font-mono text-xs">{item.journal_entry_id ?? '-'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}

          {runResult && (
            <Card className="border-emerald-200 bg-emerald-50 shadow-sm">
              <CardHeader>
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-5 w-5 text-emerald-700" />
                  <CardTitle>{t('result.title')}</CardTitle>
                </div>
                <CardDescription>{t('result.description', { period })}</CardDescription>
              </CardHeader>
              <CardContent className="grid gap-3 text-sm md:grid-cols-4">
                <Stat label={t('result.stats.pendingCreated')} value={String(generatedPendingCount)} />
                <Stat label={t('result.stats.posted')} value={String(runResult.posted ?? runResult.successful ?? 0)} />
                <Stat label={t('result.stats.skipped')} value={String(runResult.skipped ?? 0)} />
                <Stat label={t('result.stats.failed')} value={String(runResult.failed ?? 0)} />
              </CardContent>
            </Card>
          )}
        </div>
      </div>

      <Dialog open={runConfirmOpen} onOpenChange={setRunConfirmOpen}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>{t('dialogs.generateTitle')}</DialogTitle>
            <DialogDescription>
              {t('alerts.confirmGenerate', { period })}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-lg border p-3"><div className="text-muted-foreground">{t('stats.posted')}</div><div className="font-semibold">{schedule?.posted ?? 0}</div></div>
              <div className="rounded-lg border p-3"><div className="text-muted-foreground">{t('stats.pending')}</div><div className="font-semibold">{schedule?.pending ?? 0}</div></div>
              <div className="rounded-lg border p-3"><div className="text-muted-foreground">{t('result.stats.posted')}</div><div className="font-semibold flex items-center gap-1"><TrendingUp className="h-4 w-4" />{formatCurrency(schedule?.total_amount ?? 0)}</div></div>
              <div className="rounded-lg border p-3"><div className="text-muted-foreground">{t('stats.period')}</div><div className="font-semibold">{period}</div></div>
            </div>
            <Alert className="border-amber-200 bg-amber-50 text-amber-900">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {t('alerts.reviewPending')}
              </AlertDescription>
            </Alert>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRunConfirmOpen(false)}>{tCommon('actions.cancel')}</Button>
            <Button onClick={() => runMutation.mutate()} disabled={runMutation.isPending}>
              {runMutation.isPending ? `${t('buttons.generatePending')}...` : t('buttons.generatePending')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={approveConfirmOpen} onOpenChange={setApproveConfirmOpen}>
        <DialogContent className="max-w-xl">
          <DialogHeader>
            <DialogTitle>{t('dialogs.approveTitle')}</DialogTitle>
            <DialogDescription>
              {t('alerts.confirmApprove', { period })}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-3 text-sm">
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-lg border p-3"><div className="text-muted-foreground">{t('stats.pending')}</div><div className="font-semibold">{schedule?.pending ?? 0}</div></div>
              <div className="rounded-lg border p-3"><div className="text-muted-foreground">{t('stats.alreadyPosted')}</div><div className="font-semibold">{schedule?.posted ?? 0}</div></div>
              <div className="rounded-lg border p-3"><div className="text-muted-foreground">{t('result.stats.posted')}</div><div className="font-semibold flex items-center gap-1"><TrendingUp className="h-4 w-4" />{formatCurrency(schedule?.total_amount ?? 0)}</div></div>
              <div className="rounded-lg border p-3"><div className="text-muted-foreground">{t('stats.period')}</div><div className="font-semibold">{period}</div></div>
            </div>
            <Alert className="border-amber-200 bg-amber-50 text-amber-900">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {t('alerts.afterPosting')}
              </AlertDescription>
            </Alert>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setApproveConfirmOpen(false)}>{tCommon('actions.cancel')}</Button>
            <Button onClick={() => approveMutation.mutate()} disabled={approveMutation.isPending || pendingCount === 0}>
              {approveMutation.isPending ? `${t('buttons.approvePosting')}...` : t('buttons.approvePosting')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      </div>
    </PermissionGuard>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-white/10 bg-white/5 px-4 py-3 backdrop-blur-sm">
      <div className="text-xs uppercase tracking-wide text-slate-300">{label}</div>
      <div className="mt-1 text-xl font-semibold text-white">{value}</div>
    </div>
  );
}

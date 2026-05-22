import { getRequestConfig } from "next-intl/server";
import { routing } from "./routing";
import type { Locale } from "@/types/locale";
// Global/shared messages
import globalEnMessages from "./messages/en.json";
import globalIdMessages from "./messages/id.json";
// Feature messages
import { userManagementEn } from "@/features/master-data/user-management/i18n/en";
import { userManagementId } from "@/features/master-data/user-management/i18n/id";
import notificationsEnMessages from "@/features/notifications/i18n/messages/en.json";
import notificationsIdMessages from "@/features/notifications/i18n/messages/id.json";
import dashboardEnMessages from "@/features/general/dashboard/i18n/messages/en.json";
import dashboardIdMessages from "@/features/general/dashboard/i18n/messages/id.json";

import { geographicEn } from "@/features/master-data/geographic/i18n/en";
import { geographicId } from "@/features/master-data/geographic/i18n/id";
import { organizationEn } from "@/features/master-data/organization/i18n/en";
import { organizationId } from "@/features/master-data/organization/i18n/id";
import { employeeEn } from "@/features/master-data/employee/i18n/en";
import { employeeId } from "@/features/master-data/employee/i18n/id";
import { supplierEn } from "@/features/master-data/supplier/i18n/en";
import { supplierId } from "@/features/master-data/supplier/i18n/id";
import { customerEn } from "@/features/master-data/customer/i18n/en";
import { customerId } from "@/features/master-data/customer/i18n/id";
import { productEn } from "@/features/master-data/product/i18n/en";
import { productId } from "@/features/master-data/product/i18n/id";
import { warehouseEn } from "@/features/master-data/warehouse/i18n/en";
import { warehouseId } from "@/features/master-data/warehouse/i18n/id";
import { outletEn } from "@/features/master-data/outlet/i18n/en";
import { outletId } from "@/features/master-data/outlet/i18n/id";
import { currencyEn } from "@/features/master-data/currencies/i18n/en";
import { currencyId } from "@/features/master-data/currencies/i18n/id";
import { paymentTermEn } from "@/features/master-data/payment-and-couriers/payment-terms/i18n/en";
import { paymentTermId } from "@/features/master-data/payment-and-couriers/payment-terms/i18n/id";
import { courierAgencyEn } from "@/features/master-data/payment-and-couriers/courier-agency/i18n/en";
import { courierAgencyId } from "@/features/master-data/payment-and-couriers/courier-agency/i18n/id";
import { soSourceEn } from "@/features/master-data/payment-and-couriers/so-source/i18n/en";
import { soSourceId } from "@/features/master-data/payment-and-couriers/so-source/i18n/id";
import { leaveTypeEn } from "@/features/master-data/leave-type/i18n/en";
import { leaveTypeId } from "@/features/master-data/leave-type/i18n/id";
import { leaveRequestEn } from "@/features/hrd/leave-request/i18n/en";
import { leaveRequestId } from "@/features/hrd/leave-request/i18n/id";
import { quotationEn } from "@/features/sales/quotation/i18n/en";
import { quotationId } from "@/features/sales/quotation/i18n/id";
import { orderEn } from "@/features/sales/order/i18n/en";
import { orderId } from "@/features/sales/order/i18n/id";
import { deliveryEn } from "@/features/sales/delivery/i18n/en";
import { deliveryId } from "@/features/sales/delivery/i18n/id";
import { invoiceEn } from "@/features/sales/invoice/i18n/en";
import { invoiceId } from "@/features/sales/invoice/i18n/id";
import { salesReturnsEn } from "@/features/sales/returns/i18n/en";
import { salesReturnsId } from "@/features/sales/returns/i18n/id";
import { commandPaletteEn } from "@/features/command-palette/i18n/en";
import { commandPaletteId } from "@/features/command-palette/i18n/id";
import { salesTargetsEn } from "@/features/crm/sales-targets/i18n/en";
import { salesTargetsId } from "@/features/crm/sales-targets/i18n/id";
import { hrdEn } from "@/features/hrd/i18n/en";
import { hrdId } from "@/features/hrd/i18n/id";
import { inventoryEn } from "@/features/stock/inventory/i18n/en";
import { inventoryId } from "@/features/stock/inventory/i18n/id";
import { stockLedgerEn } from "@/features/stock/stock-ledger/i18n/en";
import { stockLedgerId } from "@/features/stock/stock-ledger/i18n/id";
import { stockOpnameEn } from "@/features/stock/stock-opname/i18n/en";
import { stockOpnameId } from "@/features/stock/stock-opname/i18n/id";
import { stockMovementEn } from "@/features/stock/stock-movement/i18n/en";
import { stockMovementId } from "@/features/stock/stock-movement/i18n/id";
import { settingsEn } from "@/features/settings/i18n/en";
import { settingsId } from "@/features/settings/i18n/id";
import { evaluationEn } from "@/features/hrd/evaluation/i18n/en";
import { evaluationId } from "@/features/hrd/evaluation/i18n/id";
import { recruitmentEn } from "@/features/hrd/recruitment/i18n/en";
import { recruitmentId } from "@/features/hrd/recruitment/i18n/id";

import { passwordResetEn } from "@/features/auth/password-reset/i18n/en";
import { passwordResetId } from "@/features/auth/password-reset/i18n/id";

import { purchaseRequisitionEn } from "@/features/purchase/requisitions/i18n/en";
import { purchaseRequisitionId } from "@/features/purchase/requisitions/i18n/id";

import { purchaseOrderEn } from "@/features/purchase/orders/i18n/en";
import { purchaseOrderId } from "@/features/purchase/orders/i18n/id";

import { goodsReceiptEn } from "@/features/purchase/goods-receipt/i18n/en";
import { goodsReceiptId } from "@/features/purchase/goods-receipt/i18n/id";

import { supplierInvoiceEn } from "@/features/purchase/supplier-invoices/i18n/en";
import { supplierInvoiceId } from "@/features/purchase/supplier-invoices/i18n/id";
import { purchaseReturnsEn } from "@/features/purchase/returns/i18n/en";
import { purchaseReturnsId } from "@/features/purchase/returns/i18n/id";

import { supplierInvoiceDPEn } from "@/features/purchase/supplier-invoice-down-payments/i18n/en";
import { supplierInvoiceDPId } from "@/features/purchase/supplier-invoice-down-payments/i18n/id";

import { customerInvoiceDPEn } from "@/features/sales/customer-invoice-down-payments/i18n/en";
import { customerInvoiceDPId } from "@/features/sales/customer-invoice-down-payments/i18n/id";

import { purchasePaymentEn } from "@/features/purchase/payments/i18n/en";
import { purchasePaymentId } from "@/features/purchase/payments/i18n/id";

import { salesPaymentEn } from "@/features/sales/payments/i18n/en";
import { salesPaymentId } from "@/features/sales/payments/i18n/id";

import { receivablesRecapEn } from "@/features/sales/receivables-recap/i18n/en";
import { receivablesRecapId } from "@/features/sales/receivables-recap/i18n/id";

import { payableRecapEn } from "@/features/purchase/payable-recap/i18n/en";
import { payableRecapId } from "@/features/purchase/payable-recap/i18n/id";

import { financeCoaEn } from "@/features/finance/coa/i18n/en";
import { financeCoaId } from "@/features/finance/coa/i18n/id";
import { financeJournalsEn } from "@/features/finance/journals/i18n/en";
import { financeJournalsId } from "@/features/finance/journals/i18n/id";

import { financeBankAccountsEn } from "@/features/finance/bank-accounts/i18n/en";
import { financeBankAccountsId } from "@/features/finance/bank-accounts/i18n/id";
import { financeCashBankTransactionsEn } from "@/features/finance/cash-bank-transactions/i18n/en";
import { financeCashBankTransactionsId } from "@/features/finance/cash-bank-transactions/i18n/id";
import { financeBankTransferEn } from "@/features/finance/bank-transfer/i18n/en";
import { financeBankTransferId } from "@/features/finance/bank-transfer/i18n/id";
import { financeBankReconciliationEn } from "@/features/finance/bank-reconciliation/i18n/en";
import { financeBankReconciliationId } from "@/features/finance/bank-reconciliation/i18n/id";
import { financePaymentsEn } from "@/features/finance/payments/i18n/en";
import { financePaymentsId } from "@/features/finance/payments/i18n/id";
import { financeBudgetEn } from "@/features/finance/budget/i18n/en";
import { financeBudgetId } from "@/features/finance/budget/i18n/id";

import { financeAgingReportsEn } from "@/features/finance/aging-reports/i18n/en";
import { financeAgingReportsId } from "@/features/finance/aging-reports/i18n/id";
import { financeAssetCategoriesEn } from "@/features/finance/asset-categories/i18n/en";
import { financeAssetCategoriesId } from "@/features/finance/asset-categories/i18n/id";
import { financeAssetLocationsEn } from "@/features/finance/asset-locations/i18n/en";
import { financeAssetLocationsId } from "@/features/finance/asset-locations/i18n/id";
import { financeAssetsEn } from "@/features/finance/assets/i18n/en";
import { financeAssetsId } from "@/features/finance/assets/i18n/id";
import { financeClosingEn } from "@/features/finance/closing/i18n/en";
import { financeClosingId } from "@/features/finance/closing/i18n/id";
import { financeSettingsEn } from "@/features/finance/settings/i18n/en";
import { financeSettingsId } from "@/features/finance/settings/i18n/id";
import { financeTaxInvoicesEn } from "@/features/finance/tax-invoices/i18n/en";
import { financeTaxInvoicesId } from "@/features/finance/tax-invoices/i18n/id";
import { financeNonTradePayablesEn } from "@/features/finance/non-trade-payables/i18n/en";
import { financeNonTradePayablesId } from "@/features/finance/non-trade-payables/i18n/id";
import { financeSalaryEn as hrdSalaryEn } from "@/features/hrd/salary-structures/i18n/en";
import { financeSalaryId as hrdSalaryId } from "@/features/hrd/salary-structures/i18n/id";
import { financeFixedAssetsEn } from "@/features/finance/fixed-assets/i18n/en";
import { financeFixedAssetsId } from "@/features/finance/fixed-assets/i18n/id";
import { aiChatEn } from "@/features/ai-chat/i18n/en";
import { aiChatId } from "@/features/ai-chat/i18n/id";

import { pipelineStageEn } from "@/features/crm/pipeline-stage/i18n/en";
import { pipelineStageId } from "@/features/crm/pipeline-stage/i18n/id";
import { leadSourceEn } from "@/features/crm/lead-source/i18n/en";
import { leadSourceId } from "@/features/crm/lead-source/i18n/id";
import { leadStatusEn } from "@/features/crm/lead-status/i18n/en";
import { leadStatusId } from "@/features/crm/lead-status/i18n/id";
import { contactRoleEn } from "@/features/crm/contact-role/i18n/en";
import { contactRoleId } from "@/features/crm/contact-role/i18n/id";
import { activityTypeEn } from "@/features/crm/activity-type/i18n/en";
import { activityTypeId } from "@/features/crm/activity-type/i18n/id";
import { financeReportsEn } from "@/features/finance/reports/i18n/en";
import { financeReportsId } from "@/features/finance/reports/i18n/id";
import { crmContactEn } from "@/features/crm/contact/i18n/en";
import { crmContactId } from "@/features/crm/contact/i18n/id";
import { crmLeadEn } from "@/features/crm/lead/i18n/en";
import { crmLeadId } from "@/features/crm/lead/i18n/id";
import { crmDealEn } from "@/features/crm/deal/i18n/en";
import { crmDealId } from "@/features/crm/deal/i18n/id";
import { crmVisitReportEn } from "@/features/crm/visit-report/i18n/en";
import { crmVisitReportId } from "@/features/crm/visit-report/i18n/id";
import { crmActivityEn } from "@/features/crm/activity/i18n/en";
import { crmActivityId } from "@/features/crm/activity/i18n/id";
import { crmTaskEn } from "@/features/crm/task/i18n/en";
import { crmTaskId } from "@/features/crm/task/i18n/id";
import { crmScheduleEn } from "@/features/crm/schedule/i18n/en";
import { crmScheduleId } from "@/features/crm/schedule/i18n/id";
import { areaMappingEn } from "@/features/crm/area-mapping/i18n/en";
import { areaMappingId } from "@/features/crm/area-mapping/i18n/id";
import { salesOverviewReportEn } from "@/features/reports/sales-overview/i18n/en";
import { salesOverviewReportId } from "@/features/reports/sales-overview/i18n/id";
import { productAnalysisReportEn } from "@/features/reports/product-analysis/i18n/en";
import { productAnalysisReportId } from "@/features/reports/product-analysis/i18n/id";
import { geoPerformanceReportEn } from "@/features/reports/geo-performance/i18n/en";
import { geoPerformanceReportId } from "@/features/reports/geo-performance/i18n/id";
import { customerResearchReportEn } from "@/features/reports/customer-research/i18n/en";
import { customerResearchReportId } from "@/features/reports/customer-research/i18n/id";
import { supplierResearchReportEn } from "@/features/reports/supplier-research/i18n/en";
import { supplierResearchReportId } from "@/features/reports/supplier-research/i18n/id";
import { visitPlannerEn } from "@/features/travel/visit-planner/i18n/en";
import { visitPlannerId } from "@/features/travel/visit-planner/i18n/id";
import { floorLayoutEn } from "@/features/pos/fb/floor-layout/i18n/en";
import { floorLayoutId } from "@/features/pos/fb/floor-layout/i18n/id";
import { posTerminalEn } from "@/features/pos/terminal/i18n/en";
import { posTerminalId } from "@/features/pos/terminal/i18n/id";
import { feedbackEn } from "@/features/pos/feedback/i18n/en";
import { feedbackId } from "@/features/pos/feedback/i18n/id";
import { posSelflOrderEn } from "@/features/pos/self-order/i18n/en";
import { posSelflOrderId } from "@/features/pos/self-order/i18n/id";
import { loyaltyEn } from "@/features/loyalty/i18n/en";
import { loyaltyId } from "@/features/loyalty/i18n/id";
import { landingEn } from "@/features/landing/i18n/en";
import { landingId } from "@/features/landing/i18n/id";
import { systemAdminEn } from "@/features/system-admin/i18n/en";
import { systemAdminId } from "@/features/system-admin/i18n/id";

// Merge all messages
const messages = {
  en: {
    ...globalEnMessages,
    userManagement: userManagementEn,
    ...notificationsEnMessages,
    ...dashboardEnMessages,
    ...geographicEn,
    organization: organizationEn,
    employee: employeeEn,
    supplier: supplierEn,
    customer: customerEn,
    product: productEn,
    warehouse: warehouseEn,
    outlet: outletEn,
    currency: currencyEn,
    paymentTerm: paymentTermEn,
    courierAgency: courierAgencyEn,
    soSource: soSourceEn,
    leaveType: leaveTypeEn,
    ...leaveRequestEn,
    ...quotationEn,
    ...orderEn,
    ...deliveryEn,
    ...invoiceEn,
    ...salesReturnsEn,
    customerInvoiceDP: customerInvoiceDPEn,
    ...commandPaletteEn,
    ...salesTargetsEn,
    ...hrdEn,
    ...inventoryEn,
    stock_ledger: stockLedgerEn,
    stock_opname: stockOpnameEn,
    ...stockMovementEn,
    ...settingsEn,
    ...evaluationEn,
    ...recruitmentEn,
    ...passwordResetEn,
    auth: {
      ...globalEnMessages.auth,
      ...passwordResetEn.auth,
    },
    purchaseRequisition: purchaseRequisitionEn,
    purchaseOrder: purchaseOrderEn,
    goodsReceipt: goodsReceiptEn,
    supplierInvoice: supplierInvoiceEn,
    ...purchaseReturnsEn,
    supplierInvoiceDP: supplierInvoiceDPEn,
    purchasePayment: purchasePaymentEn,
    salesPayment: salesPaymentEn,
    receivablesRecap: receivablesRecapEn,
    payableRecap: payableRecapEn,
    financeCoa: financeCoaEn,
    financeJournals: financeJournalsEn,

    financeBankAccounts: financeBankAccountsEn,
    financeCashBankTransactions: financeCashBankTransactionsEn,
    financeBankTransfer: financeBankTransferEn,
    financeBankReconciliation: financeBankReconciliationEn,
    financePayments: financePaymentsEn,
    financeBudget: financeBudgetEn,

    financeAgingReports: financeAgingReportsEn,
    financeAssetCategories: financeAssetCategoriesEn,
    financeAssetLocations: financeAssetLocationsEn,
    financeAssets: financeAssetsEn,
    financeClosing: financeClosingEn,
    financeSettings: financeSettingsEn,
    financeTaxInvoices: financeTaxInvoicesEn,
    financeNonTradePayables: financeNonTradePayablesEn,
    hrdSalary: hrdSalaryEn,
    financeFixedAssets: financeFixedAssetsEn,
    ...aiChatEn,
    pipelineStage: pipelineStageEn,
    leadSource: leadSourceEn,
    leadStatus: leadStatusEn,
    contactRole: contactRoleEn,
    activityType: activityTypeEn,
    financeReports: financeReportsEn,
    ...crmContactEn,
    ...crmLeadEn,
    ...crmDealEn,
    ...crmVisitReportEn,
    ...crmActivityEn,
    ...crmTaskEn,
    ...crmScheduleEn,
    ...areaMappingEn,
    ...salesOverviewReportEn,
    ...productAnalysisReportEn,
    ...geoPerformanceReportEn,
    ...customerResearchReportEn,
    ...supplierResearchReportEn,
    ...visitPlannerEn,
    ...floorLayoutEn,
    ...posTerminalEn,
    ...feedbackEn,
    ...posSelflOrderEn,
    ...loyaltyEn,
    ...landingEn,
    ...systemAdminEn,
  },
  id: {
    ...globalIdMessages,
    userManagement: userManagementId,
    ...notificationsIdMessages,
    ...dashboardIdMessages,
    ...geographicId,
    organization: organizationId,
    employee: employeeId,
    supplier: supplierId,
    customer: customerId,
    product: productId,
    warehouse: warehouseId,
    outlet: outletId,
    currency: currencyId,
    paymentTerm: paymentTermId,
    courierAgency: courierAgencyId,
    soSource: soSourceId,
    leaveType: leaveTypeId,
    ...leaveRequestId,
    ...quotationId,
    ...orderId,
    ...deliveryId,
    ...invoiceId,
    ...salesReturnsId,
    customerInvoiceDP: customerInvoiceDPId,
    ...commandPaletteId,
    ...salesTargetsId,
    ...hrdId,
    ...inventoryId,
    stock_ledger: stockLedgerId,
    stock_opname: stockOpnameId,
    ...stockMovementId,
    ...settingsId,
    ...evaluationId,
    ...recruitmentId,
    ...passwordResetId,
    auth: {
      ...globalIdMessages.auth,
      ...passwordResetId.auth,
    },
    purchaseRequisition: purchaseRequisitionId,
    purchaseOrder: purchaseOrderId,
    goodsReceipt: goodsReceiptId,
    supplierInvoice: supplierInvoiceId,
    ...purchaseReturnsId,
    supplierInvoiceDP: supplierInvoiceDPId,
    purchasePayment: purchasePaymentId,
    salesPayment: salesPaymentId,
    receivablesRecap: receivablesRecapId,
    payableRecap: payableRecapId,
    financeCoa: financeCoaId,
    financeJournals: financeJournalsId,

    financeBankAccounts: financeBankAccountsId,
    financeCashBankTransactions: financeCashBankTransactionsId,
    financeBankTransfer: financeBankTransferId,
    financeBankReconciliation: financeBankReconciliationId,
    financePayments: financePaymentsId,
    financeBudget: financeBudgetId,

    financeAgingReports: financeAgingReportsId,
    financeAssetCategories: financeAssetCategoriesId,
    financeAssetLocations: financeAssetLocationsId,
    financeAssets: financeAssetsId,
    financeClosing: financeClosingId,
    financeSettings: financeSettingsId,
    financeTaxInvoices: financeTaxInvoicesId,
    financeNonTradePayables: financeNonTradePayablesId,
    hrdSalary: hrdSalaryId,
    financeFixedAssets: financeFixedAssetsId,
    ...aiChatId,
    pipelineStage: pipelineStageId,
    leadSource: leadSourceId,
    leadStatus: leadStatusId,
    contactRole: contactRoleId,
    activityType: activityTypeId,
    financeReports: financeReportsId,
    ...crmContactId,
    ...crmLeadId,
    ...crmDealId,
    ...crmVisitReportId,
    ...crmActivityId,
    ...crmTaskId,
    ...crmScheduleId,
    ...areaMappingId,
    ...salesOverviewReportId,
    ...productAnalysisReportId,
    ...geoPerformanceReportId,
    ...customerResearchReportId,
    ...supplierResearchReportId,
    ...visitPlannerId,
    ...floorLayoutId,
    ...posTerminalId,
    ...feedbackId,
    ...posSelflOrderId,
    ...loyaltyId,
    ...landingId,
    ...systemAdminId,
  },
} as const;

export default getRequestConfig(async ({ requestLocale }) => {
  let locale = await requestLocale;

  if (!locale || !routing.locales.includes(locale as Locale)) {
    locale = routing.defaultLocale;
  }

  return {
    locale,
    messages: messages[locale as keyof typeof messages],
  };
});

import { redirect } from "next/navigation";

export default function LegacyFinanceNonTradePayablesRedirectPage() {
  redirect("/finance/ap/non-trade-payables");
}

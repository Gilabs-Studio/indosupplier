import { redirect } from "next/navigation";

export default function LegacyFinancePaymentsRedirectPage() {
  redirect("/finance/ap/payments");
}

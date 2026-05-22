import { redirect } from "next/navigation";

export default function FinanceAPDebitNotesRedirectPage() {
  redirect("/finance/journals/purchase");
}

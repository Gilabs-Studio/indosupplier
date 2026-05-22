import { redirect } from "next/navigation";

export default function CanonicalFinanceRouteRedirectPage() {
  redirect("/finance/aging-reports");
}
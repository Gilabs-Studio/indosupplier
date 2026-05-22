import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { ContactRoleContainer } from "@/features/crm/contact-role/components";

export default function ContactRolePage() {
  return (
    <PermissionGuard requiredPermission="crm_contact_role.read">
      <ContactRoleContainer />
    </PermissionGuard>
  );
}

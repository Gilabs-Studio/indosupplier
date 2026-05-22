import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { AreaMappingMapView } from "@/features/crm/area-mapping/components";

export default function AreaMappingPage() {
	return (
		<PermissionGuard requiredPermission="crm_area_mapping.read">
			<div className="w-full h-full">
				<AreaMappingMapView />
			</div>
		</PermissionGuard>
	);
}

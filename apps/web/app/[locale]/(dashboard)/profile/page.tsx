import { PageMotion } from "@/components/motion";
import { ProfileView } from "@/features/settings/components/profile/profile-view";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Profile Settings",
  description: "Manage your account profile settings",
};

export default function ProfilePage() {
  return (
    <PageMotion>
      <ProfileView />
    </PageMotion>
  );
}

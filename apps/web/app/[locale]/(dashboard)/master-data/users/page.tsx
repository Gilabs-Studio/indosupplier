"use client";

import { motion } from "framer-motion";
import { useTranslations } from "next-intl";
import { AuthGuard } from "@/features/auth/components/auth-guard";
import { PermissionGuard } from "@/features/auth/components/permission-guard";
import { UserManagement } from "@/features/master-data/user-management/components/user-management";

const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
    },
  },
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.4 },
  },
};

function UsersPageContent() {
  const t = useTranslations("masterData.usersPage");

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      <motion.div variants={itemVariants}>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">
              {t("title")}
            </h1>
            <p className="text-muted-foreground mt-1">
              {t("description")}
            </p>
          </div>
        </div>
      </motion.div>

      <motion.div variants={itemVariants}>
        <UserManagement />
      </motion.div>
    </motion.div>
  );
}

export default function UsersPage() {
  return (
    <AuthGuard>
      <PermissionGuard requiredPermission="user.read">
        <UsersPageContent />
      </PermissionGuard>
    </AuthGuard>
  );
}




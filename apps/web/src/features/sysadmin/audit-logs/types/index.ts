export interface AuditLog {
  id: string;
  actorName: string;
  actorEmail: string;
  action: string;
  target: string;
  metadata: string;
  createdAt: string;
}

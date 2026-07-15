export type PublicUser = {
  handle: string;
  displayName: string;
  bio: string;
  location: string;
  createdAt: string;
};

export type Me = {
  user: PublicUser;
  role: "member" | "platform_admin";
};

export type Notification = {
  id: string;
  kind: string;
  title: string;
  body: string;
  data: Record<string, unknown>;
  readAt?: string;
  createdAt: string;
};

export type Invite = {
  id: string;
  maxUses: number;
  uses: number;
  expiresAt: string;
  disabledAt?: string;
  createdAt: string;
};

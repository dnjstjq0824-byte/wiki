// packages/types/index.ts
export interface User {
  id: string;
  username: string;
  role: string;
  edit_count: number;
  profile_bio: string;
  avatar_url: string;
  created_at: string;
}

export interface Page {
  id: string;
  namespace: number;
  title: string;
  content: string;
  summary: string;
  author_id: string;
  protection_level: string;
  is_redirect: boolean;
  is_deleted: boolean;
  current_revision_id: string;
  created_at: string;
  updated_at: string;
}

export interface Revision {
  id: string;
  page_id: string;
  content: string;
  summary: string;
  author_id: string;
  is_minor: boolean;
  created_at: string;
}
export interface User {
  user_id: number;
  username?: string;
  password_hash?: string;
  role: string;
  first_name: string;
  last_name: string;
  email: string;
  phone?: string;
  department?: string;
  position?: string;
  last_login?: Date;
  created_at: Date;
  updated_at: Date;
}

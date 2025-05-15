export interface Contact {
  contact_id: number;
  customer_id: number;
  first_name: string;
  last_name: string;
  position: string;
  email: string;
  phone: string;
  created_at?: string;
  updated_at?: string;
}

export type ContactCreate = Omit<Contact, 'contact_id' | 'created_at' | 'updated_at'>;
export type ContactUpdate = Partial<ContactCreate> & { contact_id: number }; 
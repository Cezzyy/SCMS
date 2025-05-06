export interface Contact {
  contact_id: number;
  customer_id: number;
  first_name: string;
  last_name: string;
  position?: string;
  phone?: string;
  email?: string;
  created_at: string;
  updated_at: string;
}

export interface ContactCreate extends Omit<Contact, 'contact_id' | 'created_at' | 'updated_at'> {
  contact_id?: number;
}

export interface ContactUpdate extends ContactCreate {
  contact_id: number;
} 
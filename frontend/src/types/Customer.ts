export interface Customer {
  customer_id: number;
  company_name: string;
  industry?: string;
  address?: string;
  phone?: string;
  email?: string;
  website?: string;
  created_at: string;
  updated_at: string;
}

export interface CustomerCreate extends Omit<Customer, 'customer_id' | 'created_at' | 'updated_at'> {
  customer_id?: number;
}

export interface CustomerUpdate extends CustomerCreate {
  customer_id: number;
} 
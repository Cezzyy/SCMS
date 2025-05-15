export interface Product {
  product_id: number;
  product_name: string;
  model?: string;
  description?: string;
  technical_specs?: string | Record<string, any>; // Allow either JSON string or object
  certifications?: string;
  safety_standards?: string;
  warranty_period: number;
  price: number;
  created_at: string;
  updated_at: string;
}

export interface ProductCreate extends Omit<Product, 'product_id' | 'created_at' | 'updated_at'> {
  product_id?: number;
  created_at?: string; // Optional for create payloads
  updated_at?: string; // Optional for create payloads
}

export interface ProductUpdate extends Omit<ProductCreate, 'product_id'> {
  product_id: number;
  created_at?: string; // Optional for update payloads
  updated_at?: string; // Optional for update payloads
} 
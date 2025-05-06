export interface Product {
  product_id: number;
  product_name: string;
  model?: string;
  description?: string;
  technical_specs?: Record<string, any>;
  certifications?: string;
  safety_standards?: string;
  warranty_period: number;
  price: number;
  created_at: string;
  updated_at: string;
}

export interface ProductCreate extends Omit<Product, 'product_id' | 'created_at' | 'updated_at'> {
  product_id?: number;
}

export interface ProductUpdate extends ProductCreate {
  product_id: number;
} 
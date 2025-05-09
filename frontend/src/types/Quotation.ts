export interface Quotation {
  quotation_id: number;
  customer_id: number;
  quote_date: string;
  validity_date: string;
  status: 'Pending' | 'Approved' | 'Rejected' | 'Expired';
  total_amount: number;
  created_at: string;
  updated_at: string;
}

export interface QuotationItem {
  quotation_item_id: number;
  quotation_id: number;
  product_id: number;
  quantity: number;
  unit_price: number;
  discount: number;
  line_total: number;
}

export interface QuotationCreate {
  quotation_id?: number;
  customer_id: number;
  quote_date: string;
  validity_date: string;
  status: 'Pending' | 'Approved' | 'Rejected' | 'Expired';
  total_amount: number;
  items: Omit<QuotationItem, 'quotation_item_id' | 'quotation_id'>[];
}

export interface QuotationUpdate extends QuotationCreate {
  quotation_id: number;
}

export interface QuotationItemCreate extends Omit<QuotationItem, 'quotation_item_id' | 'quotation_id'> {
  quotation_item_id?: number;
}

export interface QuotationItemUpdate extends QuotationItemCreate {
  quotation_item_id: number;
  quotation_id: number;
}

export interface QuotationWithItems extends Quotation {
  items: QuotationItem[];
} 
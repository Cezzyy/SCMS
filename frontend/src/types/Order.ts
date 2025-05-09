export interface Order {
  order_id: number;
  customer_id: number;
  quotation_id?: number;
  order_date: string;
  shipping_address: string;
  status: 'Pending' | 'Shipped' | 'Delivered' | 'Cancelled';
  total_amount: number;
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  order_item_id: number;
  order_id: number;
  product_id: number;
  quantity: number;
  unit_price: number;
  discount: number;
  line_total: number;
}

// Report interfaces matching the backend models

export interface SalesTrend {
  day: string;
  total_amount: number;
}

export interface LowStockItem {
  id: number;
  product_id: number;
  name: string;
  current_stock: number;
  reorder_level: number;
  unit_price: number;
}

export interface TopCustomer {
  id: number;
  name: string;
  total_spent: number;
  orders: number;
  contact_name?: string;
  avatar?: string; // For UI display purposes
}

export interface DashboardSummary {
  total_sales: number;
  order_count: number;
  low_stock_count: number;
  sales_trends: SalesTrend[];
  low_stock_items: LowStockItem[];
  top_customers: TopCustomer[];
  period: string;
  last_updated: string;
}

// Request parameters
export interface ReportParams {
  days?: number;
  limit?: number;
} 
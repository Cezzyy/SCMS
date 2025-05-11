export interface Inventory {
  inventory_id: number;
  product_id: number;
  current_stock: number;
  reorder_level: number;
  last_restock_date?: Date;
}

export interface InventoryWithProduct extends Inventory {
  product_name?: string;
  model?: string;
  price?: number;
  isLowStock?: boolean;
}

export interface InventoryCreate extends Omit<Inventory, 'inventory_id'> {
  inventory_id?: number;
}

export interface InventoryUpdate extends InventoryCreate {
  inventory_id: number;
}

export interface StockUpdate {
  current_stock: number;
}

export interface LowStockItem extends InventoryWithProduct {
  product_name: string;
  price: number;
}

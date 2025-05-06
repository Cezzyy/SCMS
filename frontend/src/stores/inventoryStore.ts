import { defineStore } from 'pinia';
import apiClient from '@/services/api';
import type { Inventory, InventoryCreate, InventoryUpdate, StockUpdate, LowStockItem } from '@/types/Inventory';

export const useInventoryStore = defineStore('inventory', {
  state: () => ({
    inventory: [] as Inventory[],
    lowStockItems: [] as LowStockItem[],
    currentInventory: null as Inventory | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    getInventoryById: (state) => (id: number): Inventory | undefined => {
      return state.inventory.find(item => item.inventory_id === id);
    },
    getInventoryByProductId: (state) => (productId: number): Inventory | undefined => {
      return state.inventory.find(item => item.product_id === productId);
    },
    isLoading: (state) => state.loading,
    hasError: (state) => state.error !== null
  },

  actions: {
    /**
     * Fetch all inventory items
     */
    async fetchInventory() {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<Inventory[]>('/api/inventory');
        this.inventory = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch inventory';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch an inventory item by ID
     */
    async fetchInventoryById(id: number) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<Inventory>(`/api/inventory/${id}`);
        this.currentInventory = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch inventory details';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch inventory by product ID
     */
    async fetchInventoryByProductId(productId: number) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<Inventory>(`/api/inventory/product/${productId}`);
        
        // Update in local collection if exists
        const index = this.inventory.findIndex(item => item.product_id === productId);
        if (index !== -1) {
          this.inventory[index] = response.data;
        } else {
          this.inventory.push(response.data);
        }

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch inventory for product';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Create a new inventory item
     */
    async createInventory(inventory: InventoryCreate) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.post<Inventory>('/api/inventory', inventory);
        
        // Add new inventory to the list
        this.inventory.push(response.data);

        return response.data;
      } catch (err) {
        this.error = 'Failed to create inventory item';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Update an existing inventory item
     */
    async updateInventory(inventory: InventoryUpdate) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.put<Inventory>(`/api/inventory/${inventory.inventory_id}`, inventory);

        // Update inventory in the list
        const index = this.inventory.findIndex(
          (item) => item.inventory_id === inventory.inventory_id
        );

        if (index !== -1) {
          this.inventory[index] = response.data;
        }

        // Update current inventory if it's the one being edited
        if (this.currentInventory?.inventory_id === inventory.inventory_id) {
          this.currentInventory = response.data;
        }

        return response.data;
      } catch (err) {
        this.error = 'Failed to update inventory item';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Update only the stock level of an inventory item
     */
    async updateStock(inventoryId: number, stockUpdate: StockUpdate) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.put<Inventory>(
          `/api/inventory/${inventoryId}/stock`, 
          stockUpdate
        );

        // Update inventory in the list
        const index = this.inventory.findIndex(
          (item) => item.inventory_id === inventoryId
        );

        if (index !== -1) {
          this.inventory[index] = response.data;
        }

        // Update current inventory if it's the one being edited
        if (this.currentInventory?.inventory_id === inventoryId) {
          this.currentInventory = response.data;
        }

        return response.data;
      } catch (err) {
        this.error = 'Failed to update stock level';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Delete an inventory item
     */
    async deleteInventory(id: number) {
      try {
        this.loading = true;
        this.error = null;

        await apiClient.delete(`/api/inventory/${id}`);

        // Remove inventory from the list
        this.inventory = this.inventory.filter(
          (item) => item.inventory_id !== id
        );

        // Reset current inventory if it's the one being deleted
        if (this.currentInventory?.inventory_id === id) {
          this.currentInventory = null;
        }

        return true;
      } catch (err) {
        this.error = 'Failed to delete inventory item';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch all inventory items that are low on stock
     */
    async fetchLowStockItems() {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<Inventory[]>('/api/inventory/low-stock');
        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch low stock items';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch low stock items with product details
     */
    async fetchLowStockWithProductInfo() {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<LowStockItem[]>('/api/inventory/low-stock/details');
        this.lowStockItems = response.data;
        
        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch low stock items with product info';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    }
  }
}); 
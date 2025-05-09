import { defineStore } from 'pinia';
import apiClient from '@/services/api';
import type { Order, OrderItem } from '@/types/Order';

interface OrderWithItems {
  order: Order;
  items: OrderItem[];
}

export interface OrderCreate {
  order: Omit<Order, 'order_id' | 'created_at' | 'updated_at'>;
  items: Omit<OrderItem, 'order_item_id' | 'order_id'>[];
}

export interface OrderUpdate {
  order_id: number;
  customer_id: number;
  quotation_id?: number;
  order_date: string;
  shipping_address: string;
  status: 'Pending' | 'Shipped' | 'Delivered' | 'Cancelled';
  total_amount: number;
}

export const useOrderStore = defineStore('order', {
  state: () => ({
    orders: [] as Order[],
    currentOrder: null as OrderWithItems | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    getOrderById: (state) => (id: number): Order | undefined => {
      return state.orders.find(order => order.order_id === id);
    },
    isLoading: (state) => state.loading,
    hasError: (state) => state.error !== null
  },

  actions: {
    /**
     * Fetch all orders
     */
    async fetchOrders() {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<Order[]>('/api/orders');
        this.orders = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch orders';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch an order by ID with its items
     */
    async fetchOrderById(id: number) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<OrderWithItems>(`/api/orders/${id}`);
        this.currentOrder = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch order details';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Create a new order with items
     */
    async createOrder(orderData: OrderCreate) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.post<OrderWithItems>('/api/orders', orderData);

        // Add new order to the list
        this.orders.push(response.data.order);

        return response.data;
      } catch (err) {
        this.error = 'Failed to create order';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Update an existing order
     */
    async updateOrder(order: OrderUpdate) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.put<Order>(`/api/orders/${order.order_id}`, order);

        // Update order in the list
        const index = this.orders.findIndex(
          (o) => o.order_id === order.order_id
        );

        if (index !== -1) {
          this.orders[index] = response.data;
        }

        // Update current order if it's the one being edited
        if (this.currentOrder?.order.order_id === order.order_id) {
          this.currentOrder.order = response.data;
        }

        return response.data;
      } catch (err) {
        this.error = 'Failed to update order';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Delete an order
     */
    async deleteOrder(id: number) {
      try {
        this.loading = true;
        this.error = null;

        await apiClient.delete(`/api/orders/${id}`);

        // Remove order from the list
        this.orders = this.orders.filter(
          (order) => order.order_id !== id
        );

        // Reset current order if it's the one being deleted
        if (this.currentOrder?.order.order_id === id) {
          this.currentOrder = null;
        }

        return true;
      } catch (err) {
        this.error = 'Failed to delete order';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    }
  }
}); 
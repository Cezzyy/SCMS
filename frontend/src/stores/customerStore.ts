import { defineStore } from 'pinia';
import apiClient from '@/services/api';
import type { Customer, CustomerCreate, CustomerUpdate } from '@/types/Customer';

export const useCustomerStore = defineStore('customer', {
  state: () => ({
    customers: [] as Customer[],
    currentCustomer: null as Customer | null,
    loading: false,
    error: null as string | null
  }),
  
  getters: {
    getCustomerById: (state) => (id: number): Customer | undefined => {
      return state.customers.find(customer => customer.customer_id === id);
    },
    isLoading: (state) => state.loading,
    hasError: (state) => state.error !== null
  },
  
  actions: {
    /**
     * Fetch all customers
     */
    async fetchCustomers(search?: string) {
      try {
        this.loading = true;
        this.error = null;
        
        const params = search ? { search } : {};
        const response = await apiClient.get<Customer[]>('/customers', { params });
        this.customers = response.data;
        
        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch customers';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch a customer by ID
     */
    async fetchCustomerById(id: number) {
      try {
        this.loading = true;
        this.error = null;
        
        const response = await apiClient.get<Customer>(`/customers/${id}`);
        this.currentCustomer = response.data;
        
        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch customer details';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Create a new customer
     */
    async createCustomer(customer: CustomerCreate) {
      try {
        this.loading = true;
        this.error = null;
        
        const response = await apiClient.post<Customer>('/customers', customer);
        
        // Add new customer to the list
        this.customers.push(response.data);
        
        return response.data;
      } catch (err) {
        this.error = 'Failed to create customer';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Update an existing customer
     */
    async updateCustomer(customer: CustomerUpdate) {
      try {
        this.loading = true;
        this.error = null;
        
        const response = await apiClient.put<Customer>(`/customers/${customer.customer_id}`, customer);
        
        // Update customer in the list
        const index = this.customers.findIndex(
          (c) => c.customer_id === customer.customer_id
        );
        
        if (index !== -1) {
          this.customers[index] = response.data;
        }
        
        // Update current customer if it's the one being edited
        if (this.currentCustomer?.customer_id === customer.customer_id) {
          this.currentCustomer = response.data;
        }
        
        return response.data;
      } catch (err) {
        this.error = 'Failed to update customer';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Delete a customer
     */
    async deleteCustomer(id: number) {
      try {
        this.loading = true;
        this.error = null;
        
        await apiClient.delete(`/customers/${id}`);
        
        // Remove customer from the list
        this.customers = this.customers.filter(
          (customer) => customer.customer_id !== id
        );
        
        // Reset current customer if it's the one being deleted
        if (this.currentCustomer?.customer_id === id) {
          this.currentCustomer = null;
        }
        
        return true;
      } catch (err) {
        this.error = 'Failed to delete customer';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    }
  }
}); 
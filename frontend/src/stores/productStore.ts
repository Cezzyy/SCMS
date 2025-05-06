import { defineStore } from 'pinia';
import apiClient from '@/services/api';
import type { Product, ProductCreate, ProductUpdate } from '@/types/Product';

export const useProductStore = defineStore('product', {
  state: () => ({
    products: [] as Product[],
    currentProduct: null as Product | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    getProductById: (state) => (id: number): Product | undefined => {
      return state.products.find(product => product.product_id === id);
    },
    isLoading: (state) => state.loading,
    hasError: (state) => state.error !== null
  },

  actions: {
    /**
     * Fetch all products
     */
    async fetchProducts(search?: string) {
      try {
        this.loading = true;
        this.error = null;

        const params = search ? { search } : {};
        const response = await apiClient.get<Product[]>('/api/products', { params });
        this.products = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch products';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch a product by ID
     */
    async fetchProductById(id: number) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<Product>(`/api/products/${id}`);
        this.currentProduct = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch product details';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Create a new product
     */
    async createProduct(product: ProductCreate) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.post<Product>('/api/products', product);
        
        // Add new product to the list
        this.products.push(response.data);

        return response.data;
      } catch (err) {
        this.error = 'Failed to create product';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Update an existing product
     */
    async updateProduct(product: ProductUpdate) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.put<Product>(`/api/products/${product.product_id}`, product);

        // Update product in the list
        const index = this.products.findIndex(
          (p) => p.product_id === product.product_id
        );

        if (index !== -1) {
          this.products[index] = response.data;
        }

        // Update current product if it's the one being edited
        if (this.currentProduct?.product_id === product.product_id) {
          this.currentProduct = response.data;
        }

        return response.data;
      } catch (err) {
        this.error = 'Failed to update product';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Delete a product
     */
    async deleteProduct(id: number) {
      try {
        this.loading = true;
        this.error = null;

        await apiClient.delete(`/api/products/${id}`);

        // Remove product from the list
        this.products = this.products.filter(
          (product) => product.product_id !== id
        );

        // Reset current product if it's the one being deleted
        if (this.currentProduct?.product_id === id) {
          this.currentProduct = null;
        }

        return true;
      } catch (err) {
        this.error = 'Failed to delete product';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    }
  }
}); 
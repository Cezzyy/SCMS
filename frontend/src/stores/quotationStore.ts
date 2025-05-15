import { defineStore } from 'pinia';
import apiClient from '@/services/api';
import type { 
  Quotation, 
  QuotationCreate, 
  QuotationUpdate, 
  QuotationWithItems 
} from '@/types/Quotation';
import axios from 'axios';

export const useQuotationStore = defineStore('quotation', {
  state: () => ({
    quotations: [] as Quotation[],
    currentQuotation: null as QuotationWithItems | null,
    loading: false,
    error: null as string | null
  }),

  getters: {
    getQuotationById: (state) => (id: number): Quotation | undefined => {
      return state.quotations.find(quotation => quotation.quotation_id === id);
    },
    isLoading: (state) => state.loading,
    hasError: (state) => state.error !== null
  },

  actions: {
    /**
     * Fetch all quotations
     */
    async fetchQuotations(params?: { customer_id?: number, status?: string }) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<Quotation[]>('/api/quotations', { params });
        this.quotations = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch quotations';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Fetch a quotation by ID with its items
     */
    async fetchQuotationById(id: number) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.get<QuotationWithItems>(`/api/quotations/${id}`);
        this.currentQuotation = response.data;

        return response.data;
      } catch (err) {
        this.error = 'Failed to fetch quotation details';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Create a new quotation
     */
    async createQuotation(quotation: QuotationCreate) {
      try {
        this.loading = true;
        this.error = null;

        // Convert date strings to ISO format for Go's time.Time parsing
        const quoteDate = new Date(quotation.quote_date).toISOString();
        const validityDate = new Date(quotation.validity_date).toISOString();

        // Create the payload as a plain JavaScript object (not stringified)
        const payload = {
          quotation: {
            customer_id: Number(quotation.customer_id),
            quote_date: quoteDate,
            validity_date: validityDate,
            status: quotation.status,
            total_amount: quotation.total_amount
          },
          items: quotation.items ? quotation.items.map(item => ({
            product_id: item.product_id,
            quantity: item.quantity,
            unit_price: item.unit_price,
            discount: item.discount
          })) : []
        };

        console.log('Sending payload:', payload);
        
        // Use a simple axios POST with default settings
        const response = await axios.post(
          'http://localhost:8081/api/quotations',
          payload,
          {
            headers: { 'Content-Type': 'application/json' }
          }
        );
        
        console.log('Success! Response:', response.data);
        
        // Add to store if successful
        this.quotations.push(response.data);
        return response.data;
      } catch (err: any) {
        this.error = 'Failed to create quotation';
        console.error('Error details:', err);
        // Show the response error message if available
        if (err.response && err.response.data) {
          console.error('Server error:', err.response.data);
        }
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Update an existing quotation
     */
    async updateQuotation(quotation: QuotationUpdate) {
      try {
        this.loading = true;
        this.error = null;

        // Convert date strings to ISO format for Go's time.Time parsing
        const quoteDate = new Date(quotation.quote_date).toISOString();
        const validityDate = new Date(quotation.validity_date).toISOString();

        // Create the payload as a plain JavaScript object (not stringified)
        const payload = {
          quotation: {
            quotation_id: quotation.quotation_id,
            customer_id: Number(quotation.customer_id),
            quote_date: quoteDate,
            validity_date: validityDate,
            status: quotation.status,
            total_amount: quotation.total_amount
          },
          items: quotation.items ? quotation.items.map(item => ({
            product_id: item.product_id,
            quantity: item.quantity,
            unit_price: item.unit_price,
            discount: item.discount
          })) : []
        };

        console.log('Sending payload:', payload);
        
        // Use a simple axios PUT with default settings
        const response = await axios.put(
          `http://localhost:8081/api/quotations/${quotation.quotation_id}`,
          payload,
          {
            headers: { 'Content-Type': 'application/json' }
          }
        );
        
        // Update quotation in the list
        const index = this.quotations.findIndex(
          (q) => q.quotation_id === quotation.quotation_id
        );

        if (index !== -1) {
          this.quotations[index] = response.data;
        }

        // Update current quotation if it's the one being edited
        if (this.currentQuotation?.quotation_id === quotation.quotation_id) {
          // Preserve the items from the current quotation
          this.currentQuotation = {
            ...response.data,
            items: this.currentQuotation.items
          };
        }

        return response.data;
      } catch (err: any) {
        this.error = 'Failed to update quotation';
        console.error('Error details:', err);
        // Show the response error message if available
        if (err.response && err.response.data) {
          console.error('Server error:', err.response.data);
        }
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Delete a quotation
     */
    async deleteQuotation(id: number) {
      try {
        this.loading = true;
        this.error = null;

        await apiClient.delete(`/api/quotations/${id}`);

        // Remove quotation from the list
        this.quotations = this.quotations.filter(
          (quotation) => quotation.quotation_id !== id
        );

        // Reset current quotation if it's the one being deleted
        if (this.currentQuotation?.quotation_id === id) {
          this.currentQuotation = null;
        }

        return true;
      } catch (err) {
        this.error = 'Failed to delete quotation';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Change quotation status
     */
    async updateQuotationStatus(id: number, status: Quotation['status']) {
      try {
        this.loading = true;
        this.error = null;

        const response = await apiClient.post<Quotation>(
          `/api/quotations/${id}/status`, 
          { status }
        );

        // Update quotation in the list
        const index = this.quotations.findIndex(q => q.quotation_id === id);
        if (index !== -1) {
          this.quotations[index] = response.data;
        }

        // Update current quotation if it's the one being updated
        if (this.currentQuotation?.quotation_id === id) {
          this.currentQuotation = {
            ...this.currentQuotation,
            status: response.data.status
          };
        }

        return response.data;
      } catch (err) {
        this.error = 'Failed to update quotation status';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    },

    /**
     * Download quotation as PDF
     */
    async downloadPDF(id: number) {
      try {
        this.loading = true;
        this.error = null;
        
        // Request PDF with blob response type
        const response = await apiClient.get(
          `/api/quotations/${id}/pdf`, 
          { responseType: 'blob' }
        );
        
        // Create a URL for the blob
        const blob = new Blob([response.data], { type: 'application/pdf' });
        const url = window.URL.createObjectURL(blob);
        
        // Create a link element and trigger download
        const link = document.createElement('a');
        link.href = url;
        link.download = `quotation-${id}.pdf`;
        document.body.appendChild(link);
        link.click();
        
        // Clean up
        window.URL.revokeObjectURL(url);
        document.body.removeChild(link);
        
        return true;
      } catch (err) {
        this.error = 'Failed to download quotation PDF';
        console.error(err);
        throw err;
      } finally {
        this.loading = false;
      }
    }
  }
}); 
import { defineStore } from 'pinia';
import axios from 'axios';
import type { DashboardSummary, SalesTrend, LowStockItem, TopCustomer, ReportParams } from '@/types/reports';

// Make sure API_URL has the proper structure - the issue is likely here
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081';
const API_URL = `${API_BASE_URL}/api`;

interface ReportState {
  dashboard: DashboardSummary | null;
  salesTrends: SalesTrend[];
  lowStockItems: LowStockItem[];
  topCustomers: TopCustomer[];
  loading: {
    dashboard: boolean;
    sales: boolean;
    stock: boolean;
    customers: boolean;
  };
  error: {
    dashboard: string | null;
    sales: string | null;
    stock: string | null;
    customers: string | null;
  };
  period: number; // days
}

export const useReportStore = defineStore('report', {
  state: (): ReportState => ({
    dashboard: null,
    salesTrends: [],
    lowStockItems: [],
    topCustomers: [],
    loading: {
      dashboard: false,
      sales: false,
      stock: false,
      customers: false
    },
    error: {
      dashboard: null,
      sales: null,
      stock: null,
      customers: null
    },
    period: 7 // Default to 7 days
  }),
  
  getters: {
    totalSales: (state) => state.dashboard?.total_sales || 0,
    orderCount: (state) => state.dashboard?.order_count || 0,
    lowStockCount: (state) => state.dashboard?.low_stock_count || 0,
    formattedPeriod: (state) => state.dashboard?.period || `Last ${state.period} days`,
    
    // Add avatars to top customers for UI display
    topCustomersWithAvatars: (state) => {
      const avatars = ['👔', '💻', '🌐', '🛒', '⚙️'];
      return state.topCustomers.map((customer, index) => ({
        ...customer,
        avatar: customer.avatar || avatars[index % avatars.length]
      }));
    }
  },
  
  actions: {
    async fetchDashboardSummary(params: ReportParams = {}) {
      this.loading.dashboard = true;
      this.error.dashboard = null;
      
      try {
        const days = params.days || this.period;
        this.period = days;
        
        console.log(`Fetching dashboard data from: ${API_URL}/dashboard`);
        const response = await axios.get(`${API_URL}/dashboard`, {
          params: { days }
        });
        
        this.dashboard = response.data;
        this.salesTrends = response.data.sales_trends;
        this.lowStockItems = response.data.low_stock_items;
        this.topCustomers = response.data.top_customers;
        
        return response.data;
      } catch (error: any) {
        this.error.dashboard = error.response?.data?.error || 'Failed to load dashboard data';
        console.error('Error fetching dashboard data:', error);
        throw error;
      } finally {
        this.loading.dashboard = false;
      }
    },
    
    async fetchSalesTrends(params: ReportParams = {}) {
      this.loading.sales = true;
      this.error.sales = null;
      
      try {
        const days = params.days || this.period;
        this.period = days;
        
        const response = await axios.get(`${API_URL}/reports/sales-trends`, {
          params: { days }
        });
        
        this.salesTrends = response.data;
        return response.data;
      } catch (error: any) {
        this.error.sales = error.response?.data?.error || 'Failed to load sales trends';
        console.error('Error fetching sales trends:', error);
        throw error;
      } finally {
        this.loading.sales = false;
      }
    },
    
    async fetchLowStockItems() {
      this.loading.stock = true;
      this.error.stock = null;
      
      try {
        const response = await axios.get(`${API_URL}/reports/low-stock`);
        this.lowStockItems = response.data;
        return response.data;
      } catch (error: any) {
        this.error.stock = error.response?.data?.error || 'Failed to load inventory data';
        console.error('Error fetching low stock items:', error);
        throw error;
      } finally {
        this.loading.stock = false;
      }
    },
    
    async fetchTopCustomers(params: ReportParams = {}) {
      this.loading.customers = true;
      this.error.customers = null;
      
      try {
        const days = params.days || 365; // Default to 1 year for top customers
        const limit = params.limit || 5;
        
        const response = await axios.get(`${API_URL}/reports/top-customers`, {
          params: { days, limit }
        });
        
        this.topCustomers = response.data;
        return response.data;
      } catch (error: any) {
        this.error.customers = error.response?.data?.error || 'Failed to load customer data';
        console.error('Error fetching top customers:', error);
        throw error;
      } finally {
        this.loading.customers = false;
      }
    },
    
    // Utility method to refresh all data
    async refreshAll(params: ReportParams = {}) {
      try {
        await this.fetchDashboardSummary(params);
      } catch (error) {
        // Individual errors are already handled in each method
        console.error('Error refreshing all data:', error);
      }
    },
    
    // Clear error messages
    dismissError(type: 'dashboard' | 'sales' | 'stock' | 'customers') {
      this.error[type] = null;
    },
    
    // Download CSV export (placeholder - actual implementation would depend on backend)
    downloadSalesCSV() {
      window.open(`${API_URL}/reports/sales-trends/export?days=${this.period}`, '_blank');
    },
    
    downloadStockCSV() {
      window.open(`${API_URL}/reports/low-stock/export`, '_blank');
    },
    
    downloadCustomersCSV() {
      window.open(`${API_URL}/reports/top-customers/export?days=365`, '_blank');
    }
  }
}); 
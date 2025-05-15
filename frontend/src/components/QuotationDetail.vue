<script setup lang="ts">
import { ref, onMounted, computed, defineProps, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuotationStore } from '../stores/quotationStore';
import { useCustomerStore } from '../stores/customerStore';
import { useProductStore } from '../stores/productStore';

// Optional props for when used as a modal
const props = defineProps({
  isModal: {
    type: Boolean,
    default: false
  }
});

const route = useRoute();
const router = useRouter();
const quotationStore = useQuotationStore();
const customerStore = useCustomerStore();
const productStore = useProductStore();

// State
const isLoading = ref(true);
const quotationId = computed(() => {
  // Handle different data structures to ensure we always get a valid ID
  if (props.isModal) {
    const data = quotationStore.currentQuotation as any;
    if (!data) return 0;
    
    // Handle different API response structures
    if (data.quotation_id) {
      return data.quotation_id;
    } else if (data.quotation && data.quotation.quotation_id) {
      return data.quotation.quotation_id;
    }
    
    // If we still don't have an ID, check for any property that might be the ID
    for (const key in data) {
      if (key.includes('id') && typeof data[key] === 'number') {
        return data[key];
      }
    }
    
    console.error('Failed to extract quotation ID from:', data);
    return 0;
  } else {
    // In standalone mode, ID comes from the route
    return Number(route.params.id) || 0;
  }
});

// Define an interface for the quotation response which might be different based on API structure
interface QuotationData {
  quotation_id?: number;
  customer_id?: number;
  quote_date?: string;
  validity_date?: string;
  status?: string;
  total_amount?: number;
  created_at?: string;
  updated_at?: string;
  items?: any[];
  [key: string]: any; // Allow other properties to handle different API formats
}

const quotation = computed<QuotationData | null>(() => {
  if (props.isModal) {
    // In modal mode, explicitly use the currentQuotation from store
    console.log('Modal mode: accessing currentQuotation from store:', quotationStore.currentQuotation);
    const data = quotationStore.currentQuotation as any;
    
    if (!data) return null;
    
    // Handle different API response structures
    if (data.quotation && typeof data.quotation === 'object') {
      // Nested structure: {quotation: {...}, items: [...]}
      return data.quotation as QuotationData;
    } else if (data.customer_id && data.quotation_id) {
      // Direct structure
      return data as QuotationData;
    }
    
    // Fallback
    return data as QuotationData;
  } else {
    // In standalone mode, we've already fetched it ourselves
    return quotationStore.currentQuotation as QuotationData;
  }
});

const customer = computed(() => {
  if (!quotation.value) {
    console.error('Missing quotation data:', quotationStore.currentQuotation);
    return null;
  }
  
  // Get the customer ID from the quotation object, regardless of structure
  const customerId = quotation.value.customer_id;
  if (!customerId) {
    console.error('Missing customer_id in quotation:', quotation.value);
    return null;
  }
  
  console.log('Looking for customer with ID:', customerId);
  console.log('Available customers:', customerStore.customers);
  
  // Find the customer in the store
  const foundCustomer = customerStore.customers.find(c => c.customer_id === customerId);
  
  if (!foundCustomer) {
    console.error('Customer not found with ID:', customerId);
  } else {
    console.log('Found customer:', foundCustomer);
  }
  
  return foundCustomer || null;
});

// Get the items from the quotation, handling different API structures
const quotationItems = computed(() => {
  if (!quotationStore.currentQuotation) return [];
  
  const data = quotationStore.currentQuotation as any;
  
  // Handle different API response structures
  if (data.items && Array.isArray(data.items)) {
    // Items directly in the response
    console.log('Found items array in top-level data:', data.items);
    return data.items;
  } else if (data.quotation && data.quotation.items && Array.isArray(data.quotation.items)) {
    // Items in nested quotation object
    console.log('Found items array in nested quotation data:', data.quotation.items);
    return data.quotation.items;
  }
  
  console.warn('No items array found in quotation data:', data);
  return [];
});

// Add notification state variables
const showNotification = ref(false);
const notificationMessage = ref('');
const notificationIsError = ref(false);

// Load data
const loadData = async () => {
  isLoading.value = true;
  try {
    // For non-modal mode, we need to handle the complete data loading
    if (!props.isModal && quotationId.value) {
      console.log('Standalone mode: Loading quotation with ID:', quotationId.value);
      await quotationStore.fetchQuotationById(quotationId.value);
    }
    
    // Always load customer and product data
    console.log('Loading customer and product data...');
    await Promise.all([
      customerStore.fetchCustomers(),
      productStore.fetchProducts()
    ]);
    
    // Verify data after loading
    console.log('After loading - Current quotation:', quotationStore.currentQuotation);
    console.log('After loading - Customer data loaded:', customerStore.customers.length > 0);
    console.log('After loading - Product data loaded:', productStore.products.length > 0);
    
  } catch (error) {
    console.error('Error loading quotation details:', error);
  } finally {
    isLoading.value = false;
  }
};

// Format money
const formatMoney = (amount: number | undefined): string => {
  if (amount === undefined) return 'N/A';
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
};

// Download PDF
const downloadPDF = async () => {
  console.log('Attempting to download PDF with quotationId:', quotationId.value);
  
  if (!quotationId.value) {
    console.error('Cannot download PDF: No valid quotation ID', { 
      currentQuotation: quotationStore.currentQuotation,
      isModal: props.isModal
    });
    return;
  }
  
  try {
    console.log('Downloading PDF for quotation ID:', quotationId.value);
    // Use window.open to handle redirects and auth properly
    const apiUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8081'}/api/quotations/${quotationId.value}/pdf`;
    console.log('Opening PDF URL in new tab:', apiUrl);
    
    // Use window.open which handles cookies and auth context better for PDFs
    window.open(apiUrl, '_blank');
    
  } catch (error) {
    console.error('Error downloading PDF:', error);
    alert(`Failed to download PDF: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
};

// Update quotation status
const updateStatus = async (status: 'Pending' | 'Approved' | 'Rejected' | 'Expired') => {
  console.log('Attempting to update status to', status, 'with quotationId:', quotationId.value);
  
  if (!quotationId.value) {
    // Display error notification instead of alert
    notificationMessage.value = 'Cannot update status: No valid quotation ID';
    notificationIsError.value = true;
    showNotification.value = true;
    
    // Auto-hide notification after 5 seconds
    setTimeout(() => {
      showNotification.value = false;
    }, 5000);
    return;
  }
  
  try {
    console.log('Updating status to', status, 'for quotation ID:', quotationId.value);
    
    // Use the store method which now uses POST instead of PATCH
    await quotationStore.updateQuotationStatus(quotationId.value, status);
    
    // Show success notification
    notificationMessage.value = `Status has been updated to ${status}`;
    notificationIsError.value = false;
    showNotification.value = true;
    
    // Auto-hide notification after 5 seconds
    setTimeout(() => {
      showNotification.value = false;
    }, 5000);
    
    // Refresh the data if needed
    if (!props.isModal) {
      await quotationStore.fetchQuotationById(quotationId.value);
    }
  } catch (error) {
    console.error('Error updating quotation status:', error);
    
    // Show error notification instead of alert
    notificationMessage.value = `Failed to update status: ${error instanceof Error ? error.message : 'Unknown error'}`;
    notificationIsError.value = true;
    showNotification.value = true;
    
    // Auto-hide notification after 5 seconds
    setTimeout(() => {
      showNotification.value = false;
    }, 5000);
  }
};

// Format dates
const formatDate = (dateString: string | undefined): string => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleDateString();
};

// Get product name by ID
const getProductName = (productId: number): string => {
  const product = productStore.products.find(p => p.product_id === productId);
  return product ? product.product_name : 'Unknown Product';
};

// Get status badge class
const getStatusBadgeClass = (status: string): string => {
  switch (status) {
    case 'Pending':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300';
    case 'Approved':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300';
    case 'Rejected':
      return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300';
    case 'Expired':
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300';
    default:
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300';
  }
};

// Go back to list
const goBack = () => {
  if (props.isModal) {
    // We don't need to navigate when in modal mode
    return;
  }
  router.push('/quotations');
};

// Load data on component mount - only for standalone mode
onMounted(() => {
  if (!props.isModal) {
    console.log('Component mounted in standalone mode, loading data...');
    loadData();
  } else {
    console.log('Component mounted in modal mode, data should already be loaded by parent');
  }
});

// When the component is shown in modal mode, ensure we have customer data
watch(() => props.isModal, async (isModal) => {
  if (isModal) {
    console.log('QuotationDetail in modal mode, currentQuotation:', quotationStore.currentQuotation);
    
    try {
      // Make sure we have the correct quotation ID and structure
      if (quotationStore.currentQuotation) {
        console.log('Quotation data structure:', quotationStore.currentQuotation);
        console.log('Extracted quotation ID:', quotationId.value);
        
        // Check the structure of the data
        if ('quotation' in quotationStore.currentQuotation) {
          // API sometimes returns {quotation: {...}, items: [...]}
          console.log('Using nested quotation data structure');
        } else if ('customer_id' in quotationStore.currentQuotation) {
          // API sometimes returns the quotation directly
          console.log('Using direct quotation data structure');
        }
      }
      
      // Always fetch customer and product data to ensure they're available
      await Promise.all([
        customerStore.fetchCustomers(),
        productStore.fetchProducts()
      ]);
      
      console.log('After loading, currentQuotation:', quotationStore.currentQuotation);
      console.log('After loading, quotationId:', quotationId.value);
      console.log('After loading, customer data:', customerStore.customers);
      console.log('After loading, isLoading state:', isLoading.value);
      
      // Force isLoading to false
      isLoading.value = false;
    } catch (error) {
      console.error('Error in modal initialization:', error);
      isLoading.value = false;
    }
  }
}, { immediate: true });
</script>

<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
    <!-- Notification -->
    <div v-if="showNotification" 
         :class="[
           'fixed top-4 right-4 z-50 p-4 rounded-md shadow-lg flex items-center max-w-sm transition-all transform',
           notificationIsError ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200' : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
         ]">
      <div :class="['flex-shrink-0 mr-3', notificationIsError ? 'text-red-500 dark:text-red-300' : 'text-green-500 dark:text-green-300']">
        <svg v-if="notificationIsError" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
        </svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
        </svg>
      </div>
      <div>
        <p>{{ notificationMessage }}</p>
      </div>
      <button @click="showNotification = false" class="ml-auto text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
        </svg>
      </button>
    </div>
    
    <!-- Header with actions -->
    <div class="flex flex-col md:flex-row md:justify-between md:items-center p-4 md:p-6 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-gray-800 dark:to-gray-900">
      <div class="mb-4 md:mb-0">
        <div class="flex items-center">
          <button 
            v-if="!isModal"
            @click="goBack" 
            class="mr-3 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300"
            title="Back to list"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </button>
          <h2 class="text-xl md:text-2xl font-bold text-gray-800 dark:text-white">
            Quotation #{{ quotation ? quotation.quotation_id : 'Loading...' }}
          </h2>
          <span 
            v-if="quotation && quotation.status" 
            :class="['ml-3 px-2.5 py-1 text-xs font-semibold rounded-full', getStatusBadgeClass(quotation.status)]"
          >
            {{ quotation?.status }}
          </span>
        </div>
        <p class="text-gray-600 dark:text-gray-300 text-xs md:text-sm mt-1">
          {{ customer?.company_name || 'Loading customer...' }}
        </p>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="isLoading" class="flex justify-center items-center p-12">
      <div class="relative">
        <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-200 dark:border-gray-600"></div>
        <div class="animate-spin rounded-full h-12 w-12 border-t-4 border-blue-600 dark:border-blue-500 absolute top-0 left-0"></div>
      </div>
      <div class="ml-4 text-gray-600 dark:text-gray-300 text-sm font-medium">Loading quotation details...</div>
    </div>

    <!-- Debug info -->
    <div v-if="props.isModal && !isLoading && (!quotation || !customer || quotationItems.length === 0)">
      <div class="p-6 bg-red-50 dark:bg-red-900 text-red-800 dark:text-red-200 rounded-md m-4">
        <h3 class="font-bold mb-2">Debug Information</h3>
        <p>Quotation data may not be loaded properly.</p>
        <p class="mt-2">Current data state:</p>
        <ul class="list-disc pl-5 mt-1">
          <li>isLoading: {{ isLoading }}</li>
          <li>isModal: {{ props.isModal }}</li>
          <li>Store currentQuotation exists: {{ quotationStore.currentQuotation ? 'Yes' : 'No' }}</li>
          <li>Store currentQuotation structure: {{ quotationStore.currentQuotation ? Object.keys(quotationStore.currentQuotation).join(', ') : 'N/A' }}</li>
          <li>Computed quotation exists: {{ quotation ? 'Yes' : 'No' }}</li>
          <li>Customer ID from data: {{ quotation?.customer_id }}</li>
          <li>Customer store length: {{ customerStore.customers.length }}</li>
          <li>Found customer: {{ customer ? 'Yes' : 'No' }}</li>
          <li>Items found: {{ quotationItems.length }}</li>
          <li>Items source: {{ quotationStore.currentQuotation?.items ? 'Direct' : ((quotationStore.currentQuotation as any)?.quotation?.items ? 'Nested' : 'None') }}</li>
        </ul>
      </div>
    </div>

    <!-- Not found state -->
    <div v-else-if="!quotation && !isLoading" class="p-8 text-center">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto text-gray-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
      </svg>
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">Quotation Not Found</h3>
      <p class="text-gray-500 dark:text-gray-400 mb-6">The quotation you're looking for doesn't exist or has been deleted.</p>
      <button 
        @click="goBack" 
        class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700"
      >
        Back to Quotations
      </button>
    </div>

    <!-- Quotation details -->
    <div v-else-if="quotation && !isLoading" class="p-4 md:p-6">
      <!-- Quotation information section -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <!-- Customer information -->
        <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg border border-gray-200 dark:border-gray-600">
          <h3 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase mb-3">Customer Information</h3>
          <div class="space-y-2">
            <p class="text-gray-800 dark:text-gray-200 font-medium">{{ customer?.company_name }}</p>
            <p class="text-gray-600 dark:text-gray-400 text-sm">{{ customer?.address || 'No address provided' }}</p>
            <p class="text-gray-600 dark:text-gray-400 text-sm">{{ customer?.email || 'No email provided' }}</p>
            <p class="text-gray-600 dark:text-gray-400 text-sm">{{ customer?.phone || 'No phone provided' }}</p>
          </div>
        </div>

        <!-- Quotation details -->
        <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg border border-gray-200 dark:border-gray-600">
          <h3 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase mb-3">Quotation Details</h3>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">Quote Date</p>
              <p class="text-gray-800 dark:text-gray-200">{{ formatDate(quotation.quote_date) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">Valid Until</p>
              <p class="text-gray-800 dark:text-gray-200">{{ formatDate(quotation.validity_date) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">Quotation ID</p>
              <p class="text-gray-800 dark:text-gray-200">#{{ quotation.quotation_id }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-gray-400">Status</p>
              <p :class="['text-sm font-medium', 
                quotation.status === 'Approved' ? 'text-green-600 dark:text-green-400' : 
                quotation.status === 'Rejected' ? 'text-red-600 dark:text-red-400' :
                quotation.status === 'Expired' ? 'text-gray-500 dark:text-gray-400' :
                'text-yellow-600 dark:text-yellow-400'
              ]">{{ quotation.status }}</p>
            </div>
          </div>
        </div>

        <!-- Update status & actions -->
        <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg border border-gray-200 dark:border-gray-600">
          <h3 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase mb-3">Actions</h3>
          <div class="space-y-3">
            <p class="text-xs text-gray-500 dark:text-gray-400">Update Status</p>
            <div class="flex flex-wrap gap-2">
              <button 
                @click="updateStatus('Approved')" 
                :disabled="quotation.status === 'Approved'"
                :class="[
                  'px-3 py-1.5 rounded text-sm font-medium', 
                  quotation.status === 'Approved' ? 
                    'bg-gray-100 text-gray-400 dark:bg-gray-600 dark:text-gray-500 cursor-not-allowed' : 
                    'bg-green-100 text-green-700 hover:bg-green-200 dark:bg-green-900 dark:text-green-300 dark:hover:bg-green-800'
                ]"
              >
                Approve
              </button>
              <button 
                @click="updateStatus('Rejected')" 
                :disabled="quotation.status === 'Rejected'"
                :class="[
                  'px-3 py-1.5 rounded text-sm font-medium', 
                  quotation.status === 'Rejected' ? 
                    'bg-gray-100 text-gray-400 dark:bg-gray-600 dark:text-gray-500 cursor-not-allowed' : 
                    'bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-900 dark:text-red-300 dark:hover:bg-red-800'
                ]"
              >
                Reject
              </button>
              <button 
                @click="updateStatus('Pending')" 
                :disabled="quotation.status === 'Pending'"
                :class="[
                  'px-3 py-1.5 rounded text-sm font-medium', 
                  quotation.status === 'Pending' ? 
                    'bg-gray-100 text-gray-400 dark:bg-gray-600 dark:text-gray-500 cursor-not-allowed' : 
                    'bg-yellow-100 text-yellow-700 hover:bg-yellow-200 dark:bg-yellow-900 dark:text-yellow-300 dark:hover:bg-yellow-800'
                ]"
              >
                Mark Pending
              </button>
              <button 
                @click="updateStatus('Expired')" 
                :disabled="quotation.status === 'Expired'"
                :class="[
                  'px-3 py-1.5 rounded text-sm font-medium', 
                  quotation.status === 'Expired' ? 
                    'bg-gray-100 text-gray-400 dark:bg-gray-600 dark:text-gray-500 cursor-not-allowed' : 
                    'bg-gray-200 text-gray-700 hover:bg-gray-300 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
                ]"
              >
                Mark Expired
              </button>
            </div>
            <div class="pt-3 border-t border-gray-200 dark:border-gray-600 mt-3">
              <button 
                @click="downloadPDF" 
                class="w-full flex items-center justify-center px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition-colors"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-2" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M6 2a2 2 0 00-2 2v12a2 2 0 002 2h8a2 2 0 002-2V7.414A2 2 0 0015.414 6L12 2.586A2 2 0 0010.586 2H6zm5 6a1 1 0 10-2 0v3.586l-1.293-1.293a1 1 0 10-1.414 1.414l3 3a1 1 0 001.414 0l3-3a1 1 0 00-1.414-1.414L11 11.586V8z" clip-rule="evenodd" />
                </svg>
                Download PDF
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Quotation items table -->
      <div class="mb-6" v-if="quotationItems && quotationItems.length > 0">
        <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-4">Quotation Items</h3>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead class="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Product
                </th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Quantity
                </th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Unit Price
                </th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Discount
                </th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                  Line Total
                </th>
              </tr>
            </thead>
            <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-for="item in quotationItems" :key="item.quotation_item_id">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="font-medium text-gray-900 dark:text-white">{{ getProductName(item.product_id) }}</div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-gray-500 dark:text-gray-300">
                  {{ item.quantity }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-gray-500 dark:text-gray-300">
                  {{ formatMoney(item.unit_price) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right text-gray-500 dark:text-gray-300">
                  {{ item.discount || 0 }}%
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-right font-medium text-gray-900 dark:text-white">
                  {{ formatMoney(item.line_total) }}
                </td>
              </tr>
            </tbody>
            <tfoot class="bg-gray-50 dark:bg-gray-700">
              <tr>
                <td colspan="4" class="px-6 py-4 text-right font-medium text-gray-700 dark:text-gray-300">
                  Total Amount:
                </td>
                <td class="px-6 py-4 text-right font-bold text-gray-900 dark:text-white">
                  {{ formatMoney(quotation.total_amount) }}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
      <div v-else class="mb-6 bg-gray-50 dark:bg-gray-700 p-4 rounded-lg border border-gray-200 dark:border-gray-600">
        <p class="text-gray-600 dark:text-gray-400 text-center">No items in this quotation</p>
      </div>

      <!-- Notes and terms -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
        <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg border border-gray-200 dark:border-gray-600">
          <h3 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase mb-3">Notes</h3>
          <p class="text-gray-600 dark:text-gray-400 text-sm">
            This quotation is valid until {{ formatDate(quotation.validity_date) }}. Please contact us if you have any questions.
          </p>
        </div>
        <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg border border-gray-200 dark:border-gray-600">
          <h3 class="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase mb-3">Terms and Conditions</h3>
          <ul class="text-gray-600 dark:text-gray-400 text-sm list-disc pl-5 space-y-1">
            <li>Prices are subject to change without notice.</li>
            <li>Payment terms: 50% advance, balance before delivery.</li>
            <li>Delivery timeline will be confirmed upon order confirmation.</li>
            <li>Standard warranty applies to all products as per manufacturer terms.</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Responsive tweaks */
@media (max-width: 640px) {
  /* Enhanced table responsiveness on small screens */
  table {
    display: block;
    overflow-x: auto;
    white-space: nowrap;
    -webkit-overflow-scrolling: touch;
  }
}
</style> 
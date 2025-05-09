<script setup lang="ts">
import { ref, onMounted, computed, defineAsyncComponent, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useQuotationStore } from '../stores/quotationStore';
import { useCustomerStore } from '../stores/customerStore';
import type { Quotation } from '../types/Quotation';

const QuotationModal = defineAsyncComponent(() => import('../components/QuotationModal.vue'));
const QuotationDetail = defineAsyncComponent(() => import('../components/QuotationDetail.vue'));

// Initialize stores
const quotationStore = useQuotationStore();
const customerStore = useCustomerStore();

// State
const isLoading = ref(false);
const showQuotationModal = ref(false);
const showQuotationDetail = ref(false);
const quotationToEdit = ref<Partial<Quotation> | null>(null);
const quotationToView = ref<Quotation | null>(null);
const searchQuery = ref('');
const statusFilter = ref('all');

// Pagination
const currentPage = ref(1);
const itemsPerPage = 10;

// Load data
const loadData = async () => {
  isLoading.value = true;
  try {
    await Promise.all([
      quotationStore.fetchQuotations(),
      customerStore.fetchCustomers()
    ]);
  } catch (error) {
    console.error('Error loading quotation data:', error);
  } finally {
    isLoading.value = false;
  }
};

// Combined data for display
const quotationsWithCustomers = computed(() => {
  return quotationStore.quotations.map(quote => {
    const customer = customerStore.customers.find(c => c.customer_id === quote.customer_id);
    return {
      ...quote,
      customer_name: customer?.company_name || 'Unknown Customer'
    };
  });
});

// Filtered quotations
const filteredQuotations = computed(() => {
  let result = quotationsWithCustomers.value;
  
  // Apply status filter
  if (statusFilter.value !== 'all') {
    result = result.filter(quote => quote.status === statusFilter.value);
  }
  
  // Apply search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter(quote => 
      quote.customer_name.toLowerCase().includes(query) || 
      quote.quotation_id.toString().includes(query)
    );
  }
  
  return result;
});

// Paginated data
const paginatedQuotations = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  return filteredQuotations.value.slice(startIndex, endIndex);
});

// Total pages for pagination
const totalPages = computed(() => {
  return Math.ceil(filteredQuotations.value.length / itemsPerPage);
});

// Reset pagination when search or filter changes
watch([searchQuery, statusFilter], () => {
  currentPage.value = 1;
});

// Format date
const formatDate = (dateString: string): string => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleDateString();
};

// Format currency
const formatMoney = (amount: number): string => {
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
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

// Page navigation
const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
  }
};

// Open create/edit modal
const openQuotationModal = (quotation: Quotation | null = null) => {
  quotationToEdit.value = quotation ? { ...quotation } : {
    customer_id: 0,
    quote_date: new Date().toISOString().slice(0, 10),
    validity_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
    status: 'Pending',
    total_amount: 0
  };
  showQuotationModal.value = true;
};

// View quotation details
const viewQuotation = async (quotation: Quotation) => {
  try {
    // Set loading state
    isLoading.value = true;
    
    // Make sure we have the customer ID from the quotation
    const customerId = quotation.customer_id;
    if (!customerId) {
      console.error('Customer ID is missing from quotation:', quotation);
      isLoading.value = false;
      return;
    }
    
    console.log('Fetching quotation details for ID:', quotation.quotation_id);
    console.log('Customer ID:', customerId);
    
    // Fetch both quotation and customer data simultaneously
    await Promise.all([
      quotationStore.fetchQuotationById(quotation.quotation_id),
      customerStore.fetchCustomers()
    ]);
    
    // Verify we have both quotation and customer data
    if (quotationStore.currentQuotation) {
      console.log('Loaded quotation:', quotationStore.currentQuotation);
      
      // Ensure we have the specific customer for this quotation
      const customer = customerStore.customers.find(c => c.customer_id === customerId);
      if (customer) {
        console.log('Found matching customer:', customer);
        
        // Store the current quotation for the modal
        quotationToView.value = quotationStore.currentQuotation;
        
        // Short delay to ensure data is properly set in the store
        setTimeout(() => {
          showQuotationDetail.value = true;
          isLoading.value = false;
        }, 200);
      } else {
        console.error('Customer not found with ID:', customerId);
        console.log('Available customers:', customerStore.customers);
        isLoading.value = false;
      }
    } else {
      console.error('Failed to load quotation details');
      isLoading.value = false;
    }
  } catch (error) {
    console.error('Error loading quotation details:', error);
    isLoading.value = false;
  }
};

// Download PDF
const downloadQuotationPDF = async (quotationId: number) => {
  try {
    await quotationStore.downloadPDF(quotationId);
  } catch (error) {
    console.error('Error downloading PDF:', error);
  }
};

// Save quotation
const saveQuotation = async (quotation: any) => {
  try {
    console.log('QuotationList received:', JSON.stringify(quotation, null, 2));
    
    if (quotation.quotation_id) {
      console.log('Updating quotation with customer_id:', quotation.customer_id);
      await quotationStore.updateQuotation(quotation);
    } else {
      console.log('Creating quotation with customer_id:', quotation.customer_id);
      await quotationStore.createQuotation(quotation);
    }
    showQuotationModal.value = false;
    await loadData();
  } catch (error) {
    console.error('Error saving quotation:', error);
  }
};

// Load data on component mount
onMounted(loadData);
</script>

<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
    <!-- Header with actions -->
    <div class="flex flex-col md:flex-row md:justify-between md:items-center p-4 md:p-6 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-gray-800 dark:to-gray-900">
      <div class="mb-4 md:mb-0">
        <h2 class="text-xl md:text-2xl font-bold text-gray-800 dark:text-white mb-1">Quotation Management</h2>
        <p class="text-gray-600 dark:text-gray-300 text-xs md:text-sm">Create and manage customer quotations</p>
      </div>
      <div class="flex flex-wrap gap-3 sm:gap-4">
        <button
          @click="openQuotationModal()"
          class="bg-blue-600 text-white px-3 py-2 md:px-4 md:py-2 rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center shadow-sm text-sm md:text-base"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Create Quotation
        </button>
      </div>
    </div>

    <!-- Stats summary cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 p-4 bg-gray-50 dark:bg-gray-900">
      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Total Quotations</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ quotationStore.quotations.length }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-yellow-100 dark:bg-yellow-900 text-yellow-600 dark:text-yellow-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Pending Quotations</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ quotationStore.quotations.filter(q => q.status === 'Pending').length }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Approved Quotations</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ quotationStore.quotations.filter(q => q.status === 'Approved').length }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Filters section -->
    <div class="p-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800">
      <div class="flex flex-col md:flex-row gap-3 md:items-center">
        <div class="relative flex-grow">
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg class="h-4 w-4 md:h-5 md:w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
            </svg>
          </div>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by customer name or ID..."
            class="w-full pl-8 md:pl-10 px-3 md:px-4 py-2 md:py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg
                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                   transition-all duration-200
                   dark:bg-gray-700 dark:text-white text-xs md:text-sm"
          />
        </div>
        
        <div class="w-full md:w-auto">
          <select
            v-model="statusFilter"
            class="w-full md:w-auto pl-3 md:pl-4 pr-8 md:pr-10 py-2 md:py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg
                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                   bg-white dark:bg-gray-700 text-gray-700 dark:text-white
                   appearance-none text-xs md:text-sm"
          >
            <option value="all">All Statuses</option>
            <option value="Pending">Pending</option>
            <option value="Approved">Approved</option>
            <option value="Rejected">Rejected</option>
            <option value="Expired">Expired</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Loading indicator -->
    <div v-if="isLoading" class="flex justify-center items-center p-12">
      <div class="relative">
        <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-200 dark:border-gray-600"></div>
        <div class="animate-spin rounded-full h-12 w-12 border-t-4 border-blue-600 dark:border-blue-500 absolute top-0 left-0"></div>
      </div>
      <div class="ml-4 text-gray-600 dark:text-gray-300 text-sm font-medium">Loading quotations...</div>
    </div>

    <!-- Quotations table -->
    <div v-else-if="filteredQuotations.length === 0 && searchQuery" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <p class="text-gray-600 dark:text-gray-300 mb-2">No results found</p>
      <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Try adjusting your search criteria</p>
    </div>
    
    <div v-else-if="filteredQuotations.length === 0" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <p class="text-gray-600 dark:text-gray-300 mb-2">No quotations found for the selected status</p>
      <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Try changing the status filter or create a new quotation</p>
      <button
        @click="openQuotationModal()"
        class="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
        </svg>
        Create New Quotation
      </button>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 shadow-sm">
        <thead class="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th class="group px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              <div class="flex items-center">
                ID
              </div>
            </th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Customer
            </th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Date
            </th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Valid Until
            </th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Amount
            </th>
            <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
              Status
            </th>
            <th class="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider sticky right-0 bg-gray-50 dark:bg-gray-700 z-10">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
          <tr v-for="quotation in paginatedQuotations" :key="quotation.quotation_id"
              class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="font-medium text-gray-900 dark:text-white">{{ quotation.quotation_id }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="font-medium text-gray-900 dark:text-white truncate max-w-[200px]">{{ quotation.customer_name }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-300">
              {{ formatDate(quotation.quote_date) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-300">
              {{ formatDate(quotation.validity_date) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-gray-800 dark:text-gray-200 font-medium">
              {{ formatMoney(quotation.total_amount) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="['px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full', getStatusBadgeClass(quotation.status)]">
                {{ quotation.status }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium sticky right-0 bg-white dark:bg-gray-800 shadow-sticky">
              <button
                @click="viewQuotation(quotation)"
                class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-3 transition-colors duration-200"
                title="View quotation details"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                  <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
                </svg>
              </button>
              <button
                @click="downloadQuotationPDF(quotation.quotation_id)"
                class="text-green-600 hover:text-green-900 dark:text-green-400 dark:hover:text-green-300 mr-3 transition-colors duration-200"
                title="Download PDF"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M6 2a2 2 0 00-2 2v12a2 2 0 002 2h8a2 2 0 002-2V7.414A2 2 0 0015.414 6L12 2.586A2 2 0 0010.586 2H6zm5 6a1 1 0 10-2 0v3.586l-1.293-1.293a1 1 0 10-1.414 1.414l3 3a1 1 0 001.414 0l3-3a1 1 0 00-1.414-1.414L11 11.586V8z" clip-rule="evenodd" />
                </svg>
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex justify-center mt-6 mb-6">
        <nav class="inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
          <button
            @click="goToPage(currentPage - 1)"
            :disabled="currentPage === 1"
            class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm font-medium text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span class="sr-only">Previous</span>
            <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd" />
            </svg>
          </button>
          <button
            v-for="page in totalPages"
            :key="page"
            @click="goToPage(page)"
            :class="[
              'relative inline-flex items-center px-4 py-2 border text-sm font-medium',
              currentPage === page
                ? 'z-10 bg-blue-50 dark:bg-blue-900 border-blue-500 dark:border-blue-600 text-blue-600 dark:text-blue-300'
                : 'bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600'
            ]"
          >
            {{ page }}
          </button>
          <button
            @click="goToPage(currentPage + 1)"
            :disabled="currentPage === totalPages"
            class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-sm font-medium text-gray-500 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <span class="sr-only">Next</span>
            <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
              <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
            </svg>
          </button>
        </nav>
      </div>
    </div>

    <!-- Quotation Modal Component -->
    <QuotationModal
      v-if="showQuotationModal"
      :show="showQuotationModal"
      @update:show="showQuotationModal = $event"
      :quotation="quotationToEdit"
      :customers="customerStore.customers"
      @save="saveQuotation"
    />

    <!-- Modal wrapper for QuotationDetail -->
    <div v-if="showQuotationDetail" class="fixed inset-0 z-50 overflow-auto bg-black bg-opacity-50 flex items-center justify-center p-4">
      <div class="relative bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] overflow-auto">
        <button 
          @click="showQuotationDetail = false" 
          class="absolute top-4 right-4 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300 z-10"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
        
        <!-- Loading state while waiting for quotation data -->
        <div v-if="isLoading" class="p-12 flex items-center justify-center">
          <div class="relative">
            <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-200 dark:border-gray-600"></div>
            <div class="animate-spin rounded-full h-12 w-12 border-t-4 border-blue-600 dark:border-blue-500 absolute top-0 left-0"></div>
          </div>
          <div class="ml-4 text-gray-600 dark:text-gray-300 text-sm font-medium">Loading quotation details...</div>
        </div>
        
        <!-- We're embedding the QuotationDetail component here -->
        <QuotationDetail 
          v-else-if="quotationStore.currentQuotation" 
          :isModal="true" 
          :key="quotationStore.currentQuotation.quotation_id" 
        />
        
        <!-- Fallback if data loading failed -->
        <div v-else class="p-8 text-center">
          <h3 class="text-lg font-medium text-red-600 dark:text-red-400 mb-2">Failed to load quotation data</h3>
          <p class="text-gray-600 dark:text-gray-400 mb-4">There was a problem retrieving the quotation information</p>
          <button 
            @click="showQuotationDetail = false" 
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
          >
            Close
          </button>
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
  
  th, td {
    padding: 0.5rem !important;
  }

  /* Ensure sticky columns stay visible */
  .sticky {
    position: sticky;
    right: 0;
    z-index: 20;
  }
  
  /* Apply shadow for visual separation */
  .shadow-sticky {
    box-shadow: -3px 0 5px rgba(0, 0, 0, 0.1);
  }
}

/* Shadow for sticky columns */
.shadow-sticky {
  box-shadow: -3px 0 5px rgba(0, 0, 0, 0.1);
}
</style> 
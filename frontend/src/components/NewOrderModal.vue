<script setup lang="ts">
import { defineProps, defineEmits, ref, watch, computed, onMounted } from 'vue';
import type { Order, OrderItem } from '../types/Order';
import type { Product } from '../types/Product';
import { useOrderStore } from '../stores/orderStore';
import { useCustomerStore } from '../stores/customerStore';
import { useQuotationStore } from '../stores/quotationStore';
import { useProductStore } from '../stores/productStore';
import type { OrderCreate } from '../stores/orderStore';
import { storeToRefs } from 'pinia';

// Add TypeScript interfaces for the API responses
interface QuotationResponse {
  quotation?: {
    customer_id: number;
    total_amount: number;
    [key: string]: any;
  };
  items?: Array<{
    product_id: number;
    quantity: number;
    unit_price: number;
    discount?: number;
    [key: string]: any;
  }>;
  customer_id?: number;
  total_amount?: number;
  [key: string]: any;
}

const orderStore = useOrderStore();
const customerStore = useCustomerStore();
const quotationStore = useQuotationStore();
const productStore = useProductStore();

// Get reactive references to store states
const { customers, loading: customersLoading } = storeToRefs(customerStore);
const { quotations, loading: quotationsLoading } = storeToRefs(quotationStore);
const { products, loading: productsLoading } = storeToRefs(productStore);
const { orders, loading: ordersLoading } = storeToRefs(orderStore);

const props = defineProps<{
  show: boolean;
}>();

const emit = defineEmits(['update:show', 'save']);

// Initialize with default values
const orderData = ref<Partial<Order>>({
  order_date: new Date().toISOString().split('T')[0],
  status: 'Pending',
  total_amount: 0,
  shipping_address: ''
});

// Empty initial order items
const orderItems = ref<Partial<OrderItem>[]>([]);

// Form validation and UI state
const errors = ref<Record<string, string>>({});
const touched = ref<Record<string, boolean>>({});
const isLoading = ref(false);
const isQuotationBased = ref(false);
const selectedCustomer = ref<any>(null);
const shippingAddressPrefilled = ref(false);

// Load data
async function loadAllData() {
  try {
    isLoading.value = true;
    
    // Load all required data in parallel
    await Promise.all([
      customers.value.length === 0 ? customerStore.fetchCustomers() : Promise.resolve(),
      quotations.value.length === 0 ? quotationStore.fetchQuotations() : Promise.resolve(),
      products.value.length === 0 ? productStore.fetchProducts() : Promise.resolve(),
      orders.value.length === 0 ? orderStore.fetchOrders() : Promise.resolve()
    ]);
    
    console.log('Data loaded:', {
      customers: customers.value.length,
      products: products.value.length,
      quotations: quotations.value.length,
      orders: orders.value.length
    });
  } catch (error) {
    console.error('Error loading data:', error);
  } finally {
    isLoading.value = false;
  }
}

// Initial data loading
onMounted(loadAllData);

// Reset form when modal is opened/closed
watch(() => props.show, (newVal) => {
  if (newVal) {
    // Initialize with defaults when opening
    orderData.value = {
      order_date: new Date().toISOString().split('T')[0],
      status: 'Pending',
      total_amount: 0,
      shipping_address: ''
    };
    
    // Reset items
    orderItems.value = [];
    isQuotationBased.value = false;
    selectedCustomer.value = null;
    shippingAddressPrefilled.value = false;
    
    // Load data if needed
    loadAllData();
  }
  // Reset validation state
  errors.value = {};
  touched.value = {};
}, { immediate: true });

// Get product name by ID
const getProductName = (productId: number | undefined): string => {
  if (!productId) return 'Select a product';
  const product = products.value.find(p => p.product_id === productId);
  return product ? product.product_name : `Product ${productId}`;
};

// Updated formatMoney to handle undefined values
const formatMoney = (amount: number | undefined): string => {
  if (amount === undefined) return '₱0.00';
  return '₱' + amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
};

// Updated line_total calculation to handle undefined values
const calculateLineTotal = (quantity: number | undefined, unitPrice: number | undefined, discount: number | undefined): number => {
  const qty = quantity || 0;
  const price = unitPrice || 0;
  const disc = discount || 0;
  return (qty * price) - disc;
};

// Filter to only show approved quotations that don't have orders yet
const availableQuotations = computed(() => {
  // Filter quotations that are approved and don't have an associated order
  return quotations.value.filter(q => {
    // Only include approved quotations
    if (q.status !== 'Approved') return false;
    
    // Check if this quotation already has an order created from it
    const hasOrder = orders.value.some(order => order.quotation_id === q.quotation_id);
    
    // Only include quotations that don't have orders
    return !hasOrder;
  });
});

// Get the list of quotations that already have orders
const quotationsWithOrders = computed(() => {
  return quotations.value.filter(q => {
    // Only include approved quotations
    if (q.status !== 'Approved') return false;
    
    // Check if this quotation already has an order created from it
    return orders.value.some(order => order.quotation_id === q.quotation_id);
  });
});

// Get customer name from ID
const getCustomerName = (customerId: number | undefined): string => {
  if (!customerId) return '';
  const customer = customers.value.find(c => c.customer_id === customerId);
  return customer ? customer.company_name : '';
};

// Check if a product is already added to prevent duplicates within this order only
const isProductAlreadyAdded = (productId: number | undefined): boolean => {
  if (!productId) return false;
  return orderItems.value.some(item => item.product_id === productId);
};

// Get product price by ID
const getProductPrice = (productId: number | undefined): number => {
  if (!productId) return 0;
  const product = products.value.find(p => p.product_id === productId);
  return product ? product.price : 0;
};

// Update the handleProductSelection function to remove the addToAddedProducts call
const handleProductSelection = (item: Partial<OrderItem>, index: number) => {
  if (item.product_id) {
    // Set price from inventory using getProductPrice function
    item.unit_price = getProductPrice(item.product_id);
    
    // Calculate line total
    item.line_total = calculateLineTotal(item.quantity, item.unit_price, item.discount);
  }
};

// Watch for customer selection to pre-fill shipping address
watch(() => orderData.value.customer_id, (customerId) => {
  if (!customerId) {
    selectedCustomer.value = null;
    return;
  }
  
  const customer = customers.value.find(c => c.customer_id === customerId);
  selectedCustomer.value = customer;
  
  // Only pre-fill if shipping address is empty and not already prefilled
  if (!orderData.value.shipping_address && customer && customer.address && !shippingAddressPrefilled.value) {
    orderData.value.shipping_address = customer.address;
    shippingAddressPrefilled.value = true;
  }
}, { immediate: true });

// Update the watch for quotation selection to check if it's already used in an order
watch(() => orderData.value.quotation_id, async (quotationId) => {
  if (!quotationId) {
    isQuotationBased.value = false;
    shippingAddressPrefilled.value = false;
    return;
  }
  
  try {
    isLoading.value = true;
    console.log('Loading quotation details for ID:', quotationId);
    
    // Clear validation errors immediately when selecting a quotation
    errors.value = {};
    
    // Check if this quotation already has an associated order
    const quotationHasOrder = orders.value.some(order => order.quotation_id === quotationId);
    
    if (quotationHasOrder) {
      errors.value.quotation_id = 'This quotation has already been used to create an order';
      isLoading.value = false;
      return;
    }
    
    // Get quotation details to pre-fill the order
    const response = await quotationStore.fetchQuotationById(quotationId);
    console.log('Raw quotation response:', response);
    
    if (response) {
      // Extract the quotation data - handle both possible structures using type assertion
      // Some APIs return {items, quotation} and others return the data directly
      const quotationResponse = response as QuotationResponse;
      const quotationObj = quotationResponse.quotation || quotationResponse;
      const quotationItems = quotationResponse.items || [];
      
      console.log('Extracted quotation object:', quotationObj);
      console.log('Extracted items:', quotationItems);
      
      isQuotationBased.value = true;
      
      // Reset order items
      orderItems.value = [];
      
      // Check if we have the customer ID and set it
      if (quotationObj && quotationObj.customer_id) {
        const customerId = quotationObj.customer_id;
        console.log('Setting customer ID from quotation:', customerId);
        
        // Directly set the customer_id
        orderData.value.customer_id = customerId;
        
        // Try to find the customer in the already loaded customers
        let foundCustomer = customers.value.find(c => c.customer_id === customerId);
        
        if (!foundCustomer) {
          console.log('Customer not found in loaded customers, fetching...');
          // Refresh the customer list
          await customerStore.fetchCustomers();
          foundCustomer = customers.value.find(c => c.customer_id === customerId);
          
          // If still not found, fetch individually
          if (!foundCustomer) {
            try {
              foundCustomer = await customerStore.fetchCustomerById(customerId);
              console.log('Fetched customer data:', foundCustomer);
              
              // Add to customers list if needed
              if (foundCustomer && !customers.value.some(c => c.customer_id === foundCustomer!.customer_id)) {
                customers.value.push(foundCustomer);
              }
            } catch (err) {
              console.error('Failed to fetch customer:', err);
            }
          }
        }
        
        // Set the selected customer and shipping address
        if (foundCustomer) {
          console.log('Setting selected customer:', foundCustomer);
          selectedCustomer.value = foundCustomer;
          
          if (foundCustomer.address) {
            console.log('Setting shipping address from customer:', foundCustomer.address);
            orderData.value.shipping_address = foundCustomer.address;
            shippingAddressPrefilled.value = true;
          }
        }
      }
      
      // Set total amount from quotation
      if (quotationObj && quotationObj.total_amount) {
        orderData.value.total_amount = quotationObj.total_amount;
      }
      
      // Add items from quotation
      if (quotationItems && quotationItems.length > 0) {
        console.log('Adding items from quotation:', quotationItems);
        
        // Load any missing products
        const productIds = quotationItems.map(item => item.product_id);
        
        // Ensure all products are loaded
        await Promise.all(
          productIds.map(id => {
            if (!products.value.find(p => p.product_id === id)) {
              return productStore.fetchProductById(id);
            }
            return Promise.resolve();
          })
        );
        
        // Add each quotation item to our order items
        quotationItems.forEach(item => {
          orderItems.value.push({
            product_id: item.product_id,
            quantity: item.quantity,
            unit_price: item.unit_price,
            discount: item.discount || 0,
            line_total: (item.quantity * item.unit_price) - (item.discount || 0)
          });
        });
      }
      
      // Force validation with a longer delay
      setTimeout(() => {
        console.log('Running validation after delay with data:', {
          customer_id: orderData.value.customer_id,
          customer_object: selectedCustomer.value,
          shipping_address: orderData.value.shipping_address,
          order_items: orderItems.value.length
        });
        validateFormOnQuotationLoad();
      }, 1000);
    }
  } catch (error) {
    console.error('Error loading quotation details:', error);
  } finally {
    isLoading.value = false;
  }
});

// Validation rules
const validateField = (field: string, value: any) => {
  touched.value[field] = true;

  // Required field validation
  if (value === undefined || value === null || value === '') {
    errors.value[field] = `${field.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')} is required`;
    return false;
  }

  // Customer validation
  if (field === 'customer_id' && (typeof value !== 'number' || value <= 0)) {
    errors.value[field] = 'Please select a customer';
    return false;
  }

  // Date validation
  if (field === 'order_date' && !value) {
    errors.value[field] = 'Please select a valid date';
    return false;
  }

  // Address validation
  if (field === 'shipping_address' && value.trim().length < 5) {
    errors.value[field] = 'Shipping address must be at least 5 characters';
    return false;
  }

  // Total amount validation
  if (field === 'total_amount' && (typeof value !== 'number' || value < 0)) {
    errors.value[field] = 'Total amount must be a positive number';
    return false;
  }

  delete errors.value[field];
  return true;
};

// Update the validateAllFields function to add better debugging
const validateAllFields = () => {
  const requiredFields = ['customer_id', 'order_date', 'shipping_address', 'status', 'total_amount'] as const;
  
  console.log('Validating fields with values:', {
    customer_id: orderData.value.customer_id,
    order_date: orderData.value.order_date,
    shipping_address: orderData.value.shipping_address,
    status: orderData.value.status,
    total_amount: orderData.value.total_amount,
    selectedCustomer: selectedCustomer.value
  });
  
  return requiredFields.map(field => {
    // Only validate if the field has a value
    if (orderData.value[field] !== undefined && 
        orderData.value[field] !== null && 
        orderData.value[field] !== '') {
      return validateField(field, orderData.value[field]);
    }
    // Mark as touched to show validation errors
    touched.value[field] = true;
    errors.value[field] = `${field.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')} is required`;
    return false;
  });
};

// Close modal
const closeModal = () => {
  emit('update:show', false);
};

// Save order with validation
const saveOrder = async () => {
  if (isLoading.value) return;

  isLoading.value = true;

  try {
    // Final check if we're using a quotation that already has an order
    // This is important for race conditions where another user might 
    // have created an order from this quotation while this form was open
    if (orderData.value.quotation_id) {
      // Check if this quotation already has an associated order
      const quotationHasOrder = orders.value.some(
        order => order.quotation_id === orderData.value.quotation_id
      );
      
      if (quotationHasOrder) {
        errors.value.quotation_id = 'This quotation has already been used to create an order';
        isLoading.value = false;
        return;
      }
    }

    // Validate all required fields
    const validations = validateAllFields();

    if (validations.includes(false)) {
      isLoading.value = false;
      return;
    }

    // Format date in RFC 3339 format
    const formatDateForBackend = (dateStr: string) => {
      // Convert YYYY-MM-DD to YYYY-MM-DDTHH:mm:ssZ
      const date = new Date(dateStr);
      return date.toISOString(); // Returns in format: YYYY-MM-DDTHH:mm:ss.sssZ
    };

    // Create the order data in the format expected by the server
    let newOrderData: OrderCreate;

    if (isQuotationBased.value && orderData.value.quotation_id) {
      // Format for quotation-based orders - match the expected API structure
      newOrderData = {
        order: {
          customer_id: orderData.value.customer_id!,
          order_date: formatDateForBackend(orderData.value.order_date!),
          shipping_address: orderData.value.shipping_address!,
          status: orderData.value.status as 'Pending' | 'Shipped' | 'Delivered' | 'Cancelled',
          total_amount: orderData.value.total_amount!
        },
        items: orderItems.value.map(item => ({
          product_id: item.product_id!,
          quantity: item.quantity!,
          unit_price: item.unit_price!,
          discount: item.discount!,
          line_total: item.line_total!
        })),
        quotation: {
          quotation_id: orderData.value.quotation_id
        }
      };
    } else {
      // Format for standard orders
      newOrderData = {
        order: {
          customer_id: orderData.value.customer_id!,
          order_date: formatDateForBackend(orderData.value.order_date!),
          shipping_address: orderData.value.shipping_address!,
          status: orderData.value.status as 'Pending' | 'Shipped' | 'Delivered' | 'Cancelled',
          total_amount: orderData.value.total_amount!,
          quotation_id: orderData.value.quotation_id
        },
        items: orderItems.value.map(item => ({
          product_id: item.product_id!,
          quantity: item.quantity!,
          unit_price: item.unit_price!,
          discount: item.discount!,
          line_total: item.line_total!
        }))
      };
    }

    console.log('Sending order data to server:', JSON.stringify(newOrderData, null, 2));

    // Save the order using the store
    await orderStore.createOrder(newOrderData);
    
    // Emit save event for parent component
    emit('save', newOrderData);
    closeModal();
  } catch (error) {
    console.error('Error saving order:', error);
    errors.value.submit = 'Failed to save: ' + (error instanceof Error ? error.message : 'Unknown error');
  } finally {
    isLoading.value = false;
  }
};

// Update total amount based on items
const updateTotalAmount = () => {
  const total = orderItems.value.reduce((sum, item) => {
    return sum + (item.line_total || 0);
  }, 0);
  
  orderData.value.total_amount = total;
};

// Call this when items change
watch(orderItems, () => {
  updateTotalAmount();
}, { deep: true });

// Add a validateFormOnQuotationLoad function that properly handles validation with a quotation
const validateFormOnQuotationLoad = () => {
  // Force touched state for all required fields
  const requiredFields = ['customer_id', 'order_date', 'shipping_address', 'status', 'total_amount'] as const;
  requiredFields.forEach(field => {
    touched.value[field] = true;
  });
  
  // Clear any existing errors
  errors.value = {};
  
  // Validate all fields
  validateAllFields();
};
</script>

<template>
  <div v-if="show"
      class="fixed inset-0 z-50 overflow-y-auto">
    <div class="min-h-screen px-4 text-center">
      <div class="fixed inset-0 transition-opacity" @click="closeModal">
        <div class="absolute inset-0 bg-black opacity-50 dark:opacity-70"></div>
      </div>

      <span class="inline-block h-screen align-middle" aria-hidden="true">&#8203;</span>

      <div class="inline-block w-full max-w-3xl p-6 my-8 overflow-hidden text-left align-middle transition-all transform bg-white dark:bg-gray-800 shadow-xl rounded-lg border border-gray-200 dark:border-gray-700 modal-content">
        <!-- Modal Header -->
        <div class="flex justify-between items-center border-b border-gray-200 dark:border-gray-700 pb-4 mb-6">
          <h3 class="text-xl font-semibold text-gray-900 dark:text-white">
            Add New Order
          </h3>
          <button @click="closeModal" class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 focus:outline-none">
            <span class="sr-only">Close</span>
            <svg class="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Error Summary -->
        <div v-if="Object.keys(errors).length > 0" class="mb-6 p-4 bg-red-50 dark:bg-red-900 dark:bg-opacity-20 border-l-4 border-red-500 rounded">
          <div class="flex">
            <div class="flex-shrink-0">
              <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
              </svg>
            </div>
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800 dark:text-red-200">Please fix the following errors:</h3>
              <div class="mt-2 text-sm text-red-700 dark:text-red-300">
                <ul class="list-disc pl-5 space-y-1">
                  <li v-for="(value, key) in errors" :key="key">{{ value }}</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="isLoading" class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-4 border-gray-200 dark:border-gray-600 border-t-blue-600 dark:border-t-blue-500"></div>
          <p class="mt-2 text-gray-600 dark:text-gray-400">Loading data...</p>
        </div>

        <!-- Form -->
        <form v-else @submit.prevent="saveOrder" class="space-y-6">
          <!-- Quotation Selection (Optional) -->
          <div>
            <label for="quotation_id" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Based on Quotation <span class="text-gray-400">(Optional)</span>
            </label>
            <select
              id="quotation_id"
              v-model="orderData.quotation_id"
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
              :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.quotation_id}"
              :disabled="isLoading"
            >
              <option :value="undefined">None - Create from scratch</option>
              <optgroup label="Available Quotations">
                <option v-for="quotation in availableQuotations" :key="quotation.quotation_id" :value="quotation.quotation_id">
                  QT{{ quotation.quotation_id.toString().padStart(3, '0') }} - {{ formatMoney(quotation.total_amount) }}
                </option>
              </optgroup>
              <optgroup v-if="quotationsWithOrders.length > 0" label="Already Used in Orders (Unavailable)">
                <option v-for="quotation in quotationsWithOrders" :key="quotation.quotation_id" :value="quotation.quotation_id" disabled class="text-gray-400">
                  QT{{ quotation.quotation_id.toString().padStart(3, '0') }} - {{ formatMoney(quotation.total_amount) }}
                </option>
              </optgroup>
            </select>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Selecting a quotation will pre-fill the order details
            </p>
            <p v-if="errors.quotation_id" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.quotation_id }}
            </p>
          </div>

          <div>
            <label for="customer_id" class="block text-sm font-medium text-gray-700 dark:text-gray-300 flex justify-between">
              <span>Customer <span class="text-red-500">*</span></span>
              <span v-if="isQuotationBased" class="text-xs italic text-blue-600 dark:text-blue-400">From Quotation</span>
            </label>
            <select
              id="customer_id"
              v-model="orderData.customer_id"
              @blur="validateField('customer_id', orderData.customer_id)"
              required
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
              :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.customer_id && touched.customer_id, 'border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20': isQuotationBased}"
              :disabled="isLoading || isQuotationBased"
            >
              <option :value="undefined" disabled>Select a customer</option>
              <option v-for="customer in customers" :key="customer.customer_id" :value="customer.customer_id">
                {{ customer.company_name }}
              </option>
            </select>
            <p v-if="selectedCustomer" class="mt-1 text-xs text-blue-600 dark:text-blue-400">
              {{ getCustomerName(selectedCustomer.customer_id) }} - {{ selectedCustomer.phone || 'No phone' }}
            </p>
            <p v-if="errors.customer_id && touched.customer_id" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.customer_id }}
            </p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label for="order_date" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Order Date <span class="text-red-500">*</span>
              </label>
              <input
                id="order_date"
                v-model="orderData.order_date"
                @blur="validateField('order_date', orderData.order_date)"
                type="date"
                required
                class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
                :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.order_date && touched.order_date}"
                :disabled="isLoading"
              />
              <p v-if="errors.order_date && touched.order_date" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.order_date }}
              </p>
            </div>

            <div>
              <label for="status" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Status <span class="text-red-500">*</span>
              </label>
              <select
                id="status"
                v-model="orderData.status"
                @blur="validateField('status', orderData.status)"
                required
                class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
                :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.status && touched.status}"
                :disabled="isLoading"
              >
                <option value="Pending">Pending</option>
                <option value="Shipped">Shipped</option>
                <option value="Delivered">Delivered</option>
                <option value="Cancelled">Cancelled</option>
              </select>
              <p v-if="errors.status && touched.status" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.status }}
              </p>
            </div>
            </div>

            <div>
            <label for="shipping_address" class="block text-sm font-medium text-gray-700 dark:text-gray-300 flex justify-between">
              <span>Shipping Address <span class="text-red-500">*</span></span>
              <span v-if="isQuotationBased || shippingAddressPrefilled" class="text-xs italic text-blue-600 dark:text-blue-400">Prefilled from Customer/Quotation</span>
            </label>
            <textarea
              id="shipping_address"
              v-model="orderData.shipping_address"
              @blur="validateField('shipping_address', orderData.shipping_address)"
              rows="3"
              required
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
              :class="{
                'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.shipping_address && touched.shipping_address,
                'border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20': isQuotationBased || shippingAddressPrefilled
              }"
              :disabled="isLoading || isQuotationBased || shippingAddressPrefilled"
              placeholder="Enter the complete shipping address"
            ></textarea>
            <p v-if="errors.shipping_address && touched.shipping_address" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.shipping_address }}
            </p>
          </div>

          <!-- Order items section -->
          <div class="border border-gray-200 dark:border-gray-700 rounded-md p-4">
            <div class="flex justify-between items-center mb-3">
              <h4 class="font-medium text-gray-700 dark:text-gray-300 text-sm uppercase tracking-wider">Order Items</h4>
              <span v-if="isQuotationBased" class="text-xs italic text-blue-600 dark:text-blue-400">From Quotation</span>
            </div>
            
            <!-- Column headers -->
            <div class="grid grid-cols-12 gap-2 mb-2 px-2">
              <div class="col-span-4 text-xs font-medium text-gray-500 dark:text-gray-400">Product</div>
              <div class="col-span-2 text-xs font-medium text-gray-500 dark:text-gray-400">Qty</div>
              <div class="col-span-2 text-xs font-medium text-gray-500 dark:text-gray-400">Price</div>
              <div class="col-span-2 text-xs font-medium text-gray-500 dark:text-gray-400">Discount</div>
              <div class="col-span-2 text-xs font-medium text-gray-500 dark:text-gray-400">Total</div>
            </div>
            
            <!-- Basic item list -->
            <div v-for="(item, index) in orderItems" :key="index" 
                 class="grid grid-cols-12 gap-2 mb-3" 
                 :class="{'bg-blue-50 dark:bg-blue-900 dark:bg-opacity-10 p-2 rounded': isQuotationBased}">
              <div class="col-span-4">
                <div v-if="isQuotationBased" class="w-full px-2 py-1 text-sm bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20 border border-blue-300 dark:border-blue-700 rounded text-gray-900 dark:text-gray-100">
                  {{ getProductName(item.product_id) }}
                </div>
                <select
                  v-else
                  v-model="item.product_id"
                  @change="handleProductSelection(item, index)"
                  class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                  :disabled="isQuotationBased"
                >
                  <option :value="undefined" disabled>Select a product</option>
                  <option 
                    v-for="product in products" 
                    :key="product.product_id" 
                    :value="product.product_id"
                    :disabled="isProductAlreadyAdded(product.product_id) && item.product_id !== product.product_id"
                  >
                    {{ getProductName(product.product_id) }} - {{ formatMoney(getProductPrice(product.product_id)) }}
                  </option>
                </select>
              </div>
              <div class="col-span-2">
                <input
                  v-model.number="item.quantity"
                  type="number"
                  min="1"
                  class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                  @input="item.line_total = calculateLineTotal(item.quantity, item.unit_price, item.discount)"
                  :disabled="isQuotationBased || !item.product_id"
                  :class="{'opacity-90 bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20 border-blue-300 dark:border-blue-700': isQuotationBased}"
                />
              </div>
              <div class="col-span-2">
                <div v-if="isQuotationBased || item.product_id" class="w-full px-2 py-1 text-sm bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20 border border-blue-300 dark:border-blue-700 rounded text-right text-gray-900 dark:text-gray-100">
                  {{ formatMoney(item.unit_price) }}
                </div>
                <div v-else class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-gray-100 dark:bg-gray-700 text-right text-gray-400 dark:text-gray-500">
                  Select a product
                </div>
              </div>
              <div class="col-span-2">
                <div v-if="isQuotationBased" class="w-full px-2 py-1 text-sm bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20 border border-blue-300 dark:border-blue-700 rounded text-right text-gray-900 dark:text-gray-100">
                  {{ formatMoney(item.discount) }}
                </div>
                <input
                  v-else
                  v-model.number="item.discount"
                  type="number"
                  min="0"
                  step="0.01"
                  class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded dark:bg-gray-700 text-gray-900 dark:text-gray-100"
                  @input="item.line_total = calculateLineTotal(item.quantity, item.unit_price, item.discount)"
                  :disabled="!item.product_id"
                />
              </div>
              <div class="col-span-2">
                <div class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-gray-50 dark:bg-gray-800 text-right text-gray-900 dark:text-gray-100"
                   :class="{'opacity-90 bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20 border-blue-300 dark:border-blue-700': isQuotationBased}">
                  {{ formatMoney(item.line_total) }}
                </div>
              </div>
            </div>
            
            <!-- Add more items button - hidden when quotation-based -->
            <button
              v-if="!isQuotationBased"
              type="button"
              @click="orderItems.push({
                product_id: undefined,
                quantity: 1,
                unit_price: 0,
                discount: 0,
                line_total: 0
              })"
              class="mt-2 inline-flex items-center px-3 py-1 border border-gray-300 dark:border-gray-600 text-sm leading-4 font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <svg class="-ml-0.5 mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 3a1 1 0 00-1 1v5H4a1 1 0 100 2h5v5a1 1 0 102 0v-5h5a1 1 0 100-2h-5V4a1 1 0 00-1-1z" clip-rule="evenodd" />
              </svg>
              Add Item
            </button>
          </div>

          <!-- Total Amount with formatted display -->
          <div>
            <label for="total_amount" class="block text-sm font-medium text-gray-700 dark:text-gray-300 flex justify-between">
              <span>Total Amount <span class="text-red-500">*</span></span>
              <span v-if="isQuotationBased" class="text-xs italic text-blue-600 dark:text-blue-400">From Quotation</span>
              </label>
              <div class="mt-1 relative">
              <div v-if="isQuotationBased" 
                   class="block w-full px-3 py-2 bg-blue-50 dark:bg-blue-900 dark:bg-opacity-20 border border-blue-300 dark:border-blue-700 rounded-md shadow-sm text-gray-900 dark:text-gray-100 sm:text-sm">
                {{ formatMoney(orderData.total_amount) }}
              </div>
              <template v-else>
                <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <span class="text-gray-500 sm:text-sm">₱</span>
                </div>
                <input
                  id="total_amount"
                  v-model.number="orderData.total_amount"
                  @blur="validateField('total_amount', orderData.total_amount)"
                  type="number"
                  min="0"
                  step="0.01"
                  required
                  class="mt-1 block w-full pl-10 pr-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
                  :class="{
                    'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.total_amount && touched.total_amount
                  }"
                  readonly
                />
              </template>
              </div>
              <p v-if="errors.total_amount && touched.total_amount" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.total_amount }}
            </p>
          </div>

          <!-- Footer with buttons -->
          <div class="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700 flex justify-end space-x-3">
            <button
              type="button"
              @click="closeModal"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 transition-colors"
              :disabled="isLoading"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="isLoading || Object.keys(errors).length > 0 || orderItems.length === 0"
              class="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-offset-gray-800 transition-colors"
              :class="{'opacity-75 cursor-not-allowed': isLoading || Object.keys(errors).length > 0 || orderItems.length === 0}"
            >
              <span v-if="isLoading" class="flex items-center">
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Processing...
              </span>
              <span v-else>Create Order</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease-out;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.modal-content {
  animation: modal-pop 0.3s ease-out forwards;
}

@keyframes modal-pop {
  0% {
    transform: scale(0.95);
    opacity: 0;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

/* Hide number input spinners */
input[type=number]::-webkit-inner-spin-button,
input[type=number]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
input[type=number] {
  -moz-appearance: textfield;
  appearance: textfield;
}
</style>

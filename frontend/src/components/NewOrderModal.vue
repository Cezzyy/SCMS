<script setup lang="ts">
import { defineProps, defineEmits, ref, watch, computed } from 'vue';
import type { Order } from '../types/Order';

const props = defineProps<{
  show: boolean;
  customers: { customer_id: number; name: string }[];
}>();

const emit = defineEmits(['update:show', 'save']);

const orderData = ref<Partial<Order>>({
  order_date: new Date().toISOString().split('T')[0],
  status: 'Pending',
  total_amount: 0
});

// Form validation and UI state
const errors = ref<Record<string, string>>({});
const touched = ref<Record<string, boolean>>({});
const isLoading = ref(false);

// Reset form when modal is opened/closed
watch(() => props.show, (newVal) => {
  if (newVal) {
    // Initialize with defaults when opening
    orderData.value = {
      order_date: new Date().toISOString().split('T')[0],
      status: 'Pending',
      total_amount: 0
    };
  }
  // Reset validation state
  errors.value = {};
  touched.value = {};
}, { immediate: true });

// Format money with commas
const formatMoney = (amount: number): string => {
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
};

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

// Close modal
const closeModal = () => {
  emit('update:show', false);
};

// Save order with validation
const saveOrder = () => {
  if (isLoading.value) return;

  isLoading.value = true;

  try {
    // Validate all required fields
    const requiredFields = ['customer_id', 'order_date', 'shipping_address', 'status', 'total_amount'] as const;
    const validations = requiredFields.map(field =>
      validateField(field, orderData.value[field])
    );

    if (validations.includes(false)) {
      isLoading.value = false;
      return;
    }

    const finalOrderData: any = {
      ...orderData.value,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    };

    emit('save', finalOrderData);
  } catch (error) {
    console.error('Error saving order:', error);
    errors.value.submit = 'Failed to save. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Get customer name by ID
const getCustomerName = (customerId: number): string => {
  const customer = props.customers.find(c => c.customer_id === customerId);
  return customer ? customer.name : 'Unknown Customer';
};
</script>

<template>
  <div v-if="show"
      class="fixed inset-0 z-50 overflow-y-auto">
    <div class="min-h-screen px-4 text-center">
      <div class="fixed inset-0 transition-opacity" @click="closeModal">
        <div class="absolute inset-0 bg-black opacity-50 dark:opacity-60"></div>
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

        <!-- Form -->
        <form @submit.prevent="saveOrder" class="space-y-6">
          <div>
            <label for="customer_id" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Customer <span class="text-red-500">*</span>
            </label>
            <select
              id="customer_id"
              v-model="orderData.customer_id"
              @blur="validateField('customer_id', orderData.customer_id)"
              required
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
              :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.customer_id && touched.customer_id}"
              :disabled="isLoading"
            >
              <option :value="undefined" disabled>Select a customer</option>
              <option v-for="customer in customers" :key="customer.customer_id" :value="customer.customer_id">
                {{ customer.name }}
              </option>
            </select>
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
              <label for="quotation_id" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Quotation ID <span class="text-gray-400">(Optional)</span>
              </label>
              <input
                id="quotation_id"
                v-model="orderData.quotation_id"
                type="number"
                min="1"
                placeholder="Related quotation ID"
                class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
                :disabled="isLoading"
              />
            </div>
          </div>

          <div>
            <label for="shipping_address" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Shipping Address <span class="text-red-500">*</span>
            </label>
            <textarea
              id="shipping_address"
              v-model="orderData.shipping_address"
              @blur="validateField('shipping_address', orderData.shipping_address)"
              rows="3"
              required
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 text-gray-900 dark:text-gray-100 sm:text-sm"
              :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.shipping_address && touched.shipping_address}"
              :disabled="isLoading"
              placeholder="Enter the complete shipping address"
            ></textarea>
            <p v-if="errors.shipping_address && touched.shipping_address" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.shipping_address }}
            </p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
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

            <div>
              <label for="total_amount" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Total Amount <span class="text-red-500">*</span>
              </label>
              <div class="mt-1 relative">
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
                  :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.total_amount && touched.total_amount}"
                  :disabled="isLoading"
                  placeholder="0.00"
                />
              </div>
              <p v-if="errors.total_amount && touched.total_amount" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.total_amount }}
              </p>
            </div>
          </div>

          <!-- Note: In a real app, we would also have order items section here -->
          <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-md">
            <div class="flex items-center text-gray-600 dark:text-gray-300 mb-2">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span class="text-sm font-medium">Note</span>
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              Order items can be added after creating the order. In a complete implementation, you would be able to add products, quantities and pricing here.
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
              :disabled="isLoading || Object.keys(errors).length > 0"
              class="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 dark:focus:ring-offset-gray-800 transition-colors"
              :class="{'opacity-75 cursor-not-allowed': isLoading}"
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

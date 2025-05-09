<script setup lang="ts">
import { defineProps, defineEmits, ref, watch, computed } from 'vue';
import type { Order } from '../types/Order';

const props = defineProps<{
  show: boolean;
  order: Order | null;
}>();

const emit = defineEmits(['update:show', 'update']);

// Selected status
const selectedStatus = ref<string>('');

// Get available status options based on current status
const statusOptions = computed(() => {
  if (!props.order) return [];
  
  const currentStatus = props.order.status;
  
  // Base options
  const options = [
    { value: 'Pending', label: 'Pending', disabled: false },
    { value: 'Shipped', label: 'Shipped', disabled: false },
    { value: 'Delivered', label: 'Delivered', disabled: false },
    { value: 'Cancelled', label: 'Cancelled', disabled: false }
  ];
  
  // Apply validation rules
  if (currentStatus === 'Cancelled') {
    // Cancelled orders can't be edited
    return options.map(option => ({ ...option, disabled: true }));
  }
  
  if (currentStatus === 'Delivered') {
    // Delivered orders can't be edited
    return options.map(option => ({ ...option, disabled: true }));
  }
  
  if (currentStatus === 'Shipped') {
    // Shipped orders can't go back to pending
    return options.map(option => ({
      ...option,
      disabled: option.value === 'Pending'
    }));
  }
  
  return options;
});

// A message to explain why status can't be changed
const statusMessage = computed(() => {
  if (!props.order) return '';
  
  switch (props.order.status) {
    case 'Cancelled':
      return 'This order has been cancelled and cannot be updated.';
    case 'Delivered':
      return 'This order has been delivered and cannot be updated.';
    case 'Shipped':
      return 'You cannot change a shipped order back to pending status.';
    default:
      return '';
  }
});

// Reset selected status when modal opens
watch(() => props.show, (newVal) => {
  if (newVal && props.order) {
    selectedStatus.value = props.order.status;
  }
}, { immediate: true });

// Watch for order changes and update selected status
watch(() => props.order, (newOrder) => {
  if (newOrder) {
    selectedStatus.value = newOrder.status;
  }
});

// Check if status update is disabled
const isUpdateDisabled = computed(() => {
  if (!props.order) return true;
  
  // If current status is the same as selected status
  if (selectedStatus.value === props.order.status) return true;
  
  // Cancelled or delivered orders can't be updated
  if (props.order.status === 'Cancelled' || props.order.status === 'Delivered') return true;
  
  // Can't go from shipped to pending
  if (props.order.status === 'Shipped' && selectedStatus.value === 'Pending') return true;
  
  return false;
});

// Close modal
const close = () => {
  emit('update:show', false);
};

// Update status
const updateStatus = () => {
  if (isUpdateDisabled.value) {
    close();
    return;
  }
  
  emit('update', {
    orderId: props.order!.order_id,
    status: selectedStatus.value
  });
  
  close();
};

// Get status class for colors
const getStatusClass = (status: string) => {
  switch (status) {
    case 'Pending': return 'bg-yellow-100 text-yellow-800 border-yellow-300 dark:bg-yellow-900 dark:text-yellow-200 dark:border-yellow-700';
    case 'Shipped': return 'bg-indigo-100 text-indigo-800 border-indigo-300 dark:bg-indigo-900 dark:text-indigo-200 dark:border-indigo-700';
    case 'Delivered': return 'bg-green-100 text-green-800 border-green-300 dark:bg-green-900 dark:text-green-200 dark:border-green-700';
    case 'Cancelled': return 'bg-red-100 text-red-800 border-red-300 dark:bg-red-900 dark:text-red-200 dark:border-red-700';
    default: return 'bg-gray-100 text-gray-800 border-gray-300 dark:bg-gray-700 dark:text-gray-200 dark:border-gray-600';
  }
};
</script>

<template>
  <div v-if="show && order"
       class="fixed inset-0 z-50 overflow-y-auto"
       aria-labelledby="modal-title"
       role="dialog"
       aria-modal="true">
    <div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
      <!-- Background overlay -->
      <div class="fixed inset-0 bg-gray-500 bg-opacity-75 dark:bg-gray-900 dark:bg-opacity-75 transition-opacity" 
           aria-hidden="true"
           @click="close"></div>

      <!-- Modal panel -->
      <span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>
      <div class="inline-block align-bottom bg-white dark:bg-gray-800 rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
        <div class="bg-white dark:bg-gray-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
          <div class="sm:flex sm:items-start">
            <div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 dark:bg-blue-900 sm:mx-0 sm:h-10 sm:w-10">
              <!-- Status update icon -->
              <svg class="h-6 w-6 text-blue-600 dark:text-blue-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </div>
            <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full">
              <h3 class="text-lg leading-6 font-medium text-gray-900 dark:text-gray-100" id="modal-title">
                Update Order Status
              </h3>
              <div class="mt-2">
                <p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
                  Update the status for order #ORD{{ order.order_id.toString().padStart(3, '0') }}
                </p>
                
                <div class="mb-4">
                  <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Current Status:</p>
                  <div class="inline-block px-3 py-1 text-sm rounded-full border" :class="getStatusClass(order.status)">
                    {{ order.status }}
                  </div>
                </div>

                <!-- Status validation message -->
                <div v-if="statusMessage" class="mb-4 p-3 border rounded-md" :class="getStatusClass(order.status)">
                  <p class="text-sm">{{ statusMessage }}</p>
                </div>
                
                <div>
                  <label for="status" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    New Status:
                  </label>
                  <select
                    id="status"
                    v-model="selectedStatus"
                    class="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 dark:border-gray-600 focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm rounded-md dark:bg-gray-700 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    <option
                      v-for="option in statusOptions"
                      :key="option.value"
                      :value="option.value"
                      :disabled="option.disabled || option.value === order.status"
                    >
                      {{ option.label }}{{ option.disabled ? ' (Not Available)' : '' }}
                    </option>
                  </select>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="bg-gray-50 dark:bg-gray-700 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
          <button type="button" 
                  class="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm disabled:opacity-50 disabled:cursor-not-allowed disabled:bg-blue-400"
                  @click="updateStatus"
                  :disabled="isUpdateDisabled">
            Update Status
          </button>
          <button type="button" 
                  class="mt-3 w-full inline-flex justify-center rounded-md border border-gray-300 shadow-sm px-4 py-2 bg-white text-base font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm dark:bg-gray-800 dark:text-gray-200 dark:border-gray-600 dark:hover:bg-gray-700"
                  @click="close">
            Cancel
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Add any custom styling here if needed */
</style> 
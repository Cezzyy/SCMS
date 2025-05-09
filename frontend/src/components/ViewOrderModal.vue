<script setup lang="ts">
import { defineProps, defineEmits, ref, computed } from 'vue';
import type { Order } from '../types/Order';

const props = defineProps<{
  show: boolean;
  order: Order | null;
  orderItems?: {
    order_item_id: number;
    product_id: number;
    product_name: string;
    quantity: number;
    unit_price: number;
    discount: number;
    line_total: number;
  }[];
}>();

const emit = defineEmits(['update:show']);

// Close modal
const closeModal = () => {
  emit('update:show', false);
};

// Format money with commas
const formatMoney = (amount: number): string => {
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
};

// Format date from ISO string to readable format
const formatDate = (dateString: string): string => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
};

// Get appropriate status class for styling
const getStatusClass = (status: string) => {
  switch (status) {
    case 'Pending': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:bg-opacity-30 dark:text-yellow-200';
    case 'Shipped': return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:bg-opacity-30 dark:text-indigo-200';
    case 'Delivered': return 'bg-green-100 text-green-800 dark:bg-green-900 dark:bg-opacity-30 dark:text-green-200';
    case 'Cancelled': return 'bg-red-100 text-red-800 dark:bg-red-900 dark:bg-opacity-30 dark:text-red-200';
    default: return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200';
  }
};

// Calculate summary data for items
const orderSummary = computed(() => {
  if (!props.orderItems || props.orderItems.length === 0) {
    return {
      totalItems: 0,
      subtotal: 0,
      totalDiscount: 0,
      grandTotal: props.order?.total_amount || 0
    };
  }

  const totalItems = props.orderItems.reduce((sum, item) => sum + item.quantity, 0);
  const subtotal = props.orderItems.reduce((sum, item) => sum + (item.unit_price * item.quantity), 0);
  const totalDiscount = props.orderItems.reduce((sum, item) => sum + item.discount, 0);

  return {
    totalItems,
    subtotal,
    totalDiscount,
    grandTotal: subtotal - totalDiscount
  };
});

// Get customer name (in a real app, this would fetch from API)
const getCustomerName = (customerId: number): string => {
  // This would normally fetch from a customer service or API
  // For now we'll use a mock mapping
  const customerMap: Record<number, string> = {
    101: 'Alice Wonderland',
    102: 'Bob The Builder',
    103: 'Charlie Brown',
    104: 'Diana Prince',
    105: 'Edward Scissorhands'
  };

  return customerMap[customerId] || `Customer ${customerId}`;
};
</script>

<template>
  <div v-if="show && order"
      class="fixed inset-0 z-50 overflow-y-auto">
    <div class="min-h-screen px-4 text-center">
      <div class="fixed inset-0 transition-opacity" @click="closeModal">
        <div class="absolute inset-0 bg-black opacity-50 dark:opacity-60"></div>
      </div>

      <span class="inline-block h-screen align-middle" aria-hidden="true">&#8203;</span>

      <div class="inline-block w-full max-w-4xl p-6 my-8 overflow-hidden text-left align-middle transition-all transform bg-white dark:bg-gray-800 shadow-xl rounded-lg border border-gray-200 dark:border-gray-700 modal-content">
        <!-- Modal Header -->
        <div class="flex justify-between items-center border-b border-gray-200 dark:border-gray-700 pb-4 mb-6">
          <div>
            <h3 class="text-xl font-semibold text-gray-900 dark:text-white flex items-center">
              Order Details
              <span :class="[getStatusClass(order.status), 'ml-3 px-3 py-1 text-sm rounded-full']">
                {{ order.status }}
              </span>
            </h3>
            <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Order #ORD{{ order.order_id.toString().padStart(3, '0') }}
            </p>
          </div>
          <button @click="closeModal" class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 focus:outline-none">
            <span class="sr-only">Close</span>
            <svg class="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <!-- Order Info -->
          <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
            <h4 class="font-medium text-gray-700 dark:text-gray-300 text-sm uppercase tracking-wider mb-3">Order Information</h4>
            <div class="space-y-2">
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400 text-sm">Date:</span>
                <span class="text-gray-800 dark:text-gray-200 text-sm font-medium">{{ formatDate(order.order_date) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400 text-sm">Status:</span>
                <span :class="[getStatusClass(order.status), 'text-sm px-2 py-0.5 rounded-full']">{{ order.status }}</span>
              </div>
              <div class="flex justify-between" v-if="order.quotation_id">
                <span class="text-gray-500 dark:text-gray-400 text-sm">Quotation ID:</span>
                <span class="text-gray-800 dark:text-gray-200 text-sm font-medium">QT{{ order.quotation_id.toString().padStart(3, '0') }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400 text-sm">Total:</span>
                <span class="text-gray-800 dark:text-gray-200 text-sm font-semibold">{{ formatMoney(order.total_amount) }}</span>
              </div>
            </div>
          </div>

          <!-- Customer Info -->
          <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
            <h4 class="font-medium text-gray-700 dark:text-gray-300 text-sm uppercase tracking-wider mb-3">Customer Information</h4>
            <div class="space-y-2">
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400 text-sm">Customer ID:</span>
                <span class="text-gray-800 dark:text-gray-200 text-sm font-medium">{{ order.customer_id }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-gray-500 dark:text-gray-400 text-sm">Name:</span>
                <span class="text-gray-800 dark:text-gray-200 text-sm font-medium">{{ getCustomerName(order.customer_id) }}</span>
              </div>
            </div>
          </div>

          <!-- Shipping Info -->
          <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg">
            <h4 class="font-medium text-gray-700 dark:text-gray-300 text-sm uppercase tracking-wider mb-3">Shipping Address</h4>
            <p class="text-gray-800 dark:text-gray-200 text-sm whitespace-pre-line">{{ order.shipping_address }}</p>
          </div>
        </div>

        <!-- Order Items -->
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden mb-6">
          <div class="bg-gray-50 dark:bg-gray-700 p-4">
            <h4 class="font-medium text-gray-700 dark:text-gray-300 text-sm uppercase tracking-wider">Order Items</h4>
          </div>

          <div v-if="!orderItems || orderItems.length === 0" class="p-8 text-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-12 w-12 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
            <p class="mt-4 text-gray-500 dark:text-gray-400">No items in this order</p>
          </div>

          <table v-else class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
            <thead class="bg-gray-100 dark:bg-gray-800">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Product</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Quantity</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Unit Price</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Discount</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Total</th>
              </tr>
            </thead>
            <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-for="item in orderItems" :key="item.order_item_id">
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200">
                  {{ item.product_name }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200">
                  {{ item.quantity }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200 text-right">
                  {{ formatMoney(item.unit_price) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200 text-right">
                  {{ formatMoney(item.discount) }}
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-800 dark:text-gray-200 text-right font-medium">
                  {{ formatMoney(item.line_total) }}
                </td>
              </tr>
            </tbody>
            <!-- Order Summary -->
            <tfoot class="bg-gray-50 dark:bg-gray-700">
              <tr>
                <td colspan="2" class="px-6 py-3 text-sm text-gray-500 dark:text-gray-400">
                  Total Items: {{ orderSummary.totalItems }}
                </td>
                <td colspan="2" class="px-6 py-3 text-right text-sm font-medium text-gray-700 dark:text-gray-300">
                  Subtotal:
                </td>
                <td class="px-6 py-3 text-right text-sm font-medium text-gray-800 dark:text-gray-200">
                  {{ formatMoney(orderSummary.subtotal) }}
                </td>
              </tr>
              <tr>
                <td colspan="2" class="px-6 py-3"></td>
                <td colspan="2" class="px-6 py-3 text-right text-sm font-medium text-gray-700 dark:text-gray-300">
                  Total Discount:
                </td>
                <td class="px-6 py-3 text-right text-sm font-medium text-red-600 dark:text-red-400">
                  -{{ formatMoney(orderSummary.totalDiscount) }}
                </td>
              </tr>
              <tr class="border-t border-gray-200 dark:border-gray-600">
                <td colspan="2" class="px-6 py-3"></td>
                <td colspan="2" class="px-6 py-3 text-right text-sm font-bold text-gray-800 dark:text-white">
                  Grand Total:
                </td>
                <td class="px-6 py-3 text-right text-sm font-bold text-gray-800 dark:text-white">
                  {{ formatMoney(orderSummary.grandTotal) }}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>

        <!-- Timeline -->
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden mb-6">
          <div class="bg-gray-50 dark:bg-gray-700 p-4">
            <h4 class="font-medium text-gray-700 dark:text-gray-300 text-sm uppercase tracking-wider">Order Timeline</h4>
          </div>
          <div class="p-4">
            <div class="flow-root">
              <ul class="-mb-8">
                <li class="relative pb-8">
                  <div class="relative flex space-x-3">
                    <div>
                      <span class="h-8 w-8 rounded-full bg-green-500 flex items-center justify-center ring-4 ring-white dark:ring-gray-800">
                        <svg class="h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                        </svg>
                      </span>
                    </div>
                    <div class="min-w-0 flex-1 pt-1.5">
                      <div>
                        <p class="text-sm font-medium text-gray-900 dark:text-white">Order Created</p>
                        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ formatDate(order.created_at) }}</p>
                      </div>
                    </div>
                  </div>
                </li>
                <li class="relative pb-8" v-if="order.status !== 'Pending'">
                  <div class="relative flex space-x-3">
                    <div>
                      <span class="h-8 w-8 rounded-full bg-blue-500 flex items-center justify-center ring-4 ring-white dark:ring-gray-800">
                        <svg class="h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                          <path d="M8 16.5a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0zM15 16.5a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z" />
                          <path d="M3 4a1 1 0 00-1 1v10a1 1 0 001 1h1.05a2.5 2.5 0 014.9 0H10a1 1 0 001-1V5a1 1 0 00-1-1H3zM14 7a1 1 0 00-1 1v6.05A2.5 2.5 0 0115.95 16H17a1 1 0 001-1v-5a1 1 0 00-.293-.707l-2-2A1 1 0 0015 7h-1z" />
                        </svg>
                      </span>
                    </div>
                    <div class="min-w-0 flex-1 pt-1.5">
                      <div>
                        <p class="text-sm font-medium text-gray-900 dark:text-white">Order Shipped</p>
                        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ formatDate(order.updated_at) }}</p>
                      </div>
                    </div>
                  </div>
                </li>
                <li class="relative" v-if="order.status === 'Delivered'">
                  <div class="relative flex space-x-3">
                    <div>
                      <span class="h-8 w-8 rounded-full bg-green-500 flex items-center justify-center ring-4 ring-white dark:ring-gray-800">
                        <svg class="h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                          <path d="M2 10.5a1.5 1.5 0 113 0v6a1.5 1.5 0 01-3 0v-6zM6 10.333v5.43a2 2 0 001.106 1.79l.05.025A4 4 0 008.943 18h5.416a2 2 0 001.962-1.608l1.2-6A2 2 0 0015.56 8H12V4a2 2 0 00-2-2 1 1 0 00-1 1v.667a4 4 0 01-.8 2.4L6.8 7.933a4 4 0 00-.8 2.4z" />
                        </svg>
                      </span>
                    </div>
                    <div class="min-w-0 flex-1 pt-1.5">
                      <div>
                        <p class="text-sm font-medium text-gray-900 dark:text-white">Order Delivered</p>
                        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ formatDate(order.updated_at) }}</p>
                      </div>
                    </div>
                  </div>
                </li>
                <li class="relative" v-if="order.status === 'Cancelled'">
                  <div class="relative flex space-x-3">
                    <div>
                      <span class="h-8 w-8 rounded-full bg-red-500 flex items-center justify-center ring-4 ring-white dark:ring-gray-800">
                        <svg class="h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                        </svg>
                      </span>
                    </div>
                    <div class="min-w-0 flex-1 pt-1.5">
                      <div>
                        <p class="text-sm font-medium text-gray-900 dark:text-white">Order Cancelled</p>
                        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ formatDate(order.updated_at) }}</p>
                      </div>
                    </div>
                  </div>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <!-- Footer with buttons -->
        <div class="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700 flex justify-end space-x-3">
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 transition-colors"
          >
            Close
          </button>
          <!-- Additional buttons like print or edit could go here -->
        </div>
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
</style>

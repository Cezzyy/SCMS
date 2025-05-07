<script setup lang="ts">
import { defineProps, defineEmits, ref, computed } from 'vue';
import type { Product } from '../types/Product';
import type { Inventory } from '../types/Inventory';

const props = defineProps<{
  show: boolean;
  item: any; // Can be a Product or an Inventory item with product details
  type: 'product' | 'inventory';
}>();

const emit = defineEmits(['update:show', 'edit-product', 'update-stock']);

// Close modal
const closeModal = () => {
  emit('update:show', false);
};

// Edit actions
const editProduct = () => {
  emit('edit-product', product.value);
  closeModal();
};

const updateStock = () => {
  emit('update-stock', inventory.value);
  closeModal();
};

// Format money with commas
const formatMoney = (amount: number): string => {
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
};

// Format technical specifications for display
const formatTechSpecs = (specs: any): { key: string, value: string }[] => {
  if (!specs) return [];

  // Handle case where specs is a JSON string
  let specsObj = specs;
  if (typeof specs === 'string') {
    try {
      specsObj = JSON.parse(specs);
    } catch (e) {
      // If it's not valid JSON but is a string, return it as a single entry
      if (typeof specs === 'string' && specs.trim() !== '') {
        return [{ key: 'Specifications', value: specs }];
      }
      return [];
    }
  }

  // Handle empty object
  if (!specsObj || Object.keys(specsObj).length === 0) {
    return [];
  }

  // Return as key-value pairs for display
  return Object.entries(specsObj).map(([key, value]) => ({
    key,
    value: typeof value === 'string' ? value : String(value)
  }));
};

// Check if viewing a product or inventory
const isProduct = computed(() => props.type === 'product');

// Get the product information - either directly or from inventory
const product = computed(() => {
  if (!props.item) return null;

  if (isProduct.value) {
    return props.item;
  } else {
    // For inventory items, use all available product properties directly
    return {
      product_id: props.item.product_id,
      product_name: props.item.product_name || 'Unknown Product',
      model: props.item.model,
      price: props.item.price || 0,
      warranty_period: props.item.warranty_period || 0,
      technical_specs: props.item.technical_specs,
      description: props.item.description,
      certifications: props.item.certifications,
      safety_standards: props.item.safety_standards
    };
  }
});

// Get inventory information if applicable
const inventory = computed(() => {
  if (!isProduct.value && props.item) {
    return {
      inventory_id: props.item.inventory_id,
      current_stock: props.item.current_stock,
      reorder_level: props.item.reorder_level,
      last_restock_date: props.item.last_restock_date,
      isLowStock: props.item.isLowStock
    };
  }
  return null;
});

// Parse technical specs
const techSpecs = computed(() => formatTechSpecs(product.value?.technical_specs));

// Determine stock status text and color
const stockStatus = computed(() => {
  if (!inventory.value) return null;

  return inventory.value.isLowStock ? {
    label: 'Low Stock',
    color: 'red',
    message: 'Stock level is at or below reorder threshold.'
  } : {
    label: 'In Stock',
    color: 'green',
    message: 'Stock level is above reorder threshold.'
  };
});

// Helper functions to display data safely
const displayText = (value: any, defaultText = 'Not specified'): string => {
  if (value === undefined || value === null || value === '') {
    return defaultText;
  }
  return String(value);
};
</script>

<template>
  <div v-if="show && item"
      class="fixed inset-0 z-50 overflow-y-auto">
    <div class="min-h-screen px-4 text-center">
      <div class="fixed inset-0 transition-opacity" @click="closeModal">
        <div class="absolute inset-0 bg-black opacity-50 dark:opacity-60"></div>
      </div>

      <span class="inline-block h-screen align-middle" aria-hidden="true">&#8203;</span>

      <div class="inline-block w-full max-w-3xl p-6 my-8 overflow-hidden text-left align-middle transition-all transform bg-white dark:bg-gray-800 shadow-xl rounded-lg border border-gray-200 dark:border-gray-700 modal-content">
        <!-- Modal Header -->
        <div class="flex justify-between items-center border-b border-gray-200 dark:border-gray-700 pb-4 mb-6">
          <h3 class="text-xl font-semibold text-gray-900 dark:text-white flex items-center">
            <span v-if="isProduct" class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-indigo-600 dark:text-indigo-400" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M5 2a1 1 0 011-1h8a1 1 0 011 1v1H5V2zM8 5h4a1 1 0 010 2H8a1 1 0 010-2zm2 4a1 1 0 011 1v3a1 1 0 11-2 0v-3a1 1 0 011-1z" clip-rule="evenodd" />
                <path d="M14 9a1 1 0 00-1-1H7a1 1 0 00-1 1v7a1 1 0 001 1h6a1 1 0 001-1V9z" />
              </svg>
              Product Details
            </span>
            <span v-else class="flex items-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2 text-blue-600 dark:text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 6a1 1 0 011-1h6a1 1 0 110 2H7a1 1 0 01-1-1zm1 3a1 1 0 100 2h6a1 1 0 100-2H7z" clip-rule="evenodd" />
              </svg>
              Inventory Details
            </span>
          </h3>
          <button @click="closeModal" class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 focus:outline-none">
            <span class="sr-only">Close</span>
            <svg class="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Content -->
        <div class="space-y-6">
          <!-- Product Information Section -->
          <div>
            <h4 class="text-lg font-medium text-gray-900 dark:text-white mb-4 flex items-center">
              {{ product?.product_name || 'Unknown Product' }}
              <span v-if="product?.product_id" class="ml-2 text-xs text-gray-500 dark:text-gray-400">#{{ product.product_id }}</span>
            </h4>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <!-- Product basic details -->
              <div class="space-y-3 px-4 py-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Model</p>
                  <p class="text-sm text-gray-900 dark:text-white">{{ displayText(product?.model) }}</p>
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Price</p>
                  <p class="text-sm text-gray-900 dark:text-white font-medium">{{ formatMoney(product?.price || 0) }}</p>
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Warranty Period</p>
                  <p class="text-sm text-gray-900 dark:text-white">{{ product?.warranty_period || 0 }} months</p>
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Certifications</p>
                  <p class="text-sm text-gray-900 dark:text-white">{{ displayText(product?.certifications) }}</p>
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Safety Standards</p>
                  <p class="text-sm text-gray-900 dark:text-white">{{ displayText(product?.safety_standards) }}</p>
                </div>
              </div>

              <!-- Stock information if viewing inventory -->
              <div v-if="inventory" class="space-y-3 px-4 py-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Current Stock</p>
                  <div class="flex items-center mt-1">
                    <span
                      v-if="stockStatus"
                      :class="[
                        'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium',
                        stockStatus.color === 'red' ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-100' : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
                      ]"
                    >
                      {{ inventory.current_stock }} units
                    </span>
                    <span v-else class="text-sm text-gray-900 dark:text-white">
                      {{ inventory.current_stock }} units
                    </span>
                  </div>
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Reorder Level</p>
                  <p class="text-sm text-gray-900 dark:text-white">{{ inventory.reorder_level }} units</p>
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Last Restock Date</p>
                  <p class="text-sm text-gray-900 dark:text-white">
                    {{ inventory.last_restock_date ? new Date(inventory.last_restock_date).toLocaleDateString() : 'Never' }}
                  </p>
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-500 dark:text-gray-300">Stock Status</p>
                  <div
                    v-if="stockStatus"
                    :class="[
                      'mt-1 p-2 rounded-md',
                      stockStatus.color === 'red' ? 'bg-red-50 dark:bg-red-900 dark:bg-opacity-20' : 'bg-green-50 dark:bg-green-900 dark:bg-opacity-20'
                    ]"
                  >
                    <p
                      :class="[
                        'text-sm',
                        stockStatus.color === 'red' ? 'text-red-700 dark:text-red-300' : 'text-green-700 dark:text-green-300'
                      ]"
                    >
                      {{ stockStatus.message }}
                    </p>
                  </div>
                  <p v-else class="text-sm text-gray-900 dark:text-white">Unknown</p>
                </div>
              </div>
            </div>

            <!-- Description -->
            <div class="mt-4">
              <h5 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</h5>
              <div class="bg-gray-50 dark:bg-gray-700 p-3 rounded-lg">
                <p v-if="product?.description" class="text-sm text-gray-600 dark:text-gray-400">
                  {{ product.description }}
                </p>
                <p v-else class="text-sm text-gray-500 dark:text-gray-400 italic">
                  No description available.
                </p>
              </div>
            </div>

            <!-- Technical Specifications -->
            <div class="mt-4">
              <h5 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Technical Specifications</h5>
              <div v-if="techSpecs.length > 0" class="overflow-hidden bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600">
                <div class="px-4 py-2 border-b border-gray-200 dark:border-gray-600 bg-gray-100 dark:bg-gray-600">
                  <div class="grid grid-cols-2 gap-4">
                    <div class="text-xs font-medium text-gray-500 dark:text-gray-300">Specification</div>
                    <div class="text-xs font-medium text-gray-500 dark:text-gray-300">Value</div>
                  </div>
                </div>
                <div class="divide-y divide-gray-200 dark:divide-gray-600">
                  <div v-for="(spec, index) in techSpecs" :key="index" class="px-4 py-2">
                    <div class="grid grid-cols-2 gap-4">
                      <div class="text-sm text-gray-700 dark:text-gray-300">{{ spec.key }}</div>
                      <div class="text-sm text-gray-600 dark:text-gray-400">{{ spec.value }}</div>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="bg-gray-50 dark:bg-gray-700 p-3 rounded-lg text-sm text-gray-500 dark:text-gray-400 italic">
                No technical specifications available.
              </div>
            </div>
          </div>
        </div>

        <!-- Footer with action buttons -->
        <div class="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700 flex justify-between">
          <div v-if="!isProduct && inventory" class="flex space-x-3">
            <button
              type="button"
              @click="updateStock"
              class="px-4 py-2 text-sm font-medium bg-indigo-600 text-white rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 transition-colors"
            >
              Update Stock
            </button>
          </div>
          <div v-else-if="isProduct" class="flex space-x-3">
            <button
              type="button"
              @click="editProduct"
              class="px-4 py-2 text-sm font-medium bg-indigo-600 text-white rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 transition-colors"
            >
              Edit Product
            </button>
          </div>
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-content {
  animation: modal-pop 0.3s ease-out forwards;
}

@keyframes modal-pop {
  0% {
    opacity: 0;
    transform: scale(0.95);
  }
  100% {
    opacity: 1;
    transform: scale(1);
  }
}
</style>

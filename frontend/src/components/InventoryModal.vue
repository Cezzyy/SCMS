<script setup lang="ts">
import { defineProps, defineEmits, ref, watch, computed } from 'vue';
import type { Inventory } from '../types/Inventory';
import type { Product } from '../types/Product';

const props = defineProps<{
  show: boolean;
  inventory: Inventory | null;
  products: Product[];
  existingInventory?: Inventory[];
}>();

const emit = defineEmits(['update:show', 'save']);

const inventoryData = ref<Inventory | null>(null);

// Form validation and UI state
const errors = ref<Record<string, string>>({});
const touched = ref<Record<string, boolean>>({});
const isLoading = ref(false);

// Format money with commas
const formatMoney = (amount: number): string => {
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
};

// Filter out products that already have inventory items
const availableProducts = computed(() => {
  // Don't filter when editing an existing inventory item
  if (inventoryData.value?.inventory_id) {
    return props.products;
  }

  // When creating a new inventory, filter out products that already have inventory
  const inventoriedProductIds = props.existingInventory?.map(inv => inv.product_id) || [];
  return props.products.filter(p => !inventoriedProductIds.includes(p.product_id));
});

// Deep copy inventory data when props change
watch(() => props.inventory, (newVal) => {
  inventoryData.value = newVal ? JSON.parse(JSON.stringify(newVal)) : null;
  // Reset validation state
  errors.value = {};
  touched.value = {};
}, { immediate: true });

// Validation rules
const validateField = (field: string, value: number | string | undefined) => {
  touched.value[field] = true;

  // Required field validation
  if (value === undefined || value === null || value === '') {
    errors.value[field] = `${field.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')} is required`;
    return false;
  }

  // Product validation
  if (field === 'product_id' && (typeof value !== 'number' || value <= 0)) {
    errors.value[field] = 'Please select a product';
    return false;
  }

  // Number validation for stock levels
  if ((field === 'current_stock' || field === 'reorder_level') && (typeof value !== 'number' || value < 0)) {
    errors.value[field] = `${field.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')} must be a positive number`;
    return false;
  }

  delete errors.value[field];
  return true;
};

// Close modal
const closeModal = () => {
  emit('update:show', false);
};

// Save inventory with validation
const saveInventory = () => {
  if (!inventoryData.value || isLoading.value) return;

  isLoading.value = true;

  try {
    // Validate all required fields
    const requiredFields = ['product_id', 'current_stock', 'reorder_level'] as const;
    const validations = requiredFields.map(field =>
      validateField(field, inventoryData.value?.[field])
    );

    if (validations.includes(false)) {
      isLoading.value = false;
      return;
    }

    emit('save', inventoryData.value);
  } catch (error) {
    console.error('Error saving inventory:', error);
    errors.value.submit = 'Failed to save. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Get product name by ID
const getProductName = (productId: number): string => {
  const product = props.products.find(p => p.product_id === productId);
  return product ? product.product_name : 'Unknown Product';
};

// Computed property to check if stock is low
const isLowStock = computed(() => {
  return inventoryData.value &&
    inventoryData.value.current_stock !== undefined &&
    inventoryData.value.reorder_level !== undefined &&
    inventoryData.value.current_stock <= inventoryData.value.reorder_level;
});
</script>

<template>
  <div v-if="show && inventoryData"
      class="fixed inset-0 z-50 overflow-y-auto">
    <div class="min-h-screen px-4 text-center">
      <div class="fixed inset-0 transition-opacity" @click="closeModal">
        <div class="absolute inset-0 bg-black opacity-50"></div>
      </div>

      <span class="inline-block h-screen align-middle" aria-hidden="true">&#8203;</span>

      <div class="inline-block w-full max-w-3xl p-6 my-8 overflow-hidden text-left align-middle transition-all transform bg-white dark:bg-gray-800 shadow-xl rounded-lg border border-gray-200 dark:border-gray-700 modal-content">
        <!-- Modal Header -->
        <div class="flex justify-between items-center border-b border-gray-200 dark:border-gray-700 pb-4 mb-6">
          <h3 class="text-xl font-semibold text-gray-900 dark:text-white">
            {{ inventoryData.inventory_id ? 'Edit Inventory' : 'Add New Inventory' }}
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
        <form @submit.prevent="saveInventory" class="space-y-6">
          <div>
            <label for="product_id" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Product <span class="text-red-500">*</span>
            </label>
            <select
              id="product_id"
              v-model="inventoryData.product_id"
              @blur="validateField('product_id', inventoryData.product_id)"
              required
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.product_id && touched.product_id}"
              :disabled="isLoading"
            >
              <option :value="0" disabled>Select a product</option>
              <option v-for="product in availableProducts" :key="product.product_id" :value="product.product_id">
                {{ product.product_name }} ({{ formatMoney(product.price || 0) }})
              </option>
            </select>
            <p v-if="errors.product_id && touched.product_id" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.product_id }}
            </p>
            <p v-else-if="inventoryData.product_id && inventoryData.product_id > 0" class="mt-2 text-sm text-gray-500 dark:text-gray-400">
              Managing inventory for: <span class="font-medium">{{ getProductName(inventoryData.product_id) }}</span>
            </p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label for="current_stock" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Current Stock <span class="text-red-500">*</span>
              </label>
              <div class="mt-1 relative">
                <input
                  id="current_stock"
                  v-model.number="inventoryData.current_stock"
                  @blur="validateField('current_stock', inventoryData.current_stock)"
                  type="number"
                  min="0"
                  required
                  class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.current_stock && touched.current_stock}"
                  :disabled="isLoading"
                  placeholder="0"
                />
              </div>
              <p v-if="errors.current_stock && touched.current_stock" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.current_stock }}
              </p>
            </div>

            <div>
              <label for="reorder_level" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Reorder Level <span class="text-red-500">*</span>
              </label>
              <div class="mt-1 relative">
                <input
                  id="reorder_level"
                  v-model.number="inventoryData.reorder_level"
                  @blur="validateField('reorder_level', inventoryData.reorder_level)"
                  type="number"
                  min="0"
                  required
                  class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.reorder_level && touched.reorder_level}"
                  :disabled="isLoading"
                  placeholder="5"
                />
              </div>
              <p v-if="errors.reorder_level && touched.reorder_level" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.reorder_level }}
              </p>
              <p v-else class="mt-2 text-sm text-gray-500 dark:text-gray-400">
                Alert will be triggered when stock falls below this level
              </p>
            </div>
          </div>

          <!-- Low stock indicators -->
          <div v-if="inventoryData.current_stock !== undefined && inventoryData.reorder_level !== undefined">
            <div v-if="isLowStock"
                class="rounded-md bg-red-50 dark:bg-red-900 dark:bg-opacity-20 p-4 mt-4">
              <div class="flex">
                <div class="flex-shrink-0">
                  <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div class="ml-3">
                  <h3 class="text-sm font-medium text-red-800 dark:text-red-200">
                    Low Stock Warning
                  </h3>
                  <div class="mt-2 text-sm text-red-700 dark:text-red-300">
                    <p>The current stock level is at or below the reorder threshold.</p>
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="rounded-md bg-green-50 dark:bg-green-900 dark:bg-opacity-20 p-4 mt-4">
              <div class="flex">
                <div class="flex-shrink-0">
                  <svg class="h-5 w-5 text-green-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                  </svg>
                </div>
                <div class="ml-3">
                  <h3 class="text-sm font-medium text-green-800 dark:text-green-200">
                    Stock Level Good
                  </h3>
                  <div class="mt-2 text-sm text-green-700 dark:text-green-300">
                    <p>Current stock level is above the reorder threshold.</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer with buttons -->
          <div class="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700 flex justify-end space-x-3">
            <button
              type="button"
              @click="closeModal"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors"
              :disabled="isLoading"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="isLoading || Object.keys(errors).length > 0"
              class="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
              :class="{'opacity-75 cursor-not-allowed': isLoading}"
            >
              <span v-if="isLoading" class="flex items-center">
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Processing...
              </span>
              <span v-else>{{ inventoryData.inventory_id ? 'Update Inventory' : 'Add Inventory' }}</span>
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

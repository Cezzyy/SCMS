<script setup lang="ts">
import { computed } from 'vue';
import type { Product } from '../types/Product';
import type { Inventory, InventoryWithProduct } from '../types/Inventory';
import type { User } from '../types/User';
import { useUserStore } from '../stores/userStore';

// Define props
const props = defineProps<{
  show: boolean;
  item: Product | Inventory | InventoryWithProduct | User | null;
  type: 'product' | 'inventory' | 'user';
}>();

// Define emits
const emit = defineEmits<{
  'update:show': [value: boolean];
  'edit-product': [product: Product];
  'update-stock': [inventory: Inventory | InventoryWithProduct];
  'edit-user': [user: User];
}>();

// Initialize store for user details if needed
const userStore = useUserStore();

// Methods
const closeModal = () => {
  emit('update:show', false);
};

const editItem = () => {
  if (props.item) {
    if (props.type === 'product') {
      emit('edit-product', props.item as Product);
    } else if (props.type === 'inventory') {
      emit('update-stock', props.item as Inventory | InventoryWithProduct);
    } else if (props.type === 'user') {
      emit('edit-user', props.item as User);
    }
  }
  closeModal();
};

// Computed properties
const modalTitle = computed(() => {
  if (!props.item) return '';

  if (props.type === 'product') {
    return `Product: ${(props.item as Product).product_name}`;
  } else if (props.type === 'inventory') {
    const inventory = props.item as InventoryWithProduct;
    return `Inventory: ${inventory.product_name || ''}`;
  } else if (props.type === 'user') {
    const user = props.item as User;
    return `User: ${user.first_name} ${user.last_name}`;
  }

  return '';
});

// Format money with commas
const formatMoney = (amount?: number): string => {
  if (amount === undefined) return 'N/A';
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
};

// Format date for display
const formatDate = (date?: Date): string => {
  if (!date) return 'N/A';
  return new Date(date).toLocaleString();
};

// Format technical specifications
const formatTechSpecs = (specs: any): string => {
  if (!specs) return 'N/A';

  // Handle case where specs is a JSON string
  let specsObj = specs;
  if (typeof specs === 'string') {
    try {
      specsObj = JSON.parse(specs);
    } catch (e) {
      return specs || 'N/A'; // Return as is if not valid JSON
    }
  }

  // Handle empty object
  if (!specsObj || Object.keys(specsObj).length === 0) {
    return 'N/A';
  }

  // Format the specs for display
  return Object.entries(specsObj)
    .map(([key, value]) => `${key}: ${value}`)
    .join(', ');
};

// Helper functions to display data safely
const displayText = (text: any) => {
  if (text === undefined || text === null) return 'N/A';
  return text;
};
</script>

<template>
  <div v-if="show && item" class="fixed inset-0 z-50 overflow-y-auto">
    <!-- Backdrop -->
    <div class="fixed inset-0 bg-black bg-opacity-50 transition-opacity" @click="closeModal"></div>

    <!-- Modal content -->
    <div class="flex items-center justify-center min-h-screen p-4">
      <div class="relative bg-white dark:bg-gray-800 rounded-lg max-w-xl w-full mx-auto shadow-xl transform transition-all">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-lg font-medium text-gray-900 dark:text-white">{{ modalTitle }}</h3>
          <button
            @click="closeModal"
            class="absolute top-4 right-4 text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
            aria-label="Close"
          >
            <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Content -->
        <div class="px-6 py-4 max-h-[calc(100vh-200px)] overflow-y-auto">
          <!-- Product view -->
          <div v-if="type === 'product' && item" class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Product ID</h4>
                <p class="text-gray-900 dark:text-white">{{ (item as Product).product_id }}</p>
              </div>
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Price</h4>
                <p class="text-gray-900 dark:text-white">{{ formatMoney((item as Product).price) }}</p>
              </div>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Model</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as Product).model || 'N/A' }}</p>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Warranty Period</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as Product).warranty_period }} months</p>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Description</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as Product).description || 'N/A' }}</p>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Technical Specifications</h4>
              <p class="text-gray-900 dark:text-white">{{ formatTechSpecs((item as Product).technical_specs) }}</p>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Certifications</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as Product).certifications || 'N/A' }}</p>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Safety Standards</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as Product).safety_standards || 'N/A' }}</p>
            </div>
          </div>

          <!-- Inventory view -->
          <div v-else-if="type === 'inventory' && item" class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Inventory ID</h4>
                <p class="text-gray-900 dark:text-white">{{ (item as Inventory).inventory_id }}</p>
              </div>
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Product ID</h4>
                <p class="text-gray-900 dark:text-white">{{ (item as Inventory).product_id }}</p>
              </div>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Product Name</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as InventoryWithProduct).product_name || 'N/A' }}</p>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Current Stock</h4>
                <p class="text-gray-900 dark:text-white">{{ (item as Inventory).current_stock }}</p>
              </div>
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Reorder Level</h4>
                <p class="text-gray-900 dark:text-white">{{ (item as Inventory).reorder_level }}</p>
              </div>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Last Restock Date</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as Inventory).last_restock_date ? formatDate((item as Inventory).last_restock_date) : 'Never' }}</p>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Price</h4>
              <p class="text-gray-900 dark:text-white">{{ formatMoney((item as InventoryWithProduct).price) }}</p>
            </div>
          </div>

          <!-- User view -->
          <div v-else-if="type === 'user' && item" class="space-y-4">
            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">User ID</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as User).user_id }}</p>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Full Name</h4>
                <p class="text-gray-900 dark:text-white">{{ (item as User).first_name }} {{ (item as User).last_name }}</p>
              </div>
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Role</h4>
                <p class="text-gray-900 dark:text-white">
                  <span class="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full"
                        :class="{
                          'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-100': (item as User).role === 'admin',
                          'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100': (item as User).role === 'manager',
                          'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100': (item as User).role === 'staff'
                        }">
                    {{ (item as User).role.charAt(0).toUpperCase() + (item as User).role.slice(1) }}
                  </span>
                </p>
              </div>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Email</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as User).email }}</p>
            </div>

            <div>
              <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Phone</h4>
              <p class="text-gray-900 dark:text-white">{{ (item as User).phone || 'N/A' }}</p>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Department</h4>
                <p class="text-gray-900 dark:text-white">{{ (item as User).department || 'N/A' }}</p>
              </div>
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Position</h4>
                <p class="text-gray-900 dark:text-white">{{ (item as User).position || 'N/A' }}</p>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Last Login</h4>
                <p class="text-gray-900 dark:text-white">{{ formatDate((item as User).last_login) }}</p>
              </div>
              <div>
                <h4 class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Created At</h4>
                <p class="text-gray-900 dark:text-white">{{ formatDate((item as User).created_at) }}</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 bg-gray-50 dark:bg-gray-700 border-t border-gray-200 dark:border-gray-600 flex justify-end space-x-3 rounded-b-lg">
          <button
            @click="closeModal"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 sm:text-sm transition-colors"
          >
            Close
          </button>
          <button
            @click="editItem"
            class="px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 sm:text-sm transition-colors"
          >
            {{ type === 'inventory' ? 'Update Stock' : 'Edit' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Smooth transitions */
.fixed {
  transition: all 0.3s ease;
}

/* Scrollbar styles for better UX */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 8px;
}

::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 8px;
}

::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>

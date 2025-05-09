<script setup lang="ts">
import { ref, onMounted, computed, defineAsyncComponent, watch } from 'vue';
import { useProductStore } from '../stores/productStore';
import { useInventoryStore } from '../stores/inventoryStore';
import { storeToRefs } from 'pinia';
import type { Product } from '../types/Product';
import type { Inventory, LowStockItem, StockUpdate } from '../types/Inventory';
import ConfirmationModal from './ConfirmationModal.vue';

const ProductModal = defineAsyncComponent(() => import('../components/ProductModal.vue'));
const InventoryModal = defineAsyncComponent(() => import('../components/InventoryModal.vue'));
const ViewModal = defineAsyncComponent(() => import('../components/ViewModal.vue'));

// Initialize stores
const productStore = useProductStore();
const inventoryStore = useInventoryStore();

// State
const isLoading = ref(false);
const activeTab = ref('inventory'); // 'inventory' or 'low-stock' or 'products'
const showProductModal = ref(false);
const showInventoryModal = ref(false);
const showViewModal = ref(false);
const productToEdit = ref<Product | null>(null);
const inventoryToEdit = ref<Inventory | null>(null);
const itemToView = ref<any>(null);
const viewModalType = ref<'product' | 'inventory'>('product');
const newStockLevel = ref(0);
const searchQuery = ref('');

// Pagination
const currentPage = ref(1);
const itemsPerPage = 10;

// Add state for confirmation modals
const showDeleteProductConfirmation = ref(false);
const showDeleteInventoryConfirmation = ref(false);
const productToDelete = ref<number | null>(null);
const inventoryToDelete = ref<number | null>(null);

// Combined data for display
const inventoryWithProducts = computed(() => {
  return inventoryStore.inventory.map(inv => {
    const product = productStore.products.find(p => p.product_id === inv.product_id);
    return {
      ...inv,
      product_name: product?.product_name || 'Unknown Product',
      model: product?.model || '',
      price: product?.price || 0,
      warranty_period: product?.warranty_period || 0,
      technical_specs: product?.technical_specs,
      certifications: product?.certifications,
      safety_standards: product?.safety_standards,
      description: product?.description,
      isLowStock: inv.current_stock <= inv.reorder_level
    };
  });
});

// Filtered inventory
const filteredInventory = computed(() => {
  if (!searchQuery.value) return inventoryWithProducts.value;

  const query = searchQuery.value.toLowerCase();
  return inventoryWithProducts.value.filter(item =>
    item.product_name.toLowerCase().includes(query) ||
    (typeof (item as any).model === 'string' && (item as any).model.toLowerCase().includes(query))
  );
});

// Filtered products
const filteredProducts = computed(() => {
  if (!searchQuery.value) return productStore.products;

  const query = searchQuery.value.toLowerCase();
  return productStore.products.filter(item =>
    item.product_name.toLowerCase().includes(query) ||
    (item.model && item.model.toLowerCase().includes(query)) ||
    (item.description && item.description.toLowerCase().includes(query))
  );
});

// Filtered low stock items
const filteredLowStockItems = computed(() => {
  if (!searchQuery.value) return inventoryStore.lowStockItems;

  const query = searchQuery.value.toLowerCase();
  return inventoryStore.lowStockItems.filter(item =>
    item.product_name.toLowerCase().includes(query) ||
    (typeof (item as any).model === 'string' && (item as any).model.toLowerCase().includes(query))
  );
});

// Modified to type-safe paginated data
const paginatedItems = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;

  return {
    products: activeTab.value === 'products' ?
      filteredProducts.value.slice(startIndex, endIndex) : [],
    inventory: activeTab.value === 'inventory' ?
      filteredInventory.value.slice(startIndex, endIndex) : [],
    lowStock: activeTab.value === 'low-stock' ?
      filteredLowStockItems.value.slice(startIndex, endIndex) : []
  };
});

// Total pages for pagination
const totalPages = computed(() => {
  let total;

  if (activeTab.value === 'inventory') {
    total = filteredInventory.value.length;
  } else if (activeTab.value === 'low-stock') {
    total = filteredLowStockItems.value.length;
  } else { // products
    total = filteredProducts.value.length;
  }

  return Math.ceil(total / itemsPerPage);
});

// Reset pagination when tab or search changes
watch([activeTab, searchQuery], () => {
  currentPage.value = 1;
});

// Format technical specifications for display
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

// Format money with commas
const formatMoney = (amount: number): string => {
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
};

// Load data
const loadData = async () => {
  isLoading.value = true;
  try {
    await Promise.all([
      productStore.fetchProducts(),
      inventoryStore.fetchInventory(),
      inventoryStore.fetchLowStockWithProductInfo()
    ]);
  } catch (error) {
    console.error('Error loading inventory data:', error);
  } finally {
    isLoading.value = false;
  }
};

// Page navigation
const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
  }
};

// CRUD operations
const openProductModal = (product: Product | null = null) => {
  productToEdit.value = product ? { ...product } : {
    product_id: 0,
    product_name: 'New Product',
    warranty_period: 12,
    price: 1000,
    technical_specs: {},
    model: undefined,
    description: undefined,
    certifications: undefined,
    safety_standards: undefined
  } as Product;
  showProductModal.value = true;
};

const openInventoryModal = (inventory: Inventory | null = null) => {
  inventoryToEdit.value = inventory ? { ...inventory } : {
    inventory_id: 0,
    product_id: 0,
    current_stock: 0,
    reorder_level: 5
  } as Inventory;
  showInventoryModal.value = true;
};

const openUpdateStockModal = (inventory: Inventory) => {
  inventoryToEdit.value = { ...inventory };
  newStockLevel.value = inventory.current_stock;
  showInventoryModal.value = true;
};

const saveProduct = async (product: Product) => {
  try {
    console.log('Saving product in InventoryList:', JSON.stringify(product, null, 2));

    if (product.product_id) {
      await productStore.updateProduct(product);
    } else {
      await productStore.createProduct(product);
    }
    showProductModal.value = false;
    await loadData();
  } catch (error) {
    console.error('Error saving product:', error);
  }
};

const saveInventory = async (inventory: Inventory) => {
  try {
    if (inventory.inventory_id) {
      await inventoryStore.updateInventory(inventory);
    } else {
      await inventoryStore.createInventory(inventory);
    }
    showInventoryModal.value = false;
    await loadData();
  } catch (error) {
    console.error('Error saving inventory:', error);
  }
};

const updateStockLevel = async (inventoryId: number, newStock: number) => {
  try {
    const stockUpdate: StockUpdate = { current_stock: newStock };
    await inventoryStore.updateStock(inventoryId, stockUpdate);
    await loadData();
  } catch (error) {
    console.error('Error updating stock level:', error);
  }
};

const deleteProduct = async (productId: number) => {
  // Show confirmation modal instead of using confirm()
  productToDelete.value = productId;
  showDeleteProductConfirmation.value = true;
};

const confirmDeleteProduct = async () => {
  if (!productToDelete.value) return;
  
  try {
    await productStore.deleteProduct(productToDelete.value);
    await loadData();
  } catch (error) {
    console.error('Error deleting product:', error);
  } finally {
    productToDelete.value = null;
  }
};

const deleteInventory = async (inventoryId: number) => {
  // Show confirmation modal instead of using confirm()
  inventoryToDelete.value = inventoryId;
  showDeleteInventoryConfirmation.value = true;
};

const confirmDeleteInventory = async () => {
  if (!inventoryToDelete.value) return;
  
  try {
    await inventoryStore.deleteInventory(inventoryToDelete.value);
    await loadData();
  } catch (error) {
    console.error('Error deleting inventory:', error);
  } finally {
    inventoryToDelete.value = null;
  }
};

// Open view modal for product or inventory item
const openViewModal = (item: any, type: 'product' | 'inventory') => {
  itemToView.value = item;
  viewModalType.value = type;
  showViewModal.value = true;
};

// Load data on component mount
onMounted(loadData);
</script>

<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700">
    <!-- Header with tabs and actions -->
    <div class="flex flex-col md:flex-row md:justify-between md:items-center p-4 md:p-6 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-gray-800 dark:to-gray-900">
      <div class="mb-4 md:mb-0">
        <h2 class="text-xl md:text-2xl font-bold text-gray-800 dark:text-white mb-1">Inventory Management</h2>
        <p class="text-gray-600 dark:text-gray-300 text-xs md:text-sm">Manage products and stock levels efficiently</p>
      </div>
      <div class="flex flex-wrap gap-3 sm:gap-4">
        <button
          @click="openProductModal()"
          class="bg-blue-600 text-white px-3 py-2 md:px-4 md:py-2 rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center shadow-sm text-sm md:text-base"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Add Product
        </button>
        <button
          @click="openInventoryModal()"
          class="bg-indigo-600 text-white px-3 py-2 md:px-4 md:py-2 rounded-md hover:bg-indigo-700 transition-colors duration-200 flex items-center shadow-sm text-sm md:text-base"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Add Inventory
        </button>
      </div>
    </div>

    <!-- Stats summary cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 p-4 bg-gray-50 dark:bg-gray-900">
      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Total Products</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ productStore.products.length }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Inventory Items</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ inventoryStore.inventory.length }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 sm:col-span-2 md:col-span-1">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-red-100 dark:bg-red-900 text-red-600 dark:text-red-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Low Stock Items</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ inventoryStore.lowStockItems.length }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab navigation -->
    <div class="flex border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
      <button
        @click="activeTab = 'inventory'"
        class="px-3 py-2 md:px-6 md:py-3 text-xs md:text-sm font-medium transition-all duration-200 flex items-center border-b-2 whitespace-nowrap"
        :class="activeTab === 'inventory'
          ? 'border-blue-600 text-blue-600 dark:border-blue-500 dark:text-blue-400'
          : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 md:h-4 md:w-4 mr-1 md:mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
        </svg>
        All Inventory
      </button>
      <button
        @click="activeTab = 'low-stock'"
        class="px-3 py-2 md:px-6 md:py-3 text-xs md:text-sm font-medium transition-all duration-200 flex items-center border-b-2 whitespace-nowrap"
        :class="activeTab === 'low-stock'
          ? 'border-red-600 text-red-600 dark:border-red-500 dark:text-red-400'
          : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 md:h-4 md:w-4 mr-1 md:mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        Low Stock Items
        <span
          v-if="inventoryStore.lowStockItems.length"
          class="ml-1 md:ml-2 bg-red-100 text-red-600 dark:bg-red-900 dark:text-red-300 px-1.5 py-0.5 rounded-full text-xs font-medium"
        >
          {{ inventoryStore.lowStockItems.length }}
        </span>
      </button>
      <button
        @click="activeTab = 'products'"
        class="px-3 py-2 md:px-6 md:py-3 text-xs md:text-sm font-medium transition-all duration-200 flex items-center border-b-2 whitespace-nowrap"
        :class="activeTab === 'products'
          ? 'border-indigo-600 text-indigo-600 dark:border-indigo-500 dark:text-indigo-400'
          : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 md:h-4 md:w-4 mr-1 md:mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        Products
      </button>
    </div>

    <!-- Enhanced search bar -->
    <div class="p-3 md:p-4 border-b border-gray-200 dark:border-gray-700">
      <div class="relative">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <svg class="h-4 w-4 md:h-5 md:w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
          </svg>
        </div>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by name, model, or description..."
          class="w-full pl-8 md:pl-10 px-3 md:px-4 py-2 md:py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg
                 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                 transition-all duration-200
                 dark:bg-gray-700 dark:text-white text-xs md:text-sm"
        />
      </div>
    </div>

    <!-- Improved loading indicator -->
    <div v-if="isLoading" class="flex justify-center items-center p-12">
      <div class="relative">
        <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-200 dark:border-gray-600"></div>
        <div class="animate-spin rounded-full h-12 w-12 border-t-4 border-blue-600 dark:border-blue-500 absolute top-0 left-0"></div>
      </div>
      <div class="ml-4 text-gray-600 dark:text-gray-300 text-sm font-medium">Loading data...</div>
    </div>

    <!-- All Products tab with improved table -->
    <div v-else-if="activeTab === 'products'" class="p-4">
      <div v-if="filteredProducts.length === 0 && searchQuery" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <p class="text-gray-600 dark:text-gray-300 mb-2">No results found</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Try adjusting your search criteria</p>
      </div>
      <div v-else-if="!productStore.products || productStore.products.length === 0"
           class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        <p class="text-gray-600 dark:text-gray-300 mb-2">No products found</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Add some products to get started</p>
        <button
          @click="openProductModal()"
          class="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Create First Product
        </button>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 shadow-sm">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th class="group px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                <div class="flex items-center">
                  Product Name
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-1 opacity-0 group-hover:opacity-100 transition-opacity duration-200" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
                  </svg>
                </div>
              </th>
              <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                Model
              </th>
              <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                Technical Specs
              </th>
              <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                Price
              </th>
              <th class="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider sticky right-0 bg-gray-50 dark:bg-gray-700 z-10">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="product in paginatedItems.products" :key="product.product_id"
                class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="font-medium text-gray-900 dark:text-white truncate max-w-[200px]">{{ product.product_name }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-300">
                {{ product.model || 'N/A' }}
              </td>
              <td class="px-6 py-4 text-gray-500 dark:text-gray-300">
                <div class="truncate max-w-[250px]">{{ formatTechSpecs(product.technical_specs) }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="font-medium text-gray-800 dark:text-gray-200">{{ formatMoney(product.price) }}</span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium sticky right-0 bg-white dark:bg-gray-800 shadow-sticky">
                <button
                  @click="openViewModal(product, 'product')"
                  class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-3 transition-colors duration-200"
                  title="View product details"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                    <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
                  </svg>
                </button>
                <button
                  @click="openProductModal(product)"
                  class="text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-300 mr-3 transition-colors duration-200"
                  title="Edit product"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                  </svg>
                </button>
                <button
                  @click="deleteProduct(product.product_id)"
                  class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 transition-colors duration-200"
                  title="Delete product"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                  </svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div v-if="activeTab === 'products' && totalPages > 1" class="flex justify-center mt-6">
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
    </div>

    <!-- All inventory tab -->
    <div v-else-if="activeTab === 'inventory'" class="p-4">
      <div v-if="!filteredInventory.length && searchQuery" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <p class="text-gray-600 dark:text-gray-300 mb-2">No results found</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Try adjusting your search criteria</p>
      </div>
      <div v-else-if="!filteredInventory.length" class="text-center py-8 text-gray-500 dark:text-gray-400">
        No inventory items found. Add some products and inventory to get started.
      </div>
      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Product
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Model
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Price
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Current Stock
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Reorder Level
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Last Restock
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider sticky right-0 bg-gray-50 dark:bg-gray-800 z-10">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="item in paginatedItems.inventory" :key="item.inventory_id" :class="{ 'bg-red-50 dark:bg-red-900 dark:bg-opacity-20': item.isLowStock }">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="font-medium text-gray-900 dark:text-white truncate max-w-[200px]">{{ item.product_name }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-400">
                {{ item.model || 'N/A' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-400">
                {{ formatMoney(item.price) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  :class="[
                    'px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full',
                    item.isLowStock
                      ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-100'
                      : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100'
                  ]"
                >
                  {{ item.current_stock }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-400">
                {{ item.reorder_level }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-400">
                {{ item.last_restock_date ? new Date(item.last_restock_date).toLocaleDateString() : 'Never' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium sticky right-0 bg-white dark:bg-gray-800 shadow-sticky">
                <button
                  @click="openViewModal(item, 'inventory')"
                  class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-3 transition-colors duration-200"
                  title="View inventory details"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                    <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
                  </svg>
                </button>
                <button
                  @click="openUpdateStockModal(item)"
                  class="text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-300 mr-3 transition-colors duration-200"
                  title="Update inventory"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                  </svg>
                </button>
                <button
                  @click="deleteInventory(item.inventory_id)"
                  class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 transition-colors duration-200"
                  title="Delete inventory"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                  </svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="flex justify-center mt-6">
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
    </div>

    <!-- Low stock tab -->
    <div v-else-if="activeTab === 'low-stock'" class="p-4">
      <div v-if="!filteredLowStockItems.length && searchQuery" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <p class="text-gray-600 dark:text-gray-300 mb-2">No results found</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Try adjusting your search criteria</p>
      </div>
      <div v-else-if="!inventoryStore.lowStockItems.length" class="text-center py-8 text-gray-500 dark:text-gray-400">
        No low stock items found. All inventory levels are above their reorder thresholds.
      </div>

      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-800">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Product
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Current Stock
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Reorder Level
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Price
              </th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider sticky right-0 bg-gray-50 dark:bg-gray-800 z-10">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="item in paginatedItems.lowStock" :key="item.inventory_id" class="bg-red-50 dark:bg-red-900 dark:bg-opacity-20">
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="font-medium text-gray-900 dark:text-white truncate max-w-[200px]">{{ item.product_name }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-100">
                  {{ item.current_stock }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-400">
                {{ item.reorder_level }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-400">
                {{ formatMoney(item.price) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium sticky right-0 bg-red-50 dark:bg-red-900 dark:bg-opacity-20 shadow-sticky">
                <button
                  @click="openViewModal(item, 'inventory')"
                  class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-3 transition-colors duration-200"
                  title="View inventory details"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                    <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
                  </svg>
                </button>
                <button
                  @click="openUpdateStockModal(item)"
                  class="inline-flex items-center px-3 py-1.5 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M10 5a1 1 0 011 1v3h3a1 1 0 110 2h-3v3a1 1 0 11-2 0v-3H6a1 1 0 110-2h3V6a1 1 0 011-1z" clip-rule="evenodd" />
                  </svg>
                  Restock
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="flex justify-center mt-6">
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
    </div>

    <!-- Product Modal Component -->
    <ProductModal
      v-if="showProductModal"
      :show="showProductModal"
      :product="productToEdit"
      @update:show="showProductModal = $event"
      @save="saveProduct"
    />

    <!-- Inventory Modal Component -->
    <InventoryModal
      v-if="showInventoryModal"
      :show="showInventoryModal"
      :inventory="inventoryToEdit"
      :products="productStore.products"
      :existingInventory="inventoryStore.inventory"
      @save="saveInventory"
    />

    <!-- View Modal Component -->
    <ViewModal
      :show="showViewModal"
      @update:show="showViewModal = $event"
      :item="itemToView"
      :type="viewModalType"
      @edit-product="openProductModal"
      @update-stock="openUpdateStockModal"
    />

    <!-- Confirmation Modals -->
    <ConfirmationModal
      v-model:show="showDeleteProductConfirmation"
      title="Delete Product"
      message="Are you sure you want to delete this product? This action cannot be undone."
      confirmButtonText="Delete"
      @confirm="confirmDeleteProduct"
    />

    <ConfirmationModal
      v-model:show="showDeleteInventoryConfirmation"
      title="Delete Inventory Item"
      message="Are you sure you want to delete this inventory item? This action cannot be undone."
      confirmButtonText="Delete"
      @confirm="confirmDeleteInventory"
    />
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

  /* Adjust modal sizing */
  :deep(.modal-content) {
    width: 95%;
    max-width: 95vw;
    margin: 0 auto;
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

/* Make sure pagination is responsive */
@media (max-width: 480px) {
  :deep(.pagination-item) {
    margin: 0 0.1rem;
    padding: 0.3rem 0.5rem;
    font-size: 0.75rem;
  }
}

/* Original styles */
/* Add fade and pulse animations for better UI */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Improved scrollbars for better UX */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
}

::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #555;
}

/* Shadow for sticky columns */
.shadow-sticky {
  box-shadow: -3px 0 5px rgba(0, 0, 0, 0.1);
}
</style>

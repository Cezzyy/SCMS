<script setup lang="ts">
import { ref, computed, watch, defineEmits, defineProps } from 'vue';
import { useProductStore } from '../stores/productStore';
import { useInventoryStore } from '../stores/inventoryStore';
import type { Quotation, QuotationCreate, QuotationItemCreate } from '../types/Quotation';
import type { Customer } from '../types/Customer';

const props = defineProps({
  show: {
    type: Boolean,
    required: true
  },
  quotation: {
    type: Object as () => Partial<Quotation> | null,
    default: null
  },
  customers: {
    type: Array as () => Customer[],
    required: true
  }
});

const emit = defineEmits(['update:show', 'save']);

// Get product and inventory stores
const productStore = useProductStore();
const inventoryStore = useInventoryStore();

// Get today's date in YYYY-MM-DD format
const today = new Date().toISOString().slice(0, 10);

// Local state
const localQuotation = ref<Partial<QuotationCreate>>({
  customer_id: 0,  // Fix: Changed from string to number, but set to 0 to indicate not selected
  quote_date: today, // Fixed to today's date
  validity_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
  status: 'Pending',
  total_amount: 0,
  items: []
});

const newItem = ref<QuotationItemCreate>({
  product_id: 0,
  quantity: 1,
  unit_price: 0,
  discount: 0,
  line_total: 0
});

// Reset form when modal is opened
watch(() => props.show, (show) => {
  if (show && props.quotation) {
    // Clone the quotation data to avoid modifying the original
    localQuotation.value = { ...props.quotation };
    
    // Ensure items array exists
    if (!localQuotation.value.items) {
      localQuotation.value.items = [];
    } else {
      // Clone the items to avoid modifying the originals
      localQuotation.value.items = [...localQuotation.value.items];
    }
    
    // Ensure dates are formatted properly
    if (localQuotation.value.quote_date) {
      localQuotation.value.quote_date = today; // Always use today's date
    }
    if (localQuotation.value.validity_date) {
      // Ensure validity date is not in the past
      const validityDate = new Date(localQuotation.value.validity_date);
      const todayDate = new Date(today);
      if (validityDate < todayDate) {
        // If in the past, set to 30 days from today
        localQuotation.value.validity_date = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
      } else {
        localQuotation.value.validity_date = new Date(localQuotation.value.validity_date).toISOString().slice(0, 10);
      }
    }
  } else if (show) {
    // Reset for new quotation
    localQuotation.value = {
      customer_id: 0,  // Fix: Changed from string to number, but set to 0 to indicate not selected
      quote_date: today, // Fixed to today's date
      validity_date: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10),
      status: 'Pending',
      total_amount: 0,
      items: []
    };
  }
});

// Computed properties
const isNewQuotation = computed(() => !props.quotation?.quotation_id);
const modalTitle = computed(() => isNewQuotation.value ? 'Create New Quotation' : 'Edit Quotation');

const selectedProduct = computed(() => {
  if (!newItem.value.product_id) return null;
  return productStore.products.find(p => p.product_id === newItem.value.product_id);
});

// Get available stock for selected product
const availableStock = computed(() => {
  if (!selectedProduct.value) return 0;
  
  const inventory = inventoryStore.inventory.find(
    (item: any) => item.product_id === selectedProduct.value?.product_id
  );
  
  return inventory ? inventory.current_stock : 0;
});

// Check if product is already in the quotation
const isProductAlreadyAdded = computed(() => {
  if (!newItem.value.product_id || !localQuotation.value.items) return false;
  
  return localQuotation.value.items.some(
    item => item.product_id === newItem.value.product_id
  );
});

// Load products and inventory if needed
const loadData = async () => {
  try {
    const promises = [];
    
    if (productStore.products.length === 0) {
      promises.push(productStore.fetchProducts());
    }
    
    if (inventoryStore.inventory.length === 0) {
      promises.push(inventoryStore.fetchInventory());
    }
    
    await Promise.all(promises);
  } catch (error) {
    console.error('Error loading product or inventory data:', error);
  }
};

// Calculate line total for an item
const calculateLineTotal = (item: QuotationItemCreate): number => {
  const subtotal = item.quantity * item.unit_price;
  const discountAmount = subtotal * (item.discount / 100);
  return subtotal - discountAmount;
};

// Update line total when item details change
watch([() => newItem.value.quantity, () => newItem.value.unit_price, () => newItem.value.discount], () => {
  newItem.value.line_total = calculateLineTotal(newItem.value);
});

// Update item unit price when product changes
watch(() => newItem.value.product_id, () => {
  if (selectedProduct.value) {
    newItem.value.unit_price = selectedProduct.value.price;
    // Reset quantity to 1 or available stock if only 1 left
    newItem.value.quantity = availableStock.value > 0 ? 1 : 0;
    newItem.value.line_total = calculateLineTotal(newItem.value);
  }
});

// Calculate total amount for the whole quotation
const calculateTotalAmount = (): number => {
  if (!localQuotation.value.items) return 0;
  return localQuotation.value.items.reduce((total, item) => total + item.line_total, 0);
};

// Update total amount when items change
const updateTotalAmount = () => {
  localQuotation.value.total_amount = calculateTotalAmount();
};

// Add item to quotation
const addItem = () => {
  if (!localQuotation.value.items) {
    localQuotation.value.items = [];
  }

  if (newItem.value.product_id && newItem.value.quantity > 0) {
    // Check if product already exists in quotation
    if (isProductAlreadyAdded.value) {
      alert('This product is already added to the quotation. Please modify the existing entry instead.');
      return;
    }
    
    // Check if quantity is within available stock
    if (newItem.value.quantity > availableStock.value) {
      alert(`Cannot add more than the available stock (${availableStock.value} units).`);
      return;
    }
    
    // Add the item
    localQuotation.value.items.push({ ...newItem.value });
    
    // Update total
    updateTotalAmount();
    
    // Reset the new item form
    newItem.value = {
      product_id: 0,
      quantity: 1,
      unit_price: 0,
      discount: 0,
      line_total: 0
    };
  }
};

// Remove item from quotation
const removeItem = (index: number) => {
  if (localQuotation.value.items) {
    localQuotation.value.items.splice(index, 1);
    updateTotalAmount();
  }
};

// Format money for display
const formatMoney = (amount: number): string => {
  return '₱' + amount.toFixed(2).replace(/\d(?=(\d{3})+\.)/g, '$&,');
};

// Get product name by ID
const getProductName = (productId: number): string => {
  const product = productStore.products.find(p => p.product_id === productId);
  return product ? product.product_name : 'Unknown Product';
};

// Save the quotation
const saveQuotation = () => {
  // Validate required fields
  if (!localQuotation.value.customer_id || localQuotation.value.customer_id === 0) {
    alert('Please select a customer');
    return;
  }
  
  if (!localQuotation.value.items || localQuotation.value.items.length === 0) {
    alert('Please add at least one item to the quotation');
    return;
  }
  
  // Ensure customer_id is a number
  const quotationToSave = {
    ...localQuotation.value,
    customer_id: Number(localQuotation.value.customer_id),
    items: localQuotation.value.items ? localQuotation.value.items.map(item => ({...item})) : []
  };
  
  // Log the data to verify correct format
  console.log('Saving quotation modal data:', quotationToSave);
  console.log('customer_id value and type:', quotationToSave.customer_id, typeof quotationToSave.customer_id);
  
  emit('save', quotationToSave);
};

// Close the modal
const closeModal = () => {
  emit('update:show', false);
};

// Load products and inventory on component creation
loadData();
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 overflow-auto bg-black bg-opacity-50 flex items-center justify-center p-4">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-5xl w-full max-h-[90vh] flex flex-col">
      <!-- Modal header -->
      <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-lg font-medium text-gray-900 dark:text-white">{{ modalTitle }}</h3>
        <button 
          @click="closeModal" 
          class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 focus:outline-none"
        >
          <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Modal body -->
      <div class="p-6 overflow-y-auto">
        <form @submit.prevent="saveQuotation">
          <!-- Main quotation information -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Customer *</label>
              <select 
                v-model="localQuotation.customer_id"
                required
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 p-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                @change="console.log('Selected customer:', localQuotation.customer_id)"
              >
                <option :value="0" disabled selected>Select a customer</option>
                <option v-for="customer in customers" :key="customer.customer_id" :value="customer.customer_id">
                  {{ customer.company_name }}
                </option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Status</label>
              <select 
                v-model="localQuotation.status"
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 p-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              >
                <option value="Pending">Pending</option>
                <option value="Approved">Approved</option>
                <option value="Rejected">Rejected</option>
                <option value="Expired">Expired</option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Quote Date *</label>
              <input 
                v-model="localQuotation.quote_date"
                type="date"
                required
                readonly
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 p-2 bg-gray-100 dark:bg-gray-600 text-gray-900 dark:text-white cursor-not-allowed"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Valid Until *</label>
              <input 
                v-model="localQuotation.validity_date"
                type="date"
                required
                :min="today"
                class="w-full rounded-md border border-gray-300 dark:border-gray-600 p-2 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
              />
            </div>
          </div>

          <!-- Quotation items section -->
          <div class="mb-6">
            <h4 class="text-lg font-medium text-gray-900 dark:text-white mb-3">Quotation Items</h4>
            
            <!-- Existing items table -->
            <div v-if="localQuotation.items && localQuotation.items.length > 0" class="mb-6 overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead class="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Product
                    </th>
                    <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Quantity
                    </th>
                    <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Unit Price
                    </th>
                    <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Discount
                    </th>
                    <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Line Total
                    </th>
                    <th class="px-4 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                  <tr v-for="(item, index) in localQuotation.items" :key="index">
                    <td class="px-4 py-3 whitespace-nowrap">
                      <div class="font-medium text-gray-900 dark:text-white">{{ getProductName(item.product_id) }}</div>
                    </td>
                    <td class="px-4 py-3 whitespace-nowrap text-right text-gray-500 dark:text-gray-300">
                      {{ item.quantity }}
                    </td>
                    <td class="px-4 py-3 whitespace-nowrap text-right text-gray-500 dark:text-gray-300">
                      {{ formatMoney(item.unit_price) }}
                    </td>
                    <td class="px-4 py-3 whitespace-nowrap text-right text-gray-500 dark:text-gray-300">
                      {{ item.discount }}%
                    </td>
                    <td class="px-4 py-3 whitespace-nowrap text-right font-medium text-gray-900 dark:text-white">
                      {{ formatMoney(item.line_total) }}
                    </td>
                    <td class="px-4 py-3 whitespace-nowrap text-center">
                      <button 
                        @click="removeItem(index)" 
                        type="button"
                        class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                      >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                          <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                        </svg>
                      </button>
                    </td>
                  </tr>
                </tbody>
                <tfoot class="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <td colspan="4" class="px-4 py-3 text-right font-medium text-gray-700 dark:text-gray-300">
                      Total Amount:
                    </td>
                    <td class="px-4 py-3 text-right font-bold text-gray-900 dark:text-white">
                      {{ formatMoney(localQuotation.total_amount || 0) }}
                    </td>
                    <td></td>
                  </tr>
                </tfoot>
              </table>
            </div>

            <!-- Add new item form -->
            <div class="bg-gray-50 dark:bg-gray-700 p-4 rounded-lg border border-gray-200 dark:border-gray-600">
              <h5 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Add Item</h5>
              <div class="grid grid-cols-1 sm:grid-cols-5 gap-4">
                <div class="sm:col-span-2">
                  <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Product *</label>
                  <select 
                    v-model="newItem.product_id"
                    required
                    class="w-full rounded-md border border-gray-300 dark:border-gray-600 p-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    :class="{ 'border-red-500': isProductAlreadyAdded }"
                  >
                    <option value="0" disabled selected>Select a product</option>
                    <option 
                      v-for="product in productStore.products" 
                      :key="product.product_id" 
                      :value="product.product_id"
                      :disabled="localQuotation.items?.some(item => item.product_id === product.product_id)"
                    >
                      {{ product.product_name }} ({{ formatMoney(product.price) }})
                    </option>
                  </select>
                  <div v-if="isProductAlreadyAdded" class="text-xs text-red-500 mt-1">
                    This product is already in the quotation
                  </div>
                </div>
                <div>
                  <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Quantity *</label>
                  <input 
                    v-model.number="newItem.quantity"
                    type="number"
                    min="1"
                    :max="availableStock"
                    required
                    class="w-full rounded-md border border-gray-300 dark:border-gray-600 p-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    @input="e => { 
                      if (newItem.quantity > availableStock) {
                        newItem.quantity = availableStock; 
                      } else if (newItem.quantity < 1) {
                        newItem.quantity = 1;
                      }
                    }"
                  />
                  <div v-if="selectedProduct" class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    Available: {{ availableStock }}
                  </div>
                </div>
                <div>
                  <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Discount (%)</label>
                  <input 
                    v-model.number="newItem.discount"
                    type="number"
                    min="0"
                    max="100"
                    class="w-full rounded-md border border-gray-300 dark:border-gray-600 p-2 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    @input="e => {
                      if (newItem.discount > 100) {
                        newItem.discount = 100;
                      } else if (newItem.discount < 0) {
                        newItem.discount = 0;
                      }
                    }"
                  />
                </div>
                <div>
                  <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Actions</label>
                  <button 
                    @click="addItem"
                    type="button"
                    :disabled="!newItem.product_id || newItem.quantity <= 0"
                    class="w-full px-3 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Add Item
                  </button>
                </div>
              </div>
              <div v-if="newItem.product_id && newItem.quantity > 0" class="mt-3 text-sm text-gray-700 dark:text-gray-300">
                <span class="font-medium">Subtotal:</span> {{ formatMoney(newItem.line_total) }}
              </div>
            </div>

            <!-- No items message -->
            <div v-if="!localQuotation.items || localQuotation.items.length === 0" class="text-center py-4 text-gray-500 dark:text-gray-400 mt-3">
              No items added yet. Add some items using the form above.
            </div>
          </div>
        </form>
      </div>

      <!-- Modal footer -->
      <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end">
        <button 
          @click="closeModal" 
          type="button" 
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 mr-3"
        >
          Cancel
        </button>
        <button 
          @click="saveQuotation" 
          type="button" 
          :disabled="!localQuotation.customer_id || (localQuotation.items && localQuotation.items.length === 0)"
          class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Save Quotation
        </button>
      </div>
    </div>
  </div>
</template> 
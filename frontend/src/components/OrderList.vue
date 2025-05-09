<template>
  <div class="container mx-auto p-4 sm:p-6 lg:p-8 bg-white dark:bg-gray-800 shadow-lg rounded-lg">
    <div class="flex flex-col sm:flex-row justify-between items-center mb-6 pb-4 border-b border-gray-200 dark:border-gray-700">
      <div>
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-800 dark:text-white">Order Management</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400">Manage and track all customer orders efficiently.</p>
      </div>
      <button
        @click="openNewOrderModal"
        class="mt-4 sm:mt-0 flex items-center bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 rounded-lg shadow-md transition duration-150 ease-in-out transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-opacity-50"
      >
        <font-awesome-icon icon="plus" class="mr-2 h-4 w-4" />
        New Order
      </button>
    </div>

    <!-- Search and Filters -->
    <div class="mb-6 flex flex-col sm:flex-row justify-between items-center gap-4">
      <div class="relative w-full sm:w-auto flex-grow max-w-md">
        <input
          type="text"
          v-model="searchQuery"
          placeholder="Search by Order ID, Customer..."
          class="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:text-white placeholder-gray-400 dark:placeholder-gray-500"
        />
        <font-awesome-icon icon="search" class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400 dark:text-gray-500" />
      </div>
      <!-- Add filter dropdowns here if needed -->
    </div>

    <!-- Orders Table -->
    <div v-if="isLoading" class="flex justify-center items-center py-12">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      <p class="ml-3 text-gray-600 dark:text-gray-300">Loading orders...</p>
    </div>

    <div v-else-if="paginatedOrders.length === 0 && !isLoading" class="text-center py-12">
        <font-awesome-icon icon="folder-open" class="h-16 w-16 text-gray-400 dark:text-gray-500 mb-4" />
        <p class="text-xl text-gray-500 dark:text-gray-400">No orders found.</p>
        <p class="text-sm text-gray-400 dark:text-gray-500 mt-1">Try adjusting your search or filters, or create a new order.</p>
    </div>

    <div v-else class="overflow-x-auto bg-white dark:bg-gray-800 shadow-md rounded-lg">
      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
        <thead class="bg-gray-50 dark:bg-gray-700">
          <tr>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Order ID</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Customer Name</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Order Date</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Status</th>
            <th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Total Amount</th>
            <th scope="col" class="px-6 py-3 text-center text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
          <tr v-for="order in paginatedOrders" :key="order.id" class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150">
            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">#{{ order.id }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-300">{{ order.customerName }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-300">{{ formatDate(order.orderDate) }}</td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="getStatusClass(order.status)" class="px-3 py-1 inline-flex text-xs leading-5 font-semibold rounded-full">
                {{ order.status }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-300">${{ order.totalAmount.toFixed(2) }}</td>
            <td class="px-6 py-4 whitespace-nowrap text-center text-sm font-medium space-x-2">
              <button @click="viewOrderDetails(order)" class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 transition-colors duration-150" title="View Details">
                <font-awesome-icon icon="eye" class="h-5 w-5" />
              </button>
              <button @click="editOrder(order)" class="text-yellow-500 hover:text-yellow-700 dark:text-yellow-400 dark:hover:text-yellow-300 transition-colors duration-150" title="Edit Order">
                <font-awesome-icon icon="pencil-alt" class="h-5 w-5" />
              </button>
               <button @click="deleteOrder(order.id)" class="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300 transition-colors duration-150" title="Delete Order">
                <font-awesome-icon icon="trash" class="h-5 w-5" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="mt-6 flex justify-between items-center">
        <button
            @click="prevPage"
            :disabled="currentPage === 1"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 dark:bg-gray-700 dark:text-gray-200 dark:border-gray-600 dark:hover:bg-gray-600"
        >
            <font-awesome-icon icon="chevron-left" class="mr-2 h-4 w-4" /> Previous
        </button>
        <span class="text-sm text-gray-700 dark:text-gray-300">
            Page {{ currentPage }} of {{ totalPages }}
        </span>
        <button
            @click="nextPage"
            :disabled="currentPage === totalPages"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 dark:bg-gray-700 dark:text-gray-200 dark:border-gray-600 dark:hover:bg-gray-600"
        >
            Next <font-awesome-icon icon="chevron-right" class="ml-2 h-4 w-4" />
        </button>
    </div>

    <!-- Modals for New/Edit/View Order would go here -->
    <!-- e.g., <NewOrderModal v-if="showNewOrderModal" @close="showNewOrderModal = false" /> -->
    <!-- e.g., <ViewOrderModal v-if="selectedOrder" :order="selectedOrder" @close="selectedOrder = null" /> -->

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
// import NewOrderModal from './NewOrderModal.vue'; // Placeholder for modal component
// import ViewOrderModal from './ViewOrderModal.vue'; // Placeholder for modal component

interface Order {
  id: string;
  customerName: string;
  orderDate: string; // ISO string format for dates
  status: 'Pending' | 'Processing' | 'Shipped' | 'Delivered' | 'Cancelled';
  totalAmount: number;
  // Add more fields as needed, e.g., items, shippingAddress, etc.
}

const isLoading = ref(true);
const searchQuery = ref('');
const orders = ref<Order[]>([]);

// Placeholder data - replace with API call
const sampleOrders: Order[] = [
  { id: 'ORD001', customerName: 'Alice Wonderland', orderDate: new Date(2023, 0, 15).toISOString(), status: 'Delivered', totalAmount: 150.75 },
  { id: 'ORD002', customerName: 'Bob The Builder', orderDate: new Date(2023, 1, 20).toISOString(), status: 'Shipped', totalAmount: 89.99 },
  { id: 'ORD003', customerName: 'Charlie Brown', orderDate: new Date(2023, 2, 5).toISOString(), status: 'Processing', totalAmount: 220.00 },
  { id: 'ORD004', customerName: 'Diana Prince', orderDate: new Date(2023, 2, 10).toISOString(), status: 'Pending', totalAmount: 45.50 },
  { id: 'ORD005', customerName: 'Edward Scissorhands', orderDate: new Date(2023, 2, 12).toISOString(), status: 'Cancelled', totalAmount: 12.00 },
];

onMounted(() => {
  // Simulate API call
  setTimeout(() => {
    orders.value = sampleOrders;
    isLoading.value = false;
  }, 1500);
});

const filteredOrders = computed(() => {
  if (!searchQuery.value) {
    return orders.value;
  }
  const lowerSearch = searchQuery.value.toLowerCase();
  return orders.value.filter(order =>
    order.id.toLowerCase().includes(lowerSearch) ||
    order.customerName.toLowerCase().includes(lowerSearch) ||
    order.status.toLowerCase().includes(lowerSearch)
  );
});

// Pagination
const currentPage = ref(1);
const itemsPerPage = 10;
const totalPages = computed(() => Math.ceil(filteredOrders.value.length / itemsPerPage));

const paginatedOrders = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage;
  const end = start + itemsPerPage;
  return filteredOrders.value.slice(start, end);
});

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++;
  }
}

function prevPage() {
  if (currentPage.value > 1) {
    currentPage.value--;
  }
}

// Modal states (placeholders)
const showNewOrderModal = ref(false);
const selectedOrder = ref<Order | null>(null); // For viewing or editing

const openNewOrderModal = () => {
  showNewOrderModal.value = true;
  // console.log('Open new order modal');
};

const viewOrderDetails = (order: Order) => {
  selectedOrder.value = order;
  // console.log('View order:', order);
  // Implement logic to show a view modal, e.g., by setting a ref for a modal component
};

const editOrder = (order: Order) => {
  selectedOrder.value = order;
  // console.log('Edit order:', order);
  // Implement logic to show an edit modal
};

const deleteOrder = (orderId: string) => {
  if(confirm(`Are you sure you want to delete order #${orderId}?`)){
    orders.value = orders.value.filter(o => o.id !== orderId);
    // console.log('Delete order:', orderId);
    // Add API call for deletion here
  }
};

const formatDate = (dateString: string) => {
  const options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'long', day: 'numeric' };
  return new Date(dateString).toLocaleDateString(undefined, options);
};

const getStatusClass = (status: Order['status']) => {
  switch (status) {
    case 'Pending':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-700 dark:text-yellow-100';
    case 'Processing':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-700 dark:text-blue-100';
    case 'Shipped':
      return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-700 dark:text-indigo-100';
    case 'Delivered':
      return 'bg-green-100 text-green-800 dark:bg-green-700 dark:text-green-100';
    case 'Cancelled':
      return 'bg-red-100 text-red-800 dark:bg-red-700 dark:text-red-100';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-600 dark:text-gray-200';
  }
};

</script>

<style scoped>
/* Add any component-specific styles here if needed */
/* Tailwind utility classes should cover most styling */
.container {
  max-width: 1280px; /* Or your preferred max width */
}
</style>

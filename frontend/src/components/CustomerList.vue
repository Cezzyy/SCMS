<script setup lang="ts">
import { ref, onMounted, computed, defineAsyncComponent } from 'vue';
import { useCustomerStore } from '../stores/customerStore';
import { useContactStore } from '../stores/contactStore';
import type { Customer, CustomerUpdate } from '../types/Customer';
import type { Contact, ContactUpdate } from '../types/Contact';

const customerForm = defineAsyncComponent(() => import('./CustomerForm.vue'));
const editModal = defineAsyncComponent(() => import('./EditModal.vue'));

// Initialize stores
const customerStore = useCustomerStore();
const contactStore = useContactStore();

// Search state
const searchQuery = ref('');
const selectedCustomer = ref<Customer | null>(null);
const initialLoading = ref(true);

// Modal state
const showCustomerModal = ref(false);
const showContactModal = ref(false);
const showEditModal = ref(false);
const editingCustomer = ref<Customer | null>(null);
const editingContact = ref<Contact | null>(null);
const selectedContacts = ref<Contact[]>([]);
const contactsLoading = ref(false);
const selectedCustomerForModal = ref<Customer | null>(null);
const editModalMode = ref<'customer' | 'contact'>('customer');

// Pagination
const currentPage = ref(1);
const itemsPerPage = 10;

// Get total contacts count
const totalContacts = ref(0);

// Load customers on component mount
onMounted(async () => {
  try {
    await customerStore.fetchCustomers();
    // Count total contacts
    countTotalContacts();
  } catch (error) {
    console.error('Failed to load customers:', error);
  } finally {
    initialLoading.value = false;
  }
});

// Count total contacts across all customers
async function countTotalContacts() {
  try {
    let count = 0;
    for (const customer of customerStore.customers.slice(0, 5)) { // Limit to first 5 for performance
      const contacts = await contactStore.fetchContactsByCustomer(customer.customer_id);
      count += contacts.length;
    }
    totalContacts.value = count;
  } catch (error) {
    console.error('Failed to count contacts:', error);
  }
}

// Computed property for filtered customers
const filteredCustomers = computed(() => {
  if (!searchQuery.value) return customerStore.customers;

  const query = searchQuery.value.toLowerCase();
  return customerStore.customers.filter(customer =>
    customer.company_name.toLowerCase().includes(query) ||
    (customer.industry && customer.industry.toLowerCase().includes(query)) ||
    (customer.email && customer.email.toLowerCase().includes(query))
  );
});

// Paginated customers
const paginatedCustomers = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  return filteredCustomers.value.slice(startIndex, endIndex);
});

// Total pages for pagination
const totalPages = computed(() => {
  return Math.ceil(filteredCustomers.value.length / itemsPerPage);
});

// Page navigation
const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
  }
};

// Search handler
async function handleSearch() {
  currentPage.value = 1;
  try {
    await customerStore.fetchCustomers(searchQuery.value);
  } catch (error) {
    console.error('Search failed:', error);
  }
}

// Customer modal handlers
function openAddCustomerModal() {
  editingCustomer.value = null;
  showCustomerModal.value = true;
}

function openEditCustomerModal(customer: Customer) {
  editingCustomer.value = customer;
  editModalMode.value = 'customer';
  showEditModal.value = true;
}

function openEditContactModal(contact: Contact) {
  editingContact.value = contact;
  editModalMode.value = 'contact';
  showEditModal.value = true;
}

// Handle customer save from modal
async function handleSaveCustomer(customer: Customer, contact: Contact) {
  try {
    if (editingCustomer.value) {
      // This is an update
      await customerStore.updateCustomer({
        ...customer,
        customer_id: editingCustomer.value.customer_id
      });
    } else {
      // This is a create, already handled by the form
      // The form updates the customer with contact info
      await customerStore.fetchCustomers(); // Refresh the list to get the new customer
    }
    
    showCustomerModal.value = false;
    countTotalContacts(); // Update contact count
  } catch (error) {
    console.error('Failed to save customer:', error);
  }
}

// Handle edit modal save
async function handleEditSave(data: CustomerUpdate | ContactUpdate) {
  try {
    if (editModalMode.value === 'customer') {
      await customerStore.updateCustomer(data as CustomerUpdate);
      await customerStore.fetchCustomers(); // Refresh the list
    } else {
      if (!selectedCustomerForModal.value) throw new Error('No customer selected');
      await contactStore.updateContact(selectedCustomerForModal.value.customer_id, data as ContactUpdate);
      const contacts = await contactStore.fetchContactsByCustomer(selectedCustomerForModal.value.customer_id);
      selectedContacts.value = contacts;
      countTotalContacts(); // Update contact count
    }
    showEditModal.value = false;
  } catch (error) {
    console.error('Failed to save:', error);
  }
}

// View contacts modal handler
async function viewCustomerContacts(customer: Customer) {
  selectedCustomerForModal.value = customer;
  contactsLoading.value = true;
  showContactModal.value = true;
  
  try {
    const contacts = await contactStore.fetchContactsByCustomer(customer.customer_id);
    selectedContacts.value = contacts;
  } catch (error) {
    console.error('Failed to load contacts:', error);
  } finally {
    contactsLoading.value = false;
  }
}
</script>

<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 min-h-[600px] flex flex-col">
    <!-- Header with search and actions -->
    <div class="flex flex-col md:flex-row md:justify-between md:items-center p-4 md:p-6 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-gray-800 dark:to-gray-900">
      <div class="mb-4 md:mb-0">
        <h2 class="text-xl md:text-2xl font-bold text-gray-800 dark:text-white mb-1">Customer Management</h2>
        <p class="text-gray-600 dark:text-gray-300 text-xs md:text-sm">Manage your customers and their contacts</p>
      </div>
      <div class="flex flex-wrap gap-3 sm:gap-4">
        <button
          @click="openAddCustomerModal"
          class="bg-blue-600 text-white px-3 py-2 md:px-4 md:py-2 rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center shadow-sm text-sm md:text-base"
        >
          <font-awesome-icon icon="plus" class="mr-1.5 h-4 w-4" />
          Add Customer
        </button>
      </div>
    </div>

    <!-- Stats summary cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 p-4 bg-gray-50 dark:bg-gray-900">
      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 mr-3 md:mr-4">
            <font-awesome-icon icon="users" class="h-5 w-5 md:h-6 md:w-6" />
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Total Customers</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ customerStore.customers.length }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-300 mr-3 md:mr-4">
            <font-awesome-icon icon="address-book" class="h-5 w-5 md:h-6 md:w-6" />
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Total Contacts</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ totalContacts }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Enhanced search bar -->
    <div class="p-3 md:p-4 border-b border-gray-200 dark:border-gray-700">
      <div class="relative">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <font-awesome-icon icon="search" class="h-4 w-4 md:h-5 md:w-5 text-gray-400" />
        </div>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by company, email, or industry..."
          class="w-full pl-8 md:pl-10 px-3 md:px-4 py-2 md:py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg
                 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                 transition-all duration-200
                 dark:bg-gray-700 dark:text-white text-xs md:text-sm"
          @keyup.enter="handleSearch"
        />
      </div>
    </div>

    <!-- Customer Table -->
    <div class="p-0 sm:p-4 flex-grow">
      <!-- Initial Loading -->
      <div v-if="initialLoading" class="flex justify-center items-center p-12 min-h-[400px]">
        <div class="relative">
          <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-200 dark:border-gray-600"></div>
          <div class="animate-spin rounded-full h-12 w-12 border-t-4 border-blue-600 dark:border-blue-500 absolute top-0 left-0"></div>
        </div>
        <div class="ml-4 text-gray-600 dark:text-gray-300 text-sm font-medium">Loading data...</div>
      </div>

      <!-- API Loading -->
      <div v-else-if="customerStore.loading" class="flex justify-center py-8 min-h-[400px]">
        <div class="animate-spin rounded-full h-6 w-6 sm:h-8 sm:w-8 border-b-2 border-blue-600"></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredCustomers.length === 0" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700 min-h-[500px]">
        <font-awesome-icon icon="users" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" />
        <p class="text-gray-600 dark:text-gray-300 mb-2">No customers found</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Try adjusting your search or add a new customer to get started</p>
        <button
          @click="openAddCustomerModal"
          class="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center">
          <font-awesome-icon icon="plus" class="mr-1.5 h-4 w-4" />
          Create First Customer
        </button>
      </div>

      <!-- Customer Table (only shown when there are customers) -->
      <div v-else class="overflow-x-auto w-full min-h-[400px]">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 shadow-sm">
          <thead class="bg-gray-50 dark:bg-gray-700">
            <tr>
              <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-1/3">
                <div class="flex items-center">
                  Company
                  <font-awesome-icon icon="sort" class="ml-1 h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity duration-200" />
                </div>
              </th>
              <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider hidden sm:table-cell w-1/4">
                Email
              </th>
              <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider hidden md:table-cell w-1/6">
                Industry
              </th>
              <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider hidden lg:table-cell w-1/6">
                Phone
              </th>
              <th class="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider sticky right-0 bg-gray-50 dark:bg-gray-700 z-10 w-1/6">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr
              v-for="customer in paginatedCustomers"
              :key="customer.customer_id"
              class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150"
            >
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="font-medium text-gray-900 dark:text-white truncate max-w-[200px]">{{ customer.company_name }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-300 hidden sm:table-cell">
                <div class="truncate max-w-[200px]">{{ customer.email || 'N/A' }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-300 hidden md:table-cell">
                {{ customer.industry || 'N/A' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-gray-500 dark:text-gray-300 hidden lg:table-cell">
                {{ customer.phone || 'N/A' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium sticky right-0 bg-white dark:bg-gray-800 shadow-sticky">
                <button
                  @click="viewCustomerContacts(customer)"
                  class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 mr-3 transition-colors duration-200"
                  title="View contacts"
                >
                  <font-awesome-icon icon="address-book" class="h-5 w-5" />
                </button>
                <button
                  @click="openEditCustomerModal(customer)"
                  class="text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-300 mr-3 transition-colors duration-200"
                  title="Edit customer"
                >
                  <font-awesome-icon icon="edit" class="h-5 w-5" />
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
              <font-awesome-icon icon="chevron-left" class="h-5 w-5" />
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
              <font-awesome-icon icon="chevron-right" class="h-5 w-5" />
            </button>
          </nav>
        </div>
      </div>
    </div>

    <!-- Customer Form Modal -->
    <customerForm
      :show="showCustomerModal"
      :customer="editingCustomer"
      :is-edit="!!editingCustomer"
      @close="showCustomerModal = false"
      @save="handleSaveCustomer"
    />

    <!-- Edit Modal -->
    <editModal
      :show="showEditModal"
      :customer="editingCustomer"
      :contact="editingContact"
      :mode="editModalMode"
      @close="showEditModal = false"
      @save="handleEditSave"
      class="z-50"
    />

    <!-- Contact Modal -->
    <div v-if="showContactModal" class="fixed inset-0 z-40 overflow-y-auto">
      <div class="flex min-h-screen items-center justify-center px-2 pt-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="showContactModal = false" />

        <!-- Modal panel -->
        <div class="fixed inset-x-0 top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-full max-w-[95%] sm:max-w-lg md:max-w-xl lg:max-w-2xl overflow-hidden rounded-lg bg-white dark:bg-gray-800 text-left shadow-xl">
          <!-- Header -->
          <div class="bg-gradient-to-r from-blue-600 to-blue-800 px-3 sm:px-6 py-3 sm:py-4">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-base font-medium leading-6 text-white truncate max-w-[200px] sm:max-w-full">
                  Contacts for {{ selectedCustomerForModal?.company_name }}
                </h3>
                <p class="mt-0.5 text-xs text-blue-100 hidden sm:block">
                  View all contacts for this customer
                </p>
              </div>
              <button
                @click="showContactModal = false"
                class="rounded-md bg-blue-700 p-1.5 text-blue-200 hover:text-white focus:outline-none"
              >
                <span class="sr-only">Close</span>
                <font-awesome-icon icon="times" class="h-5 w-5" />
              </button>
            </div>
          </div>

          <!-- Contact details -->
          <div class="bg-white dark:bg-gray-800 px-3 sm:px-6 py-3 sm:py-4">
            <div v-if="contactsLoading" class="flex justify-center py-8">
              <div class="animate-spin rounded-full h-6 w-6 sm:h-8 sm:w-8 border-b-2 border-accent2"></div>
            </div>

            <div v-else-if="selectedContacts.length === 0" class="flex flex-col items-center justify-center py-6 sm:py-8 px-4">
              <font-awesome-icon icon="users" class="h-10 w-10 sm:h-12 sm:w-12 text-gray-300 dark:text-gray-600 mb-3" />
              <p class="text-sm text-gray-500 dark:text-gray-400 text-center">No contacts found for this customer</p>
            </div>

            <div v-else class="overflow-x-auto">
              <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead class="bg-gray-50 dark:bg-gray-700">
                  <tr>
                    <th class="px-3 py-2 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-1/3">Name</th>
                    <th class="px-3 py-2 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider hidden sm:table-cell w-1/3">Email</th>
                    <th class="px-3 py-2 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider hidden md:table-cell w-1/6">Position</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider sticky right-0 bg-gray-50 dark:bg-gray-700 z-10 w-1/6">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white dark:bg-gray-800">
                  <tr v-for="contact in selectedContacts" :key="contact.contact_id" class="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td class="px-3 py-2 text-xs text-gray-900 dark:text-gray-200 truncate">
                      {{ contact.first_name }} {{ contact.last_name }}
                      <div class="text-xs text-gray-500 mt-0.5 sm:hidden">
                        {{ contact.position || '' }}
                      </div>
                    </td>
                    <td class="px-3 py-2 text-xs text-gray-600 dark:text-gray-300 hidden sm:table-cell truncate">
                      <a
                        v-if="contact.email"
                        :href="`mailto:${contact.email}`"
                        class="text-blue-600 dark:text-blue-400 hover:underline"
                      >
                        {{ contact.email }}
                      </a>
                      <span v-else>-</span>
                    </td>
                    <td class="px-3 py-2 text-xs text-gray-600 dark:text-gray-300 hidden md:table-cell truncate">
                      {{ contact.position || '-' }}
                    </td>
                    <td class="px-3 py-2 text-xs text-gray-600 dark:text-gray-300 text-right sticky right-0 bg-white dark:bg-gray-800 shadow-sticky">
                      <button
                        @click="openEditContactModal(contact)"
                        class="text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-300"
                        title="Edit contact"
                      >
                        <font-awesome-icon icon="edit" class="h-4 w-4" />
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Footer -->
          <div class="bg-gray-50 dark:bg-gray-700 px-3 sm:px-6 py-3 sm:py-4">
            <div class="flex justify-end">
              <button
                type="button"
                class="rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 px-3 py-1.5 text-sm font-medium text-gray-700 dark:text-gray-300 shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                @click="showContactModal = false"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Transition animations */
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

.modal-enter-to,
.modal-leave-from {
  opacity: 1;
  transform: scale(1);
}

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
  
  /* Optimize column widths */
  th:first-child, td:first-child {
    width: 60%;
  }
  
  th:last-child, td:last-child {
    width: 40%;
  }
}

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

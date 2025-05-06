<script setup lang="ts">
import { ref, onMounted, computed, defineAsyncComponent } from 'vue';
import { useCustomerStore } from '@/stores/customerStore';
import { useContactStore } from '@/stores/contactStore';
import type { Customer, CustomerUpdate } from '@/types/Customer';
import type { Contact, ContactUpdate } from '@/types/Contact';

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

// Load customers on component mount
onMounted(async () => {
  try {
    await customerStore.fetchCustomers();
  } catch (error) {
    console.error('Failed to load customers:', error);
  } finally {
    initialLoading.value = false;
  }
});

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

// Search handler
async function handleSearch() {
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
  <div class="bg-white rounded-lg shadow">
    <!-- Header and Search -->
    <div class="p-6 border-b border-gray-200">
      <div class="flex justify-between items-center">
        <h2 class="text-2xl font-semibold text-primary">Customers</h2>
        <div class="flex items-center space-x-4">
          <!-- Search Bar -->
          <div class="relative">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search customers..."
              class="w-64 pl-10 pr-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-accent1"
              @keyup.enter="handleSearch"
            />
            <div class="absolute left-3 top-2.5 text-gray-400">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </div>

          <!-- Add Customer Button -->
          <button
            @click="openAddCustomerModal"
            class="bg-accent1 hover:bg-accent2 text-white px-4 py-2 rounded-lg flex items-center"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            Add Customer
          </button>
        </div>
      </div>
    </div>

    <!-- Customer Table -->
    <div class="p-6">
      <!-- Initial Loading -->
      <div v-if="initialLoading" class="flex justify-center py-16">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-accent1"></div>
      </div>

      <!-- API Loading -->
      <div v-else-if="customerStore.loading" class="flex justify-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-accent1"></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="filteredCustomers.length === 0" class="flex flex-col items-center justify-center py-12 px-4">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h-3.5c.83 0 1.5-.67 1.5-1.5s-.67-1.5-1.5-1.5H9c-.83 0-1.5.67-1.5 1.5S8.17 21 9 21h10zm-3-4h.01M12 14h.01M9 14h.01" />
        </svg>
        <p class="text-lg font-medium text-gray-500">No customers found</p>
        <p class="text-gray-400 mt-1">Try adjusting your search or add a new customer to get started</p>
      </div>

      <!-- Customer Table (only shown when there are customers) -->
      <div v-else class="overflow-x-auto">
        <table class="min-w-full">
          <thead class="bg-bg-alt">
            <tr>
              <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Company</th>
              <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Industry</th>
              <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Email</th>
              <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Phone</th>
              <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Website</th>
              <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr
              v-for="customer in filteredCustomers"
              :key="customer.customer_id"
              class="hover:bg-gray-50"
            >
              <td class="px-4 py-3 text-sm text-gray-900">{{ customer.company_name }}</td>
              <td class="px-4 py-3 text-sm text-gray-600">{{ customer.industry || '-' }}</td>
              <td class="px-4 py-3 text-sm text-gray-600">{{ customer.email || '-' }}</td>
              <td class="px-4 py-3 text-sm text-gray-600">{{ customer.phone || '-' }}</td>
              <td class="px-4 py-3 text-sm text-gray-600">
                <a
                  v-if="customer.website"
                  :href="customer.website"
                  target="_blank"
                  class="text-accent2 hover:underline"
                  @click.stop
                >
                  {{ customer.website }}
                </a>
                <span v-else>-</span>
              </td>
              <td class="px-4 py-3 text-sm text-gray-600">
                <div class="flex space-x-2">
                  <button 
                    @click="viewCustomerContacts(customer)"
                    class="text-blue-600 hover:text-blue-800 font-medium"
                  >
                    View Contacts
                  </button>
                  <span class="text-gray-300">|</span>
                  <button 
                    @click="openEditCustomerModal(customer)"
                    class="text-green-600 hover:text-green-800 font-medium"
                  >
                    Edit
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
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
      <div class="flex min-h-screen items-center justify-center px-4 pt-4 pb-20 text-center sm:block sm:p-0">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="showContactModal = false" />

        <!-- Modal panel -->
        <div class="fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-full max-w-2xl overflow-hidden rounded-lg bg-white text-left shadow-xl">
          <!-- Header -->
          <div class="bg-gradient-to-r from-blue-600 to-blue-800 px-6 py-4">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-lg font-medium leading-6 text-white">
                  Contacts for {{ selectedCustomerForModal?.company_name }}
                </h3>
                <p class="mt-1 text-sm text-blue-100">
                  View all contacts for this customer
                </p>
              </div>
              <button
                @click="showContactModal = false"
                class="rounded-md bg-blue-700 p-2 text-blue-200 hover:text-white focus:outline-none"
              >
                <span class="sr-only">Close</span>
                <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>

          <!-- Contact details -->
          <div class="bg-white px-6 py-4">
            <div v-if="contactsLoading" class="flex justify-center py-8">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-accent2"></div>
            </div>

            <div v-else-if="selectedContacts.length === 0" class="flex flex-col items-center justify-center py-8 px-4">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-gray-300 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
              <p class="text-gray-500">No contacts found for this customer</p>
            </div>

            <div v-else class="overflow-x-auto">
              <table class="min-w-full">
                <thead class="bg-white">
                  <tr>
                    <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Name</th>
                    <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Position</th>
                    <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Email</th>
                    <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Phone</th>
                    <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Actions</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 bg-white">
                  <tr v-for="contact in selectedContacts" :key="contact.contact_id">
                    <td class="px-4 py-3 text-sm text-gray-900">
                      {{ contact.first_name }} {{ contact.last_name }}
                    </td>
                    <td class="px-4 py-3 text-sm text-gray-600">{{ contact.position || '-' }}</td>
                    <td class="px-4 py-3 text-sm text-gray-600">
                      <a
                        v-if="contact.email"
                        :href="`mailto:${contact.email}`"
                        class="text-accent2 hover:underline"
                      >
                        {{ contact.email }}
                      </a>
                      <span v-else>-</span>
                    </td>
                    <td class="px-4 py-3 text-sm text-gray-600">{{ contact.phone || '-' }}</td>
                    <td class="px-4 py-3 text-sm text-gray-600">
                      <button
                        @click="openEditContactModal(contact)"
                        class="text-blue-600 hover:text-blue-800 font-medium"
                      >
                        Edit
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Footer -->
          <div class="bg-gray-50 px-6 py-4">
            <div class="flex justify-end">
              <button
                type="button"
                class="rounded-md border border-gray-300 bg-white px-4 py-2 text-base font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
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
</style> 

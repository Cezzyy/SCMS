<script setup lang="ts">
import { ref, onMounted, computed, defineAsyncComponent } from 'vue';
import type { User } from '../types/User';

const UserModal = defineAsyncComponent(() => import('../components/UserModal.vue'));
const ViewModal = defineAsyncComponent(() => import('../components/ViewModal.vue'));
const ConfirmationModal = defineAsyncComponent(() => import('../components/ConfirmationModal.vue'));

// State
const isLoading = ref(false);
const showUserModal = ref(false);
const showViewModal = ref(false);
const showConfirmModal = ref(false);
const userToEdit = ref<User | null>(null);
const userToView = ref<User | null>(null);
const userToDelete = ref<User | null>(null);
const searchQuery = ref('');

// Pagination
const currentPage = ref(1);
const itemsPerPage = 10;

// Mock data - would normally come from an API
const users = ref<User[]>([
  {
    user_id: 1,
    username: 'admin',
    role: 'admin',
    first_name: 'Admin',
    last_name: 'User',
    email: 'admin@example.com',
    phone: '123-456-7890',
    department: 'IT',
    position: 'System Administrator',
    last_login: new Date('2023-09-30T10:30:00Z'),
    created_at: new Date('2023-01-01T00:00:00Z'),
    updated_at: new Date('2023-09-30T10:30:00Z')
  },
  {
    user_id: 2,
    username: 'jsmith',
    role: 'manager',
    first_name: 'John',
    last_name: 'Smith',
    email: 'john.smith@example.com',
    phone: '123-456-7891',
    department: 'Sales',
    position: 'Sales Manager',
    last_login: new Date('2023-09-29T14:20:00Z'),
    created_at: new Date('2023-01-15T00:00:00Z'),
    updated_at: new Date('2023-09-29T14:20:00Z')
  },
  {
    user_id: 3,
    username: 'mjohnson',
    role: 'staff',
    first_name: 'Mary',
    last_name: 'Johnson',
    email: 'mary.johnson@example.com',
    phone: '123-456-7892',
    department: 'Customer Service',
    position: 'Customer Support Representative',
    last_login: new Date('2023-09-28T09:45:00Z'),
    created_at: new Date('2023-02-01T00:00:00Z'),
    updated_at: new Date('2023-09-28T09:45:00Z')
  },
  {
    user_id: 4,
    username: 'rwilliams',
    role: 'staff',
    first_name: 'Robert',
    last_name: 'Williams',
    email: 'robert.williams@example.com',
    phone: '123-456-7893',
    department: 'Warehouse',
    position: 'Inventory Specialist',
    last_login: new Date('2023-09-27T11:15:00Z'),
    created_at: new Date('2023-02-15T00:00:00Z'),
    updated_at: new Date('2023-09-27T11:15:00Z')
  },
  {
    user_id: 5,
    username: 'jbrown',
    role: 'manager',
    first_name: 'Jennifer',
    last_name: 'Brown',
    email: 'jennifer.brown@example.com',
    phone: '123-456-7894',
    department: 'Finance',
    position: 'Financial Analyst',
    last_login: new Date('2023-09-26T13:40:00Z'),
    created_at: new Date('2023-03-01T00:00:00Z'),
    updated_at: new Date('2023-09-26T13:40:00Z')
  }
]);

// Filtered users based on search
const filteredUsers = computed(() => {
  if (!searchQuery.value) return users.value;

  const query = searchQuery.value.toLowerCase();
  return users.value.filter(user =>
    user.username.toLowerCase().includes(query) ||
    user.first_name.toLowerCase().includes(query) ||
    user.last_name.toLowerCase().includes(query) ||
    user.email.toLowerCase().includes(query) ||
    user.department?.toLowerCase().includes(query) ||
    user.position?.toLowerCase().includes(query) ||
    user.role.toLowerCase().includes(query)
  );
});

// Paginated users
const paginatedUsers = computed(() => {
  const startIndex = (currentPage.value - 1) * itemsPerPage;
  const endIndex = startIndex + itemsPerPage;
  return filteredUsers.value.slice(startIndex, endIndex);
});

// Total pages for pagination
const totalPages = computed(() => {
  return Math.ceil(filteredUsers.value.length / itemsPerPage);
});

// Format date for display
const formatDate = (date: Date | undefined): string => {
  if (!date) return 'Never';
  return new Date(date).toLocaleString();
};

// Page navigation
const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page;
  }
};

// CRUD operations
const openUserModal = (user: User | null = null) => {
  userToEdit.value = user ? { ...user } : {
    user_id: 0,
    username: '',
    role: 'Sales Staff',
    first_name: '',
    last_name: '',
    email: '',
    phone: '',
    department: '',
    position: '',
    created_at: new Date(),
    updated_at: new Date()
  } as User;
  showUserModal.value = true;
};

const openViewModal = (user: User) => {
  userToView.value = { ...user };
  showViewModal.value = true;
};

const saveUser = (user: User) => {
  // In a real app, this would be an API call
  if (user.user_id) {
    // Update existing user
    const index = users.value.findIndex(u => u.user_id === user.user_id);
    if (index !== -1) {
      users.value[index] = { ...user, updated_at: new Date() };
    }
  } else {
    // Create new user
    const maxId = Math.max(0, ...users.value.map(u => u.user_id));
    const newUser = {
      ...user,
      user_id: maxId + 1,
      created_at: new Date(),
      updated_at: new Date()
    };
    users.value.push(newUser);
  }
  showUserModal.value = false;
};

const confirmDelete = (user: User) => {
  userToDelete.value = user;
  showConfirmModal.value = true;
};

const deleteUser = () => {
  if (!userToDelete.value) return;

  // In a real app, this would be an API call
  users.value = users.value.filter(u => u.user_id !== userToDelete.value?.user_id);

  // Close modal
  showConfirmModal.value = false;
  userToDelete.value = null;

  // Check if we need to adjust current page after deletion
  if (paginatedUsers.value.length === 0 && currentPage.value > 1) {
    currentPage.value--;
  }
};

// Load data (in a real app, this would fetch from an API)
const loadData = async () => {
  isLoading.value = true;
  try {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500));
    // In a real app, you would fetch data here
  } catch (error) {
    console.error('Error loading user data:', error);
  } finally {
    isLoading.value = false;
  }
};

// Load data on component mount
onMounted(loadData);
</script>

<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 overflow-hidden h-full flex flex-col">
    <!-- Header with tabs and actions -->
    <div class="flex flex-col md:flex-row md:justify-between md:items-center p-4 md:p-6 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-gray-800 dark:to-gray-900 flex-shrink-0">
      <div class="mb-4 md:mb-0">
        <h2 class="text-xl md:text-2xl font-bold text-gray-800 dark:text-white mb-1">User Management</h2>
        <p class="text-gray-600 dark:text-gray-300 text-xs md:text-sm">Manage system users and access control</p>
      </div>
      <div class="flex flex-wrap gap-3 sm:gap-4">
        <button
          @click="openUserModal()"
          class="bg-blue-600 text-white px-3 py-2 md:px-4 md:py-2 rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center shadow-sm text-sm md:text-base"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Add User
        </button>
      </div>
    </div>

    <!-- Stats summary cards - non-scrollable -->
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3 p-4 bg-gray-50 dark:bg-gray-900 flex-shrink-0">
      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Total Users</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ users.length }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Admin Users</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ users.filter(user => user.role === 'admin').length }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white dark:bg-gray-800 p-3 md:p-4 rounded-lg shadow-sm border border-gray-100 dark:border-gray-700 sm:col-span-2 md:col-span-1">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-indigo-100 dark:bg-indigo-900 text-indigo-600 dark:text-indigo-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-500 dark:text-gray-400">Staff Users</div>
            <div class="text-lg md:text-2xl font-semibold text-gray-800 dark:text-white">
              {{ users.filter(user => user.role === 'staff').length }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Enhanced search bar - non-scrollable -->
    <div class="p-3 md:p-4 border-b border-gray-200 dark:border-gray-700 flex-shrink-0">
      <div class="relative">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <svg class="h-4 w-4 md:h-5 md:w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
          </svg>
        </div>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search by name, email, department, role..."
          class="w-full pl-8 md:pl-10 px-3 md:px-4 py-2 md:py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg
                 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                 transition-all duration-200
                 dark:bg-gray-700 dark:text-white text-xs md:text-sm"
        />
      </div>
    </div>

    <!-- Content area - scrollable -->
    <div class="flex-1 overflow-auto">
      <!-- Improved loading indicator -->
      <div v-if="isLoading" class="flex justify-center items-center p-12">
        <div class="relative">
          <div class="animate-spin rounded-full h-12 w-12 border-4 border-gray-200 dark:border-gray-600"></div>
          <div class="animate-spin rounded-full h-12 w-12 border-t-4 border-blue-600 dark:border-blue-500 absolute top-0 left-0"></div>
        </div>
        <div class="ml-4 text-gray-600 dark:text-gray-300 text-sm font-medium">Loading data...</div>
      </div>

      <!-- Users List -->
      <div v-else-if="filteredUsers.length === 0 && searchQuery" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <p class="text-gray-600 dark:text-gray-300 mb-2">No results found</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Try adjusting your search criteria</p>
      </div>

      <div v-else-if="!users.length" class="flex flex-col items-center justify-center py-12 rounded-lg bg-gray-50 dark:bg-gray-700">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-gray-300 dark:text-gray-600 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <p class="text-gray-600 dark:text-gray-300 mb-2">No users found</p>
        <p class="text-gray-500 dark:text-gray-400 text-sm mb-4">Add some users to get started</p>
        <button
          @click="openUserModal()"
          class="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 transition-colors duration-200 flex items-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
          </svg>
          Create First User
        </button>
      </div>

      <div v-else class="overflow-x-auto p-4 w-full">
        <div class="max-w-full overflow-hidden">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 shadow-sm table-fixed">
            <thead class="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-1/6">
                  Name
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-1/8">
                  Username
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-1/6">
                  Email
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-1/10">
                  Role
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-1/6">
                  Department
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-1/6">
                  Last Login
                </th>
                <th class="px-6 py-3.5 text-right text-xs font-semibold text-gray-500 dark:text-gray-300 uppercase tracking-wider w-24 sticky right-0 bg-gray-50 dark:bg-gray-700 z-10 shadow-sticky-l">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-for="user in paginatedUsers" :key="user.user_id"
                  class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors duration-150">
                <td class="px-6 py-4 whitespace-nowrap">
                  <div class="flex items-center">
                    <div class="h-8 w-8 rounded-full bg-indigo-100 dark:bg-indigo-900 flex items-center justify-center text-indigo-600 dark:text-indigo-300 flex-shrink-0">
                      {{ user.first_name.charAt(0) }}{{ user.last_name.charAt(0) }}
                    </div>
                    <div class="ml-3 min-w-0">
                      <div class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ user.first_name }} {{ user.last_name }}</div>
                      <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
                        {{ user.position || 'N/A' }}
                      </div>
                    </div>
                  </div>
                </td>
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  <div class="truncate max-w-full">{{ user.username }}</div>
                </td>
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  <div class="truncate max-w-full" :title="user.email">{{ user.email }}</div>
                </td>
                <td class="px-6 py-4">
                  <span class="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-full"
                        :class="{
                          'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-100': user.role === 'admin',
                          'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100': user.role === 'manager',
                          'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-100': user.role === 'staff'
                        }">
                    {{ user.role.charAt(0).toUpperCase() + user.role.slice(1) }}
                  </span>
                </td>
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  <div class="truncate max-w-full">{{ user.department || 'N/A' }}</div>
                </td>
                <td class="px-6 py-4 text-sm text-gray-500 dark:text-gray-400">
                  <div class="truncate max-w-full">{{ formatDate(user.last_login) }}</div>
                </td>
                <td class="px-6 py-4 text-right text-sm font-medium sticky right-0 bg-white dark:bg-gray-800 shadow-sticky-l">
                  <div class="flex justify-end space-x-2">
                    <button
                      @click="openViewModal(user)"
                      class="text-blue-600 hover:text-blue-900 dark:text-blue-400 dark:hover:text-blue-300 transition-colors duration-200"
                      title="View user details"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path d="M10 12a2 2 0 100-4 2 2 0 000 4z" />
                        <path fill-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
                      </svg>
                    </button>
                    <button
                      @click="openUserModal(user)"
                      class="text-indigo-600 hover:text-indigo-900 dark:text-indigo-400 dark:hover:text-indigo-300 transition-colors duration-200"
                      title="Edit user"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                      </svg>
                    </button>
                    <button
                      @click="confirmDelete(user)"
                      class="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300 transition-colors duration-200"
                      title="Delete user"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                        <path fill-rule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

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

    <!-- User Modal Component -->
    <UserModal
      v-if="showUserModal"
      :show="showUserModal"
      @update:show="showUserModal = $event"
      :user="userToEdit"
      @save="saveUser"
    />

    <!-- View Modal Component -->
    <ViewModal
      v-if="showViewModal"
      :show="showViewModal"
      @update:show="showViewModal = $event"
      :item="userToView"
      type="user"
      @edit-user="openUserModal"
    />

    <!-- Confirmation Modal for Delete -->
    <ConfirmationModal
      v-if="showConfirmModal"
      :show="showConfirmModal"
      @update:show="showConfirmModal = $event"
      title="Delete User"
      :message="`Are you sure you want to delete ${userToDelete?.first_name} ${userToDelete?.last_name}? This action cannot be undone.`"
      confirmButtonText="Delete"
      @confirm="deleteUser"
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
    max-width: 100%;
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
.shadow-sticky-l {
  box-shadow: -4px 0 5px -2px rgba(0, 0, 0, 0.15);
}

/* Ensure proper sticky column behavior */
.sticky {
  position: sticky;
  z-index: 2;
  right: 0;
}

/* Proper hover states on sticky columns */
tr:hover .sticky {
  background-color: rgba(249, 250, 251, 1); /* bg-gray-50 */
}

tr:hover .dark .sticky {
  background-color: rgba(55, 65, 81, 1); /* dark:bg-gray-700 */
}

/* Fix for text overflow */
td, th {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

/* Responsive container for mobile */
@media (max-width: 768px) {
  .bg-white.dark\:bg-gray-800 {
    overflow-x: hidden;
    max-width: 100%;
  }
}

/* Allow multiline wrap for specific cells if needed */
.can-wrap {
  white-space: normal;
  word-break: break-word;
}

/* Ensure proper column widths */
.table-fixed {
  table-layout: fixed;
}

/* Tooltip style for truncated text */
[title] {
  position: relative;
  cursor: help;
}

/* Responsive container for mobile */
@media (max-width: 768px) {
  td, th {
    max-width: 150px;
  }
}

@media (max-width: 640px) {
  td, th {
    max-width: 100px;
  }
}
</style>

<script setup lang="ts">
import { ref, onMounted, computed, defineAsyncComponent } from 'vue';
import type { User } from '../types/User';
import { useUserStore } from '../stores/userStore';

const UserModal = defineAsyncComponent(() => import('../components/UserModal.vue'));
const ViewModal = defineAsyncComponent(() => import('../components/ViewModal.vue'));
const ConfirmationModal = defineAsyncComponent(() => import('../components/ConfirmationModal.vue'));

// Initialize user store
const userStore = useUserStore();

// State
const showUserModal = ref(false);
const showViewModal = ref(false);
const showConfirmModal = ref(false);
const userToEdit = ref<User | null>(null);
const userToView = ref<User | null>(null);
const userToDelete = ref<User | null>(null);
const searchQuery = ref('');
const selectedRole = ref('all');

// Pagination
const currentPage = ref(1);
const itemsPerPage = 10;

// Computed props
const isLoading = computed(() => userStore.loading);
const error = computed(() => userStore.error);
const users = computed(() => userStore.users);

// Role counts
const roleCounts = computed(() => {
  const counts = {
    admin: users.value.filter(user => user.role.toLowerCase().includes('admin')).length,
    manager: users.value.filter(user => user.role.toLowerCase().includes('manager')).length,
    sales: users.value.filter(user => user.role.toLowerCase().includes('sales')).length,
    inventory: users.value.filter(user => user.role.toLowerCase().includes('inventory')).length
  };
  return counts;
});

// Filtered users based on search and role filter
const filteredUsers = computed(() => {
  let filtered = users.value;
  
  // Apply search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(user =>
      user.first_name.toLowerCase().includes(query) ||
      user.last_name.toLowerCase().includes(query) ||
      user.email.toLowerCase().includes(query) ||
      user.department?.toLowerCase().includes(query) ||
      user.position?.toLowerCase().includes(query) ||
      user.role.toLowerCase().includes(query)
    );
  }
  
  // Apply role filter
  if (selectedRole.value !== 'all') {
    filtered = filtered.filter(user => 
      user.role.toLowerCase().includes(selectedRole.value.toLowerCase())
    );
  }
  
  return filtered;
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

const saveUser = async (user: User) => {
  try {
    if (user.user_id) {
      // Update existing user
      await userStore.updateUser(user.user_id, user);
    } else {
      // Create new user
      await userStore.createUser(user);
    }
    showUserModal.value = false;
  } catch (err) {
    console.error('Error saving user:', err);
  }
};

const confirmDelete = (user: User) => {
  userToDelete.value = user;
  showConfirmModal.value = true;
};

const deleteUser = async () => {
  if (!userToDelete.value) return;

  try {
    await userStore.deleteUser(userToDelete.value.user_id);
    
    // Close modal
    showConfirmModal.value = false;
    userToDelete.value = null;

    // Check if we need to adjust current page after deletion
    if (paginatedUsers.value.length === 0 && currentPage.value > 1) {
      currentPage.value--;
    }
  } catch (err) {
    console.error('Error deleting user:', err);
  }
};

// Load data on component mount
onMounted(async () => {
  try {
    await userStore.fetchUsers();
  } catch (err) {
    console.error('Error loading user data:', err);
  }
});
</script>

<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 overflow-hidden h-full flex flex-col">
    <!-- Header with tabs and actions -->
    <div class="flex flex-col md:flex-row md:justify-between md:items-center p-4 md:p-6 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-gray-700 to-gray-800 text-white flex-shrink-0">
      <div class="mb-4 md:mb-0">
        <h2 class="text-xl md:text-2xl font-bold mb-1">User Management</h2>
        <p class="text-xs md:text-sm text-gray-300">Manage system users and access control</p>
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
    <div class="grid grid-cols-1 sm:grid-cols-4 gap-3 p-4 bg-gray-800 flex-shrink-0">
      <div class="bg-gray-700 p-3 md:p-4 rounded-lg shadow-sm border border-gray-600">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-indigo-900 text-indigo-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-300">Admin Users</div>
            <div class="text-lg md:text-2xl font-semibold text-white">
              {{ roleCounts.admin }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-gray-700 p-3 md:p-4 rounded-lg shadow-sm border border-gray-600">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-blue-900 text-blue-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-300">Manager Users</div>
            <div class="text-lg md:text-2xl font-semibold text-white">
              {{ roleCounts.manager }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-gray-700 p-3 md:p-4 rounded-lg shadow-sm border border-gray-600">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-yellow-800 text-yellow-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-300">Sales Users</div>
            <div class="text-lg md:text-2xl font-semibold text-white">
              {{ roleCounts.sales }}
            </div>
          </div>
        </div>
      </div>

      <div class="bg-gray-700 p-3 md:p-4 rounded-lg shadow-sm border border-gray-600">
        <div class="flex items-center">
          <div class="p-2 md:p-3 rounded-full bg-purple-900 text-purple-300 mr-3 md:mr-4">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 md:h-6 md:w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
            </svg>
          </div>
          <div>
            <div class="text-xs md:text-sm font-medium text-gray-300">Inventory Users</div>
            <div class="text-lg md:text-2xl font-semibold text-white">
              {{ roleCounts.inventory }}
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Enhanced search bar with role filter - non-scrollable -->
    <div class="p-3 md:p-4 border-b border-gray-600 bg-gray-700 flex-shrink-0">
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="relative flex-grow">
          <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <svg class="h-4 w-4 md:h-5 md:w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
            </svg>
          </div>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search by name, email, department..."
            class="w-full pl-10 px-4 py-2.5 border border-gray-600 rounded-md
                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                   transition-all duration-200
                   bg-gray-600 text-white text-sm"
          />
        </div>
        
        <div class="sm:w-48">
          <select
            v-model="selectedRole"
            class="w-full px-4 py-2.5 border border-gray-600 rounded-md
                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                   transition-all duration-200
                   bg-gray-600 text-white text-sm appearance-none"
          >
            <option value="all">All Roles</option>
            <option value="admin">Admin</option>
            <option value="manager">Manager</option>
            <option value="sales">Sales</option>
            <option value="inventory">Inventory</option>
          </select>
        </div>
      </div>
      
      <!-- Display filtered count -->
      <div class="mt-2 text-sm text-gray-300">
        Showing {{ filteredUsers.length }} of {{ users.length }} users
      </div>
    </div>

    <!-- Content area - scrollable -->
    <div class="flex-1 overflow-auto bg-gray-900">
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
            <thead class="bg-gray-700">
              <tr>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-300 uppercase tracking-wider w-1/6">
                  Name
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-300 uppercase tracking-wider w-1/6">
                  Email
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-300 uppercase tracking-wider w-1/10">
                  Role
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-300 uppercase tracking-wider w-1/6">
                  Department
                </th>
                <th class="px-6 py-3.5 text-left text-xs font-semibold text-gray-300 uppercase tracking-wider w-1/6">
                  Last Login
                </th>
                <th class="px-6 py-3.5 text-right text-xs font-semibold text-gray-300 uppercase tracking-wider w-24 sticky right-0 bg-gray-700 z-10 shadow-sticky-l">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody class="bg-gray-800 divide-y divide-gray-700">
              <tr v-for="user in paginatedUsers" :key="user.user_id"
                  class="hover:bg-gray-700 transition-colors duration-150">
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
                  <div class="truncate max-w-full" :title="user.email">{{ user.email }}</div>
                </td>
                <td class="px-6 py-4">
                  <span class="px-2 py-1 inline-flex text-xs leading-5 font-semibold rounded-md"
                        :class="{
                          'bg-indigo-900 text-indigo-100': user.role === 'admin' || user.role === 'Administrator',
                          'bg-blue-900 text-blue-100': user.role === 'manager' || user.role === 'Branch Manager',
                          'bg-green-900 text-green-100': user.role === 'staff' || user.role === 'Sales Staff',
                          'bg-purple-900 text-purple-100': user.role === 'inventory manager' || user.role === 'Inventory Manager',
                          'bg-yellow-800 text-yellow-100': user.role.includes('sales') && !user.role.includes('staff')
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
                <td class="px-6 py-4 text-right text-sm font-medium sticky right-0 bg-gray-900 z-10 shadow-sticky-l">
                  <div class="flex justify-end space-x-4">
                    <button
                      @click="openViewModal(user)"
                      class="text-blue-400 hover:text-blue-300"
                      title="View user details"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    </button>
                    <button
                      @click="openUserModal(user)"
                      class="text-blue-400 hover:text-blue-300"
                      title="Edit user"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    </button>
                    <button
                      @click="confirmDelete(user)"
                      class="text-red-400 hover:text-red-300"
                      title="Delete user"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
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
  z-index: 10;
  background-color: #111827 !important; /* dark bg-gray-900 */
}

/* Proper hover states on sticky columns */
tr:hover .sticky {
  background-color: #111827 !important; /* dark bg-gray-900 */
}

/* Make sure the table heading for actions stays consistent */
thead .sticky {
  background-color: #374151 !important; /* dark bg-gray-700 */
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

/* Dark mode scrollbar */
.dark ::-webkit-scrollbar-track {
  background: #1f2937;
}

.dark ::-webkit-scrollbar-thumb {
  background: #4b5563;
  border-radius: 4px;
}

.dark ::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}

/* Role badge styling */
.role-badge {
  padding: 0.25rem 0.5rem;
  display: inline-flex;
  font-size: 0.75rem;
  line-height: 1.25rem;
  font-weight: 500;
  border-radius: 0.375rem;
}

/* Shadow for sticky columns in dark mode */
.dark .shadow-sticky-l {
  box-shadow: -4px 0 5px -2px rgba(0, 0, 0, 0.3);
}

/* Proper hover states on sticky columns in dark mode */
.dark tr:hover .sticky {
  background-color: #111827 !important; /* dark:bg-gray-900 */
}
</style>

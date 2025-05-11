<script setup lang="ts">
import { ref, computed, watch } from 'vue';

// Props
interface User {
  user_id: number;
  username: string;
  password_hash?: string;
  role: string;
  first_name: string;
  last_name: string;
  email: string;
  phone?: string;
  department?: string;
  position?: string;
  last_login?: Date;
  created_at: Date;
  updated_at: Date;
}

const props = defineProps<{
  show: boolean;
  user: User | null;
}>();

const emit = defineEmits<{
  'update:show': [value: boolean];
  'save': [user: User];
}>();

// Form state
const formData = ref<User>({
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
});

// Form validation
const errors = ref<Record<string, string>>({});
const isFormTouched = ref(false);
const showPassword = ref(false);

// Role options
const roleOptions = [
  { value: 'Sales Staff', label: 'Sales Staff' },
  { value: 'Inventory Manager', label: 'Inventory Manager' },
  { value: 'Branch Manager', label: 'Branch Manager' }
];

// Department options based on role
const departmentOptions = computed(() => {
  switch (formData.value.role) {
    case 'Sales Staff':
      return ['Sales', 'Marketing', 'Customer Service'];
    case 'Inventory Manager':
      return ['Warehouse', 'Logistics', 'Supply Chain'];
    case 'Branch Manager':
      return ['Operations', 'Administration', 'Executive'];
    default:
      return [];
  }
});

// Position options based on role
const positionOptions = computed(() => {
  switch (formData.value.role) {
    case 'Sales Staff':
      return ['Sales Representative', 'Sales Associate', 'Account Manager', 'Customer Service Rep'];
    case 'Inventory Manager':
      return ['Inventory Specialist', 'Stock Controller', 'Logistics Coordinator'];
    case 'Branch Manager':
      return ['Store Manager', 'Branch Director', 'Regional Manager', 'Operations Lead'];
    default:
      return [];
  }
});

// Set department and position defaults when role changes
watch(() => formData.value.role, (newRole) => {
  if (newRole && departmentOptions.value.length > 0) {
    formData.value.department = departmentOptions.value[0];
  }
  if (newRole && positionOptions.value.length > 0) {
    formData.value.position = positionOptions.value[0];
  }
});

// Computed properties
const isNewUser = computed(() => !formData.value.user_id);
const modalTitle = computed(() => isNewUser.value ? 'Create New User' : 'Edit User');
const isFormValid = computed(() => {
  validateForm();
  return Object.keys(errors.value).length === 0;
});

// Methods
const closeModal = () => {
  isFormTouched.value = false;
  errors.value = {};
  emit('update:show', false);
};

const validateForm = (): boolean => {
  const newErrors: Record<string, string> = {};

  // Username validation
  if (!formData.value.username.trim()) {
    newErrors.username = 'Username is required';
  } else if (formData.value.username.length < 3) {
    newErrors.username = 'Username must be at least 3 characters';
  } else if (!/^[a-zA-Z0-9_]+$/.test(formData.value.username)) {
    newErrors.username = 'Username can only contain letters, numbers, and underscores';
  }

  // Name validations
  if (!formData.value.first_name.trim()) {
    newErrors.first_name = 'First name is required';
  } else if (formData.value.first_name.length < 2) {
    newErrors.first_name = 'First name must be at least 2 characters';
  }

  if (!formData.value.last_name.trim()) {
    newErrors.last_name = 'Last name is required';
  } else if (formData.value.last_name.length < 2) {
    newErrors.last_name = 'Last name must be at least 2 characters';
  }

  // Email validation
  if (!formData.value.email.trim()) {
    newErrors.email = 'Email is required';
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.value.email)) {
    newErrors.email = 'Please enter a valid email address';
  }

  // Role validation
  if (!formData.value.role) {
    newErrors.role = 'Role is required';
  } else if (!roleOptions.some(option => option.value === formData.value.role)) {
    newErrors.role = 'Invalid role selected';
  }

  // Department validation
  if (!formData.value.department) {
    newErrors.department = 'Department is required';
  }

  // Position validation
  if (!formData.value.position) {
    newErrors.position = 'Position is required';
  }

  // Philippine Phone validation (if provided)
  if (formData.value.phone) {
    // Should start with +63 followed by 10 digits
    if (!/^\+63[0-9]{10}$/.test(formData.value.phone)) {
      newErrors.phone = 'Please enter a valid Philippine phone number (+63XXXXXXXXXX)';
    }
  }

  // Check for password if creating new user
  if (isNewUser.value && !formData.value.password_hash) {
    newErrors.password_hash = 'Password is required for new users';
  } else if (isNewUser.value && formData.value.password_hash && formData.value.password_hash.length < 6) {
    newErrors.password_hash = 'Password must be at least 6 characters';
  }

  errors.value = newErrors;
  return Object.keys(newErrors).length === 0;
};

const handleSubmit = () => {
  isFormTouched.value = true;
  if (validateForm()) {
    emit('save', { ...formData.value });
  }
};

const touchField = () => {
  isFormTouched.value = true;
};

const togglePasswordVisibility = () => {
  showPassword.value = !showPassword.value;
}

// Watch for changes in props
watch(() => props.user, (newUser) => {
  if (newUser) {
    formData.value = { ...newUser };
    // Reset validation state
    isFormTouched.value = false;
    errors.value = {};
  }
}, { immediate: true });

// Watch for modal visibility
watch(() => props.show, (isVisible) => {
  if (isVisible && props.user) {
    formData.value = { ...props.user };
    // Reset validation state
    isFormTouched.value = false;
    errors.value = {};
  }
});
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto">
    <!-- Backdrop -->
    <div class="fixed inset-0 bg-black bg-opacity-50 transition-opacity" @click="closeModal"></div>

    <!-- Modal content -->
    <div class="flex items-center justify-center min-h-screen p-4">
      <div class="relative bg-white dark:bg-gray-800 rounded-lg max-w-xl w-full mx-auto shadow-xl transform transition-all sm:max-w-lg modal-content">
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

        <!-- Form -->
        <div class="px-6 py-4 max-h-[calc(100vh-200px)] overflow-y-auto">
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <!-- Basic info section -->
            <div class="space-y-4">
              <h4 class="font-medium text-gray-900 dark:text-white text-sm">Account Information</h4>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <!-- Username -->
                <div>
                  <label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Username *</label>
                  <input
                    id="username"
                    v-model="formData.username"
                    @input="touchField"
                    type="text"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                    :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.username}"
                  />
                  <p v-if="isFormTouched && errors.username" class="mt-1 text-sm text-red-500">{{ errors.username }}</p>
                </div>

                <!-- Role -->
                <div>
                  <label for="role" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Role *</label>
                  <select
                    id="role"
                    v-model="formData.role"
                    @change="touchField"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                    :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.role}"
                    placeholder="Select a role"
                  >
                    <option disabled value="">Select a role</option>
                    <option v-for="option in roleOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </option>
                  </select>
                  <p v-if="isFormTouched && errors.role" class="mt-1 text-sm text-red-500">{{ errors.role }}</p>
                </div>
              </div>

              <!-- Password (only for new users) -->
              <div v-if="isNewUser">
                <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Password *</label>
                <div class="relative">
                  <input
                    id="password"
                    v-model="formData.password_hash"
                    @input="touchField"
                    :type="showPassword ? 'text' : 'password'"
                    class="w-full pl-3 pr-10 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                    :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.password_hash}"
                  />
                  <button
                    type="button"
                    @click="togglePasswordVisibility"
                    class="absolute inset-y-0 right-0 px-3 flex items-center"
                  >
                    <svg v-if="!showPassword" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                  </button>
                </div>
                <p v-if="isFormTouched && errors.password_hash" class="mt-1 text-sm text-red-500">{{ errors.password_hash }}</p>
                <p v-else class="mt-1 text-xs text-gray-500 dark:text-gray-400">Password must be at least 6 characters</p>
              </div>
            </div>

            <!-- Personal info section -->
            <div class="pt-3 space-y-4">
              <h4 class="font-medium text-gray-900 dark:text-white text-sm">Personal Information</h4>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <!-- First Name -->
                <div>
                  <label for="firstName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">First Name *</label>
                  <input
                    id="firstName"
                    v-model="formData.first_name"
                    @input="touchField"
                    type="text"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                    :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.first_name}"
                  />
                  <p v-if="isFormTouched && errors.first_name" class="mt-1 text-sm text-red-500">{{ errors.first_name }}</p>
                </div>

                <!-- Last Name -->
                <div>
                  <label for="lastName" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Last Name *</label>
                  <input
                    id="lastName"
                    v-model="formData.last_name"
                    @input="touchField"
                    type="text"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                    :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.last_name}"
                  />
                  <p v-if="isFormTouched && errors.last_name" class="mt-1 text-sm text-red-500">{{ errors.last_name }}</p>
                </div>
              </div>

              <!-- Email -->
              <div>
                <label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Email *</label>
                <input
                  id="email"
                  v-model="formData.email"
                  @input="touchField"
                  type="email"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                  :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.email}"
                />
                <p v-if="isFormTouched && errors.email" class="mt-1 text-sm text-red-500">{{ errors.email }}</p>
              </div>

              <!-- Phone -->
              <div>
                <label for="phone" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Phone Number</label>
                <input
                  id="phone"
                  v-model="formData.phone"
                  @input="touchField"
                  type="tel"
                  placeholder="+63XXXXXXXXXX"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                  :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.phone}"
                />
                <p v-if="isFormTouched && errors.phone" class="mt-1 text-sm text-red-500">{{ errors.phone }}</p>
                <p v-else class="mt-1 text-xs text-gray-500 dark:text-gray-400">Format: +63XXXXXXXXXX (Philippines number only)</p>
              </div>
            </div>

            <!-- Work info section -->
            <div class="pt-3 space-y-4">
              <h4 class="font-medium text-gray-900 dark:text-white text-sm">Work Information</h4>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <!-- Department -->
                <div>
                  <label for="department" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Department *</label>
                  <select
                    id="department"
                    v-model="formData.department"
                    @change="touchField"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                    :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.department}"
                  >
                    <option v-for="department in departmentOptions" :key="department" :value="department">
                      {{ department }}
                    </option>
                  </select>
                  <p v-if="isFormTouched && errors.department" class="mt-1 text-sm text-red-500">{{ errors.department }}</p>
                </div>

                <!-- Position -->
                <div>
                  <label for="position" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Position *</label>
                  <select
                    id="position"
                    v-model="formData.position"
                    @change="touchField"
                    class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm dark:bg-gray-700 dark:text-white"
                    :class="{'border-red-500 dark:border-red-500': isFormTouched && errors.position}"
                  >
                    <option v-for="position in positionOptions" :key="position" :value="position">
                      {{ position }}
                    </option>
                  </select>
                  <p v-if="isFormTouched && errors.position" class="mt-1 text-sm text-red-500">{{ errors.position }}</p>
                </div>
              </div>
            </div>
          </form>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 bg-gray-50 dark:bg-gray-700 border-t border-gray-200 dark:border-gray-600 flex justify-end space-x-3 rounded-b-lg">
          <button
            @click="closeModal"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 sm:text-sm transition-colors"
          >
            Cancel
          </button>
          <button
            @click="handleSubmit"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 sm:text-sm transition-colors"
            :disabled="isFormTouched && !isFormValid"
            :class="{'opacity-60 cursor-not-allowed': isFormTouched && !isFormValid}"
          >
            {{ isNewUser ? 'Create User' : 'Save Changes' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Smooth transitions */
.modal-content {
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

/* Form input focus styles */
input:focus, select:focus {
  outline: none;
  ring: 2px;
  ring-color: rgb(59, 130, 246);
  border-color: rgb(59, 130, 246);
}
</style>

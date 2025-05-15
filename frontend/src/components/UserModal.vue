<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useUserStore } from '../stores/userStore';

// Props
interface User {
  user_id: number;
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

// Initialize store
const userStore = useUserStore();

// Form state
const formData = ref<User>({
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
});

// Form validation
const errors = ref<Record<string, string>>({});
const isFormTouched = ref(false);
const showPassword = ref(false);
const activeTab = ref('account');

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

  // Department validation (optional in backend)
  if (formData.value.department && formData.value.department.trim() === '') {
    newErrors.department = 'Department cannot be empty if provided';
  }

  // Position validation (optional in backend)
  if (formData.value.position && formData.value.position.trim() === '') {
    newErrors.position = 'Position cannot be empty if provided';
  }

  // Philippine Phone validation (if provided)
  if (formData.value.phone && formData.value.phone.trim() !== '') {
    // Should start with +63 followed by 10 digits
    if (!/^\+63[0-9]{10}$/.test(formData.value.phone)) {
      newErrors.phone = 'Please enter a valid Philippine phone number (+63XXXXXXXXXX)';
    }
  }

  // Check for password if creating new user
  if (isNewUser.value && !formData.value.password_hash) {
    newErrors.password_hash = 'Password is required for new users';
  } else if (isNewUser.value && formData.value.password_hash && formData.value.password_hash.length < 8) {
    newErrors.password_hash = 'Password must be at least 8 characters';
  }

  errors.value = newErrors;
  return Object.keys(newErrors).length === 0;
};

const handleSubmit = async () => {
  isFormTouched.value = true;
  if (validateForm()) {
    try {
      // Call user store to save the user
      let savedUser: User;
      
      if (isNewUser.value) {
        savedUser = await userStore.createUser(formData.value) as User;
      } else {
        savedUser = await userStore.updateUser(formData.value.user_id, formData.value) as User;
      }
      
      emit('save', savedUser);
    } catch (err) {
      console.error('Error saving user:', err);
      errors.value.submit = userStore.error || 'Failed to save user';
    }
  }
};

const touchField = () => {
  isFormTouched.value = true;
};

const togglePasswordVisibility = () => {
  showPassword.value = !showPassword.value;
}

const setActiveTab = (tab: string) => {
  activeTab.value = tab;
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
    activeTab.value = 'account';
  }
});
</script>

<template>
  <transition name="modal-fade">
    <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto">
      <!-- Backdrop with animation -->
      <div class="fixed inset-0 bg-black bg-opacity-70 transition-opacity backdrop-blur-sm" @click="closeModal"></div>

      <!-- Modal content -->
      <div class="flex items-center justify-center min-h-screen p-4">
        <div class="relative bg-gray-800 rounded-lg max-w-xl w-full mx-auto shadow-2xl transform transition-all sm:max-w-lg modal-content">
          <!-- Header with gradient background -->
          <div class="relative px-6 py-4 border-b border-gray-700 bg-gradient-to-r from-gray-700 to-gray-800 rounded-t-lg">
            <div class="flex justify-between items-center">
              <div>
                <h3 class="text-lg font-medium text-white">{{ modalTitle }}</h3>
                <p class="text-sm text-gray-300">{{ isNewUser ? 'Create a new system user' : 'Update existing user' }}</p>
              </div>
              <button
                @click="closeModal"
                class="flex items-center justify-center h-8 w-8 rounded-md text-gray-400 hover:text-white hover:bg-gray-600 transition-colors focus:outline-none focus:ring-2 focus:ring-gray-500"
                aria-label="Close"
              >
                <svg class="h-5 w-5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <!-- Navigation tabs -->
            <div class="flex space-x-1 mt-4 border-b border-gray-700">
              <button 
                @click="setActiveTab('account')" 
                class="px-4 py-2 text-sm font-medium focus:outline-none transition-colors rounded-t-md"
                :class="activeTab === 'account' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-700' : 'text-gray-400 hover:text-gray-300'"
              >
                <div class="flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  Account
                </div>
              </button>
              <button 
                @click="setActiveTab('personal')" 
                class="px-4 py-2 text-sm font-medium focus:outline-none transition-colors rounded-t-md"
                :class="activeTab === 'personal' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-700' : 'text-gray-400 hover:text-gray-300'"
              >
                <div class="flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  Contact
                </div>
              </button>
              <button 
                @click="setActiveTab('work')" 
                class="px-4 py-2 text-sm font-medium focus:outline-none transition-colors rounded-t-md"
                :class="activeTab === 'work' ? 'text-blue-400 border-b-2 border-blue-400 bg-gray-700' : 'text-gray-400 hover:text-gray-300'"
              >
                <div class="flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                  Work
                </div>
              </button>
            </div>
          </div>

          <!-- Form -->
          <div class="px-6 py-5 max-h-[calc(100vh-250px)] overflow-y-auto bg-gray-850">
            <form @submit.prevent="handleSubmit" class="space-y-6">
              <!-- Account Tab -->
              <div v-if="activeTab === 'account'" class="space-y-5">
                <!-- Role -->
                <div class="form-group">
                  <label for="role" class="form-label">
                    <span class="text-sm font-medium text-gray-300">User Role</span>
                    <span class="text-red-400 ml-1">*</span>
                  </label>
                  <div class="mt-1.5 relative">
                    <select
                      id="role"
                      v-model="formData.role"
                      @change="touchField"
                      class="form-select"
                      :class="{'error-input': isFormTouched && errors.role}"
                    >
                      <option disabled value="">Select role</option>
                      <option v-for="option in roleOptions" :key="option.value" :value="option.value">
                        {{ option.label }}
                      </option>
                    </select>
                    <div class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                      <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
                      </svg>
                    </div>
                  </div>
                  <p v-if="isFormTouched && errors.role" class="error-message">{{ errors.role }}</p>
                </div>

                <!-- Password (only for new users) -->
                <div v-if="isNewUser" class="form-group">
                  <label for="password" class="form-label">
                    <span class="text-sm font-medium text-gray-300">Password</span>
                    <span class="text-red-400 ml-1">*</span>
                  </label>
                  <div class="mt-1.5 relative">
                    <input
                      id="password"
                      v-model="formData.password_hash"
                      @input="touchField"
                      :type="showPassword ? 'text' : 'password'"
                      class="form-input pr-10"
                      :class="{'error-input': isFormTouched && errors.password_hash}"
                      placeholder="Minimum 8 characters"
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
                  <p v-if="isFormTouched && errors.password_hash" class="error-message">{{ errors.password_hash }}</p>
                  <p v-else class="helper-text">Strong password recommended with numbers and special characters</p>
                </div>
              </div>

              <!-- Personal Tab -->
              <div v-if="activeTab === 'personal'" class="space-y-5">
                <!-- Name fields in a row -->
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <!-- First Name -->
                  <div class="form-group">
                    <label for="firstName" class="form-label">
                      <span class="text-sm font-medium text-gray-300">First Name</span>
                      <span class="text-red-400 ml-1">*</span>
                    </label>
                    <div class="mt-1.5">
                      <input
                        id="firstName"
                        v-model="formData.first_name"
                        @input="touchField"
                        type="text"
                        placeholder="John"
                        class="form-input"
                        :class="{'error-input': isFormTouched && errors.first_name}"
                      />
                    </div>
                    <p v-if="isFormTouched && errors.first_name" class="error-message">{{ errors.first_name }}</p>
                  </div>

                  <!-- Last Name -->
                  <div class="form-group">
                    <label for="lastName" class="form-label">
                      <span class="text-sm font-medium text-gray-300">Last Name</span>
                      <span class="text-red-400 ml-1">*</span>
                    </label>
                    <div class="mt-1.5">
                      <input
                        id="lastName"
                        v-model="formData.last_name"
                        @input="touchField"
                        type="text"
                        placeholder="Doe"
                        class="form-input"
                        :class="{'error-input': isFormTouched && errors.last_name}"
                      />
                    </div>
                    <p v-if="isFormTouched && errors.last_name" class="error-message">{{ errors.last_name }}</p>
                  </div>
                </div>

                <!-- Email -->
                <div class="form-group">
                  <label for="email" class="form-label">
                    <span class="text-sm font-medium text-gray-300">Email Address</span>
                    <span class="text-red-400 ml-1">*</span>
                  </label>
                  <div class="mt-1.5">
                    <input
                      id="email"
                      v-model="formData.email"
                      @input="touchField"
                      type="email"
                      placeholder="john.doe@example.com"
                      class="form-input"
                      :class="{'error-input': isFormTouched && errors.email}"
                    />
                  </div>
                  <p v-if="isFormTouched && errors.email" class="error-message">{{ errors.email }}</p>
                </div>

                <!-- Phone -->
                <div class="form-group">
                  <label for="phone" class="form-label">
                    <span class="text-sm font-medium text-gray-300">Phone Number</span>
                  </label>
                  <div class="mt-1.5">
                    <input
                      id="phone"
                      v-model="formData.phone"
                      @input="touchField"
                      type="tel"
                      placeholder="+63XXXXXXXXXX"
                      class="form-input"
                      :class="{'error-input': isFormTouched && errors.phone}"
                    />
                  </div>
                  <p v-if="isFormTouched && errors.phone" class="error-message">{{ errors.phone }}</p>
                  <p v-else class="helper-text">Format: +63XXXXXXXXXX (Philippines number)</p>
                </div>
              </div>

              <!-- Work Tab -->
              <div v-if="activeTab === 'work'" class="space-y-5">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <!-- Department -->
                  <div class="form-group">
                    <label for="department" class="form-label">
                      <span class="text-sm font-medium text-gray-300">Department</span>
                      <span class="text-red-400 ml-1">*</span>
                    </label>
                    <div class="mt-1.5 relative">
                      <select
                        id="department"
                        v-model="formData.department"
                        @change="touchField"
                        class="form-select"
                        :class="{'error-input': isFormTouched && errors.department}"
                      >
                        <option v-for="department in departmentOptions" :key="department" :value="department">
                          {{ department }}
                        </option>
                      </select>
                      <div class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
                        </svg>
                      </div>
                    </div>
                    <p v-if="isFormTouched && errors.department" class="error-message">{{ errors.department }}</p>
                  </div>

                  <!-- Position -->
                  <div class="form-group">
                    <label for="position" class="form-label">
                      <span class="text-sm font-medium text-gray-300">Position</span>
                      <span class="text-red-400 ml-1">*</span>
                    </label>
                    <div class="mt-1.5 relative">
                      <select
                        id="position"
                        v-model="formData.position"
                        @change="touchField"
                        class="form-select"
                        :class="{'error-input': isFormTouched && errors.position}"
                      >
                        <option v-for="position in positionOptions" :key="position" :value="position">
                          {{ position }}
                        </option>
                      </select>
                      <div class="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l4-4 4 4m0 6l-4 4-4-4" />
                        </svg>
                      </div>
                    </div>
                    <p v-if="isFormTouched && errors.position" class="error-message">{{ errors.position }}</p>
                  </div>
                </div>
              </div>
            </form>
          </div>

          <!-- API Error Message -->
          <div v-if="userStore.error" class="px-6 py-3 bg-gray-900">
            <div class="bg-red-900/50 border border-red-700 text-red-200 px-4 py-3 rounded-md shadow-sm" role="alert">
              <div class="flex">
                <svg class="h-5 w-5 text-red-400 mr-2 mt-0.5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
                </svg>
                <span>{{ userStore.error }}</span>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-6 py-4 bg-gray-900 border-t border-gray-700 flex justify-end space-x-3 rounded-b-lg">
            <div class="flex-grow flex items-center">
              <div v-if="activeTab !== 'account'" class="mr-auto">
                <button
                  @click="activeTab === 'personal' ? setActiveTab('account') : setActiveTab('personal')"
                  class="flex items-center text-sm text-blue-400 hover:text-blue-300 transition-colors"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                  </svg>
                  Previous
                </button>
              </div>
              <div v-if="activeTab !== 'work'" class="ml-auto mr-3">
                <button
                  @click="activeTab === 'account' ? setActiveTab('personal') : setActiveTab('work')"
                  class="flex items-center text-sm text-blue-400 hover:text-blue-300 transition-colors"
                >
                  Next
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                  </svg>
                </button>
              </div>
            </div>
            <button
              @click="closeModal"
              class="px-4 py-2 bg-gray-700 text-gray-300 rounded-md hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 transition-colors font-medium text-sm"
            >
              Cancel
            </button>
            <button
              @click="handleSubmit"
              class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors font-medium text-sm flex items-center"
              :disabled="isFormTouched && !isFormValid"
              :class="{'opacity-60 cursor-not-allowed': isFormTouched && !isFormValid}"
            >
              <svg v-if="activeTab === 'work'" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
              {{ isNewUser ? 'Create User' : 'Save Changes' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<style scoped>
/* Animations */
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.3s;
}

.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}

.modal-content {
  animation: modal-appear 0.3s ease-out;
}

@keyframes modal-appear {
  from {
    opacity: 0;
    transform: translateY(-50px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Scrollbar styles */
.overflow-y-auto::-webkit-scrollbar {
  width: 4px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: #1f2937;
  border-radius: 8px;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: #4b5563;
  border-radius: 8px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}

/* Form styles */
.form-group {
  position: relative;
}

.form-label {
  display: block;
  margin-bottom: 0.25rem;
}

.form-input,
.form-select {
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: 1px solid #4b5563;
  color: white;
  background-color: #374151;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #60a5fa;
  box-shadow: 0 0 0 1px rgba(96, 165, 250, 0.5);
}

.form-input::placeholder {
  color: #9ca3af;
}

.error-input {
  border-color: #f87171 !important;
}

.error-message {
  color: #f87171;
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

.helper-text {
  color: #9ca3af;
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

/* Custom background colors */
.bg-gray-850 {
  background-color: #1a1e2a;
}

/* Tab transitions */
.tab-content-enter-active,
.tab-content-leave-active {
  transition: opacity 0.2s, transform 0.2s;
}

.tab-content-enter-from,
.tab-content-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>

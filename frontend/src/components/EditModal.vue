<script setup lang="ts">
import { ref, defineProps, defineEmits, watch, computed } from 'vue';
import { useCustomerStore } from '@/stores/customerStore';
import { useContactStore } from '@/stores/contactStore';
import type { CustomerUpdate, Customer } from '@/types/Customer';
import type { Contact, ContactUpdate } from '@/types/Contact';

const customerStore = useCustomerStore();
const contactStore = useContactStore();

const props = defineProps<{
  show: boolean;
  customer?: Customer | null;
  contact?: Contact | null;
  mode: 'customer' | 'contact';
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'save', data: CustomerUpdate | ContactUpdate): void;
}>();

// Form state for customer data
const customerData = ref<CustomerUpdate>({
  customer_id: 0,
  company_name: '',
  industry: '',
  address: '',
  website: '',
  email: '',
  phone: ''
});

// Form state for contact data
const contactData = ref<ContactUpdate>({
  contact_id: 0,
  first_name: '',
  last_name: '',
  position: '',
  email: '',
  phone: ''
});

// Form validation and UI state
const errors = ref<Record<string, string>>({});
const touched = ref<Record<string, boolean>>({});
const isLoading = ref(false);

// Backend validation states
const isCheckingEmail = ref(false);

// Validation rules
const validateField = async (field: string, value: string | undefined) => {
  touched.value[field] = true;
  
  // Required field validation
  if (!value) {
    errors.value[field] = `${field.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')} is required`;
    return false;
  }
  
  // Email format validation
  if (field === 'email' && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
    errors.value[field] = 'Please enter a valid email address';
    return false;
  }

  // Email uniqueness check for contacts
  if (field === 'email' && props.mode === 'contact') {
    isCheckingEmail.value = true;
    try {
      const exists = await contactStore.checkEmailExists(value);
      if (exists && value !== props.contact?.email) {
        errors.value[field] = 'This email is already in use';
        return false;
      }
    } catch (error) {
      console.error('Error checking email:', error);
      errors.value[field] = 'Unable to verify email';
      return false;
    } finally {
      isCheckingEmail.value = false;
    }
  }

  delete errors.value[field];
  return true;
};

// Handle form submission
const handleSubmit = async () => {
  if (isLoading.value) return;
  
  isLoading.value = true;
  
  try {
    if (props.mode === 'customer') {
      // Validate all customer fields
      const customerFields = ['company_name', 'industry', 'address', 'website', 'email', 'phone'] as const;
      const validations = await Promise.all(
        customerFields.map(field => validateField(field, customerData.value[field]))
      );

      if (validations.includes(false)) {
        isLoading.value = false;
        return;
      }

      emit('save', customerData.value);
    } else {
      // Validate all contact fields
      const contactFields = ['first_name', 'last_name', 'position', 'email', 'phone'] as const;
      const validations = await Promise.all(
        contactFields.map(field => validateField(field, contactData.value[field]))
      );

      if (validations.includes(false)) {
        isLoading.value = false;
        return;
      }

      emit('save', contactData.value);
    }
  } catch (error) {
    console.error('Error saving data:', error);
    errors.value.submit = 'Failed to save. Please try again.';
  } finally {
    isLoading.value = false;
  }
};

// Reset form state
const resetForm = () => {
  if (props.mode === 'customer') {
    customerData.value = {
      customer_id: 0,
      company_name: '',
      industry: '',
      address: '',
      website: '',
      email: '',
      phone: ''
    };
  } else {
    contactData.value = {
      contact_id: 0,
      first_name: '',
      last_name: '',
      position: '',
      email: '',
      phone: ''
    };
  }
  
  errors.value = {};
  touched.value = {};
};

// Close modal handler
const closeModal = () => {
  resetForm();
  emit('close');
};

// Watch for data prop changes
watch(() => props.customer, (newCustomer) => {
  if (newCustomer && props.mode === 'customer') {
    customerData.value = {
      customer_id: newCustomer.customer_id,
      company_name: newCustomer.company_name || '',
      industry: newCustomer.industry || '',
      address: newCustomer.address || '',
      website: newCustomer.website || '',
      email: newCustomer.email || '',
      phone: newCustomer.phone || ''
    };
  }
}, { immediate: true });

watch(() => props.contact, (newContact) => {
  if (newContact && props.mode === 'contact') {
    contactData.value = {
      contact_id: newContact.contact_id,
      first_name: newContact.first_name || '',
      last_name: newContact.last_name || '',
      position: newContact.position || '',
      email: newContact.email || '',
      phone: newContact.phone || ''
    };
  }
}, { immediate: true });

// Computed properties for UI state
const modalTitle = computed(() => {
  return props.mode === 'customer' 
    ? 'Edit Company Information' 
    : 'Edit Contact Information';
});

const modalDescription = computed(() => {
  return props.mode === 'customer'
    ? 'Update the company details'
    : 'Update contact person information';
});
</script>

<template>
  <transition name="modal">
    <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto">
      <div class="min-h-screen px-4 text-center">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" @click="closeModal" />

        <!-- This element is to trick the browser into centering the modal contents. -->
        <span class="inline-block h-screen align-middle" aria-hidden="true">&#8203;</span>

        <!-- Modal panel -->
        <div class="inline-block w-full align-middle text-left transform transition-all z-60">
          <div class="relative mx-auto max-w-lg bg-white rounded-lg shadow-xl">
            <!-- Header -->
            <div class="bg-gradient-to-r from-blue-600 to-blue-800 px-4 py-3 sm:px-6 sm:py-4 rounded-t-lg">
              <div class="flex items-start sm:items-center justify-between">
                <div class="pr-12">
                  <h3 class="text-base sm:text-lg font-medium leading-6 text-white">
                    {{ modalTitle }}
                  </h3>
                  <p class="mt-1 text-xs sm:text-sm text-blue-100">
                    {{ modalDescription }}
                  </p>
                </div>
                <button
                  @click="closeModal"
                  class="absolute top-3 right-3 rounded-md bg-blue-700 p-1.5 text-blue-200 hover:text-white focus:outline-none"
                >
                  <span class="sr-only">Close</span>
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            </div>

            <!-- Form content -->
            <div class="px-4 py-4 sm:px-6 sm:py-5 max-h-[calc(100vh-12rem)] overflow-y-auto">
              <div v-if="mode === 'customer'" class="space-y-4">
                <!-- Company Information Form -->
                <div class="space-y-4">
                  <!-- Company Name -->
                  <div>
                    <label for="company_name" class="block text-sm font-medium text-gray-700">
                      Company Name <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <input
                        type="text"
                        id="company_name"
                        v-model="customerData.company_name"
                        @blur="validateField('company_name', customerData.company_name)"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.company_name && touched.company_name }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="errors.company_name && touched.company_name" class="mt-1 text-xs text-red-600">
                        {{ errors.company_name }}
                      </p>
                    </div>
                  </div>

                  <!-- Industry -->
                  <div>
                    <label for="industry" class="block text-sm font-medium text-gray-700">
                      Industry <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <input
                        type="text"
                        id="industry"
                        v-model="customerData.industry"
                        @blur="validateField('industry', customerData.industry)"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.industry && touched.industry }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="errors.industry && touched.industry" class="mt-1 text-xs text-red-600">
                        {{ errors.industry }}
                      </p>
                    </div>
                  </div>

                  <!-- Address -->
                  <div>
                    <label for="address" class="block text-sm font-medium text-gray-700">
                      Address <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <textarea
                        id="address"
                        v-model="customerData.address"
                        @blur="validateField('address', customerData.address)"
                        rows="3"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.address && touched.address }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="errors.address && touched.address" class="mt-1 text-xs text-red-600">
                        {{ errors.address }}
                      </p>
                    </div>
                  </div>

                  <!-- Website -->
                  <div>
                    <label for="website" class="block text-sm font-medium text-gray-700">
                      Website <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <input
                        type="url"
                        id="website"
                        v-model="customerData.website"
                        @blur="validateField('website', customerData.website)"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.website && touched.website }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="errors.website && touched.website" class="mt-1 text-xs text-red-600">
                        {{ errors.website }}
                      </p>
                    </div>
                  </div>

                  <!-- Company Email -->
                  <div>
                    <label for="email" class="block text-sm font-medium text-gray-700">
                      Company Email <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <input
                        type="email"
                        id="email"
                        v-model="customerData.email"
                        @blur="validateField('email', customerData.email)"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.email && touched.email }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="errors.email && touched.email" class="mt-1 text-xs text-red-600">
                        {{ errors.email }}
                      </p>
                    </div>
                  </div>

                  <!-- Company Phone -->
                  <div>
                    <label for="phone" class="block text-sm font-medium text-gray-700">
                      Company Phone <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <input
                        type="tel"
                        id="phone"
                        v-model="customerData.phone"
                        @blur="validateField('phone', customerData.phone)"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.phone && touched.phone }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="errors.phone && touched.phone" class="mt-1 text-xs text-red-600">
                        {{ errors.phone }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <div v-else class="space-y-4">
                <!-- Contact Information Form -->
                <div class="space-y-4">
                  <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <!-- First Name -->
                    <div>
                      <label for="first_name" class="block text-sm font-medium text-gray-700">
                        First Name <span class="text-red-500">*</span>
                      </label>
                      <div class="mt-1">
                        <input
                          type="text"
                          id="first_name"
                          v-model="contactData.first_name"
                          @blur="validateField('first_name', contactData.first_name)"
                          class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                          :class="{ 'border-red-300': errors.first_name && touched.first_name }"
                          :disabled="isLoading"
                          required
                        />
                        <p v-if="errors.first_name && touched.first_name" class="mt-1 text-xs text-red-600">
                          {{ errors.first_name }}
                        </p>
                      </div>
                    </div>

                    <!-- Last Name -->
                    <div>
                      <label for="last_name" class="block text-sm font-medium text-gray-700">
                        Last Name <span class="text-red-500">*</span>
                      </label>
                      <div class="mt-1">
                        <input
                          type="text"
                          id="last_name"
                          v-model="contactData.last_name"
                          @blur="validateField('last_name', contactData.last_name)"
                          class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                          :class="{ 'border-red-300': errors.last_name && touched.last_name }"
                          :disabled="isLoading"
                          required
                        />
                        <p v-if="errors.last_name && touched.last_name" class="mt-1 text-xs text-red-600">
                          {{ errors.last_name }}
                        </p>
                      </div>
                    </div>
                  </div>

                  <!-- Position -->
                  <div>
                    <label for="position" class="block text-sm font-medium text-gray-700">
                      Position <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <input
                        type="text"
                        id="position"
                        v-model="contactData.position"
                        @blur="validateField('position', contactData.position)"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.position && touched.position }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="errors.position && touched.position" class="mt-1 text-xs text-red-600">
                        {{ errors.position }}
                      </p>
                    </div>
                  </div>

                  <!-- Contact Email -->
                  <div>
                    <label for="contact_email" class="block text-sm font-medium text-gray-700">
                      Contact Email <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <input
                        type="email"
                        id="contact_email"
                        v-model="contactData.email"
                        @blur="validateField('email', contactData.email)"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.email && touched.email }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="isCheckingEmail" class="mt-1 text-xs text-blue-500">
                        Checking availability...
                      </p>
                      <p v-else-if="errors.email && touched.email" class="mt-1 text-xs text-red-600">
                        {{ errors.email }}
                      </p>
                    </div>
                  </div>

                  <!-- Contact Phone -->
                  <div>
                    <label for="contact_phone" class="block text-sm font-medium text-gray-700">
                      Contact Phone <span class="text-red-500">*</span>
                    </label>
                    <div class="mt-1">
                      <input
                        type="tel"
                        id="contact_phone"
                        v-model="contactData.phone"
                        @blur="validateField('phone', contactData.phone)"
                        class="block w-full px-3 py-2 text-sm border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
                        :class="{ 'border-red-300': errors.phone && touched.phone }"
                        :disabled="isLoading"
                        required
                      />
                      <p v-if="errors.phone && touched.phone" class="mt-1 text-xs text-red-600">
                        {{ errors.phone }}
                      </p>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Error message -->
              <div v-if="errors.submit" class="mt-4 rounded-md bg-red-50 p-4">
                <div class="flex">
                  <div class="flex-shrink-0">
                    <svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                      <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                    </svg>
                  </div>
                  <div class="ml-3">
                    <p class="text-sm text-red-700">{{ errors.submit }}</p>
                  </div>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div class="bg-gray-50 px-4 py-3 sm:px-6 rounded-b-lg">
              <div class="flex justify-end space-x-3">
                <button
                  type="button"
                  class="inline-flex justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                  :disabled="isLoading"
                  @click="closeModal"
                >
                  Cancel
                </button>
                <button
                  type="button"
                  class="inline-flex justify-center items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                  :disabled="isLoading || Object.keys(errors).length > 0"
                  @click="handleSubmit"
                >
                  <svg
                    v-if="isLoading"
                    class="mr-2 h-4 w-4 animate-spin"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  Save Changes
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </transition>
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
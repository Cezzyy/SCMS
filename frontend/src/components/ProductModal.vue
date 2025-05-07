<script setup lang="ts">
import { defineProps, defineEmits, ref, watch, computed } from 'vue';
import type { Product } from '../types/Product';

const props = defineProps<{
  show: boolean;
  product: Product | null;
}>();

const emit = defineEmits(['update:show', 'save']);

const productData = ref<Product | null>(null);
const techSpecs = ref<{key: string, value: string}[]>([{key: '', value: ''}]);

// Form validation and UI state
const errors = ref<Record<string, string>>({});
const touched = ref<Record<string, boolean>>({});
const isLoading = ref(false);

// Deep copy product data when props change
watch(() => props.product, (newVal) => {
  productData.value = newVal ? JSON.parse(JSON.stringify(newVal)) : null;
  // Reset validation state
  errors.value = {};
  touched.value = {};
}, { immediate: true });

// Update the watch for technical specs
watch(() => productData.value, (newVal) => {
  if (newVal && newVal.technical_specs) {
    try {
      // Parse specs from JSON string
      let specs: {key: string, value: string}[] = [];

      if (typeof newVal.technical_specs === 'string') {
        try {
          // Try to parse as JSON
          const parsed = JSON.parse(newVal.technical_specs);
          specs = Object.entries(parsed).map(([key, value]) => ({
            key,
            value: typeof value === 'string' ? value : String(value)
          }));
        } catch (e) {
          console.error('Error parsing technical_specs JSON:', e);
          specs = [];
        }
      } else if (typeof newVal.technical_specs === 'object') {
        // Handle case where it's already an object
        specs = Object.entries(newVal.technical_specs).map(([key, value]) => ({
          key,
          value: typeof value === 'string' ? value : String(value)
        }));
      }

      // Always ensure there's at least one empty row
      if (specs.length === 0) {
        specs.push({key: '', value: ''});
      }

      techSpecs.value = specs;
    } catch (e) {
      console.error('Error processing technical specs:', e);
      techSpecs.value = [{key: '', value: ''}];
    }
  } else {
    techSpecs.value = [{key: '', value: ''}];
  }
}, { immediate: true });

// Validation rules
const validateField = (field: string, value: string | number | Record<string, any> | undefined) => {
  touched.value[field] = true;

  // Required field validation
  if (value === undefined || value === null || value === '') {
    errors.value[field] = `${field.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')} is required`;
    return false;
  }

  // Number validation for price and warranty period
  if (field === 'price' && (typeof value !== 'number' || value <= 0)) {
    errors.value[field] = `Price must be greater than zero`;
    return false;
  }

  if (field === 'warranty_period' && (typeof value !== 'number' || value < 0)) {
    errors.value[field] = `Warranty period must be a positive number`;
    return false;
  }

  // If the field has an existing error but is now valid, remove the error
  if (errors.value[field]) {
    delete errors.value[field];
  }

  return true;
};

// Close modal
const closeModal = () => {
  emit('update:show', false);
};

// Save product with validation
const saveProduct = () => {
  if (!productData.value || isLoading.value) return;

  isLoading.value = true;

  try {
    // Validate all required fields
    const requiredFields = ['product_name', 'model', 'price', 'description', 'warranty_period', 'certifications', 'safety_standards'] as const;

    requiredFields.forEach(field => {
      validateField(field, productData.value?.[field as keyof typeof productData.value]);
    });

    // Validate tech specs separately
    const techSpecsValid = validateTechSpecs();

    if (Object.keys(errors.value).length > 0 || !techSpecsValid) {
      isLoading.value = false;
      return;
    }

    // Check for duplicate product
    if (!productData.value.product_id) {
      // This is where you could add duplicate checking logic
      // For now, we'll just proceed with the save
    }

    emit('save', productData.value);
  } catch (error) {
    console.error('Error saving product:', error);
  } finally {
    isLoading.value = false;
  }
};

// Add a method to add a new specification row
const addSpecRow = () => {
  techSpecs.value.push({key: '', value: ''});
};

// Add a method to remove a specification row
const removeSpecRow = (index: number) => {
  if (techSpecs.value.length > 1) {
    techSpecs.value.splice(index, 1);
  } else {
    // If it's the last row, just clear it
    techSpecs.value[0] = {key: '', value: ''};
  }
};

// Update the technical specs validator
const validateTechSpecs = () => {
  if (!productData.value) return false;

  touched.value['technical_specs'] = true;

  // Filter out empty rows - must have both key and value
  const validSpecs = techSpecs.value.filter(spec => spec.key.trim() !== '' && spec.value.trim() !== '');

  if (validSpecs.length === 0) {
    errors.value['technical_specs'] = 'At least one technical specification is required';
    return false;
  }

  // Check for duplicate keys
  const keys = validSpecs.map(spec => spec.key.trim());
  const hasDuplicates = keys.some((key, index) => keys.indexOf(key) !== index);

  if (hasDuplicates) {
    errors.value['technical_specs'] = 'Duplicate specification keys are not allowed';
    return false;
  }

  // Convert to object directly - don't stringify
  const specsObject: Record<string, string> = {};
  validSpecs.forEach(spec => {
    specsObject[spec.key.trim()] = spec.value.trim();
  });

  // Save as object directly, not as JSON string
  productData.value.technical_specs = specsObject;

  console.log('Technical specs as object:', productData.value.technical_specs);

  delete errors.value['technical_specs'];
  return true;
};

// Update the handleTechnicalSpecsBlur function
const handleTechnicalSpecsBlur = () => {
  validateTechSpecs();
};

// Computed property to check if form is valid
const isFormValid = computed(() => {
  return Object.keys(errors.value).length === 0;
});
</script>

<template>
  <div v-if="show && productData"
      class="fixed inset-0 z-50 overflow-y-auto">
    <div class="min-h-screen px-4 text-center">
      <div class="fixed inset-0 transition-opacity" @click="closeModal">
        <div class="absolute inset-0 bg-black opacity-50"></div>
      </div>

      <span class="inline-block h-screen align-middle" aria-hidden="true">&#8203;</span>

      <div class="inline-block w-full max-w-3xl p-6 my-8 overflow-hidden text-left align-middle transition-all transform bg-white dark:bg-gray-800 shadow-xl rounded-lg border border-gray-200 dark:border-gray-700 modal-content">
        <!-- Modal Header -->
        <div class="flex justify-between items-center border-b border-gray-200 dark:border-gray-700 pb-4 mb-6">
          <h3 class="text-xl font-semibold text-gray-900 dark:text-white">
            {{ productData.product_id ? 'Edit Product' : 'Create New Product' }}
          </h3>
          <button @click="closeModal" class="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300 focus:outline-none">
            <span class="sr-only">Close</span>
            <svg class="h-6 w-6" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Error Summary -->
        <div v-if="Object.keys(errors).length > 0" class="mb-6 p-4 bg-red-50 dark:bg-red-900 dark:bg-opacity-20 border-l-4 border-red-500 rounded">
          <div class="flex">
            <div class="flex-shrink-0">
              <svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
              </svg>
            </div>
            <div class="ml-3">
              <h3 class="text-sm font-medium text-red-800 dark:text-red-200">Please fix the following errors:</h3>
              <div class="mt-2 text-sm text-red-700 dark:text-red-300">
                <ul class="list-disc pl-5 space-y-1">
                  <li v-for="(value, key) in errors" :key="key">{{ value }}</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <!-- Form -->
        <form @submit.prevent="saveProduct" class="space-y-6">
          <!-- Product Name -->
          <div>
            <label for="product_name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Product Name <span class="text-red-500">*</span>
            </label>
            <input
              id="product_name"
              v-model="productData.product_name"
              @blur="validateField('product_name', productData.product_name)"
              type="text"
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.product_name && touched.product_name}"
            />
            <p v-if="errors.product_name && touched.product_name" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.product_name }}
            </p>
          </div>

          <!-- Model -->
          <div>
            <label for="model" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Model <span class="text-red-500">*</span>
            </label>
            <input
              id="model"
              v-model="productData.model"
              @blur="validateField('model', productData.model)"
              type="text"
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.model && touched.model}"
            />
            <p v-if="errors.model && touched.model" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.model }}
            </p>
          </div>

          <!-- Price and Warranty Period in a row -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label for="price" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Price <span class="text-red-500">*</span>
              </label>
              <div class="mt-1 relative rounded-md shadow-sm">
                <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <span class="text-gray-500 sm:text-sm">₱</span>
                </div>
                <input
                  id="price"
                  v-model.number="productData.price"
                  @blur="validateField('price', productData.price)"
                  type="number"
                  min="0"
                  step="0.01"
                  class="block w-full px-3 py-2 pl-7 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.price && touched.price}"
                />
              </div>
              <p v-if="errors.price && touched.price" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.price }}
              </p>
            </div>

            <div>
              <label for="warranty_period" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Warranty Period (months) <span class="text-red-500">*</span>
              </label>
              <input
                id="warranty_period"
                v-model.number="productData.warranty_period"
                @blur="validateField('warranty_period', productData.warranty_period)"
                type="number"
                min="0"
                class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.warranty_period && touched.warranty_period}"
              />
              <p v-if="errors.warranty_period && touched.warranty_period" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.warranty_period }}
              </p>
            </div>
          </div>

          <!-- Description -->
          <div>
            <label for="description" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
              Description <span class="text-red-500">*</span>
            </label>
            <textarea
              id="description"
              v-model="productData.description"
              @blur="validateField('description', productData.description)"
              rows="3"
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.description && touched.description}"
            ></textarea>
            <p v-if="errors.description && touched.description" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.description }}
            </p>
          </div>

          <!-- Technical Specifications -->
          <div>
            <div class="flex justify-between items-center mb-2">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Technical Specifications <span class="text-red-500">*</span>
              </label>
              <button
                type="button"
                @click="addSpecRow"
                class="px-2 py-1 text-xs text-white bg-blue-600 rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
              >
                Add Specification
              </button>
            </div>

            <div v-for="(spec, index) in techSpecs" :key="index" class="flex items-center space-x-2 mb-2">
              <input
                v-model="spec.key"
                @blur="handleTechnicalSpecsBlur"
                type="text"
                placeholder="Key"
                class="block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.technical_specs && touched.technical_specs}"
              />
              <input
                v-model="spec.value"
                @blur="handleTechnicalSpecsBlur"
                type="text"
                placeholder="Value"
                class="block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.technical_specs && touched.technical_specs}"
              />
              <button
                type="button"
                @click="removeSpecRow(index)"
                class="text-red-600 hover:text-red-800 focus:outline-none"
                :disabled="techSpecs.length <= 1"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                </svg>
              </button>
            </div>

            <p v-if="errors.technical_specs && touched.technical_specs" class="mt-2 text-sm text-red-600 dark:text-red-400">
              {{ errors.technical_specs }}
            </p>
          </div>

          <!-- Certifications and Safety Standards -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label for="certifications" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Certifications <span class="text-red-500">*</span>
              </label>
              <input
                id="certifications"
                v-model="productData.certifications"
                @blur="validateField('certifications', productData.certifications)"
                type="text"
                class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.certifications && touched.certifications}"
              />
              <p v-if="errors.certifications && touched.certifications" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.certifications }}
              </p>
            </div>

            <div>
              <label for="safety_standards" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                Safety Standards <span class="text-red-500">*</span>
              </label>
              <input
                id="safety_standards"
                v-model="productData.safety_standards"
                @blur="validateField('safety_standards', productData.safety_standards)"
                type="text"
                class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                :class="{'border-red-300 dark:border-red-500 bg-red-50 dark:bg-red-900 dark:bg-opacity-20': errors.safety_standards && touched.safety_standards}"
              />
              <p v-if="errors.safety_standards && touched.safety_standards" class="mt-2 text-sm text-red-600 dark:text-red-400">
                {{ errors.safety_standards }}
              </p>
            </div>
          </div>

          <!-- Footer with buttons -->
          <div class="mt-8 pt-6 border-t border-gray-200 dark:border-gray-700 flex justify-end space-x-3">
            <button
              type="button"
              @click="closeModal"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="isLoading"
              class="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
              :class="{'opacity-75 cursor-not-allowed': isLoading}"
            >
              <span v-if="isLoading" class="flex items-center">
                <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Processing...
              </span>
              <span v-else>{{ productData.product_id ? 'Update Product' : 'Create Product' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease-out;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.modal-content {
  animation: modal-pop 0.3s ease-out forwards;
}

@keyframes modal-pop {
  0% {
    transform: scale(0.95);
    opacity: 0;
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

/* Hide number input spinners */
input[type=number]::-webkit-inner-spin-button,
input[type=number]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
input[type=number] {
  -moz-appearance: textfield;
  appearance: textfield;
}
</style>

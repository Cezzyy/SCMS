<script setup lang="ts">
import { defineProps, defineEmits } from 'vue';

const props = defineProps<{
  show: boolean;
  title: string;
  message: string;
  confirmButtonText?: string;
  cancelButtonText?: string;
  confirmButtonClass?: string;
}>();

const emit = defineEmits<{
  'update:show': [value: boolean];
  'confirm': [];
  'cancel': [];
}>();

const closeModal = () => {
  emit('update:show', false);
  emit('cancel');
};

const confirmAction = () => {
  emit('confirm');
  emit('update:show', false);
};
</script>

<template>
  <div v-if="show" class="fixed inset-0 z-50 overflow-y-auto">
    <!-- Backdrop -->
    <div class="fixed inset-0 bg-black bg-opacity-50 transition-opacity" @click="closeModal"></div>

    <!-- Modal content -->
    <div class="flex items-center justify-center min-h-screen p-4">
      <div class="relative bg-white dark:bg-gray-800 rounded-lg max-w-md w-full mx-auto shadow-xl transform transition-all">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 class="text-lg font-medium text-gray-900 dark:text-white">{{ title }}</h3>
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

        <!-- Content -->
        <div class="px-6 py-4">
          <p class="text-gray-700 dark:text-gray-300">{{ message }}</p>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 bg-gray-50 dark:bg-gray-700 border-t border-gray-200 dark:border-gray-600 flex justify-end space-x-3 rounded-b-lg">
          <button
            @click="closeModal"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500 sm:text-sm transition-colors"
          >
            {{ cancelButtonText || 'Cancel' }}
          </button>
          <button
            @click="confirmAction"
            :class="[
              'px-4 py-2 text-white rounded-md focus:outline-none focus:ring-2 sm:text-sm transition-colors',
              confirmButtonClass || 'bg-red-600 hover:bg-red-700 focus:ring-red-500'
            ]"
          >
            {{ confirmButtonText || 'Confirm' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Smooth transitions */
.fixed {
  transition: all 0.3s ease;
}
</style>

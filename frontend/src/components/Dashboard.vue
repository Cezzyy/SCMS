<script setup lang="ts">
import { ref } from 'vue';
import RevenueChart from './RevenueChart.vue';
</script>

<template>
  <div class="p-6">
    <h1 class="text-2xl font-semibold text-gray-800 mb-6">Dashboard</h1>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-sm text-gray-500">Total Revenue</p>
            <h3 class="text-2xl font-bold text-gray-800">₱52.6k</h3>
            <span class="text-xs text-green-500">+3.4%</span>
          </div>
          <button class="text-gray-400 hover:text-gray-600">...</button>
        </div>
      </div>

      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-sm text-gray-500">Today Revenue</p>
            <h3 class="text-2xl font-bold text-gray-800">₱12.87k</h3>
            <span class="text-xs text-red-500">-2.5%</span>
          </div>
          <button class="text-gray-400 hover:text-gray-600">...</button>
        </div>
      </div>

      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-sm text-gray-500">Total Products</p>
            <h3 class="text-2xl font-bold text-gray-800">236</h3>
          </div>
          <button class="text-gray-400 hover:text-gray-600">...</button>
        </div>
      </div>

      <div class="bg-white rounded-lg shadow p-4">
        <div class="flex justify-between items-start">
          <div>
            <p class="text-sm text-gray-500">Items Sold</p>
            <h3 class="text-2xl font-bold text-gray-800">22</h3>
          </div>
          <button class="text-gray-400 hover:text-gray-600">...</button>
        </div>
      </div>
    </div>

    <!-- Revenue Section -->
    <div class="bg-white rounded-lg shadow p-6 mb-6">
      <div class="flex justify-between items-center mb-4">
        <div>
          <h3 class="text-lg font-semibold text-gray-800">Revenue</h3>
          <div class="flex items-center">
            <span class="text-2xl font-bold text-gray-800">₱11,642</span>
            <span class="ml-2 text-xs text-green-500">+3.4% from last period</span>
          </div>
        </div>

        <div class="flex space-x-2">
          <button class="px-3 py-1 text-sm rounded-md bg-blue-500 text-white">Week</button>
          <button class="px-3 py-1 text-sm rounded-md text-gray-600 hover:bg-gray-100">Month</button>
          <button class="px-3 py-1 text-sm rounded-md text-gray-600 hover:bg-gray-100">Year</button>
        </div>
      </div>

      <div class="h-64 border-t pt-4">
        <RevenueChart />
      </div>
    </div>

    <!-- Recent Orders -->
    <div class="bg-white rounded-lg shadow overflow-hidden">
      <div class="flex justify-between items-center p-4 border-b">
        <h3 class="text-lg font-semibold text-gray-800">Recent Orders</h3>
        <span class="text-sm text-green-500">+1 new order</span>
      </div>

      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Item</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Qty</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Order Date</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Amount</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">I293DSA39</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">Iphone 16 Pro Max</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">4</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">April 18, 2025</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">₱84,990</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-blue-100 text-blue-800">Pending</span>
              </td>
            </tr>
            <tr>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">U2349SD12</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">Steam Deck 256GB</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">1</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">April 18, 2025</td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">₱25,950</td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">Approved</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

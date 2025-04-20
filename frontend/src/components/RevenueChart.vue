<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { Chart, registerables } from 'chart.js';

Chart.register(...registerables);

const chartCanvas = ref<HTMLCanvasElement | null>(null);
const chartInstance = ref<Chart | null>(null);


const revenueData = {
  labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'June', 'July'],
  datasets: [
    {
      label: 'Revenue',
      data: [5000, 12000, 19000, 15000, 22000, 18000, 26000],
      borderColor: '#3b82f6',
      backgroundColor: 'rgba(59, 130, 246, 0.1)',
      borderWidth: 2,
      fill: true,
      tension: 0.4,
      pointBackgroundColor: '#3b82f6',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
      pointRadius: 4,
      pointHoverRadius: 6,
    },
  ],
};

onMounted(() => {
  if (chartCanvas.value) {

    chartInstance.value = new Chart(chartCanvas.value, {
      type: 'line',
      data: revenueData,
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          tooltip: {
            backgroundColor: 'white',
            titleColor: '#111827',
            bodyColor: '#111827',
            borderColor: '#e5e7eb',
            borderWidth: 1,
            padding: 12,
            displayColors: false,
            callbacks: {
              title: (tooltipItems) => {
                return 'Revenue';
              },
              label: (context) => {
                return `₱${context.parsed.y.toLocaleString()}`;
              },
              afterLabel: (context) => {
                return `${revenueData.labels[context.dataIndex]}`;
              },
            },
          },
          legend: {
            display: false,
          },
        },
        scales: {
          x: {
            grid: {
              display: false,
            },
            ticks: {
              color: '#6b7280',
            },
          },
          y: {
            grid: {
              color: '#e5e7eb',
            },
            ticks: {
              color: '#6b7280',
              callback: (value) => {
                if (value === undefined || value === null || isNaN(Number(value))) {
                  return '';
                }
                return `₱${Number(value)/1000}k`;
              },
            },
            beginAtZero: true,
          },
        },
      },
    });
  }
});
</script>

<template>
  <div class="h-full w-full">
    <canvas ref="chartCanvas"></canvas>
  </div>
</template>

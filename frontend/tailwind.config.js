/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: '#0C2C33',
        accent1: '#00B9D8',
        accent2: '#1C7FAE',
        'text-primary': '#FFFFFF',
        'text-secondary': '#CCCCCC',
        'bg-alt': '#F5F7FA'
      }
    },
  },
  plugins: [],
}

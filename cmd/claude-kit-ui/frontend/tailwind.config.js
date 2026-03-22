/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{svelte,js,ts}', './index.html'],
  theme: {
    extend: {
      colors: {
        ck: {
          pink: '#FF69B4',
          rose: '#E91E8B',
          gold: '#FFD700',
          green: '#00FF87',
          dim: '#6C6C6C',
          dark: '#1a1a2e',
          bg: '#0d1117'
        }
      }
    }
  },
  plugins: []
}

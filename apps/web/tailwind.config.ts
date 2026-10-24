import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  content: [
    './components/**/*.{vue,js,ts}',
    './layouts/**/*.vue',
    './pages/**/*.vue',
    './app.vue'
  ],
  theme: {
    extend: {
      colors: {
        forest: {
          50: '#F0FDF4',
          100: '#DCFCE7',
          500: '#10B981',
          600: '#059669',
          700: '#0F766E',
          800: '#115E59',
          900: '#134E4A'
        },
        wiki: {
          link: '#0284C7',
          redlink: '#DC2626',
          darkbg: '#0B192C'
        }
      }
    }
  },
  plugins: []
} satisfies Config

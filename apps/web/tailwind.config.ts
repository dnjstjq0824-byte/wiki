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
          link: '#0645AD',
          visited: '#0B0080',
          redlink: '#BA0000',
          pagebg: '#E9ECEF',
          border: '#C8CCD1',
          notice: '#FEF6E7',
          darkbg: '#0B192C'
        },
        namu: {
          900: '#00473F',
          800: '#00564C',
          700: '#00685C',
          600: '#008275',
          500: '#00A08A',
          100: '#D7F0EB',
          50: '#EFF8F6'
        }
      },
      fontFamily: {
        wiki: ['Georgia', "'Times New Roman'", 'Times', 'serif']
      }
    }
  },
  plugins: []
} satisfies Config


/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Inter"', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"SF Mono"', '"Fira Code"', 'monospace'],
      },
      colors: {
        bg: '#000000',
        surface: '#080808',
        'surface-2': '#0e0e0e',
        'surface-3': '#141414',
        border: '#1a1a1a',
        'border-2': '#262626',
        fg: '#e0e0e0',
        'fg-2': '#888888',
        'fg-3': '#555555',
        'fg-4': '#333333',
      },
    },
  },
  plugins: [],
}

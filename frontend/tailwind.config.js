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
        n: {
          0: 'rgb(var(--n0-rgb) / <alpha-value>)',
          1: 'rgb(var(--n1-rgb) / <alpha-value>)',
          2: 'rgb(var(--n2-rgb) / <alpha-value>)',
          3: 'rgb(var(--n3-rgb) / <alpha-value>)',
          4: 'rgb(var(--n4-rgb) / <alpha-value>)',
          5: 'rgb(var(--n5-rgb) / <alpha-value>)',
          6: 'rgb(var(--n6-rgb) / <alpha-value>)',
          7: 'rgb(var(--n7-rgb) / <alpha-value>)',
          8: 'rgb(var(--n8-rgb) / <alpha-value>)',
          9: 'rgb(var(--n9-rgb) / <alpha-value>)',
          10: 'rgb(var(--n10-rgb) / <alpha-value>)',
        },
      },
    },
  },
  plugins: [],
}

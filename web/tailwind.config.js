/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        velux: {
          blue: '#0066B3',
          dark: '#004C8C',
          light: '#E8F4FC',
        }
      }
    },
  },
  plugins: [],
}

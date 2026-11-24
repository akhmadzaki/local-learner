/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#e8f0fe",
          100: "#c7d9f7",
          200: "#9bb4e8",
          300: "#6d8fd6",
          400: "#4a6ccf",
          500: "#2b4dbf",
          600: "#1a3a9e",
          700: "#122b7a",
          800: "#0b1d56",
          900: "#061238",
        },
        surface: {
          50: "#f8f9fb",
          100: "#eef0f4",
          200: "#e1e4ea",
          300: "#cdd2dc",
          400: "#b4bac6",
          500: "#9ca2ae",
          600: "#7a808e",
          700: "#5a5f6e",
          800: "#3c404e",
          900: "#1e212e",
        },
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
        mono: ["JetBrains Mono", "Fira Code", "monospace"],
      },
    },
  },
  plugins: [],
};
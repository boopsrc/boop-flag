/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        boop: {
          50: "#f2f0ff",
          100: "#e9e5ff",
          200: "#d6ccff",
          300: "#b7a4ff",
          400: "#9470ff",
          500: "#7442ff",
          600: "#661ff7",
          700: "#5713e3",
          800: "#4813bf",
          900: "#3d139c",
        },
      },
      fontFamily: {
        display: ["Nunito", "system-ui", "sans-serif"],
      },
      keyframes: {
        "gradient-pan": {
          "0%, 100%": { backgroundPosition: "0% 50%" },
          "50%": { backgroundPosition: "100% 50%" },
        },
      },
      animation: {
        "gradient-pan": "gradient-pan 12s ease infinite",
      },
    },
  },
  plugins: [],
};

import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        primary: {
          50:  "#f0fdf4",
          100: "#dcfce7",
          200: "#bbf7d0",
          300: "#86efac",
          400: "#4ade80",
          500: "#22c55e",
          600: "#16a34a",
          700: "#15803d",
          800: "#166534",
          900: "#14532d",
        },
        sidebar: "#0f172a",
        surface: "var(--surface)",
        card:    "var(--card)",
        border:  "var(--border)",
        muted:   "var(--muted)",
        ink:     "var(--ink)",
      },
      fontFamily: {
        sans: ["Plus Jakarta Sans", "system-ui", "sans-serif"],
      },
      boxShadow: {
        card:       "0 1px 4px 0 rgb(0 0 0 / .06)",
        "card-hover": "0 6px 24px 0 rgb(0 0 0 / .12)",
      },
    },
  },
  plugins: [],
};

export default config;

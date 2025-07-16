/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
    "./src/style.css",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
        // Semantic status colors
        status: {
          success: {
            DEFAULT: '#10b981',
            light: '#d1fae5',
            dark: '#059669',
          },
          warning: {
            DEFAULT: '#f59e0b',
            light: '#fef3c7',
            dark: '#d97706',
          },
          error: {
            DEFAULT: '#ef4444',
            light: '#fee2e2',
            dark: '#dc2626',
          },
          info: {
            DEFAULT: '#3b82f6',
            light: '#dbeafe',
            dark: '#2563eb',
          },
        },
        // Surface colors for consistent backgrounds
        surface: {
          card: '#ffffff',
          background: '#f9fafb',
          muted: '#f3f4f6',
          hover: '#f9fafb',
        },
        // Text colors for consistency
        content: {
          primary: '#111827',
          secondary: '#6b7280',
          muted: '#9ca3af',
          inverse: 'white',
        },
      },
      spacing: {
        // Consistent spacing scale
        'card': '1.5rem',
        'section': '2rem',
      },
      borderRadius: {
        'card': '0.5rem',
      },
      boxShadow: {
        'card': '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
        'card-hover': '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
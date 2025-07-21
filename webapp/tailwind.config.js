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
        // Modern dark/neutral color palette
        neutral: {
          50: '#fafafa',
          100: '#f4f4f5',
          150: '#ebebec',
          200: '#e4e4e7',
          250: '#d9d9dc',
          300: '#d4d4d8',
          350: '#b8b8bc',
          400: '#a1a1aa',
          450: '#8e8e95',
          500: '#71717a',
          550: '#5e5e66',
          600: '#52525b',
          650: '#46464f',
          700: '#3f3f46',
          750: '#35353b',
          800: '#27272a',
          850: '#202023',
          900: '#18181b',
          950: '#0f0f11',
        },
        // Accent colors with better contrast
        primary: {
          DEFAULT: '#6366f1',
          hover: '#5558e8',
          active: '#4f46e5',
          light: '#818cf8',
          lighter: '#a5b4fc',
          dark: '#4338ca',
          darker: '#3730a3',
        },
        // Status colors optimized for dark backgrounds
        status: {
          success: {
            DEFAULT: '#22c55e',
            light: '#4ade80',
            dark: '#16a34a',
            bg: 'rgba(34, 197, 94, 0.1)',
            border: 'rgba(34, 197, 94, 0.2)',
          },
          warning: {
            DEFAULT: '#f59e0b',
            light: '#fbbf24',
            dark: '#d97706',
            bg: 'rgba(245, 158, 11, 0.1)',
            border: 'rgba(245, 158, 11, 0.2)',
          },
          error: {
            DEFAULT: '#ef4444',
            light: '#f87171',
            dark: '#dc2626',
            bg: 'rgba(239, 68, 68, 0.1)',
            border: 'rgba(239, 68, 68, 0.2)',
          },
          info: {
            DEFAULT: '#3b82f6',
            light: '#60a5fa',
            dark: '#2563eb',
            bg: 'rgba(59, 130, 246, 0.1)',
            border: 'rgba(59, 130, 246, 0.2)',
          },
        },
        // UI Surface colors
        surface: {
          DEFAULT: '#18181b',
          secondary: '#202023',
          tertiary: '#27272a',
          elevated: '#2a2a2d',
          overlay: 'rgba(0, 0, 0, 0.5)',
        },
        // Content colors
        content: {
          DEFAULT: '#fafafa',
          secondary: '#a1a1aa',
          tertiary: '#71717a',
          muted: '#52525b',
          inverse: '#18181b',
        },
        // Border colors
        border: {
          DEFAULT: 'rgba(244, 244, 245, 0.1)',
          strong: 'rgba(244, 244, 245, 0.2)',
          interactive: 'rgba(99, 102, 241, 0.4)',
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
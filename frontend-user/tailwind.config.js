/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        bg: '#070b09',
        panel: '#101714',
        line: '#24352c',
        acid: '#c6ff3d',
        cyan: '#3ee0c4',
        amber: '#ffb020',
        rose: '#ff5d73',
        fog: '#8aa396',
        paper: '#e8f5e4',
      },
      fontFamily: {
        display: ['Syne', 'sans-serif'],
        sans: ['IBM Plex Sans', 'sans-serif'],
        mono: ['IBM Plex Mono', 'ui-monospace', 'monospace'],
      },
    },
  },
  plugins: [],
}

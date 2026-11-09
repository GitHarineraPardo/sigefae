import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  base: '/sigefae/Frontend/',
  build: {
    chunkSizeWarningLimit: 1400,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return;
          if (id.includes('/react/') || id.includes('/react-dom/') || id.includes('react-router')) {
            return 'react';
          }
          if (id.includes('pdfjs-dist')) return 'pdfjs';
          if (id.includes('jspdf') || id.includes('pdf-lib')) return 'pdf';
          if (id.includes('@fortawesome')) return 'icons';
        },
      },
    },
  },
})

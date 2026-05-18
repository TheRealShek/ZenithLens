import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:38471',
      '/media': 'http://localhost:38471',
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
});

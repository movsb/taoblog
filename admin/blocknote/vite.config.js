import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks: ()=>'everything.js',
        entryFileNames: 'blocknote.js',
        chunkFileNames: 'blocknote.js',
        assetFileNames: (assetInfo) => {
          if (assetInfo.name && assetInfo.name.endsWith('.css')) {
            return 'blocknote.css';
          }
          return 'assets/[name][extname]';
        }
      }
    }
  },
})

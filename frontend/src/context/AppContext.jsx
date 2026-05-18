import { createContext, useState, useEffect, useCallback } from 'react';
import { fetchAPI } from '../api';

export const AppContext = createContext();

export function AppProvider({ children }) {
  const [folders, setFolders] = useState([]);
  const [favorites, setFavorites] = useState(new Set());
  const [scanning, setScanning] = useState(false);

  const loadFolders = useCallback(async () => {
    const { data } = await fetchAPI('/api/folders');
    if (!data) return;
    setFolders(data);
    const anyScanning = data.some(f => f.scanning);
    setScanning(anyScanning);
  }, []);

  const loadFavorites = useCallback(async () => {
    const { data } = await fetchAPI('/api/favorites?limit=10000');
    if (data?.items) {
      setFavorites(new Set(data.items.map(i => i.path)));
    }
  }, []);

  const toggleFavorite = useCallback(async (path) => {
    if (!path) return;
    if (favorites.has(path)) {
      await fetchAPI(`/api/favorites?path=${encodeURIComponent(path)}`, { method: 'DELETE' });
      setFavorites(prev => { const next = new Set(prev); next.delete(path); return next; });
    } else {
      await fetchAPI('/api/favorites', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path }),
      });
      setFavorites(prev => new Set(prev).add(path));
    }
  }, [favorites]);

  // Polling when scanning.
  useEffect(() => {
    if (!scanning) return;
    const id = setInterval(async () => {
      const { data } = await fetchAPI('/api/folders');
      if (!data) return;
      setFolders(data);
      if (!data.some(f => f.scanning)) setScanning(false);
    }, 2000);
    return () => clearInterval(id);
  }, [scanning]);

  // Initial load.
  useEffect(() => {
    loadFolders();
    loadFavorites();
  }, [loadFolders, loadFavorites]);

  return (
    <AppContext.Provider value={{
      folders, loadFolders,
      favorites, toggleFavorite, loadFavorites,
      scanning, setScanning,
    }}>
      {children}
    </AppContext.Provider>
  );
}

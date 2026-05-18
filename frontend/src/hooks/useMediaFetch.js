import { useState, useEffect } from 'react';
import { fetchAPI } from '../api';

function buildEndpoint(route, page, folderId, searchQuery, seed, limit) {
  switch (route) {
    case 'home':
      return `/api/home?page=${page}&limit=${limit}&seed=${seed}`;
    case 'folder':
      return `/api/folder/${folderId}?page=${page}&limit=${limit}&seed=${seed}`;
    case 'favorites':
      return `/api/favorites?page=${page}&limit=${limit}`;
    case 'search':
      return `/api/search?q=${encodeURIComponent(searchQuery)}&page=${page}&limit=${limit}&seed=${seed}`;
    default:
      return `/api/home?page=${page}&limit=${limit}&seed=${seed}`;
  }
}

export function useMediaFetch(route, page, folderId, searchQuery, seed, limit) {
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      setLoading(true);
      const url = buildEndpoint(route, page, folderId, searchQuery, seed, limit);
      const { data } = await fetchAPI(url);
      if (cancelled) return;
      setItems(data?.items || []);
      setTotal(data?.total || 0);
      setLoading(false);
    }

    load();
    return () => { cancelled = true; };
  }, [route, page, folderId, searchQuery, seed, limit]);

  return { items, total, loading };
}

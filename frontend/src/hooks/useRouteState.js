import { useState, useEffect, useCallback } from 'react';

function parseURL() {
  const p = new URLSearchParams(window.location.search);
  return {
    route: p.get('route') || 'home',
    page: parseInt(p.get('page')) || 1,
    folderId: p.get('folder_id') || null,
    searchQuery: p.get('q') || null,
  };
}

function buildURL(route, page, folderId, searchQuery, seed) {
  const p = new URLSearchParams();
  p.set('route', route);
  if (page > 1) p.set('page', page);
  if (folderId) p.set('folder_id', folderId);
  if (searchQuery) p.set('q', searchQuery);
  p.set('seed', seed);
  return '?' + p.toString();
}

export function useRouteState() {
  const [seed] = useState(() => Date.now());
  const [route, setRoute] = useState('home');
  const [page, setPage] = useState(1);
  const [folderId, setFolderId] = useState(null);
  const [searchQuery, setSearchQuery] = useState(null);

  // Hydrate from URL on mount.
  useEffect(() => {
    const s = parseURL();
    setRoute(s.route);
    setPage(s.page);
    setFolderId(s.folderId);
    setSearchQuery(s.searchQuery);
  }, []);

  // Listen for back/forward navigation.
  useEffect(() => {
    const handler = () => {
      const s = parseURL();
      setRoute(s.route);
      setPage(s.page);
      setFolderId(s.folderId);
      setSearchQuery(s.searchQuery);
    };
    window.addEventListener('popstate', handler);
    return () => window.removeEventListener('popstate', handler);
  }, []);

  const pushURL = useCallback((r, p, fid, q) => {
    window.history.pushState(null, '', buildURL(r, p, fid, q, seed));
  }, [seed]);

  const navigate = useCallback((newRoute, newFolderId = null) => {
    setRoute(newRoute);
    setPage(1);
    setFolderId(newFolderId);
    setSearchQuery(null);
    pushURL(newRoute, 1, newFolderId, null);
  }, [pushURL]);

  const goToPage = useCallback((p) => {
    setPage(p);
    pushURL(route, p, folderId, searchQuery);
  }, [pushURL, route, folderId, searchQuery]);

  const search = useCallback((q) => {
    if (!q) {
      navigate('home');
      return;
    }
    setRoute('search');
    setPage(1);
    setSearchQuery(q);
    pushURL('search', 1, folderId, q);
  }, [pushURL, folderId, navigate]);

  return { route, page, folderId, searchQuery, seed, navigate, goToPage, search };
}

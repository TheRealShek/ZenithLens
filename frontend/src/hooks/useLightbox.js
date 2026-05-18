import { useState, useEffect, useCallback } from 'react';

function downloadItem(item) {
  if (!item) return;
  const a = document.createElement('a');
  a.href = `/media/file?path=${encodeURIComponent(item.path)}`;
  a.download = item.name;
  document.body.appendChild(a);
  a.click();
  a.remove();
}

export function useLightbox(items, onToggleFavorite) {
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(null);

  const open = useCallback((index) => {
    setActiveIndex(index);
    setIsOpen(true);
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
    setActiveIndex(null);
  }, []);

  const prev = useCallback(() => {
    setActiveIndex(i => Math.max(i - 1, 0));
  }, []);

  const next = useCallback(() => {
    setActiveIndex(i => Math.min(i + 1, items.length - 1));
  }, [items.length]);

  // Keyboard bindings — only active when lightbox is open.
  useEffect(() => {
    if (!isOpen) return;

    function handleKey(e) {
      switch (e.key) {
        case 'ArrowRight': next(); break;
        case 'ArrowLeft': prev(); break;
        case 'Escape': close(); break;
        case 'f': case 'F':
          if (items[activeIndex]) onToggleFavorite(items[activeIndex].path);
          break;
        case 'd': case 'D':
          downloadItem(items[activeIndex]);
          break;
      }
    }

    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [isOpen, activeIndex, items, onToggleFavorite, next, prev, close]);

  const currentItem = isOpen && activeIndex !== null ? items[activeIndex] : null;

  return { isOpen, activeIndex, open, close, prev, next, currentItem };
}

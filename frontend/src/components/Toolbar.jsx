import { useState, useRef, useEffect } from 'react';
import './Toolbar.css';

export default function Toolbar({ onSearch, total, page, limit, loading, searchQuery }) {
  const [query, setQuery] = useState(searchQuery || '');
  const timerRef = useRef(null);

  // Sync local input with external route state (URL hydration, back/forward, sidebar nav).
  useEffect(() => {
    setQuery(searchQuery || '');
  }, [searchQuery]);

  function handleInput(e) {
    const val = e.target.value;
    setQuery(val);
    clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => onSearch(val.trim()), 300);
  }

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="toolbar">
      <input
        className="search-input"
        type="text"
        placeholder="Search files..."
        value={query}
        onChange={handleInput}
      />
      <div className="view-info">
        {loading ? 'Loading...' : `${total} files${totalPages > 1 ? ` • Page ${page}/${totalPages}` : ''}`}
      </div>
    </div>
  );
}

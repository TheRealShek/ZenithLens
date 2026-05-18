import { useState, useRef } from 'react';
import './Toolbar.css';

export default function Toolbar({ onSearch, total, page, limit, loading }) {
  const [query, setQuery] = useState('');
  const timerRef = useRef(null);

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

import { useContext } from 'react';
import { AppContext } from '../context/AppContext';
import { fetchAPI } from '../api';
import './Sidebar.css';

export default function Sidebar({ route, folderId, onNavigate, onAddFolder }) {
  const { folders, loadFolders, setScanning } = useContext(AppContext);

  async function handleRescan(e, id) {
    e.stopPropagation();
    const { error } = await fetchAPI(`/api/folders/${id}/rescan`, { method: 'POST' });
    if (error) return alert(error);
    setScanning(true);
    loadFolders();
  }

  async function handleDelete(e, id) {
    e.stopPropagation();
    if (!confirm('Remove this folder?')) return;
    await fetchAPI(`/api/folders/${id}`, { method: 'DELETE' });
    loadFolders();
    if (folderId === id) onNavigate('home');
  }

  return (
    <aside className="sidebar">
      <h1 className="sidebar-title">ZenithLens</h1>
      <nav className="sidebar-nav">
        <div
          className={`nav-item ${route === 'home' ? 'active' : ''}`}
          onClick={() => onNavigate('home')}
        >
          <span className="icon">H</span> Home
        </div>
        <div
          className={`nav-item ${route === 'favorites' ? 'active' : ''}`}
          onClick={() => onNavigate('favorites')}
        >
          <span className="icon">F</span> Favorites
        </div>

        <div className="folder-section-label">Folders</div>
        {folders.map(f => (
          <div
            key={f.id}
            className={`nav-item ${route === 'folder' && folderId === f.id ? 'active' : ''}`}
            onClick={() => onNavigate('folder', f.id)}
          >
            <span className="icon">D</span>
            <span className="folder-name">{f.name}</span>
            {f.scanning ? (
              <span className="scanning-badge">scanning</span>
            ) : (
              <span className="media-count">{f.media_count}</span>
            )}
            <span className="folder-actions">
              <button onClick={(e) => handleRescan(e, f.id)} title="Rescan">R</button>
              <button onClick={(e) => handleDelete(e, f.id)} title="Remove">X</button>
            </span>
          </div>
        ))}
      </nav>
      <button className="add-folder-btn" onClick={onAddFolder}>+ Add Folder</button>
    </aside>
  );
}

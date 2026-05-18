import { useState, useContext } from 'react';
import { fetchAPI } from '../api';
import { AppContext } from '../context/AppContext';
import './AddFolderModal.css';

export default function AddFolderModal({ onClose }) {
  const [path, setPath] = useState('');
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [picking, setPicking] = useState(false);
  const { loadFolders, setScanning } = useContext(AppContext);

  async function handleChooseFolder() {
    setError('');
    setPicking(true);
    const { data, error: apiErr } = await fetchAPI('/api/folders/pick', {
      method: 'POST',
    });
    setPicking(false);
    if (apiErr) {
      setError(apiErr);
      return;
    }
    if (data?.path) setPath(data.path);
  }

  async function handleSubmit() {
    if (!path.trim()) return;
    setError('');
    const { data, error: apiErr } = await fetchAPI('/api/folders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: path.trim(), name: name.trim() }),
    });
    if (apiErr) { setError(apiErr); return; }
    setScanning(true);
    loadFolders();
    onClose();
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={e => e.stopPropagation()}>
        <h2>Add Folder</h2>
        <input
          placeholder="Folder path (e.g. /home/user/Pictures)"
          value={path}
          onChange={e => setPath(e.target.value)}
          autoFocus
        />
        <button
          className="modal-picker"
          onClick={handleChooseFolder}
          disabled={picking}
        >
          {picking ? 'Opening Folder Picker...' : 'Choose Folder'}
        </button>
        <input
          placeholder="Display name (optional)"
          value={name}
          onChange={e => setName(e.target.value)}
        />
        {error && <p className="modal-error">{error}</p>}
        <div className="modal-actions">
          <button className="modal-cancel" onClick={onClose}>Cancel</button>
          <button className="modal-confirm" onClick={handleSubmit}>Add</button>
        </div>
      </div>
    </div>
  );
}

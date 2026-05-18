import { useState } from 'react';
import './MediaItem.css';

export default function MediaItem({ item, isFavorite, onClick }) {
  const [imgError, setImgError] = useState(false);
  const thumbURL = `/media/thumb?path=${encodeURIComponent(item.path)}`;
  const fileURL = `/media/file?path=${encodeURIComponent(item.path)}`;
  const ext = item.name.split('.').pop()?.toUpperCase() || '?';

  // Fallback placeholder for unsupported or broken thumbnails.
  if (imgError && item.type !== 'video') {
    return (
      <div className="media-item" onClick={onClick}>
        <div className="placeholder">
          <span className="ext-badge">.{ext}</span>
          <span className="placeholder-name">{item.name}</span>
          <a className="dl-link" href={fileURL} download onClick={e => e.stopPropagation()}>Download</a>
        </div>
        {isFavorite && <span className="fav-indicator">❤️</span>}
      </div>
    );
  }

  return (
    <div className="media-item" onClick={onClick}>
      {imgError && item.type === 'video' ? (
        <video className="media-thumb-video" src={fileURL} muted preload="metadata" />
      ) : (
        <img
          className="media-thumb"
          src={thumbURL}
          alt={item.name}
          loading="lazy"
          onError={() => setImgError(true)}
        />
      )}
      {item.type === 'video' && !imgError && <span className="video-badge">▶</span>}
      {isFavorite && <span className="fav-indicator">❤️</span>}
    </div>
  );
}

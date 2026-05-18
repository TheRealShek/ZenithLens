import './Lightbox.css';

export default function Lightbox({ item, onClose, onPrev, onNext, hasPrev, hasNext, isFavorite, onToggleFavorite }) {
  if (!item) return null;

  const fileURL = `/media/file?path=${encodeURIComponent(item.path)}`;

  function handleDownload() {
    const a = document.createElement('a');
    a.href = fileURL;
    a.download = item.name;
    document.body.appendChild(a);
    a.click();
    a.remove();
  }

  return (
    <div className="lightbox" onClick={onClose}>
      <div className="lightbox-content" onClick={e => e.stopPropagation()}>
        <button className="lb-close" onClick={onClose}>✕</button>

        {hasPrev && <button className="lb-nav lb-nav-left" onClick={onPrev}>‹</button>}
        {hasNext && <button className="lb-nav lb-nav-right" onClick={onNext}>›</button>}

        {item.type === 'image' ? (
          <img className="lb-media" src={fileURL} alt={item.name} />
        ) : item.type === 'video' ? (
          <video className="lb-media" src={fileURL} controls autoPlay />
        ) : (
          <div className="lb-unsupported">
            <span className="ext-badge">.{item.name.split('.').pop()?.toUpperCase()}</span>
            <span>{item.name}</span>
            <button onClick={handleDownload}>Download</button>
          </div>
        )}

        <div className="lb-info">
          <span className="lb-name">{item.name}</span>
          <button
            className={`lb-fav-btn ${isFavorite ? 'active' : ''}`}
            onClick={onToggleFavorite}
            title="Toggle favorite (F)"
          >
            {isFavorite ? '❤️' : '♡'}
          </button>
          <button className="lb-dl-btn" onClick={handleDownload} title="Download (D)">⬇</button>
        </div>
      </div>
    </div>
  );
}

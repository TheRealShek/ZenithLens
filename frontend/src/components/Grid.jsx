import MediaItem from './MediaItem';
import './Grid.css';

export default function Grid({ items, favorites, onItemClick }) {
  if (!items.length) {
    return <div className="grid-empty">No media files found</div>;
  }

  return (
    <div className="grid">
      {items.map((item, i) => (
        <MediaItem
          key={item.path}
          item={item}
          isFavorite={favorites.has(item.path)}
          onClick={() => onItemClick(i)}
        />
      ))}
    </div>
  );
}

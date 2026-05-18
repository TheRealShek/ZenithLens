import { useState, useContext } from 'react';
import { useRouteState } from './hooks/useRouteState';
import { useMediaFetch } from './hooks/useMediaFetch';
import { useLightbox } from './hooks/useLightbox';
import { AppContext } from './context/AppContext';
import Sidebar from './components/Sidebar';
import Toolbar from './components/Toolbar';
import Grid from './components/Grid';
import Pagination from './components/Pagination';
import Lightbox from './components/Lightbox';
import AddFolderModal from './components/AddFolderModal';

const LIMIT = 50;

export default function App() {
  const { route, page, folderId, searchQuery, seed, navigate, goToPage, search } = useRouteState();
  const { items, total, loading } = useMediaFetch(route, page, folderId, searchQuery, seed, LIMIT);
  const { favorites, toggleFavorite } = useContext(AppContext);
  const lightbox = useLightbox(items, toggleFavorite);
  const [modalOpen, setModalOpen] = useState(false);

  return (
    <div className="app">
      <Sidebar
        route={route}
        folderId={folderId}
        onNavigate={navigate}
        onAddFolder={() => setModalOpen(true)}
      />
      <main className="main-content">
        <Toolbar onSearch={search} total={total} page={page} limit={LIMIT} loading={loading} searchQuery={searchQuery} />
        <div className="grid-container">
          <Grid items={items} favorites={favorites} onItemClick={lightbox.open} />
        </div>
        <Pagination page={page} total={total} limit={LIMIT} onPageChange={goToPage} />
      </main>
      {lightbox.isOpen && (
        <Lightbox
          item={lightbox.currentItem}
          onClose={lightbox.close}
          onPrev={lightbox.prev}
          onNext={lightbox.next}
          hasPrev={lightbox.activeIndex > 0}
          hasNext={lightbox.activeIndex < items.length - 1}
          isFavorite={favorites.has(lightbox.currentItem?.path)}
          onToggleFavorite={() => toggleFavorite(lightbox.currentItem?.path)}
        />
      )}
      {modalOpen && <AddFolderModal onClose={() => setModalOpen(false)} />}
    </div>
  );
}

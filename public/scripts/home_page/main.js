    // ------- Helpers -------
    const fmtDuration = (sec = 0) => {
      sec = Math.floor(sec || 0);
      const h = Math.floor(sec / 3600);
      const m = Math.floor((sec % 3600) / 60);
      return h ? `${h}h ${m}m` : `${m}m`;
    };

    const fmtRes = (w, h) => (w && h) ? `${w}×${h}` : '—';
    
    const posterPlaceholder = `data:image/svg+xml;charset=utf-8,${encodeURIComponent(`
      <svg xmlns="http://www.w3.org/2000/svg" width="400" height="600" viewBox="0 0 400 600">
        <defs>
          <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" style="stop-color:#1a1a1a"/>
            <stop offset="100%" style="stop-color:#2a2a2a"/>
          </linearGradient>
        </defs>
        <rect width="100%" height="100%" fill="url(#bg)"/>
        <circle cx="200" cy="250" r="50" fill="rgba(255,255,255,0.1)"/>
        <polygon points="180,230 180,270 220,250" fill="rgba(255,255,255,0.2)"/>
        <text x="200" y="400" text-anchor="middle" font-family="Inter, sans-serif" 
              font-size="20" fill="rgba(255,255,255,0.3)" font-weight="500">Myt-V</text>
      </svg>
    `)}`;

    // ------- State -------
    const state = {
      all: [],
      page: 1,
      per: 24, // Aumenté el número por página ya que los pósters ocupan menos espacio horizontal
      query: ''
    };

    // ------- DOM Elements -------
    const elements = {
      searchInput: document.getElementById('searchInput'),
      stats: document.getElementById('stats'),
      movieGrid: document.getElementById('movieGrid'),
      prevButton: document.getElementById('prevButton'),
      nextButton: document.getElementById('nextButton'),
      pageInfo: document.getElementById('pageInfo')
    };

    // ------- Load Data -------
    async function loadMovies() {
      // Mostrar loading cards
      const loadingCards = Array(12).fill(0).map(() => '<div class="loading-card"></div>').join('');
      elements.movieGrid.innerHTML = loadingCards;
      
      try {
        const response = await fetch('/catalog', { cache: 'no-store' });
        
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        state.all = await response.json();
        render();
      } catch (error) {
        console.error('Error loading movies:', error);
        elements.movieGrid.innerHTML = `
          <div class="empty-state">
            <div class="empty-state-icon">⚠️</div>
            <h3>Error al cargar el catálogo</h3>
            <p>No se pudo conectar con el servidor.</p>
          </div>
        `;
      }
    }

    // ------- Filter & Pagination -------
    function getFilteredMovies() {
      const query = state.query.trim().toLowerCase();
      if (!query) return [...state.all];
      
      return state.all.filter(movie => 
        (movie.Title || '').toLowerCase().includes(query)
      );
    }

    function paginateMovies(movies) {
      const totalPages = Math.max(1, Math.ceil(movies.length / state.per));
      state.page = Math.min(state.page, totalPages);
      
      const start = (state.page - 1) * state.per;
      const pageMovies = movies.slice(start, start + state.per);
      
      return { movies: pageMovies, totalPages };
    }

    // ------- Create Movie Card -------
    function createMovieCard(movie) {
      const title = movie.Title || `Película ${movie.ID}`;
      const posterUrl = `/poster/${movie.ID}`;

      return `
        <a href="/watch/${movie.ID}" class="movie-card">
          <img
            src="${posterUrl}"
            alt="${title}"
            class="movie-poster"
            loading="lazy"
            onerror="this.onerror=null; this.src='${posterPlaceholder}';"
          />
          <div class="play-icon"></div>
          <div class="movie-overlay">
            <h3 class="movie-title">${title}</h3>
          </div>
        </a>
      `;
    }

    // ------- Render -------
    function render() {
      const filteredMovies = getFilteredMovies();
      const { movies: pageMovies, totalPages } = paginateMovies(filteredMovies);

      // Update stats
      const startItem = pageMovies.length > 0 ? ((state.page - 1) * state.per + 1) : 0;
      const endItem = (state.page - 1) * state.per + pageMovies.length;
      elements.stats.textContent = `${filteredMovies.length.toLocaleString()} películas • ${startItem}–${endItem}`;

      // Update grid
      if (pageMovies.length === 0) {
        elements.movieGrid.innerHTML = `
          <div class="empty-state">
            <div class="empty-state-icon">🎬</div>
            <h3>No se encontraron películas</h3>
            <p>Intenta con otros términos de búsqueda</p>
          </div>
        `;
      } else {
        elements.movieGrid.innerHTML = pageMovies.map(createMovieCard).join('');
      }

      // Update pagination
      elements.pageInfo.textContent = `${state.page} / ${totalPages}`;
      elements.prevButton.disabled = state.page <= 1;
      elements.nextButton.disabled = state.page >= totalPages;
    }

    // ------- Event Listeners -------
    let searchTimeout;
    elements.searchInput.addEventListener('input', (e) => {
      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(() => {
        state.query = e.target.value;
        state.page = 1;
        render();
      }, 300);
    });

    elements.prevButton.addEventListener('click', () => {
      if (state.page > 1) {
        state.page--;
        render();
        window.scrollTo({ top: 0, behavior: 'smooth' });
      }
    });

    elements.nextButton.addEventListener('click', () => {
      const totalPages = Math.max(1, Math.ceil(getFilteredMovies().length / state.per));
      if (state.page < totalPages) {
        state.page++;
        render();
        window.scrollTo({ top: 0, behavior: 'smooth' });
      }
    });

    // ------- Keyboard Navigation -------
    document.addEventListener('keydown', (e) => {
      if (e.target === elements.searchInput) return;
      
      switch(e.key) {
        case 'ArrowLeft':
          if (!elements.prevButton.disabled) {
            elements.prevButton.click();
          }
          break;
        case 'ArrowRight':
          if (!elements.nextButton.disabled) {
            elements.nextButton.click();
          }
          break;
        case '/':
          e.preventDefault();
          elements.searchInput.focus();
          break;
      }
    });

    // ------- Initialize -------
    loadMovies();
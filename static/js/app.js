const PLACEHOLDER_POSTER = "data:image/svg+xml;charset=utf-8," + encodeURIComponent(`
<svg xmlns="http://www.w3.org/2000/svg" width="300" height="450">
  <rect width="100%" height="100%" fill="#0f0f10"/>
  <rect x="18" y="18" width="264" height="414" rx="18" fill="#151518" stroke="rgba(255,255,255,0.12)"/>
  <text x="50%" y="52%" dominant-baseline="middle" text-anchor="middle"
        fill="rgba(255,255,255,0.55)" font-family="system-ui, -apple-system, Segoe UI, Roboto, Arial"
        font-size="26">No poster</text>
</svg>
`);
const HOME_LIMIT = 6;
function getUID() {
  return Number(document.body?.dataset?.uid || "0");
}
function getPage() {
  return String(document.body?.dataset?.page || "");
}
function getParam() {
  return String(document.body?.dataset?.param || "");
}

function renderGrid(gridEl, movies) {
  gridEl.innerHTML = (movies || []).map(m => `
    <a class="gridcard movie" href="/movie/${m.id}">
      <img src="${(m.poster_url || PLACEHOLDER_POSTER)}"
           onerror="this.onerror=null;this.src='${PLACEHOLDER_POSTER}'" />
      <div class="meta">
        <div class="title">${escapeHtml(normalizeTitle(m.title || "Untitled"))}</div>
        <div class="sub">${escapeHtml(m.year ?? "")}</div>
      </div>
    </a>
  `).join("");
}

async function initListPage() {
  const grid = document.getElementById("grid-list");
  if (!grid) return;

  const uid = getUID();
  const page = getPage();
  const param = getParam();

  try {
    let movies = [];

    if (page === "top200") {
      movies = await fetchJSON("/api/movies/top200");
    } else if (page === "watched") {
      movies = await fetchJSON(`/api/users/${uid}/watched`);
    } else if (page === "genre") {
      movies = await fetchJSON(`/api/movies/genre/${encodeURIComponent(param)}`);
    } else if (page === "genres") {
      const genres = await fetchJSON("/api/movies/genres");
      const clean = (genres || []).filter(g => g !== "(no genres listed)");
      grid.innerHTML = clean.map(g =>
        `<a class="genre-pill" href="/genre/${encodeURIComponent(g)}">${escapeHtml(g)}</a>`
      ).join("");
      return;
    } else if (page === "recommended") {
      movies = await fetchJSON(`/api/users/${uid}/recommend`);
    } else {
      grid.innerHTML = `<div class="muted">Unknown page</div>`;
      return;
    }

    if (!Array.isArray(movies) && movies && Array.isArray(movies.items)) {
      movies = movies.items;
    }

    if (!movies || movies.length === 0) {
      grid.innerHTML = `<div class="muted">Пока здесь не фильмов</div>`;
      return;
    }

    renderGrid(grid, movies);

  } catch (e) {
    console.error("list load error:", e);
    grid.innerHTML = `<div class="muted">Failed to load</div>`;
  }
}

function escapeHtml(s) {
  return String(s ?? "")
    .replaceAll("&","&amp;")
    .replaceAll("<","&lt;")
    .replaceAll(">","&gt;");
}

function debounce(fn, ms) {
  let t = null;
  return (...args) => {
    clearTimeout(t);
    t = setTimeout(() => fn(...args), ms);
  };
}

function normalizeTitle(title) {
  if (!title) return "";

  const idx = title.indexOf(",");
  if (idx === -1) return title;

  const main = title.slice(0, idx).trim();
  const rest = title.slice(idx + 1).trim();

  const match = rest.match(/^([A-Za-z']+)\b/);
  if (!match) return title;

  const article = match[1];

  const articles = ["The", "A", "An", "Le", "La", "Les"];

  if (articles.includes(article)) {
    return article + " " + main;
  }

  return title;
}

function movieCard(m) {
  const poster = m.poster_url || m.Poster || "";
  const title = normalizeTitle(m.title || "");
  const year = m.year ?? "";
  return `
    <a class="card movie" href="/movie/${m.id}">
      <img src="${poster || PLACEHOLDER_POSTER}" alt="${escapeHtml(title)}"
           onerror="this.onerror=null;this.src='${PLACEHOLDER_POSTER}'" />
      <div class="meta">
        <div class="title">${escapeHtml(title)}</div>
        <div class="sub">${escapeHtml(year)}</div>
      </div>
    </a>
  `;
}

async function fetchJSON(url, opts) {
  const res = await fetch(url, opts);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return await res.json();
}

/* menu */
function initMenu() {
  const btn = document.getElementById("menuBtn");
  const menu = document.getElementById("sideMenu");
  if (!btn || !menu) return;

  btn.addEventListener("click", () => {
    menu.classList.toggle("hidden");
  });

  document.addEventListener("click", (e) => {
    if (menu.classList.contains("hidden")) return;
    const inside = menu.contains(e.target) || btn.contains(e.target);
    if (!inside) menu.classList.add("hidden");
  });
}

/* home loader */
async function initHome() {
  const uid = getUID();
  const rowRec = document.getElementById("row-recommended");
  const rowTop = document.getElementById("row-top200");
  const rowChg = document.getElementById("row-changing");
  const genresRow = document.getElementById("genres-row");
  const changingTitleEl = document.getElementById("changingTitle");

  if (!uid) return;

  try {
    const data = await fetchJSON(`/api/users/${uid}/home`);

    const forYou = data.for_you || [];
    const top200 = data.top_200_pick || [];
    const changing = data.changing?.movies || [];
    if (changingTitleEl) {
      changingTitleEl.textContent = data.changing?.title || "Changing";
    }
    const genres = (data.genres || [])
      .filter(g => g !== "(no genres listed)");

    if (rowRec) {
      rowRec.innerHTML = forYou.slice(0,HOME_LIMIT).map(movieCard).join("");
    }

    if (rowTop) {
      rowTop.innerHTML = top200.slice(0,HOME_LIMIT).map(movieCard).join("");
    }

    if (rowChg) {
      rowChg.innerHTML = changing.slice(0,HOME_LIMIT).map(movieCard).join("");
    }

    if (genresRow) {
      genresRow.innerHTML = genres.map(g =>
        `<a class="genre-pill" href="/genre/${encodeURIComponent(g)}">${escapeHtml(g)}</a>`
      ).join("");
    }

  } catch (e) {
    console.error("home load error:", e);
  }
}

/* movie page loader + rating */
async function initMoviePage() {
  const el = document.getElementById("moviePage");
  if (!el) return;

  const uid = getUID();
  const movieID = el.getAttribute("data-movie-id");
  if (!uid || !movieID) return;

  const titleEl = document.getElementById("movieTitle");
  const metaEl = document.getElementById("movieMeta");
  const posterEl = document.getElementById("moviePoster");
  const plotEl = document.getElementById("moviePlot");
  const userRatingEl = document.getElementById("userRating");
  const similarRow = document.getElementById("row-similar");
  const imdbEl = document.getElementById("imdbRating");
  const runtimeEl = document.getElementById("runtime");
  const directorEl = document.getElementById("director");
  const actorsEl = document.getElementById("actors");
  const platformRatingEl = document.getElementById("platformRating");

  try {
    const data = await fetchJSON(`/api/users/${uid}/movies/${movieID}`);

    const m = data.movie || {};
    const details = data.details || {};
    const releasedEl = document.getElementById("released");
    const countryEl = document.getElementById("country");
    const languageEl = document.getElementById("language");
    const writerEl = document.getElementById("writer");
    const awardsEl = document.getElementById("awards");
    const boxofficeEl = document.getElementById("boxoffice");
    const ratingsListEl = document.getElementById("ratingsList");

    if (releasedEl) releasedEl.textContent = details.Released ?? "—";
    if (countryEl) countryEl.textContent = details.Country ?? "—";
    if (languageEl) languageEl.textContent = details.Language ?? "—";
    if (writerEl) writerEl.textContent = details.Writer ?? "—";
    if (awardsEl) awardsEl.textContent = details.Awards ?? "—";
    if (boxofficeEl) boxofficeEl.textContent = details.BoxOffice ?? "—";

    if (ratingsListEl) {
      const ratings = details.Ratings || [];

      let html = "";

      html += ratings.map(r =>
        `<div>${escapeHtml(r.Source)}: ${escapeHtml(r.Value)}</div>`
      ).join("");

      ratingsListEl.innerHTML = html || "—";
    }
    const poster = data.poster_url || details.Poster || "";

    if (titleEl) {
      const niceTitle = normalizeTitle(m.title || "Movie");
      titleEl.textContent = `${niceTitle}${m.year ? " ("+m.year+")" : ""}`;
    }
    if (metaEl) metaEl.textContent = (m.genres || []).join(" • ");

    if (posterEl) {
      posterEl.src = poster || PLACEHOLDER_POSTER;
      posterEl.onerror = () => { posterEl.src = PLACEHOLDER_POSTER; };
    }

    if (plotEl) plotEl.textContent = details.Plot || "N/A";
    if (userRatingEl) userRatingEl.textContent = (data.user_rating ?? "—");
    const existingRating = data.user_rating;

    if (existingRating !== null && existingRating !== undefined) {
      const controls = document.querySelector(".rate-controls");
      if (controls) {
        controls.remove(); 
      }
    }
    if (platformRatingEl) {
      const pr = data.platform_rating; 
      platformRatingEl.textContent = (pr === null || pr === undefined)
        ? "—"
        : Number(pr).toFixed(1);
    }

    if (imdbEl) imdbEl.textContent = details.imdbRating ?? "—";
    if (runtimeEl) runtimeEl.textContent = details.Runtime ?? "—";
    if (directorEl) directorEl.textContent = details.Director ?? "—";
    if (actorsEl) actorsEl.textContent = details.Actors ?? "—";

    const sim = data.similar || [];
    if (similarRow) {
      similarRow.innerHTML = sim.map(movieCard).join("") || `<div class="muted">No similar</div>`;
    }
  } catch (e) {
    console.error("movie load error:", e);
  }

  const rateBtn = document.getElementById("rateBtn");
  const rateInput = document.getElementById("rateValue");
  if (!rateBtn || !rateInput) return;

  rateBtn.addEventListener("click", async () => {
    const v = parseFloat(rateInput.value);
    if (Number.isNaN(v)) return;

    try {
      await fetch(`/api/users/${uid}/ratings/${movieID}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ value: v }),
      });

      const data = await fetchJSON(`/api/users/${uid}/movies/${movieID}`);
      const userRatingEl2 = document.getElementById("userRating");
      if (userRatingEl2) userRatingEl2.textContent = (data.user_rating ?? "—");
    } catch (e) {
      console.error("rate error:", e);
    }
  });
}

/* autocomplete search */
function initSearchAutocomplete() {
  const input = document.getElementById("searchInput");
  const dd = document.getElementById("searchDropdown");
  const box = document.getElementById("searchbox");
  if (!input || !dd || !box) return;

  let items = [];
  let active = -1;

  function close() {
    dd.classList.add("hidden");
    dd.innerHTML = "";
    items = [];
    active = -1;
  }

  function render(list) {
    items = list || [];
    active = -1;

    if (!items.length) return close();

    dd.innerHTML = items.slice(0,8).map((m, idx) => {
    const title = normalizeTitle(m.title || "");
    const year = m.year ?? "";

    return `
      <div class="search-item" data-idx="${idx}" data-id="${m.id}">
        <div>
          <div class="search-title">${escapeHtml(title)}</div>
          <div class="search-sub">${escapeHtml(year)}</div>
        </div>
      </div>
    `;
  }).join("");

    dd.classList.remove("hidden");
  }

  const load = debounce(async () => {
    const q = input.value.trim();
    if (q.length < 2) return close();

    try {
      const data = await fetchJSON(`/api/movies/search?q=${encodeURIComponent(q)}`);
      // ожидаем массив фильмов [{id,title,year,poster_url?}, ...]
      render(Array.isArray(data) ? data : (data.items || []));
    } catch (e) {
      close();
    }
  }, 180);

  input.addEventListener("input", load);

  dd.addEventListener("click", (e) => {
    const item = e.target.closest(".search-item");
    if (!item) return;
    const id = item.getAttribute("data-id");
    if (id) window.location.href = `/movie/${id}`;
  });

  input.addEventListener("keydown", (e) => {
    if (dd.classList.contains("hidden")) return;

    const els = Array.from(dd.querySelectorAll(".search-item"));
    if (!els.length) return;

    if (e.key === "ArrowDown") {
      e.preventDefault();
      active = Math.min(active + 1, els.length - 1);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      active = Math.max(active - 1, 0);
    } else if (e.key === "Enter") {
      e.preventDefault();
      const el = els[Math.max(active, 0)];
      const id = el?.getAttribute("data-id");
      if (id) window.location.href = `/movie/${id}`;
      return;
    } else if (e.key === "Escape") {
      close();
      return;
    } else {
      return;
    }

    els.forEach(x => x.classList.remove("active"));
    if (els[active]) els[active].classList.add("active");
  });

  document.addEventListener("click", (e) => {
    const inside = box.contains(e.target);
    if (!inside) close();
  });
}

function initRegister() {

  const btn = document.getElementById("registerBtn");
  if (!btn) return;

  const msg = document.getElementById("registerMsg");
  const err = document.getElementById("loginError");

  btn.addEventListener("click", async () => {

    if (err) err.textContent = "";
    if (msg) msg.textContent = "Регистрация...";

    try {
      const res = await fetch("/api/users", {
        method: "POST"
      });
      if (!res.ok) {
        throw new Error("register failed");
      }
      const data = await res.json();
      if (msg) {
        msg.textContent = `Успешно! Ваш id: ${data.user_id}`;
      }
    } catch (e) {
      if (msg) msg.textContent = "";
      if (err) err.textContent = "Ошибка регистрации";
      console.error(e);
    }

  });
}
document.addEventListener("DOMContentLoaded", () => {
  initRegister();
  initMenu();
  initSearchAutocomplete();
  initHome();
  initMoviePage();
  initListPage();
});
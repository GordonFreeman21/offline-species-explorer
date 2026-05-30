let currentSpecies = null;

document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-input');
    const searchBtn = document.getElementById('search-btn');
    const syncBtn = document.getElementById('sync-btn');
    const fetchBtn = document.getElementById('fetch-btn');
    const fetchInput = document.getElementById('fetch-input');

    searchBtn.addEventListener('click', doSearch);
    searchInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') doSearch(); });
    syncBtn.addEventListener('click', doSync);
    fetchBtn.addEventListener('click', doFetch);
    fetchInput.addEventListener('keydown', (e) => { if (e.key === 'Enter') doFetch(); });
});

async function doSearch() {
    const input = document.getElementById('search-input');
    const query = input.value.trim();
    if (!query) return;

    showGlobalStatus('Searching...', '');
    try {
        const results = (await window.go.main.App.SearchSpecies(query)) || [];
        displayResults(results);
    } catch (err) {
        const msg = formatError(err);
        showGlobalStatus(msg, 'error');
        displayResults([]);
    }
}

function displayResults(species) {
    const list = document.getElementById('results-list');
    const noResults = document.getElementById('no-results');

    list.innerHTML = '';

    if (!species || species.length === 0) {
        noResults.classList.remove('hidden');
        return;
    }

    noResults.classList.add('hidden');

    species.forEach(s => {
        const item = document.createElement('div');
        item.className = 'result-item';
        item.innerHTML =
            '<strong>' + escapeHtml(s.common_name || 'Unknown') + '</strong>' +
            '<em>' + escapeHtml(s.scientific_name) + '</em>';
        item.addEventListener('click', function() { selectSpecies(s, this); });
        list.appendChild(item);
    });

    hideGlobalStatus();
}

function selectSpecies(species, element) {
    currentSpecies = species;

    document.querySelectorAll('.result-item').forEach(el => el.classList.remove('selected'));
    element.classList.add('selected');

    document.getElementById('empty-state').classList.add('hidden');
    document.getElementById('tree-view').classList.remove('hidden');

    renderTree(species);
}

function renderTree(species) {
    document.getElementById('species-title').textContent = species.scientific_name || 'Unknown';

    const tree = document.getElementById('taxonomy-tree');
    tree.innerHTML = '';

    const levels = [
        { label: 'Kingdom', value: species.kingdom },
        { label: 'Phylum',   value: species.phylum },
        { label: 'Class',    value: species.class },
        { label: 'Order',    value: species.order },
        { label: 'Family',   value: species.family },
    ];

    levels.forEach(level => {
        const node = document.createElement('div');
        node.className = 'tree-node';
        node.innerHTML = '<span class="tree-label">' + level.label + ':</span> ' + escapeHtml(level.value || 'Unknown');
        tree.appendChild(node);
    });

    const speciesNode = document.createElement('div');
    speciesNode.className = 'tree-node species-node';
    speciesNode.innerHTML =
        '<strong>' + escapeHtml(species.scientific_name) + '</strong>' +
        '<div class="common-name">' + escapeHtml(species.common_name || 'Unknown') + '</div>';
    tree.appendChild(speciesNode);

    if (species.last_synced) {
        const syncInfo = document.createElement('div');
        syncInfo.style.cssText = 'font-size:0.75rem;color:var(--text-dim);margin-top:8px';
        syncInfo.textContent = 'Last synced: ' + species.last_synced;
        tree.appendChild(syncInfo);
    }

    document.getElementById('status-msg').classList.add('hidden');
}

async function doSync() {
    if (!currentSpecies || !currentSpecies.scientific_name) return;

    const btn = document.getElementById('sync-btn');
    btn.disabled = true;
    showStatus('Syncing from Nimbus API...', 'loading');

    try {
        const updated = await window.go.main.App.SyncSpecies(currentSpecies.scientific_name);
        currentSpecies = updated;
        renderTree(updated);
        refreshResultItem(updated);
        showStatus('Synced successfully!', 'success');
    } catch (err) {
        showStatus(formatError(err), 'error');
    } finally {
        btn.disabled = false;
    }
}

async function doFetch() {
    const input = document.getElementById('fetch-input');
    const name = input.value.trim();
    if (!name) return;

    const btn = document.getElementById('fetch-btn');
    btn.disabled = true;
    showGlobalStatus('Fetching from API...', '');

    try {
        const species = await window.go.main.App.FetchAndSaveFromAPI(name);
        currentSpecies = species;

        document.getElementById('empty-state').classList.add('hidden');
        document.getElementById('tree-view').classList.remove('hidden');
        document.getElementById('no-results').classList.add('hidden');
        document.getElementById('results-list').classList.remove('hidden');

        renderTree(species);

        const list = document.getElementById('results-list');
        list.innerHTML = '';
        const item = document.createElement('div');
        item.className = 'result-item selected';
        item.innerHTML =
            '<strong>' + escapeHtml(species.common_name || 'Unknown') + '</strong>' +
            '<em>' + escapeHtml(species.scientific_name) + '</em>';
        list.appendChild(item);

        showGlobalStatus('Fetched and saved successfully!', 'success');
        input.value = '';
    } catch (err) {
        showGlobalStatus(formatError(err), 'error');
    } finally {
        btn.disabled = false;
    }
}

function refreshResultItem(species) {
    const items = document.querySelectorAll('.result-item');
    items.forEach(item => {
        const em = item.querySelector('em');
        if (em && em.textContent === species.scientific_name) {
            const strong = item.querySelector('strong');
            if (strong) strong.textContent = species.common_name || 'Unknown';
        }
    });
}

function showStatus(msg, type) {
    const el = document.getElementById('status-msg');
    el.textContent = msg;
    el.className = 'status-msg ' + type;
    el.classList.remove('hidden');
}

function showGlobalStatus(msg, type) {
    const el = document.getElementById('status-global');
    el.textContent = msg;
    el.className = 'status-msg ' + type;
    el.classList.remove('hidden');
}

function hideGlobalStatus() {
    document.getElementById('status-global').classList.add('hidden');
}

function formatError(err) {
    if (!err) return 'Unknown error';
    if (typeof err === 'string') return err;
    if (err.message) return err.message;
    return String(err);
}

function escapeHtml(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

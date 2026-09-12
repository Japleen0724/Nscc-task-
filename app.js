// Main application script for Library Management System
let currentUser = null;
let currentTab = 'catalog';
let searchDebounceTimer = null;

// DOM Elements
const userStatusEl = document.getElementById('user-status');
const authBtnEl = document.getElementById('auth-btn');
const adminTabBtn = document.getElementById('tab-admin-btn');
const myLoansTabBtn = document.getElementById('tab-my-loans-btn');

// Initial setup
document.addEventListener('DOMContentLoaded', () => {
  initAuth();
  setupEventListeners();
  loadCatalog();
});

function toast(msg, type = 'info') {
  const toastEl = document.getElementById('toast');
  toastEl.textContent = msg;
  toastEl.className = 'show ' + (type === 'error' ? 'toast-error' : type === 'success' ? 'toast-success' : '');
  setTimeout(() => {
    toastEl.className = '';
  }, 3500);
}

// Auth Handlers
function initAuth() {
  currentUser = api.getUser();
  renderAuthHeader();
  updateRoleVisibility();

  window.addEventListener('auth-changed', () => {
    currentUser = api.getUser();
    renderAuthHeader();
    updateRoleVisibility();
    switchTab('catalog');
  });
}

function renderAuthHeader() {
  if (currentUser) {
    userStatusEl.innerHTML = `
      <div class="user-badge">
        <span>👤 ${escapeHtml(currentUser.name)}</span>
        <span class="role-pill ${currentUser.role === 'admin' ? 'role-admin' : 'role-user'}">${currentUser.role}</span>
      </div>
    `;
    authBtnEl.textContent = 'Sign Out';
    authBtnEl.className = 'btn btn-secondary btn-sm';
  } else {
    userStatusEl.innerHTML = `
      <span style="font-size: 0.85rem; color: var(--gray-500);">Not signed in</span>
    `;
    authBtnEl.textContent = 'Sign In';
    authBtnEl.className = 'btn btn-primary btn-sm';
  }
}

function updateRoleVisibility() {
  const isAdmin = currentUser && currentUser.role === 'admin';
  if (adminTabBtn) {
    adminTabBtn.style.display = isAdmin ? 'inline-flex' : 'none';
  }
  if (myLoansTabBtn) {
    myLoansTabBtn.style.display = currentUser ? 'inline-flex' : 'none';
  }
  if (!isAdmin && currentTab === 'admin') {
    switchTab('catalog');
  }
}

function handleAuthClick() {
  if (currentUser) {
    api.clearAuth();
    currentUser = null;
    renderAuthHeader();
    updateRoleVisibility();
    toast('Signed out successfully');
    loadCatalog();
  } else {
    openModal('auth-modal');
  }
}

// Quick Demo Login
async function quickLogin(email, password) {
  try {
    const res = await api.login(email, password);
    api.setAuth(res.token, res.user);
    currentUser = res.user;
    closeModal('auth-modal');
    renderAuthHeader();
    updateRoleVisibility();
    toast(`Welcome back, ${res.user.name}!`, 'success');
    if (res.user.role === 'admin') {
      switchTab('admin');
    } else {
      switchTab('catalog');
      loadCatalog();
    }
  } catch (err) {
    toast(err.message, 'error');
  }
}

// Navigation Tabs
function switchTab(tab) {
  currentTab = tab;
  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.tab === tab);
  });
  document.querySelectorAll('.section-view').forEach(view => {
    view.classList.toggle('active', view.id === `view-${tab}`);
  });

  if (tab === 'catalog') {
    loadCatalog();
  } else if (tab === 'my-loans') {
    loadMyLoans();
  } else if (tab === 'admin') {
    loadAdminDashboard();
    loadAdminBooks();
    loadAdminTransactions();
  }
}

// Catalog View
async function loadCatalog() {
  const container = document.getElementById('catalog-grid');
  const search = document.getElementById('catalog-search').value;
  const category = document.getElementById('catalog-category').value;
  const status = document.getElementById('catalog-status').value;

  container.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 2rem;">Loading catalog...</div>';

  try {
    const books = await api.getBooks(search, category, status);
    if (!books || books.length === 0) {
      container.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 2rem; color: var(--gray-500);">No books found matching criteria.</div>';
      return;
    }

    container.innerHTML = books.map(book => `
      <div class="kpi-card" style="border-top: 4px solid ${book.status === 'available' ? 'var(--success)' : 'var(--warning)'};">
        <div style="display: flex; justify-content: space-between; align-items: flex-start;">
          <span class="badge ${book.status === 'available' ? 'badge-available' : 'badge-issued'}">${book.status}</span>
          <button class="btn btn-secondary btn-sm" onclick="showQRModal(${book.id}, '${escapeHtml(book.title)}', '${book.qr_code}')">📱 QR</button>
        </div>
        <h3 style="font-size: 1.15rem; margin-top: 0.5rem; color: var(--dark);">${escapeHtml(book.title)}</h3>
        <p style="font-size: 0.85rem; color: var(--gray-500); margin-bottom: 0.5rem;">by ${escapeHtml(book.author)}</p>
        <div style="font-size: 0.8rem; background: var(--gray-100); padding: 0.2rem 0.5rem; border-radius: 4px; display: inline-block; width: fit-content; margin-bottom: 1rem;">
          📂 ${escapeHtml(book.category)}
        </div>
        <div style="margin-top: auto; display: flex; justify-content: space-between; align-items: center; border-top: 1px solid var(--gray-100); pt-2; padding-top: 0.75rem;">
          <code style="font-size: 0.75rem; color: var(--gray-500);">${escapeHtml(book.qr_code)}</code>
          ${book.status === 'available' ? `
            <button class="btn btn-primary btn-sm" onclick="handleIssueBookDirect(${book.id}, '${escapeHtml(book.title)}')">Issue Book</button>
          ` : `
            <button class="btn btn-secondary btn-sm" disabled style="opacity: 0.6; cursor: not-allowed;">Issued</button>
          `}
        </div>
      </div>
    `).join('');
  } catch (err) {
    container.innerHTML = `<div style="grid-column: 1/-1; color: var(--danger); text-align: center;">${err.message}</div>`;
  }
}

// My Borrowed Books
async function loadMyLoans() {
  if (!currentUser) return;
  const activeContainer = document.getElementById('active-loans-list');
  const historyContainer = document.getElementById('loans-history-list');

  try {
    const transactions = await api.getMyTransactions();
    const active = transactions.filter(t => t.status === 'issued');
    const past = transactions.filter(t => t.status === 'returned');

    if (active.length === 0) {
      activeContainer.innerHTML = '<tr><td colspan="5" style="text-align: center; color: var(--gray-500);">You currently have no active borrowed books.</td></tr>';
    } else {
      activeContainer.innerHTML = active.map(t => `
        <tr>
          <td><strong>${escapeHtml(t.book_title)}</strong><br><small style="color:var(--gray-500);">${escapeHtml(t.book_author)}</small></td>
          <td><code>${escapeHtml(t.book_qr_code)}</code></td>
          <td>${formatDate(t.issue_date)}</td>
          <td><span class="badge badge-issued">ISSUED</span></td>
          <td>
            <button class="btn btn-success btn-sm" onclick="handleReturnBookDirect(${t.id}, '${escapeHtml(t.book_title)}')">Return Book</button>
          </td>
        </tr>
      `).join('');
    }

    if (past.length === 0) {
      historyContainer.innerHTML = '<tr><td colspan="5" style="text-align: center; color: var(--gray-500);">No return history yet.</td></tr>';
    } else {
      historyContainer.innerHTML = past.map(t => `
        <tr>
          <td><strong>${escapeHtml(t.book_title)}</strong></td>
          <td><code>${escapeHtml(t.book_qr_code)}</code></td>
          <td>${formatDate(t.issue_date)}</td>
          <td>${formatDate(t.return_date)}</td>
          <td><span class="badge badge-returned">RETURNED</span></td>
        </tr>
      `).join('');
    }
  } catch (err) {
    toast(err.message, 'error');
  }
}

// Admin Dashboard & Management
async function loadAdminDashboard() {
  try {
    const data = await api.getAdminDashboard();
    document.getElementById('stat-total-books').textContent = data.total_books;
    document.getElementById('stat-avail-books').textContent = data.available_books;
    document.getElementById('stat-issued-books').textContent = data.issued_books;
    document.getElementById('stat-active-borrowers').textContent = data.active_borrowers;

    const recentEl = document.getElementById('admin-recent-transactions');
    if (!data.recent_transactions || data.recent_transactions.length === 0) {
      recentEl.innerHTML = '<tr><td colspan="5" style="text-align: center; color: var(--gray-500);">No recent transactions.</td></tr>';
    } else {
      recentEl.innerHTML = data.recent_transactions.map(t => `
        <tr>
          <td>${escapeHtml(t.user_name)}<br><small style="color:var(--gray-500);">${escapeHtml(t.user_email)}</small></td>
          <td>${escapeHtml(t.book_title)}</td>
          <td><code>${escapeHtml(t.book_qr_code)}</code></td>
          <td>${formatDate(t.issue_date)}</td>
          <td><span class="badge ${t.status === 'issued' ? 'badge-issued' : 'badge-returned'}">${t.status}</span></td>
        </tr>
      `).join('');
    }
  } catch (err) {
    toast(err.message, 'error');
  }
}

async function loadAdminBooks() {
  const tableBody = document.getElementById('admin-books-table');
  try {
    const books = await api.getBooks();
    tableBody.innerHTML = books.map(b => `
      <tr>
        <td>#${b.id}</td>
        <td><strong>${escapeHtml(b.title)}</strong></td>
        <td>${escapeHtml(b.author)}</td>
        <td>${escapeHtml(b.category)}</td>
        <td>
          <button class="btn btn-secondary btn-sm" onclick="showQRModal(${b.id}, '${escapeHtml(b.title)}', '${b.qr_code}')">
            📱 ${escapeHtml(b.qr_code)}
          </button>
        </td>
        <td><span class="badge ${b.status === 'available' ? 'badge-available' : 'badge-issued'}">${b.status}</span></td>
        <td>
          <div style="display: flex; gap: 0.5rem;">
            <button class="btn btn-secondary btn-sm" onclick="openEditBookModal(${b.id}, '${escapeHtml(b.title)}', '${escapeHtml(b.author)}', '${escapeHtml(b.category)}', '${b.status}')">Edit</button>
            <button class="btn btn-danger btn-sm" onclick="handleDeleteBook(${b.id}, '${escapeHtml(b.title)}')">Delete</button>
          </div>
        </td>
      </tr>
    `).join('');
  } catch (err) {
    tableBody.innerHTML = `<tr><td colspan="7" style="color:var(--danger); text-align:center;">${err.message}</td></tr>`;
  }
}

async function loadAdminTransactions() {
  const tableBody = document.getElementById('admin-transactions-table');
  const status = document.getElementById('admin-trans-status').value;
  const search = document.getElementById('admin-trans-search').value;

  try {
    const transactions = await api.getAdminTransactions(status, search);
    if (transactions.length === 0) {
      tableBody.innerHTML = '<tr><td colspan="7" style="text-align: center; color: var(--gray-500);">No transactions found.</td></tr>';
      return;
    }
    tableBody.innerHTML = transactions.map(t => `
      <tr>
        <td>#${t.id}</td>
        <td>${escapeHtml(t.user_name)}<br><small style="color:var(--gray-500);">${escapeHtml(t.user_email)}</small></td>
        <td>${escapeHtml(t.book_title)}<br><small style="color:var(--gray-500);">${escapeHtml(t.book_author)}</small></td>
        <td><code>${escapeHtml(t.book_qr_code)}</code></td>
        <td>${formatDate(t.issue_date)}</td>
        <td>${formatDate(t.return_date)}</td>
        <td><span class="badge ${t.status === 'issued' ? 'badge-issued' : 'badge-returned'}">${t.status}</span></td>
      </tr>
    `).join('');
  } catch (err) {
    tableBody.innerHTML = `<tr><td colspan="7" style="color:var(--danger); text-align:center;">${err.message}</td></tr>`;
  }
}

// Issue / Return Actions
async function handleIssueBookDirect(bookId, title) {
  if (!currentUser) {
    toast('Please log in first to issue a book', 'error');
    openModal('auth-modal');
    return;
  }
  if (!confirm(`Confirm issuing "${title}"?`)) return;

  try {
    await api.issueBook({ book_id: bookId });
    toast(`Successfully issued "${title}"!`, 'success');
    loadCatalog();
  } catch (err) {
    toast(err.message, 'error');
  }
}

async function handleReturnBookDirect(transactionId, title) {
  if (!confirm(`Confirm returning "${title}"?`)) return;

  try {
    await api.returnBook({ transaction_id: transactionId });
    toast(`Successfully returned "${title}"!`, 'success');
    loadMyLoans();
  } catch (err) {
    toast(err.message, 'error');
  }
}

// Quick Scan Handlers
async function handleQuickIssue() {
  const qr = document.getElementById('quick-qr-input').value.trim();
  if (!qr) {
    toast('Please enter or scan a book QR code', 'error');
    return;
  }
  if (!currentUser) {
    toast('Please log in to issue books', 'error');
    openModal('auth-modal');
    return;
  }

  try {
    const res = await api.issueBook({ qr_code: qr });
    toast(`Success: Issued "${res.book.title}"!`, 'success');
    document.getElementById('quick-qr-input').value = '';
    loadCatalog();
  } catch (err) {
    toast(err.message, 'error');
  }
}

async function handleQuickReturn() {
  const qr = document.getElementById('quick-qr-input').value.trim();
  if (!qr) {
    toast('Please enter or scan a book QR code', 'error');
    return;
  }
  if (!currentUser) {
    toast('Please log in to return books', 'error');
    openModal('auth-modal');
    return;
  }

  try {
    await api.returnBook({ qr_code: qr });
    toast('Book successfully returned to library catalog!', 'success');
    document.getElementById('quick-qr-input').value = '';
    loadCatalog();
  } catch (err) {
    toast(err.message, 'error');
  }
}

// Modals & Helpers
function openModal(id) {
  const modal = document.getElementById(id);
  if (modal) modal.classList.add('open');
}

function closeModal(id) {
  const modal = document.getElementById(id);
  if (modal) modal.classList.remove('open');
}

function showQRModal(bookId, title, qrCode) {
  document.getElementById('qr-modal-title').textContent = title;
  document.getElementById('qr-modal-code').textContent = qrCode;
  const qrImg = document.getElementById('qr-modal-image');
  qrImg.src = api.getQRCodeURL(bookId);

  const downloadBtn = document.getElementById('qr-modal-download');
  downloadBtn.onclick = () => {
    const a = document.createElement('a');
    a.href = api.getQRCodeURL(bookId);
    a.download = `QR_${qrCode}.png`;
    a.click();
  };

  openModal('qr-modal');
}

function openAddBookModal() {
  document.getElementById('add-book-form').reset();
  openModal('add-book-modal');
}

function openEditBookModal(id, title, author, category, status) {
  document.getElementById('edit-book-id').value = id;
  document.getElementById('edit-book-title').value = title;
  document.getElementById('edit-book-author').value = author;
  document.getElementById('edit-book-category').value = category;
  document.getElementById('edit-book-status').value = status;
  openModal('edit-book-modal');
}

async function handleDeleteBook(id, title) {
  if (!confirm(`Are you sure you want to delete "${title}"?`)) return;
  try {
    await api.deleteBook(id);
    toast(`Book "${title}" deleted`, 'success');
    loadAdminBooks();
    loadAdminDashboard();
  } catch (err) {
    toast(err.message, 'error');
  }
}

// Event Listeners setup
function setupEventListeners() {
  authBtnEl.addEventListener('click', handleAuthClick);

  // Search & Filter debouncing
  document.getElementById('catalog-search').addEventListener('input', () => {
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = setTimeout(loadCatalog, 300);
  });
  document.getElementById('catalog-category').addEventListener('change', loadCatalog);
  document.getElementById('catalog-status').addEventListener('change', loadCatalog);

  // Admin filter
  document.getElementById('admin-trans-status').addEventListener('change', loadAdminTransactions);
  document.getElementById('admin-trans-search').addEventListener('input', () => {
    clearTimeout(searchDebounceTimer);
    searchDebounceTimer = setTimeout(loadAdminTransactions, 300);
  });

  // Auth Form Submit
  document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('login-email').value.trim();
    const pass = document.getElementById('login-password').value;
    try {
      const res = await api.login(email, pass);
      api.setAuth(res.token, res.user);
      currentUser = res.user;
      closeModal('auth-modal');
      renderAuthHeader();
      updateRoleVisibility();
      toast(`Welcome back, ${res.user.name}!`, 'success');
      if (res.user.role === 'admin') switchTab('admin');
      else loadCatalog();
    } catch (err) {
      toast(err.message, 'error');
    }
  });

  document.getElementById('register-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('reg-name').value.trim();
    const email = document.getElementById('reg-email').value.trim();
    const pass = document.getElementById('reg-password').value;
    const role = document.getElementById('reg-role').value;
    try {
      const res = await api.register(name, email, pass, role);
      api.setAuth(res.token, res.user);
      currentUser = res.user;
      closeModal('auth-modal');
      renderAuthHeader();
      updateRoleVisibility();
      toast(`Account registered successfully, welcome ${res.user.name}!`, 'success');
      if (res.user.role === 'admin') switchTab('admin');
      else loadCatalog();
    } catch (err) {
      toast(err.message, 'error');
    }
  });

  // Add Book Form
  document.getElementById('add-book-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const title = document.getElementById('new-book-title').value.trim();
    const author = document.getElementById('new-book-author').value.trim();
    const category = document.getElementById('new-book-category').value.trim();

    try {
      await api.createBook({ title, author, category });
      closeModal('add-book-modal');
      toast(`Book "${title}" added successfully!`, 'success');
      loadAdminBooks();
      loadAdminDashboard();
    } catch (err) {
      toast(err.message, 'error');
    }
  });

  // Edit Book Form
  document.getElementById('edit-book-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('edit-book-id').value;
    const title = document.getElementById('edit-book-title').value.trim();
    const author = document.getElementById('edit-book-author').value.trim();
    const category = document.getElementById('edit-book-category').value.trim();
    const status = document.getElementById('edit-book-status').value;

    try {
      await api.updateBook(id, { title, author, category, status });
      closeModal('edit-book-modal');
      toast('Book updated successfully', 'success');
      loadAdminBooks();
      loadAdminDashboard();
    } catch (err) {
      toast(err.message, 'error');
    }
  });

  // Export CSV button
  document.getElementById('admin-export-csv-btn').addEventListener('click', async () => {
    try {
      toast('Generating CSV report...');
      await api.downloadTransactionsCSV();
      toast('CSV report downloaded successfully', 'success');
    } catch (err) {
      toast(err.message, 'error');
    }
  });
}

function formatDate(dateStr) {
  if (!dateStr) return '-';
  const d = new Date(dateStr);
  return isNaN(d.getTime()) ? dateStr : d.toLocaleString();
}

function escapeHtml(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

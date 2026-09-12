// API Client for Library Management System
const API_BASE = window.location.origin.includes('http') ? `${window.location.origin}/api` : 'http://localhost:8080/api';

const api = {
  getToken() {
    return localStorage.getItem('lib_token');
  },

  getUser() {
    const raw = localStorage.getItem('lib_user');
    return raw ? JSON.parse(raw) : null;
  },

  setAuth(token, user) {
    localStorage.setItem('lib_token', token);
    localStorage.setItem('lib_user', JSON.stringify(user));
  },

  clearAuth() {
    localStorage.removeItem('lib_token');
    localStorage.removeItem('lib_user');
  },

  async request(endpoint, options = {}) {
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    const token = this.getToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    });

    if (response.status === 401) {
      this.clearAuth();
      window.dispatchEvent(new Event('auth-changed'));
      throw new Error('Session expired or unauthorized. Please log in.');
    }

    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || `Error ${response.status}: ${response.statusText}`);
    }

    return data;
  },

  // Auth
  login(email, password) {
    return this.request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  },

  register(name, email, password, role = 'user') {
    return this.request('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ name, email, password, role }),
    });
  },

  getMe() {
    return this.request('/auth/me');
  },

  // Books
  getBooks(search = '', category = '', status = '') {
    const params = new URLSearchParams();
    if (search) params.append('search', search);
    if (category) params.append('category', category);
    if (status) params.append('status', status);
    return this.request(`/books?${params.toString()}`);
  },

  getBook(id) {
    return this.request(`/books/${id}`);
  },

  createBook(book) {
    return this.request('/books', {
      method: 'POST',
      body: JSON.stringify(book),
    });
  },

  updateBook(id, book) {
    return this.request(`/books/${id}`, {
      method: 'PUT',
      body: JSON.stringify(book),
    });
  },

  deleteBook(id) {
    return this.request(`/books/${id}`, {
      method: 'DELETE',
    });
  },

  getQRCodeURL(bookId) {
    return `${API_BASE}/books/${bookId}/qr`;
  },

  // Transactions
  issueBook(payload) {
    return this.request('/transactions/issue', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  returnBook(payload) {
    return this.request('/transactions/return', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  getMyTransactions() {
    return this.request('/transactions/my');
  },

  // Admin Dashboard & Reports
  getAdminDashboard() {
    return this.request('/admin/dashboard');
  },

  getAdminTransactions(status = '', search = '') {
    const params = new URLSearchParams();
    if (status) params.append('status', status);
    if (search) params.append('search', search);
    return this.request(`/admin/transactions?${params.toString()}`);
  },

  getExportCSVURL() {
    const token = this.getToken();
    return `${API_BASE}/admin/transactions/export?token=${token || ''}`;
  },

  async downloadTransactionsCSV() {
    const token = this.getToken();
    const response = await fetch(`${API_BASE}/admin/transactions/export`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    if (!response.ok) throw new Error('Failed to export CSV');
    const blob = await response.blob();
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `library_transactions_${new Date().toISOString().slice(0,10)}.csv`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(url);
  }
};

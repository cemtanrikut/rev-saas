const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// Token storage key
const TOKEN_KEY = 'revalyze_token';

// Get token from localStorage
export const getToken = () => {
  return localStorage.getItem(TOKEN_KEY);
};

// Set token in localStorage
export const setToken = (token) => {
  if (token) {
    localStorage.setItem(TOKEN_KEY, token);
  } else {
    localStorage.removeItem(TOKEN_KEY);
  }
};

// Remove token from localStorage
export const removeToken = () => {
  localStorage.removeItem(TOKEN_KEY);
};

// Build headers with optional auth
const buildHeaders = (includeAuth = true) => {
  const headers = {
    'Content-Type': 'application/json',
  };

  if (includeAuth) {
    const token = getToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  return headers;
};

// Custom error class for limit errors
export class LimitError extends Error {
  constructor(errorCode, reason, plan, limit, current) {
    super(reason || errorCode);
    this.name = 'LimitError';
    this.errorCode = errorCode;
    this.reason = reason;
    this.plan = plan;
    this.limit = limit;
    this.current = current;
  }
}

// Generic fetch wrapper with error handling
const fetchWithError = async (url, options = {}) => {
  try {
    const response = await fetch(url, options);
    
    // Handle non-JSON responses
    const contentType = response.headers.get('content-type');
    
    if (response.status === 204) {
      return { ok: true, data: null };
    }

    let data = null;
    if (contentType && contentType.includes('application/json')) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    if (!response.ok) {
      // Check for limit error (403 with specific error codes)
      if (response.status === 403 && typeof data === 'object' && data.error) {
        const limitErrorCodes = ['LIMIT_COMPETITORS', 'LIMIT_PLANS', 'LIMIT_ANALYSES', 'LIMIT_TRIAL_EXPIRED'];
        if (limitErrorCodes.includes(data.error)) {
          throw new LimitError(
            data.error,
            data.reason,
            data.plan,
            data.limit,
            data.current
          );
        }
      }
      
      const errorMessage = typeof data === 'string' ? data : data.error || data.message || 'Request failed';
      throw new Error(errorMessage);
    }

    return { ok: true, data };
  } catch (error) {
    if (error.name === 'TypeError' && error.message === 'Failed to fetch') {
      throw new Error('Unable to connect to server. Please check if the backend is running.');
    }
    throw error;
  }
};

// POST request with JSON body
export const postJson = async (path, body = {}, options = {}) => {
  const url = `${API_BASE_URL}${path}`;
  const includeAuth = options.includeAuth !== false;

  return fetchWithError(url, {
    method: 'POST',
    headers: buildHeaders(includeAuth),
    body: JSON.stringify(body),
    ...options,
  });
};

// GET request
export const getJson = async (path, options = {}) => {
  const url = `${API_BASE_URL}${path}`;
  const includeAuth = options.includeAuth !== false;

  return fetchWithError(url, {
    method: 'GET',
    headers: buildHeaders(includeAuth),
    ...options,
  });
};

// DELETE request
export const deleteJson = async (path, options = {}) => {
  const url = `${API_BASE_URL}${path}`;
  const includeAuth = options.includeAuth !== false;

  return fetchWithError(url, {
    method: 'DELETE',
    headers: buildHeaders(includeAuth),
    ...options,
  });
};

// Auth-specific API calls
export const authApi = {
  signup: (signupData) => 
    postJson('/auth/signup', {
      email: signupData.email,
      password: signupData.password,
      full_name: signupData.fullName,
      role: signupData.role,
      company_name: signupData.companyName,
      company_website: signupData.companyWebsite,
      mrr_range: signupData.mrrRange,
      heard_from: signupData.heardFrom,
    }, { includeAuth: false }),
  
  login: (email, password) => 
    postJson('/auth/login', { email, password }, { includeAuth: false }),
  
  me: () => 
    getJson('/auth/me'),
};

// Plans API calls
export const plansApi = {
  list: () => getJson('/api/plans'),
  create: (name, price, currency = 'USD', billingCycle = 'monthly') => 
    postJson('/api/plans', { name, price, currency, billing_cycle: billingCycle }),
  delete: (id) => deleteJson(`/api/plans/${id}`),
};

// Competitors API calls
export const competitorsApi = {
  list: () => getJson('/api/competitors'),
  create: (name, url, plans) => postJson('/api/competitors', { name, url, plans }),
  delete: (id) => deleteJson(`/api/competitors/${id}`),
};

// Fetch binary data (for PDF downloads)
export const fetchBlob = async (path, options = {}) => {
  const url = `${API_BASE_URL}${path}`;
  const includeAuth = options.includeAuth !== false;
  
  const headers = {};
  if (includeAuth) {
    const token = getToken();
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
  }

  try {
    const response = await fetch(url, {
      method: 'GET',
      headers,
      ...options,
    });

    if (!response.ok) {
      // Try to read error message
      const contentType = response.headers.get('content-type');
      let errorMessage = 'Download failed';
      if (contentType && contentType.includes('application/json')) {
        const data = await response.json();
        errorMessage = data.error || data.message || errorMessage;
      } else {
        errorMessage = await response.text() || errorMessage;
      }
      throw new Error(errorMessage);
    }

    const blob = await response.blob();
    return { ok: true, blob };
  } catch (error) {
    if (error.name === 'TypeError' && error.message === 'Failed to fetch') {
      throw new Error('Unable to connect to server. Please check if the backend is running.');
    }
    throw error;
  }
};

// Trigger browser download from blob
export const downloadBlob = (blob, filename) => {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
};

// Analysis API calls
export const analysisApi = {
  run: () => postJson('/api/analysis/run', {}),
  list: () => getJson('/api/analysis'),
  exportPdf: async (analysisId) => {
    const { ok, blob } = await fetchBlob(`/api/analysis/${analysisId}/export-pdf`);
    return { ok, blob };
  },
};

// Business Metrics API calls
export const businessMetricsApi = {
  get: () => getJson('/api/business-metrics'),
  set: (metrics) => {
    // Use fetch directly with PUT method since we don't have putJson
    const url = `${API_BASE_URL}/api/business-metrics`;
    return fetchWithError(url, {
      method: 'PUT',
      headers: buildHeaders(true),
      body: JSON.stringify(metrics),
    });
  },
};

export default {
  postJson,
  getJson,
  deleteJson,
  fetchBlob,
  downloadBlob,
  getToken,
  setToken,
  removeToken,
  authApi,
  plansApi,
  competitorsApi,
  analysisApi,
  businessMetricsApi,
};


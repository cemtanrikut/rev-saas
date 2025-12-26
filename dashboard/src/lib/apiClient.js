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

// Custom error class for AI credits errors
export class AICreditsError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'AICreditsError';
    this.code = code;
    this.message = message;
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
      
      // Check for AI credits errors (402 or 403 with code field)
      if ((response.status === 402 || response.status === 403) && typeof data === 'object' && data.code) {
        const aiErrorCodes = ['AI_QUOTA_EXCEEDED', 'SIMULATION_NOT_AVAILABLE'];
        if (aiErrorCodes.includes(data.code)) {
          throw new AICreditsError(data.code, data.message);
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

// PUT request with JSON body
export const putJson = async (path, body = {}, options = {}) => {
  const url = `${API_BASE_URL}${path}`;
  const includeAuth = options.includeAuth !== false;

  return fetchWithError(url, {
    method: 'PUT',
    headers: buildHeaders(includeAuth),
    body: JSON.stringify(body),
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
  create: (name, price, currency = 'USD', billingCycle = 'monthly', stripePriceId = '') => 
    postJson('/api/plans', { name, price, currency, billing_cycle: billingCycle, stripe_price_id: stripePriceId || undefined }),
  update: (id, data) => putJson(`/api/plans/${id}`, data),
  delete: (id) => deleteJson(`/api/plans/${id}`),
};

// Competitors API calls
export const competitorsApi = {
  list: () => getJson('/api/competitors'),
  create: (name, url, plans) => postJson('/api/competitors', { name, url, plans }),
  update: (id, name, url, plans) => putJson(`/api/competitors/${id}`, { name, url, plans }),
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

// Analysis API calls (legacy V1)
export const analysisApi = {
  run: () => postJson('/api/analysis/run', {}),
  list: () => getJson('/api/analysis'),
  exportPdf: async (analysisId) => {
    const { ok, blob } = await fetchBlob(`/api/analysis/${analysisId}/export-pdf`);
    return { ok, blob };
  },
};

// Analysis V2 API calls (new deterministic engine)
export const analysisV2Api = {
  run: () => postJson('/api/analysis/v2', {}),
  list: (limit = 20) => getJson(`/api/analysis/v2?limit=${limit}`),
  get: (id) => getJson(`/api/analysis/v2/${id}`),
  exportPdf: async (analysisId) => {
    const { ok, blob } = await fetchBlob(`/api/analysis/v2/${analysisId}/pdf`);
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

// Simulation API calls
export const simulationApi = {
  run: (data) => postJson('/api/simulations', {
    plan_id: data.planId,
    current_price: data.currentPrice,
    new_price: data.newPrice,
    currency: data.currency,
    active_customers_on_plan: data.activeCustomersOnPlan,
    global_mrr: data.globalMrr,
    global_churn_rate: data.globalChurnRate,
    pricing_goal: data.pricingGoal,
  }),
  list: (limit = 10, planId = null) => {
    const params = new URLSearchParams({ limit: limit.toString() });
    if (planId) params.append('plan_id', planId);
    return getJson(`/api/simulations?${params}`);
  },
  get: (id) => getJson(`/api/simulations/${id}`),
  exportPdf: async (simulationId) => {
    const { ok, blob } = await fetchBlob(`/api/simulations/${simulationId}/pdf`);
    return { ok, blob };
  },
};

// AI Credits API calls
export const aiCreditsApi = {
  get: () => getJson('/api/ai-credits'),
};

// Stripe Connect API calls
export const stripeApi = {
  getStatus: () => getJson('/api/stripe/status'),
  connect: () => postJson('/api/stripe/connect'),
  disconnect: () => postJson('/api/stripe/disconnect'),
  syncMetrics: () => postJson('/api/stripe/sync/metrics'),
};

// Billing API calls (Stripe subscription management)
export const billingApi = {
  getStatus: () => getJson('/api/billing/status'),
  createCheckoutSession: (planKey) => postJson('/api/billing/checkout-session', { plan_key: planKey }),
  createPortalSession: () => postJson('/api/billing/portal'),
};

// Competitors V2 API calls (AI discovery)
export const competitorsV2Api = {
  discover: (websiteUrl) => postJson('/api/v2/competitors/discover', { website_url: websiteUrl }),
  save: (competitors) => postJson('/api/v2/competitors/save', { competitors }),
  list: () => getJson('/api/v2/competitors'),
  delete: (id) => deleteJson(`/api/v2/competitors/${id}`),
};

// Pricing V2 API calls (auto-import from website)
export const pricingV2Api = {
  discover: (data) => postJson('/api/pricing-v2/discover', data),
  extract: (data) => postJson('/api/pricing-v2/extract', data),
  extractFromText: (data) => postJson('/api/pricing-v2/extract-from-text', data),
  save: (data) => postJson('/api/pricing-v2/save', data),
  list: () => getJson('/api/pricing-v2'),
  delete: (id) => deleteJson(`/api/pricing-v2/${id}`),
};

export default {
  postJson,
  getJson,
  deleteJson,
  putJson,
  fetchBlob,
  downloadBlob,
  getToken,
  setToken,
  removeToken,
  authApi,
  plansApi,
  competitorsApi,
  competitorsV2Api,
  pricingV2Api,
  analysisApi,
  analysisV2Api,
  businessMetricsApi,
  simulationApi,
  aiCreditsApi,
  stripeApi,
  billingApi,
  AICreditsError,
};


import axios from 'axios';
import type { InternalAxiosRequestConfig, AxiosResponse, AxiosError } from 'axios';

// Create a single axios instance for the application
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Important for handling authentication cookies
});

// Request interceptor for adding auth token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
    const token = localStorage.getItem('token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    // More detailed logging
    console.log(`Request: ${config.method?.toUpperCase()} ${config.url}`, config.data);
    console.log(`Request headers:`, config.headers);
    
    if (config.data) {
      console.log(`FINAL PAYLOAD being sent to server:`, JSON.stringify(config.data, null, 2));
      // Check if data has a quotation property
      if (config.data.quotation) {
        console.log(`✅ STRUCTURE CHECK: Properly nested with quotation key`);
        console.log(`✅ customer_id: ${config.data.quotation.customer_id}, type: ${typeof config.data.quotation.customer_id}`);
      } else {
        console.log(`❌ STRUCTURE CHECK: Missing nested quotation key, flat object found instead`);
        if (config.data.customer_id) {
          console.log(`❌ Flat customer_id: ${config.data.customer_id}, type: ${typeof config.data.customer_id}`);
        }
      }
    }
    
    return config;
  },
  (error: AxiosError): Promise<AxiosError> => {
    console.error('Request error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor for handling errors
api.interceptors.response.use(
  (response: AxiosResponse): AxiosResponse => {
    console.log(`Response from ${response.config.url}:`, response.status, response.data);
    return response;
  },
  (error: AxiosError): Promise<AxiosError> => {
    console.error('Response error:', {
      url: error.config?.url,
      status: error.response?.status,
      statusText: error.response?.statusText,
      data: error.response?.data
    });
    
    // Handle unauthorized errors (status 401)
    if (error.response && error.response.status === 401) {
      const isLoginAttempt = error.config?.url?.includes('/login') || 
                            error.config?.url?.includes('/admin/login');
      
      if (!isLoginAttempt) {
        console.warn('Unauthorized access - clearing credentials');
        // Clear local storage and redirect to login
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/login';
      } else {
        console.warn('Login attempt failed with 401 - letting component handle the error');
      }
    }
    
    // Handle forbidden errors (status 403)
    if (error.response && error.response.status === 403) {
      console.warn('Forbidden access - may be role-related issue');
    }
    
    return Promise.reject(error);
  }
);

export default api; 
const API_BASE = 'http://localhost:3000/v1';

class ApiService {
  constructor() {
    this.accessToken = localStorage.getItem('accessToken');
  }

  setToken(token) {
    this.accessToken = token;
    localStorage.setItem('accessToken', token);
  }

  clearToken() {
    this.accessToken = null;
    localStorage.removeItem('accessToken');
  }

  getHeaders() {
    const token = localStorage.getItem('accessToken');
    const headers = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
  }

  async request(endpoint, options = {}) {
    const url = `${API_BASE}${endpoint}`;
    const config = {
      ...options,
      headers: {
        ...this.getHeaders(),
        ...options.headers,
      },
    };

    try {
      const response = await fetch(url, config);
      
      if (response.status === 204) {
        return {};
      }

      const data = await response.json();

      if (!response.ok) {
        throw { 
          status: response.status, 
          message: data.message || data.Message || 'Ошибка',
          ...data 
        };
      }

      return data;
    } catch (error) {
      if (error.status === 401) {
        this.clearToken();
      }
      throw error;
    }
  }

  // Auth
  async requestCode(phone) {
    return this.request('/auth/request-code', {
      method: 'POST',
      body: JSON.stringify({ phone }),
    });
  }

  async verifyCode(phone, code) {
    const data = await this.request('/auth/verify', {
      method: 'POST',
      body: JSON.stringify({ phone, code }),
    });
    this.setToken(data.accessToken);
    return data;
  }

  logout() {
    this.clearToken();
  }

  // Slots
  async getSlots(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/slots${query ? `?${query}` : ''}`);
  }

  async getSlotById(slotId) {
    return this.request(`/slots/${slotId}`);
  }

  async createSlot(data) {
    return this.request('/slots', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // Reference Data
  async getZones() {
    return this.request('/zones');
  }

  async getInstructors() {
    return this.request('/instructors');
  }

  async getEquipment() {
    return this.request('/equipment');
  }

  // Bookings
  async getBookings(params = {}) {
    const query = new URLSearchParams(params).toString();
    return this.request(`/bookings${query ? `?${query}` : ''}`);
  }

  async getBookingById(bookingId) {
    return this.request(`/bookings/${bookingId}`);
  }

  async createBooking(data) {
    return this.request('/bookings', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async cancelBooking(bookingId) {
    return this.request(`/bookings/${bookingId}/cancel`, {
      method: 'PATCH',
    });
  }

  // Profile
  async getProfile() {
    return this.request('/profile');
  }

  async updateProfile(data) {
    return this.request('/profile', {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  // Admin
  async getAllUsers() {
    return this.request('/admin/users');
  }

  async updateUserRole(userId, role) {
    return this.request(`/admin/users/${userId}/role`, {
      method: 'PATCH',
      body: JSON.stringify({ role }),
    });
  }

  async assignTrainer(phone, instructorId) {
    return this.request('/admin/assign-trainer', {
      method: 'POST',
      body: JSON.stringify({ phone, instructorId }),
    });
  }

  async getAllBookings() {
    return this.request('/admin/bookings');
  }

  async adminCancelBooking(bookingId) {
    return this.request(`/admin/bookings/${bookingId}/cancel`, {
      method: 'PATCH',
    });
  }
}

export const api = new ApiService();
import { createContext, useContext, useState, useEffect } from 'react';
import { api } from '../services/api';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      api.getProfile()
        .then(setUser)
        .catch((err) => {
          console.log('Profile load error:', err);
        })
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (phone, code) => {
    const { client } = await api.verifyCode(phone, code);
    setUser(client);
    return client;
  };

  const requestCode = async (phone) => {
    return api.requestCode(phone);
  };

  const logout = () => {
    const bookings = localStorage.getItem('bookings');
    if (bookings) {
      localStorage.setItem('savedBookings', bookings);
    }
    api.logout();
    setUser(null);
  };

  const restoreBookings = () => {
    const saved = localStorage.getItem('savedBookings');
    if (saved) {
      localStorage.setItem('bookings', saved);
      localStorage.removeItem('savedBookings');
    }
  };

  const updateProfile = async (data) => {
    const updated = await api.updateProfile(data);
    setUser(updated);
    return updated;
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, requestCode, logout, updateProfile, restoreBookings }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
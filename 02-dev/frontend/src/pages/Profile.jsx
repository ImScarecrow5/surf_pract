import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card } from '../components/Card';
import { Input } from '../components/Input';
import { Button } from '../components/Button';
import { Badge } from '../components/Badge';
import { useAuth } from '../context/AuthContext';
import { api } from '../services/api';
import './Profile.css';

function formatDateTime(dateStr) {
  const date = new Date(dateStr);
  return date.toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });
}

export function ProfilePage() {
  const navigate = useNavigate();
  const { user, logout, updateProfile } = useAuth();
  const [name, setName] = useState(user?.name || '');
  const [level, setLevel] = useState(user?.level || 'novice');
  const [loading, setLoading] = useState(false);
  const [saved, setSaved] = useState(false);
  const [bookings, setBookings] = useState([]);
  const [bookingsLoading, setBookingsLoading] = useState(true);

  useEffect(() => {
    api.getBookings({})
      .then(data => setBookings(data.data || []))
      .catch(err => console.error(err))
      .finally(() => setBookingsLoading(false));
  }, []);

  const handleSave = async () => {
    setLoading(true);
    try {
      await updateProfile({ name, level: level === 'intermediate' ? 'experienced' : level });
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    logout();
    navigate('/auth');
  };

  const getStatusBadge = (status) => {
    const variants = {
      confirmed: 'success',
      cancelled_by_client_early: 'error',
      cancelled_by_client_late: 'error',
      cancelled_by_gym: 'muted'
    };
    const texts = {
      confirmed: 'Подтверждено',
      cancelled_by_client_early: 'Отменено',
      cancelled_by_client_late: 'Отменено (штраф)',
      cancelled_by_gym: 'Отменено скалодромом'
    };
    return { variant: variants[status] || 'default', text: texts[status] || status };
  };

  return (
    <div className="profile-page">
      <header className="profile-header">
        <button className="back-btn" onClick={() => navigate(-1)}>
          ← Назад
        </button>
        <h1>Профиль</h1>
      </header>

      <main className="profile-content">
        <Card padding="lg">
          <div className="profile-avatar">
            {user?.name?.charAt(0) || 'У'}
          </div>

          <div className="profile-field">
            <Input
              label="Имя"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Ваше имя"
            />
          </div>

          <div className="profile-field">
            <label className="input-label">Уровень</label>
            <div className="level-options">
              <label className={`level-option ${level === 'novice' ? 'selected' : ''}`}>
                <input
                  type="radio"
                  name="level"
                  checked={level === 'novice'}
                  onChange={() => setLevel('novice')}
                />
                <span>Новичок</span>
              </label>
              <label className={`level-option ${level === 'experienced' ? 'selected' : ''}`}>
                <input
                  type="radio"
                  name="level"
                  checked={level === 'experienced'}
                  onChange={() => setLevel('experienced')}
                />
                <span>Опытный</span>
              </label>
            </div>
          </div>

          <div className="profile-info">
            <div className="info-row">
              <span className="info-label">Телефон</span>
              <span className="info-value">{user?.phone}</span>
            </div>
          </div>

          <Button fullWidth onClick={handleSave} loading={loading}>
            {saved ? 'Сохранено' : 'Сохранить'}
          </Button>
        </Card>

        {bookings.length > 0 && (
          <Card padding="md">
            <h3>Мои бронирования</h3>
            {bookings.map(booking => {
              const statusBadge = getStatusBadge(booking.status);
              return (
                <div key={booking.id} className="profile-booking" onClick={() => navigate(`/booking/${booking.id}`)}>
                  <div className="booking-info">
                    <span className="booking-date">{formatDateTime(booking.slot?.startTime)}</span>
                    <span className="booking-zone">{booking.slot?.zone?.name || 'Тренировка'}</span>
                  </div>
                  <Badge variant={statusBadge.variant} size="sm">
                    {statusBadge.text}
                  </Badge>
                </div>
              );
            })}
          </Card>
        )}

        <Button variant="ghost" fullWidth onClick={handleLogout} className="logout-btn">
          Выйти из аккаунта
        </Button>
      </main>
    </div>
  );
}
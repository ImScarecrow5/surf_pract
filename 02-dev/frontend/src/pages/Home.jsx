import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card } from '../components/Card';
import { Badge } from '../components/Badge';
import { Button } from '../components/Button';
import { api } from '../services/api';
import { useAuth } from '../context/AuthContext';
import './Home.css';

function formatDate(dateStr) {
  const date = new Date(dateStr);
  const today = new Date();
  const tomorrow = new Date(today);
  tomorrow.setDate(tomorrow.getDate() + 1);

  if (date.toDateString() === today.toDateString()) return 'Сегодня';
  if (date.toDateString() === tomorrow.toDateString()) return 'Завтра';
  return date.toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'short' });
}

function formatTime(dateStr) {
  return new Date(dateStr).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

function SlotCard({ slot, onClick }) {
  const zoneType = slot.zone?.name?.toLowerCase().includes('болот') ? 'boulder' : 'rope';
  const isLowCapacity = slot.freePlaces < 3 && slot.freePlaces > 0;
  const level = slot.level || (zoneType === 'boulder' ? 'novice' : 'experienced');
  const levelBadge = level === 'experienced' 
    ? { label: 'Опытные', variant: 'warning' }
    : { label: 'Новички', variant: 'success' };

  return (
    <Card className="slot-card" onClick={onClick}>
      <div className="slot-time">
        <span className="slot-date">{formatDate(slot.startTime)}</span>
        <span className="slot-hours">{formatTime(slot.startTime)}</span>
      </div>
      <div className="slot-info">
        <div className="slot-title">
          <span>{slot.zone?.name || 'Тренировка'}</span>
          <Badge variant={levelBadge.variant} size="sm">
            {levelBadge.label}
          </Badge>
        </div>
        <div className="slot-instructor">
          {slot.instructor?.name || 'Инструктор'}
          {slot.instructor?.rating && <span className="slot-rating">★ {slot.instructor.rating}</span>}
        </div>
        <div className="slot-places">
          <span className={isLowCapacity ? 'places-warning' : ''}>
            {slot.freePlaces} из {slot.totalPlaces} мест
          </span>
          <span className="slot-price">{slot.price} ₽</span>
        </div>
      </div>
    </Card>
  );
}

function BookingCard({ booking, onClick }) {
  const statusVariant = {
    confirmed: 'success',
    cancelled_by_client_early: 'error',
    cancelled_by_client_late: 'error',
    cancelled_by_gym: 'muted'
  }[booking.status] || 'default';

  const statusText = {
    confirmed: 'Подтверждено',
    cancelled_by_client_early: 'Отменено',
    cancelled_by_client_late: 'Отменено (штраф)',
    cancelled_by_gym: 'Отменено скалодромом'
  }[booking.status] || booking.status;

  return (
    <Card className={`booking-card ${booking.status}`} onClick={onClick}>
      <div className="booking-time">
        <span className="booking-date">{formatDate(booking.slot?.startTime)}</span>
        <span className="booking-hours">{formatTime(booking.slot?.startTime)}</span>
      </div>
      <div className="booking-info">
        <div className="booking-title">
          <span>{booking.slot?.zone?.name || 'Тренировка'}</span>
          <Badge variant={statusVariant} size="sm">{statusText}</Badge>
        </div>
        <div className="booking-instructor">
          {booking.slot?.instructor?.name || 'Инструктор'}
        </div>
        <div className="booking-price">
          <span>{booking.price} ₽</span>
          {booking.equipmentType === 'rental' && <span className="booking-equipment">+ снаряжение</span>}
        </div>
      </div>
    </Card>
  );
}

export function HomePage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [activeTab, setActiveTab] = useState('schedule');
  const [slots, setSlots] = useState([]);
  const [bookings, setBookings] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const isAdminOrTrainer = user?.role === 'admin' || user?.role === 'trainer';
  const userLevel = user?.level || 'novice';

  useEffect(() => {
    if (activeTab === 'schedule') {
      setLoading(true);
      api.getSlots({ onlyAvailable: true })
        .then(data => {
          const allSlots = data.data || [];
          const filteredSlots = userLevel === 'novice'
            ? allSlots.filter(slot => slot.level !== 'experienced')
            : allSlots;
          setSlots(filteredSlots);
        })
        .catch(err => setError(err.message))
        .finally(() => setLoading(false));
    } else {
      setLoading(true);
      api.getBookings({})
        .then(data => setBookings(data.data || []))
        .catch(err => setError(err.message))
        .finally(() => setLoading(false));
    }
  }, [activeTab, userLevel]);

  const groupSlotsByDate = (slots) => {
    const groups = {};
    slots.forEach(slot => {
      const date = new Date(slot.startTime).toDateString();
      if (!groups[date]) groups[date] = [];
      groups[date].push(slot);
    });
    return groups;
  };

  return (
    <div className="home-page">
      <header className="home-header">
        <h1>{activeTab === 'schedule' ? 'Расписание' : 'Мои бронирования'}</h1>
        <div style={{ display: 'flex', gap: '8px' }}>
          {isAdminOrTrainer && (
            <Button variant="ghost" size="sm" onClick={() => navigate('/admin')}>
              Админ
            </Button>
          )}
          <Button variant="ghost" size="sm" onClick={() => navigate('/profile')}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
              <circle cx="12" cy="7" r="4"/>
            </svg>
          </Button>
        </div>
      </header>

      <main className="home-content">
        {loading ? (
          <div className="home-loading">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="skeleton-card" />
            ))}
          </div>
        ) : error ? (
          <div className="home-error">
            <p>{error}</p>
            <Button onClick={() => window.location.reload()}>Повторить</Button>
          </div>
        ) : activeTab === 'schedule' ? (
          slots.length === 0 ? (
            <div className="home-empty">
              <p>Нет доступных слотов</p>
            </div>
          ) : (
            Object.entries(groupSlotsByDate(slots)).map(([date, dateSlots]) => (
              <div key={date} className="slots-group">
                <h3 className="slots-group-title">{formatDate(date)}</h3>
                <div className="slots-list">
                  {dateSlots.map(slot => (
                    <SlotCard
                      key={slot.id}
                      slot={slot}
                      onClick={() => navigate(`/slot/${slot.id}`)}
                    />
                  ))}
                </div>
              </div>
            ))
          )
        ) : bookings.length === 0 ? (
          <div className="home-empty">
            <p>У вас пока нет бронирований</p>
            <Button onClick={() => setActiveTab('schedule')}>Записаться</Button>
          </div>
        ) : (
          <div className="bookings-list">
            {bookings.map(booking => (
              <BookingCard
                key={booking.id}
                booking={booking}
                onClick={() => navigate(`/booking/${booking.id}`)}
              />
            ))}
          </div>
        )}
      </main>

      <nav className="home-nav">
        <button
          className={`nav-tab ${activeTab === 'schedule' ? 'active' : ''}`}
          onClick={() => setActiveTab('schedule')}
        >
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
            <line x1="16" y1="2" x2="16" y2="6"/>
            <line x1="8" y1="2" x2="8" y2="6"/>
            <line x1="3" y1="10" x2="21" y2="10"/>
          </svg>
          <span>Расписание</span>
        </button>
        <button
          className={`nav-tab ${activeTab === 'bookings' ? 'active' : ''}`}
          onClick={() => setActiveTab('bookings')}
        >
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
            <polyline points="10 9 9 9 8 9"/>
          </svg>
          <span>Брони</span>
        </button>
      </nav>
    </div>
  );
}
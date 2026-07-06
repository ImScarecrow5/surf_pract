import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card } from '../components/Card';
import { Badge } from '../components/Badge';
import { Button } from '../components/Button';
import { api } from '../services/api';
import './BookingDetail.css';

function formatDateTime(dateStr) {
  const date = new Date(dateStr);
  return date.toLocaleDateString('ru-RU', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    hour: '2-digit',
    minute: '2-digit'
  });
}

export function BookingDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [booking, setBooking] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [cancelling, setCancelling] = useState(false);
  const [showCancelConfirm, setShowCancelConfirm] = useState(false);

  useEffect(() => {
    api.getBookingById(id)
      .then(setBooking)
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  const handleCancel = async () => {
    setCancelling(true);
    try {
      const updated = await api.cancelBooking(id);
      setBooking(updated);
      setShowCancelConfirm(false);
    } catch (err) {
      setError(err.message);
    } finally {
      setCancelling(false);
    }
  };

  const canCancel = booking?.status === 'confirmed';
  const hoursUntilStart = booking?.slot?.startTime
    ? (new Date(booking.slot.startTime) - new Date()) / (1000 * 60 * 60)
    : 0;
  const isFreeCancellation = hoursUntilStart >= 2;

  if (loading) {
    return (
      <div className="booking-detail-page">
        <div className="booking-detail-loading">
          <div className="skeleton-card" style={{ height: 200 }} />
        </div>
      </div>
    );
  }

  if (error && !booking) {
    return (
      <div className="booking-detail-page">
        <div className="booking-detail-error">
          <p>{error}</p>
          <Button onClick={() => navigate(-1)}>Назад</Button>
        </div>
      </div>
    );
  }

  const statusVariant = {
    confirmed: 'success',
    cancelled_by_client_early: 'error',
    cancelled_by_client_late: 'error',
    cancelled_by_gym: 'muted'
  }[booking.status] || 'default';

  const statusText = {
    confirmed: 'Подтверждено',
    cancelled_by_client_early: 'Отменено (бесплатно)',
    cancelled_by_client_late: 'Отменено (штраф 10%)',
    cancelled_by_gym: 'Отменено скалодромом'
  }[booking.status] || booking.status;

  return (
    <div className="booking-detail-page">
      <header className="booking-detail-header">
        <button className="back-btn" onClick={() => navigate(-1)}>
          ← Назад
        </button>
        <h1>Детали бронирования</h1>
      </header>

      <main className="booking-detail-content">
        <Card padding="lg">
          <div className="booking-status">
            <Badge variant={statusVariant} size="lg">{statusText}</Badge>
          </div>

          <h2>{booking?.slot?.zone?.name}</h2>
          <p className="booking-datetime">{formatDateTime(booking?.slot?.startTime)}</p>

          <div className="booking-instructor">
            <div className="instructor-avatar">
              {booking?.slot?.instructor?.name?.charAt(0) || 'И'}
            </div>
            <span>{booking?.slot?.instructor?.name}</span>
          </div>

          <div className="booking-details">
            <div className="detail-row">
              <span>Тип снаряжения</span>
              <span>{booking?.equipmentType === 'own' ? 'Своё' : 'В аренду'}</span>
            </div>
            {booking?.equipment && (
              <div className="detail-row">
                <span>Снаряжение</span>
                <span>{booking.equipment.name}</span>
              </div>
            )}
            <div className="detail-row total">
              <span>Итого</span>
              <span>{booking?.price} ₽</span>
            </div>
            {booking?.cancellationPenalty > 0 && (
              <div className="detail-row penalty">
                <span>Штраф</span>
                <span>{booking.cancellationPenalty} ₽</span>
              </div>
            )}
          </div>
        </Card>

        {canCancel && (
          <>
            {!showCancelConfirm ? (
              <Button variant="secondary" fullWidth onClick={() => setShowCancelConfirm(true)}>
                Отменить бронирование
              </Button>
            ) : (
              <Card padding="md" className="cancel-confirm">
                <p>Вы уверены, что хотите отменить бронирование?</p>
                {!isFreeCancellation && (
                  <p className="cancel-penalty">Будет наложен штраф 10%</p>
                )}
                <div className="cancel-actions">
                  <Button variant="secondary" onClick={() => setShowCancelConfirm(false)}>
                    Нет
                  </Button>
                  <Button variant="primary" onClick={handleCancel} loading={cancelling}>
                    Да, отменить
                  </Button>
                </div>
              </Card>
            )}
          </>
        )}
      </main>
    </div>
  );
}
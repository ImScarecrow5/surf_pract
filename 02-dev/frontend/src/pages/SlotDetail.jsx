import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card } from '../components/Card';
import { Badge } from '../components/Badge';
import { Button } from '../components/Button';
import { api } from '../services/api';
import './SlotDetail.css';

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

export function SlotDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [slot, setSlot] = useState(null);
  const [equipment, setEquipment] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [bookingStep, setBookingStep] = useState('select');
  const [equipmentType, setEquipmentType] = useState('own');
  const [selectedEquipment, setSelectedEquipment] = useState(null);
  const [booking, setBooking] = useState(null);
  const [bookingLoading, setBookingLoading] = useState(false);

  useEffect(() => {
    Promise.all([
      api.getSlotById(id),
      api.getEquipment()
    ])
      .then(([slotData, equipmentData]) => {
        setSlot(slotData);
        setEquipment(equipmentData || []);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  const handleBooking = async () => {
    setBookingLoading(true);
    try {
      const data = await api.createBooking({
        slotId: parseInt(id),
        equipmentType,
        equipmentId: equipmentType === 'rental' && selectedEquipment ? selectedEquipment.id : null
      });
      setBooking(data);
      setBookingStep('success');
    } catch (err) {
      setError(err.message || 'Ошибка при бронировании');
    } finally {
      setBookingLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="slot-detail-page">
        <div className="slot-detail-loading">
          <div className="skeleton-card" style={{ height: 200 }} />
          <div className="skeleton-card" style={{ height: 100 }} />
        </div>
      </div>
    );
  }

  if (error && !slot) {
    return (
      <div className="slot-detail-page">
        <div className="slot-detail-error">
          <p>{error}</p>
          <Button onClick={() => navigate(-1)}>Назад</Button>
        </div>
      </div>
    );
  }

  if (bookingStep === 'success') {
    return (
      <div className="slot-detail-page">
        <div className="booking-success">
          <div className="success-icon">✓</div>
          <h2>Бронирование подтверждено</h2>
          <p>Вы забронировали место на тренировку</p>
          <p className="success-price">{booking?.price} ₽</p>
          <div className="success-actions">
            <Button onClick={() => navigate('/')}>К расписанию</Button>
            <Button variant="secondary" onClick={() => navigate('/profile')}>
              К профилю
            </Button>
          </div>
        </div>
      </div>
    );
  }

  const zoneType = slot?.zone?.name?.toLowerCase().includes('болот') ? 'boulder' : 'rope';
  const totalPrice = slot?.price + (equipmentType === 'rental' && selectedEquipment ? selectedEquipment.pricePerSlot : 0);

  return (
    <div className="slot-detail-page">
      <header className="slot-detail-header">
        <button className="back-btn" onClick={() => navigate(-1)}>
          ← Назад
        </button>
      </header>

      <main className="slot-detail-content">
        <Card className="slot-detail-card" padding="lg">
          <div className="slot-detail-zone">
            <Badge variant={zoneType === 'boulder' ? 'info' : 'primary'}>
              {zoneType === 'boulder' ? 'Болдеринг' : 'Трассы с верёвкой'}
            </Badge>
          </div>
          <h1>{slot?.zone?.name}</h1>
          <p className="slot-detail-datetime">{formatDateTime(slot?.startTime)}</p>
          <p className="slot-detail-duration">{slot?.zone?.durationMinutes} минут</p>

          <div className="slot-detail-instructor">
            <div className="instructor-avatar">
              {slot?.instructor?.name?.charAt(0) || 'И'}
            </div>
            <div className="instructor-info">
              <span className="instructor-name">{slot?.instructor?.name}</span>
              {slot?.instructor?.rating && (
                <span className="instructor-rating">★ {slot.instructor.rating}</span>
              )}
            </div>
          </div>

          <div className="slot-detail-places">
            <span>Свободно мест: {slot?.freePlaces} из {slot?.totalPlaces}</span>
          </div>
        </Card>

        {bookingStep === 'select' && (
          <>
            <Card className="equipment-card" padding="md">
              <h3>Снаряжение</h3>
              <div className="equipment-options">
                <label className={`equipment-option ${equipmentType === 'own' ? 'selected' : ''}`}>
                  <input
                    type="radio"
                    name="equipment"
                    checked={equipmentType === 'own'}
                    onChange={() => { setEquipmentType('own'); setSelectedEquipment(null); }}
                  />
                  <span className="option-title">Своё</span>
                  <span className="option-price">0 ₽</span>
                </label>
                <label className={`equipment-option ${equipmentType === 'rental' ? 'selected' : ''}`}>
                  <input
                    type="radio"
                    name="equipment"
                    checked={equipmentType === 'rental'}
                    onChange={() => setEquipmentType('rental')}
                  />
                  <span className="option-title">В аренду</span>
                </label>
              </div>

              {equipmentType === 'rental' && (
                <div className="equipment-list">
                  {equipment.map(item => (
                    <label
                      key={item.id}
                      className={`equipment-item ${selectedEquipment?.id === item.id ? 'selected' : ''}`}
                    >
                      <input
                        type="radio"
                        name="equipmentItem"
                        checked={selectedEquipment?.id === item.id}
                        onChange={() => setSelectedEquipment(item)}
                      />
                      <span className="equipment-name">{item.name}</span>
                      <span className="equipment-price">{item.pricePerSlot} ₽</span>
                      <span className="equipment-available">×{item.availableCount}</span>
                    </label>
                  ))}
                </div>
              )}
            </Card>

            <div className="slot-detail-footer">
              <div className="total-price">
                <span>Итого</span>
                <span className="price-value">{totalPrice} ₽</span>
              </div>
              <Button
                fullWidth
                disabled={equipmentType === 'rental' && !selectedEquipment}
                loading={bookingLoading}
                onClick={handleBooking}
              >
                Забронировать
              </Button>
            </div>
          </>
        )}
      </main>
    </div>
  );
}
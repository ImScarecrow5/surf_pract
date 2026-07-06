import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { useAuth } from '../context/AuthContext';
import './Admin.css';

export function AdminPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('bookings');
  const [users, setUsers] = useState([]);
  const [bookings, setBookings] = useState([]);
  const [mySlots, setMySlots] = useState([]);
  const [instructors, setInstructors] = useState([]);
  const [loading, setLoading] = useState(true);
  const [assignPhone, setAssignPhone] = useState('');
  const [assignInstructorId, setAssignInstructorId] = useState('');
  const [zones, setZones] = useState([]);
  const [newSlot, setNewSlot] = useState({ zoneId: '', startTime: '', totalPlaces: 8, price: 1000 });

  useEffect(() => {
    if (!user || (user.role !== 'admin' && user.role !== 'trainer')) {
      navigate('/');
      return;
    }
    loadData();
  }, [user]);

  const loadData = async () => {
    try {
      const isTrainer = user?.role === 'trainer';
      
      let usersData = [];
      let bookingsData = [];
      let slotsData = [];
      let instructorsData = [];

      if (user.role === 'admin') {
        [usersData, instructorsData] = await Promise.all([
          api.getAllUsers(),
          api.getInstructors()
        ]);
      }

      if (isTrainer) {
        const instructorId = user.instructorId;
        const [myBookings, allSlots, fetchedZones] = await Promise.all([
          api.getAllBookings(),
          api.getSlots({ instructorId: instructorId }),
          api.getZones()
        ]);
        
        bookingsData = myBookings.filter(b => b.slot?.instructor?.id === instructorId);
        slotsData = allSlots.data || [];
        setZones(fetchedZones || []);
      } else {
        bookingsData = await api.getAllBookings();
      }

      setUsers(usersData || []);
      setBookings(bookingsData || []);
      setMySlots(slotsData || []);
      setInstructors(instructorsData || []);
    } catch (err) {
      console.error('Load data error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleRoleChange = async (userId, newRole) => {
    try {
      await api.updateUserRole(userId, newRole);
      loadData();
    } catch (err) {
      alert('Ошибка изменения роли');
    }
  };

  const handleAssignTrainer = async () => {
    if (!assignPhone || !assignInstructorId) {
      alert('Заполните все поля');
      return;
    }
    try {
      await api.assignTrainer(assignPhone, parseInt(assignInstructorId));
      alert('Тренер назначен');
      setAssignPhone('');
      setAssignInstructorId('');
      loadData();
    } catch (err) {
      alert('Ошибка назначения тренера');
    }
  };

  const handleCancelBooking = async (bookingId) => {
    if (!confirm('Отменить бронирование?')) return;
    try {
      await api.adminCancelBooking(bookingId);
      loadData();
    } catch (err) {
      alert('Ошибка отмены');
    }
  };

  const handleCreateSlot = async () => {
    if (!newSlot.zoneId || !newSlot.startTime) {
      alert('Заполните зону и время');
      return;
    }
    try {
      await api.createSlot({
        zoneId: parseInt(newSlot.zoneId),
        startTime: newSlot.startTime,
        totalPlaces: parseInt(newSlot.totalPlaces),
        price: parseInt(newSlot.price)
      });
      alert('Слот создан');
      setNewSlot({ zoneId: '', startTime: '', totalPlaces: 8, price: 1000 });
      loadData();
    } catch (err) {
      alert('Ошибка создания слота');
    }
  };

  if (loading) {
    return <div className="admin-loading">Загрузка...</div>;
  }

  const isAdmin = user?.role === 'admin';
  const isTrainer = user?.role === 'trainer';

  return (
    <div className="admin-page">
      <div className="admin-header">
        <h1>{isTrainer ? 'Мои тренировки' : 'Админ-панель'}</h1>
        <button className="admin-back" onClick={() => navigate('/')}>На главную</button>
      </div>

      <div className="admin-tabs">
        <button 
          className={`admin-tab ${activeTab === 'bookings' ? 'active' : ''}`}
          onClick={() => setActiveTab('bookings')}
        >
          Записавшиеся
        </button>
        {isTrainer && (
          <button 
            className={`admin-tab ${activeTab === 'schedule' ? 'active' : ''}`}
            onClick={() => setActiveTab('schedule')}
          >
            Мое расписание
          </button>
        )}
        {isTrainer && (
          <button 
            className={`admin-tab ${activeTab === 'createSlot' ? 'active' : ''}`}
            onClick={() => setActiveTab('createSlot')}
          >
            Создать слот
          </button>
        )}
        {isAdmin && (
          <button 
            className={`admin-tab ${activeTab === 'users' ? 'active' : ''}`}
            onClick={() => setActiveTab('users')}
          >
            Пользователи
          </button>
        )}
        {isAdmin && (
          <button 
            className={`admin-tab ${activeTab === 'trainers' ? 'active' : ''}`}
            onClick={() => setActiveTab('trainers')}
          >
            Назначить тренера
          </button>
        )}
      </div>

      <div className="admin-content">
        {activeTab === 'bookings' && (
          <div className="admin-bookings">
            <h2>{isTrainer ? 'Записавшиеся на мои тренировки' : 'Все бронирования'}</h2>
            {bookings.length === 0 ? (
              <p className="admin-empty">Нет бронирований</p>
            ) : (
              <table className="admin-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Клиент</th>
                    <th>Телефон</th>
                    <th>Дата</th>
                    <th>Зона</th>
                    {!isTrainer && <th>Тренер</th>}
                    <th>Статус</th>
                    <th>Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {bookings.map(b => (
                    <tr key={b.id}>
                      <td>{b.id}</td>
                      <td>{b.clientName || '-'}</td>
                      <td>{b.clientPhone}</td>
                      <td>{new Date(b.slot?.startTime).toLocaleString()}</td>
                      <td>{b.slot?.zone?.name}</td>
                      {!isTrainer && <td>{b.slot?.instructor?.name}</td>}
                      <td>{b.status}</td>
                      <td>
                        <button 
                          className="admin-btn admin-btn-danger"
                          onClick={() => handleCancelBooking(b.id)}
                        >
                          Отменить
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {activeTab === 'schedule' && isTrainer && (
          <div className="admin-schedule">
            <h2>Мои сллоты</h2>
            {mySlots.length === 0 ? (
              <p className="admin-empty">Нет слотов в расписании</p>
            ) : (
              <table className="admin-table">
                <thead>
                  <tr>
                    <th>Дата</th>
                    <th>Время</th>
                    <th>Зона</th>
                    <th>Мест</th>
                    <th>Цена</th>
                  </tr>
                </thead>
                <tbody>
                  {mySlots.map(s => (
                    <tr key={s.id}>
                      <td>{new Date(s.startTime).toLocaleDateString('ru-RU')}</td>
                      <td>{new Date(s.startTime).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}</td>
                      <td>{s.zone?.name}</td>
                      <td>{s.freePlaces} / {s.totalPlaces}</td>
                      <td>{s.price} ₽</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {activeTab === 'createSlot' && isTrainer && (
          <div className="admin-create-slot">
            <h2>Создать новый слот</h2>
            <div className="assign-form">
              <select 
                value={newSlot.zoneId}
                onChange={(e) => setNewSlot({...newSlot, zoneId: e.target.value})}
                className="admin-select"
              >
                <option value="">Выберите зону</option>
                {zones.map(z => (
                  <option key={z.id} value={z.id}>{z.name} ({z.durationMinutes} мин)</option>
                ))}
              </select>
              <input
                type="datetime-local"
                value={newSlot.startTime}
                onChange={(e) => setNewSlot({...newSlot, startTime: e.target.value})}
                className="admin-input"
              />
              <input
                type="number"
                placeholder="Мест"
                value={newSlot.totalPlaces}
                onChange={(e) => setNewSlot({...newSlot, totalPlaces: e.target.value})}
                className="admin-input"
                style={{width: '100px'}}
              />
              <input
                type="number"
                placeholder="Цена"
                value={newSlot.price}
                onChange={(e) => setNewSlot({...newSlot, price: e.target.value})}
                className="admin-input"
                style={{width: '100px'}}
              />
              <button className="admin-btn" onClick={handleCreateSlot}>
                Создать
              </button>
            </div>
          </div>
        )}

        {activeTab === 'users' && isAdmin && (
          <div className="admin-users">
            <h2>Все пользователи</h2>
            <table className="admin-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Телефон</th>
                  <th>Имя</th>
                  <th>Роль</th>
                  <th>Тренер</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {users.map(u => (
                  <tr key={u.id}>
                    <td>{u.id}</td>
                    <td>{u.phone}</td>
                    <td>{u.name || '-'}</td>
                    <td>
                      <select 
                        value={u.role || 'client'} 
                        onChange={(e) => handleRoleChange(u.id, e.target.value)}
                      >
                        <option value="client">Клиент</option>
                        <option value="trainer">Тренер</option>
                        <option value="admin">Админ</option>
                      </select>
                    </td>
                    <td>{u.instructorName || '-'}</td>
                    <td>
                      <button 
                        className="admin-btn admin-btn-danger"
                        onClick={() => handleRoleChange(u.id, 'client')}
                      >
                        Сбросить
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {activeTab === 'trainers' && isAdmin && (
          <div className="admin-trainers">
            <h2>Назначить тренера по номеру телефона</h2>
            <div className="assign-form">
              <input
                type="text"
                placeholder="Номер телефона (например, +79991234567)"
                value={assignPhone}
                onChange={(e) => setAssignPhone(e.target.value)}
                className="admin-input"
              />
              <select 
                value={assignInstructorId}
                onChange={(e) => setAssignInstructorId(e.target.value)}
                className="admin-select"
              >
                <option value="">Выберите тренера</option>
                {instructors.map(i => (
                  <option key={i.id} value={i.id}>{i.name}</option>
                ))}
              </select>
              <button className="admin-btn" onClick={handleAssignTrainer}>
                Назначить
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
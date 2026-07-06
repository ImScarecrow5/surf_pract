import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PhoneInput, Input } from '../components/Input';
import { Button } from '../components/Button';
import { useAuth } from '../context/AuthContext';
import './Auth.css';

export function AuthPage() {
  const navigate = useNavigate();
  const { requestCode, login } = useAuth();
  const [step, setStep] = useState('phone');
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const isValidPhone = phone.replace(/\D/g, '').length === 11;
  const isValidCode = code.length >= 4;

  const handleRequestCode = async () => {
    if (!isValidPhone) return;
    setLoading(true);
    setError('');
    try {
      const rawPhone = '+' + phone.replace(/\D/g, '');
      const response = await requestCode(rawPhone);
      console.log('SMS Code:', response.code);
      setStep('code');
    } catch (err) {
      setError(err.message || 'Ошибка отправки кода');
    } finally {
      setLoading(false);
    }
  };

  const handleVerify = async () => {
    if (!isValidCode) return;
    setLoading(true);
    setError('');
    try {
      const rawPhone = '+' + phone.replace(/\D/g, '');
      await login(rawPhone, code);
      navigate('/');
    } catch (err) {
      setError(err.message || 'Неверный код');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page">
      <div className="auth-card">
        <div className="auth-logo">
          <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
            <path d="M24 4L4 24L24 44L44 24L24 4Z" fill="#ff6b35"/>
            <path d="M24 14L14 24L24 34L34 24L24 14Z" fill="#fff"/>
          </svg>
          <span className="auth-logo-text">Вертикаль</span>
        </div>

        <h1 className="auth-title">Добро пожаловать</h1>
        <p className="auth-subtitle">Скалодром «Вертикаль»</p>

        <div className="auth-admin-hint">
          <p>Для входа как админ используйте: <strong>+79999999999</strong></p>
          <p>Код: 1234</p>
        </div>

        {step === 'phone' ? (
          <div className="auth-form">
            <PhoneInput
              label="Номер телефона"
              value={phone}
              onChange={setPhone}
              error={error}
              required
            />
            <Button
              fullWidth
              disabled={!isValidPhone}
              loading={loading}
              onClick={handleRequestCode}
            >
              Продолжить
            </Button>
            <p className="auth-terms">
              Нажимая, вы принимаете{' '}
              <a href="#">условия использования</a>
            </p>
          </div>
        ) : (
          <div className="auth-form">
            <p className="auth-code-sent">Код отправлен на {phone}</p>
            <Input
              label="Код из SMS"
              type="text"
              value={code}
              onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
              placeholder="0000"
              error={error}
              required
              autoComplete="one-time-code"
            />
            <Button
              fullWidth
              disabled={!isValidCode}
              loading={loading}
              onClick={handleVerify}
            >
              Войти
            </Button>
            <button
              className="auth-back"
              onClick={() => { setStep('phone'); setCode(''); setError(''); }}
            >
              Изменить номер
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
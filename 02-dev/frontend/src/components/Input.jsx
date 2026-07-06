import './Input.css';

export function Input({
  label,
  type = 'text',
  value,
  onChange,
  placeholder,
  error,
  disabled = false,
  required = false,
  name,
  id,
  className = '',
  autoComplete
}) {
  const inputId = id || name;

  return (
    <div className={`input-wrapper ${className}`}>
      {label && (
        <label htmlFor={inputId} className="input-label">
          {label}
          {required && <span className="input-required">*</span>}
        </label>
      )}
      <input
        type={type}
        id={inputId}
        name={name}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        disabled={disabled}
        required={required}
        autoComplete={autoComplete}
        className={`input-field ${error ? 'input-error' : ''}`}
      />
      {error && <span className="input-error-text">{error}</span>}
    </div>
  );
}

export function PhoneInput({
  label,
  value,
  onChange,
  error,
  disabled = false,
  required = false,
  className = ''
}) {
  const formatPhone = (input) => {
    const digits = input.replace(/\D/g, '').slice(0, 11);
    if (digits.length === 0) return '';
    if (digits.length <= 1) return '+' + digits;
    if (digits.length <= 4) return `+${digits.slice(0, 1)} (${digits.slice(1)})`;
    if (digits.length <= 7) return `+${digits.slice(0, 1)} (${digits.slice(1, 4)}) ${digits.slice(4)}`;
    if (digits.length <= 9) return `+${digits.slice(0, 1)} (${digits.slice(1, 4)}) ${digits.slice(4, 7)}-${digits.slice(7)}`;
    return `+${digits.slice(0, 1)} (${digits.slice(1, 4)}) ${digits.slice(4, 7)}-${digits.slice(7, 9)}-${digits.slice(9)}`;
  };

  const handleChange = (e) => {
    const formatted = formatPhone(e.target.value);
    onChange(formatted);
  };

  PhoneInput.getRawValue = (value) => {
    return value.replace(/\D/g, '');
  };

  return (
    <Input
      label={label}
      type="tel"
      value={value}
      onChange={handleChange}
      placeholder="+7 (999) 000-00-00"
      error={error}
      disabled={disabled}
      required={required}
      name="phone"
      autoComplete="tel"
      className={className}
    />
  );
}
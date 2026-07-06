import './Card.css';

export function Card({
  children,
  variant = 'default',
  padding = 'md',
  className = '',
  onClick,
  disabled = false
}) {
  const classes = [
    'card',
    `card-${variant}`,
    `card-padding-${padding}`,
    onClick && 'card-clickable',
    disabled && 'card-disabled',
    className
  ].filter(Boolean).join(' ');

  return (
    <div className={classes} onClick={disabled ? undefined : onClick}>
      {children}
    </div>
  );
}

export function CardHeader({ children, className = '' }) {
  return <div className={`card-header ${className}`}>{children}</div>;
}

export function CardBody({ children, className = '' }) {
  return <div className={`card-body ${className}`}>{children}</div>;
}

export function CardFooter({ children, className = '' }) {
  return <div className={`card-footer ${className}`}>{children}</div>;
}
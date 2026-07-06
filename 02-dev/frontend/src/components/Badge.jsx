import './Badge.css';

export function Badge({
  children,
  variant = 'default',
  size = 'md',
  className = ''
}) {
  const classes = [
    'badge',
    `badge-${variant}`,
    `badge-${size}`,
    className
  ].filter(Boolean).join(' ');

  return <span className={classes}>{children}</span>;
}
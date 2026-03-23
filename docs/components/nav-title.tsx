'use client';

export function NavTitle() {
  return (
    <span
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        window.location.href = '/';
      }}
      style={{ display: 'flex', alignItems: 'center', gap: '0.625rem', cursor: 'pointer' }}
    >
      <img src="/docs/logo.png" alt="Nebula API" style={{ height: 24 }} />
      <span>Nebula API</span>
    </span>
  );
}

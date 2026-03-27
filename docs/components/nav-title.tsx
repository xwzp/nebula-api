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
      <div
        style={{
          width: 28,
          height: 28,
          borderRadius: 6,
          background: 'linear-gradient(135deg, #ff3bff, #ec4899, #8b5cf6, #3b82f6)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: '0 1px 2px rgba(0, 0, 0, 0.05)',
        }}
      >
        <svg
          width={16}
          height={16}
          viewBox="0 0 24 24"
          fill="white"
        >
          <path d="M12 2l2.4 7.2L22 12l-7.6 2.8L12 22l-2.4-7.2L2 12l7.6-2.8L12 2z" />
        </svg>
      </div>
      <span
        style={{
          background: 'linear-gradient(to right, #ff3bff, #ec4899, #8b5cf6, #3b82f6)',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
          backgroundClip: 'text',
          fontWeight: 700,
          fontSize: '1.1rem',
          letterSpacing: '-0.02em',
        }}
      >
        Nebula API
      </span>
    </span>
  );
}

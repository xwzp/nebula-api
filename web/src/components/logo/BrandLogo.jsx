import React, { useState } from 'react';
import { getLogo } from '../../helpers/utils';

const DefaultLogoSvg = ({ size }) => {
  const iconSize = Math.round(size * 0.5625);
  return (
    <div
      className='relative flex items-center justify-center'
      style={{
        width: size,
        height: size,
        borderRadius: Math.round(size * 0.25),
        background:
          'linear-gradient(135deg, #ff3bff, #ec4899, #8b5cf6, #3b82f6)',
        boxShadow: '0 1px 3px rgba(0, 0, 0, 0.1)',
      }}
    >
      <svg
        width={iconSize}
        height={iconSize}
        viewBox='0 0 24 24'
        fill='white'
      >
        <path d='M12 2l2.4 7.2L22 12l-7.6 2.8L12 22l-2.4-7.2L2 12l7.6-2.8L12 2z' />
      </svg>
    </div>
  );
};

const BrandLogo = ({ size = 32, className = '' }) => {
  const [imgFailed, setImgFailed] = useState(false);
  const customLogo = getLogo();

  if (customLogo && !imgFailed) {
    return (
      <img
        src={customLogo}
        alt='Logo'
        className={className}
        style={{ width: size, height: size }}
        onError={() => setImgFailed(true)}
      />
    );
  }

  return <DefaultLogoSvg size={size} />;
};

export default BrandLogo;

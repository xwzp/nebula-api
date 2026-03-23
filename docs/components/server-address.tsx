'use client';

import { useEffect, useState } from 'react';

export function ServerAddress() {
  const [address, setAddress] = useState('');

  useEffect(() => {
    fetch('/api/status')
      .then((res) => res.json())
      .then((data) => {
        if (data.success && data.data?.server_address) {
          setAddress(data.data.server_address.replace(/\/$/, ''));
        } else {
          setAddress(window.location.origin);
        }
      })
      .catch(() => {
        setAddress(window.location.origin);
      });
  }, []);

  if (!address) return <code>loading...</code>;
  return <code className="font-semibold">{address}</code>;
}

export function ServerAddressText() {
  const [address, setAddress] = useState('https://...');

  useEffect(() => {
    fetch('/api/status')
      .then((res) => res.json())
      .then((data) => {
        if (data.success && data.data?.server_address) {
          setAddress(data.data.server_address.replace(/\/$/, ''));
        } else {
          setAddress(window.location.origin);
        }
      })
      .catch(() => {
        setAddress(window.location.origin);
      });
  }, []);

  return <>{address}</>;
}

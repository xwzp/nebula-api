'use client';

import { useEffect, useState } from 'react';

function useServerAddress() {
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

  return address;
}

export function ServerAddress() {
  const address = useServerAddress();
  if (!address) return <code>loading...</code>;
  return <code className="font-semibold">{address}</code>;
}

export function ServerAddressCurl() {
  const address = useServerAddress();
  const domain = address || 'https://...';

  return (
    <pre className="rounded-lg border bg-fd-secondary/50 p-4 text-sm overflow-x-auto">
      <code>{`curl ${domain}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-你的令牌" \\
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "user", "content": "你好"}
    ]
  }'`}</code>
    </pre>
  );
}

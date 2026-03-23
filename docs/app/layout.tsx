import './global.css';
import { RootProvider } from 'fumadocs-ui/provider';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import type { ReactNode } from 'react';
import type { Metadata } from 'next';
import { source } from '@/lib/source';

export const metadata: Metadata = {
  title: {
    template: '%s | Nebula API',
    default: 'Nebula API 文档',
  },
  description: 'Nebula API 文档 - AI API 中转站',
};

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="zh-CN" suppressHydrationWarning>
      <body className="flex flex-col min-h-screen">
        <RootProvider>
          <DocsLayout
            tree={source.pageTree}
            nav={{
              url: '/',
              title: (
                <>
                  <img src="/docs/logo.png" alt="Nebula API" style={{ height: 24 }} />
                  <span>Nebula API</span>
                </>
              ),
            }}
          >
            {children}
          </DocsLayout>
        </RootProvider>
      </body>
    </html>
  );
}

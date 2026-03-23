import './global.css';
import { RootProvider } from 'fumadocs-ui/provider';
import { DocsLayout } from 'fumadocs-ui/layouts/docs';
import type { ReactNode } from 'react';
import type { Metadata } from 'next';
import { source } from '@/lib/source';
import { NavTitle } from '@/components/nav-title';

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
              title: <NavTitle />,
            }}
            links={[
              { text: '首页', url: '/', external: true },
              { text: '控制台', url: '/console', external: true },
              { text: '关于', url: '/about', external: true },
            ]}
          >
            {children}
          </DocsLayout>
        </RootProvider>
      </body>
    </html>
  );
}

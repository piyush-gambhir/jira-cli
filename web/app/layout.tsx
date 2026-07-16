import type { Metadata } from 'next';
import localFont from 'next/font/local';
import type { CSSProperties } from 'react';
import { Provider } from '@/components/provider';
import { site } from '@/lib/site';
import './global.css';

const satoshi = localFont({
  src: '../fonts/Satoshi-Variable.woff2',
  weight: '300 900',
  display: 'swap',
  variable: '--font-satoshi',
});

const clashDisplay = localFont({
  src: '../fonts/ClashDisplay-Variable.woff2',
  weight: '200 700',
  display: 'swap',
  variable: '--font-clash-display',
});

const jetBrainsMono = localFont({
  src: [
    { path: '../fonts/JetBrainsMono-Regular.woff2', weight: '400' },
    { path: '../fonts/JetBrainsMono-Medium.woff2', weight: '500' },
    { path: '../fonts/JetBrainsMono-SemiBold.woff2', weight: '600' },
  ],
  display: 'swap',
  variable: '--font-jetbrains-mono',
});

export const metadata: Metadata = {
  metadataBase: new URL(`https://${site.repo.split('/')[1]}.pages.dev`),
  title: {
    default: `${site.name} — ${site.tagline}`,
    template: `%s · ${site.name}`,
  },
  description: site.description,
};

export default function Layout({ children }: LayoutProps<'/'>) {
  const rootStyle = {
    '--font-display': 'var(--font-clash-display)',
    '--font-mono': 'var(--font-jetbrains-mono)',
    ...(site.accent ? { '--accent': site.accent } : {}),
  } as CSSProperties;

  return (
    <html
      lang="en"
      className={`${satoshi.variable} ${clashDisplay.variable} ${jetBrainsMono.variable} ${satoshi.className}`}
      data-accent={site.accentName}
      style={rootStyle}
      suppressHydrationWarning
    >
      <head>
        <script
          dangerouslySetInnerHTML={{
            __html: "document.documentElement.classList.add('js')",
          }}
        />
      </head>
      <body className="flex flex-col min-h-screen">
        <Provider>{children}</Provider>
      </body>
    </html>
  );
}

import type { Metadata } from 'next';
import localFont from 'next/font/local';
import type { CSSProperties } from 'react';
import { Provider } from '@/components/provider';
import { defaultSocialImage } from '@/lib/metadata';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';
import './global.css';

const metadataDescription =
  'Agent-ready Jira CLI for Claude Code, OpenAI Codex, Cursor, or any shell harness. Issues, JQL, sprints, and boards via JSON/YAML, read-only, and no-input flags.';
const openGraphDescription =
  'Use Jira CLI from any coding agent or shell harness. JSON/YAML, read-only safety, and no-input automation cover issues, JQL, sprints, and boards.';
const twitterDescription =
  'Harness-agnostic Jira CLI for coding agents. Automate issues, JQL, sprints, and boards from the shell with JSON/YAML, read-only mode, and no-input flags.';

const hafferVF = localFont({
  src: '../fonts/haffer-vf-thin-2.ttf',
  variable: '--font-haffer-vf',
  weight: '100 1000',
  style: 'normal',
  display: 'swap',
});

const hafferXH = localFont({
  src: '../fonts/haffer-xh-regular-2.woff2',
  variable: '--font-haffer-xh',
  weight: '400',
  style: 'normal',
  display: 'swap',
});

const hafferMono = localFont({
  src: [
    {
      path: '../fonts/haffer-mono-regular-2.woff2',
      weight: '400',
      style: 'normal',
    },
    {
      path: '../fonts/haffer-mono-medium-2.woff2',
      weight: '500',
      style: 'normal',
    },
  ],
  variable: '--font-haffer-mono',
  display: 'swap',
});

const brisa = localFont({
  src: '../fonts/brisa-pro-regular-2.woff2',
  variable: '--font-brisa',
  weight: '400',
  style: 'normal',
  display: 'swap',
});

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  applicationName: site.name,
  title: {
    default: `${site.name}: ${site.tagline}`,
    template: `%s · ${site.name}`,
  },
  description: metadataDescription,
  alternates: {
    canonical: siteUrl,
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: siteUrl,
    siteName: site.name,
    title: `${site.name}: ${site.tagline}`,
    description: openGraphDescription,
    images: [
      {
        url: `${siteUrl}${defaultSocialImage}`,
        width: 1200,
        height: 630,
        alt: `${site.name} documentation`,
      },
    ],
  },
  twitter: {
    card: 'summary_large_image',
    title: `${site.name}: ${site.tagline}`,
    description: twitterDescription,
    images: [
      {
        url: `${siteUrl}${defaultSocialImage}`,
        width: 1200,
        height: 630,
        alt: `${site.name} documentation`,
      },
    ],
  },
  icons: {
    icon: [{ url: '/jira-cli/favicon.svg', type: 'image/svg+xml' }],
  },
};

export default function Layout({ children }: LayoutProps<'/'>) {
  const rootStyle = {
    ...(site.accent ? { '--site-accent': site.accent } : {}),
  } as CSSProperties;

  const fontVariables = [
    hafferVF.variable,
    hafferXH.variable,
    hafferMono.variable,
    brisa.variable,
  ].join(' ');

  return (
    <html
      lang="en"
      className={`${fontVariables} ${hafferVF.className}`}
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

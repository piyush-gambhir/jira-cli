import type { Metadata } from 'next';
import '@fontsource-variable/inter';
import '@fontsource-variable/jetbrains-mono';
import '@fontsource/instrument-serif';
import type { CSSProperties } from 'react';
import { Provider } from '@/components/provider';
import { homeSocialImage } from '@/lib/metadata';
import { site } from '@/lib/site';
import { siteUrl } from '@/lib/shared';
import './global.css';

const metadataDescription =
  'Agent-ready Jira CLI for Claude Code, OpenAI Codex, Cursor, or any shell harness. Issues, JQL, sprints, and boards via JSON/YAML, read-only, and no-input flags.';
const openGraphDescription =
  'Use Jira CLI from any coding agent or shell harness. JSON/YAML, read-only safety, and no-input automation cover issues, JQL, sprints, and boards.';
const twitterDescription =
  'Harness-agnostic Jira CLI for coding agents. Automate issues, JQL, sprints, and boards from the shell with JSON/YAML, read-only mode, and no-input flags.';

export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  applicationName: site.name,
  title: {
    default: `${site.name}: ${site.tagline}`,
    template: `%s · ${site.name}`,
  },
  description: metadataDescription,
  authors: [{ name: 'Piyush Gambhir', url: 'https://github.com/piyush-gambhir' }],
  creator: 'Piyush Gambhir',
  publisher: 'Piyush Gambhir',
  alternates: {
    canonical: siteUrl,
  },
  icons: {
    icon: [{ url: '/jira-cli/favicon.svg', type: 'image/svg+xml' }],
  },
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: siteUrl,
    siteName: site.name,
    title: `${site.name}: ${site.tagline}`,
    description: openGraphDescription,
    images: [homeSocialImage],
  },
  twitter: {
    card: 'summary_large_image',
    title: `${site.name}: ${site.tagline}`,
    description: twitterDescription,
    images: [homeSocialImage],
  },
};

export default function Layout({ children }: LayoutProps<'/'>) {
  const rootStyle = {
    ...(site.accent ? { '--site-accent': site.accent } : {}),
  } as CSSProperties;

  return (
    <html
      lang="en"
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

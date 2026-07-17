export const dynamic = 'force-static';
import type { MetadataRoute } from 'next';
import { source } from '@/lib/source';
import { siteUrl } from '@/lib/shared';

const legalRoutes = ['/privacy', '/terms', '/contact'] as const;

export default function sitemap(): MetadataRoute.Sitemap {
  const paths = [
    '/',
    ...source.getPages().map((page) => page.url),
    ...legalRoutes,
  ];

  return [...new Set(paths)].map((path) => ({
    url: `${siteUrl}${path}`,
  }));
}

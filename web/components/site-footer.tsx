import Link from 'next/link';
import type { ReactNode } from 'react';
import { InstallCommand } from '@/components/install-command';
import { site } from '@/lib/site';
import { getOtherSuiteProjects } from '@/lib/suite';

function FooterHeading({ children }: { children: ReactNode }) {
  return (
    <h2 className="flex items-center gap-2 font-mono text-[0.6875rem] font-medium uppercase tracking-[0.1em] text-muted-foreground">
      <span
        aria-hidden
        className="size-1.5 rounded-full bg-[var(--accent)]"
      />
      {children}
    </h2>
  );
}

const internalLinkClass =
  'text-sm text-muted-foreground hover:text-[var(--accent)]';
const externalLinkClass =
  'text-sm text-muted-foreground hover:text-[var(--accent)]';

export function SiteFooter() {
  const repoUrl = `https://github.com/${site.repo}`;
  const otherTools = getOtherSuiteProjects(site.repo);
  const year = new Date().getFullYear();

  return (
    <footer className="flex min-h-0 flex-col justify-between gap-16 overflow-hidden border-t border-border bg-fd-muted/20 pt-16 md:min-h-[70svh] md:gap-20 md:pt-24">
      <div className="mx-auto grid w-full max-w-6xl grid-cols-2 gap-x-6 gap-y-12 px-4 sm:px-6 lg:grid-cols-[0.8fr_0.8fr_1fr_2fr] lg:gap-10">
        <div>
          <FooterHeading>Documentation</FooterHeading>
          <div className="mt-5 flex flex-col items-start gap-3">
            <Link href="/docs" className={internalLinkClass}>
              Introduction
            </Link>
            <Link href="/docs/installation" className={internalLinkClass}>
              Installation
            </Link>
            <Link href="/docs/authentication" className={internalLinkClass}>
              Authentication
            </Link>
            <Link href="/docs/quickstart" className={internalLinkClass}>
              Quick start
            </Link>
          </div>
        </div>

        <div>
          <FooterHeading>Project</FooterHeading>
          <div className="mt-5 flex flex-col items-start gap-3">
            <a
              href={repoUrl}
              target="_blank"
              rel="noreferrer"
              className={externalLinkClass}
            >
              GitHub
            </a>
            <a
              href={`${repoUrl}/releases`}
              target="_blank"
              rel="noreferrer"
              className={externalLinkClass}
            >
              Releases
            </a>
            <a
              href={`${repoUrl}/issues`}
              target="_blank"
              rel="noreferrer"
              className={externalLinkClass}
            >
              Issues
            </a>
            <a
              href={`${repoUrl}/blob/main/LICENSE`}
              target="_blank"
              rel="noreferrer"
              className={externalLinkClass}
            >
              License
            </a>
          </div>
        </div>

        <div>
          <FooterHeading>More tools</FooterHeading>
          <div className="mt-5 flex flex-col items-start gap-3 font-mono">
            {otherTools.map((project) => (
              <a
                key={project.name}
                href={project.href}
                target="_blank"
                rel="noreferrer"
                className={externalLinkClass}
              >
                {project.name}
              </a>
            ))}
          </div>
        </div>

        <div className="col-span-2 lg:col-span-1">
          <FooterHeading>Get started in seconds</FooterHeading>
          <p className="mt-5 max-w-md text-sm leading-relaxed text-muted-foreground">
            Install {site.binary}, then follow the quick start to connect your
            first profile.
          </p>
          <InstallCommand
            command={site.installCommand}
            className="mt-5 max-w-none border border-border bg-background/60"
          />
        </div>
      </div>

      <div className="mx-auto w-full max-w-6xl px-4 sm:px-6">
        <p className="max-w-prose text-xs leading-relaxed text-muted-foreground">
          {site.name} is an independent, open-source project. It is{' '}
          <span className="font-medium text-fd-foreground/80">
            not affiliated with, endorsed by, or sponsored by
          </span>{' '}
          the makers of the underlying software. All product names, logos, and
          trademarks are the property of their respective owners and are used
          for identification purposes only.
        </p>
      </div>

      <div>
        <div className="mx-auto flex w-full max-w-6xl flex-col gap-2 px-4 font-mono text-[0.6875rem] text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-6">
          <span>© {year} {site.name}</span>
          <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
            <a
              href="https://github.com/piyush-gambhir"
              target="_blank"
              rel="noreferrer"
              className="hover:text-[var(--accent)]"
            >
              Built by Piyush Gambhir
            </a>
            <a
              href={`${repoUrl}/blob/main/LICENSE`}
              target="_blank"
              rel="noreferrer"
              className="hover:text-[var(--accent)]"
            >
              MIT licensed
            </a>
          </div>
        </div>

        <div
          aria-hidden="true"
          className="-mb-[4vw] flex select-none justify-center overflow-visible pt-8 font-mono text-[20vw] leading-none font-medium tracking-[-0.08em] whitespace-nowrap text-foreground opacity-[0.05] dark:opacity-[0.07]"
        >
          {site.binary}
        </div>
      </div>
    </footer>
  );
}

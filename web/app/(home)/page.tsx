import Link from 'next/link';
import { HomeHero } from '@/components/home-hero';
import { Reveal } from '@/components/reveal';
import { SiteFooter } from '@/components/site-footer';
import {
  licenseUrl,
  repositoryUrl,
  site,
  siteMetadataDescription,
  type SiteStep,
} from '@/lib/site';
import { siteUrl } from '@/lib/shared';
import { getOtherSuiteProjects } from '@/lib/suite';

const revealDelays = ['0s', '0.075s', '0.15s'] as const;
const contextualBodyLinks: Record<
  string,
  { href: string; text: string }
> = {
  Install: {
    href: '/docs/installation',
    text: 'Install the latest jira-cli binary',
  },
  Authenticate: {
    href: '/docs/authentication',
    text: 'Create a Jira profile',
  },
  Run: {
    href: '/docs/quickstart',
    text: 'Search, update, and automate Jira',
  },
  'JQL & issues': {
    href: '/docs/commands/issues',
    text: 'Search with raw JQL',
  },
  'Agent-friendly': {
    href: '/docs/agents',
    text: '-o json|yaml on every read',
  },
  'Agile built in': {
    href: '/docs/commands/agile',
    text: 'Boards, sprints, epics, backlogs',
  },
};

export default function HomePage() {
  const firstExampleCommand =
    site.example
      .split('\n')
      .find((line) => line.trim() && !line.trimStart().startsWith('#')) ??
    `${site.binary} --help`;
  const fallbackSteps: SiteStep[] = [
    {
      title: 'Install',
      body: `Install ${site.name} and make the binary available from your shell.`,
      snippet: site.installCommand,
    },
    {
      title: 'Authenticate',
      body: `Connect ${site.binary} to your account with the credentials your deployment supports.`,
      snippet: `${site.binary} auth login`,
    },
    {
      title: 'Run',
      body: 'Start with a real command, then compose it into scripts and agent workflows.',
      snippet: firstExampleCommand,
    },
  ];
  const steps = site.steps?.length ? site.steps : fallbackSteps;
  const relatedLinks = getOtherSuiteProjects(site.repo).map(({ href }) => href);
  const softwareApplication = {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    '@id': `${siteUrl}/#software-application`,
    name: site.name,
    applicationCategory: 'DeveloperApplication',
    operatingSystem: ['macOS', 'Linux', 'Windows'],
    license: licenseUrl,
    url: siteUrl,
    sameAs: [repositoryUrl],
    description: siteMetadataDescription,
    relatedLink: relatedLinks,
    featureList: [
      'Structured JSON and YAML output for list and get commands used by coding agents',
      'Read-only safety mode that blocks write commands',
      'Non-interactive automation with no-input and quiet flags',
      'Works with any coding agent or agent harness that can run shell commands',
    ],
    keywords: [
      'coding agent',
      'AI agent CLI',
      'agent harness',
      'MCP-free shell integration',
      'terminal automation',
      'jira automation',
    ],
  };
  const website = {
    '@context': 'https://schema.org',
    '@type': 'WebSite',
    '@id': `${siteUrl}/#website`,
    name: site.name,
    url: siteUrl,
    inLanguage: 'en',
    sameAs: [repositoryUrl],
    description: siteMetadataDescription,
    relatedLink: relatedLinks,
  };

  return (
    <main className="osmo-home flex flex-1 flex-col">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: serializeJsonLd(softwareApplication) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: serializeJsonLd(website) }}
      />
      <HomeHero />

      {/* Getting started */}
      <section className="osmo-section osmo-section--steps">
        <div className="osmo-container">
          <Reveal className="osmo-section__header">
            <h2 className="osmo-section__title">
              Up and running in three moves
            </h2>
          </Reveal>

          <div className="osmo-card-grid osmo-card-grid--steps">
            {steps.map(({ title, body, snippet }, index) => (
              <Reveal
                key={`${title}-${index}`}
                delay={revealDelays[index % revealDelays.length]}
                className="osmo-card osmo-step-card"
              >
                <span className="osmo-eyebrow osmo-card__number">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <h3 className="osmo-card__title">{title}</h3>
                <p className="osmo-card__body">
                  <ContextualBody body={body} link={contextualBodyLinks[title]} />
                </p>
                {snippet ? (
                  <code className="osmo-card__snippet">
                    <span aria-hidden>$</span>
                    {snippet}
                  </code>
                ) : null}
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* Capabilities */}
      <section
        className="osmo-section osmo-section--features"
        data-theme-section="dark"
        aria-labelledby="capabilities-heading"
      >
        <div className="osmo-container">
          <Reveal className="osmo-section__header">
            <h2 id="capabilities-heading" className="osmo-section__title">
              {site.featuresTitle ?? 'Everything, from one binary'}
            </h2>
          </Reveal>

          <div className="osmo-card-grid osmo-card-grid--features">
            {site.features.map(({ title, body }, index) => (
              <Reveal
                key={title}
                delay={revealDelays[index % revealDelays.length]}
                className="osmo-card osmo-feature-card"
              >
                <span className="osmo-eyebrow osmo-card__number">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <h3 className="osmo-card__title">{title}</h3>
                <p className="osmo-card__body">
                  <ContextualBody body={body} link={contextualBodyLinks[title]} />
                </p>
              </Reveal>
            ))}
          </div>

          {site.compatible?.length ? (
            <Reveal className="compatible-marquee">
              <div className="compatible-marquee__track">
                {Array.from(
                  {
                    // Each copy is one full set, so the scroll speed stays fixed
                    // per site; enough copies guarantee the track always overruns
                    // the widest container so no blank gap sweeps through.
                    length: Math.max(
                      4,
                      Math.ceil(24 / site.compatible.length) + 1,
                    ),
                  },
                  (_, copyIndex) => (
                    <span
                      className="compatible-marquee__list"
                      aria-hidden={copyIndex > 0 || undefined}
                      key={copyIndex}
                    >
                      {site.compatible?.map((item) => (
                        <span className="compatible-marquee__item" key={item}>
                          {item}
                          <span aria-hidden>{' · '}</span>
                        </span>
                      ))}
                    </span>
                  ),
                )}
              </div>
            </Reveal>
          ) : null}
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}

function ContextualBody({
  body,
  link,
}: {
  body: string;
  link?: { href: string; text: string };
}) {
  if (!link || !body.includes(link.text)) return body;

  const [before, after] = body.split(link.text, 2);

  return (
    <>
      {before}
      <Link href={link.href}>{link.text}</Link>
      {after}
    </>
  );
}

function serializeJsonLd(value: object): string {
  return JSON.stringify(value).replace(/</g, '\\u003c');
}

import { HomeHero, HomeStats } from '@/components/home-hero';
import { Reveal } from '@/components/reveal';
import { SiteFooter } from '@/components/site-footer';
import { site, type SiteStep } from '@/lib/site';

const eyebrowClass =
  'font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-[var(--site-accent)]';

function splitDescription(
  description: string,
  highlights: string[] | undefined,
) {
  const matchedHighlights = Array.from(
    new Set(
      (highlights ?? []).filter(
        (highlight) => highlight && description.includes(highlight),
      ),
    ),
  ).sort((a, b) => b.length - a.length);

  if (!matchedHighlights.length) {
    return [{ text: description, highlighted: false }];
  }

  const escapedHighlights = matchedHighlights.map((highlight) =>
    highlight.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'),
  );
  const pattern = new RegExp(`(${escapedHighlights.join('|')})`, 'g');
  const highlightSet = new Set(matchedHighlights);

  return description
    .split(pattern)
    .filter(Boolean)
    .map((text) => ({ text, highlighted: highlightSet.has(text) }));
}

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
  const descriptionParts = splitDescription(
    site.description,
    site.descriptionHighlights,
  );

  return (
    <main className="flex flex-1 flex-col">
      <HomeHero descriptionParts={descriptionParts} />
      <HomeStats stats={site.stats} />

      {/* Getting started */}
      <section className="mx-auto w-full max-w-6xl px-4 py-24 sm:py-[7.5rem]">
        <Reveal className="max-w-2xl">
          <p className={eyebrowClass}>Getting started</p>
          <h2 className="mt-4 text-balance font-display text-4xl font-[550] leading-[1.02] tracking-[-0.04em] sm:text-5xl">
            Up and running in three moves
          </h2>
        </Reveal>

        <div className="mt-14 grid gap-5 md:grid-cols-3">
          {steps.map(({ title, body, snippet }, index) => (
            <Reveal
              key={`${title}-${index}`}
              delay={index * 100}
              className="relative overflow-hidden rounded-xl bg-fd-muted/45 p-7 sm:p-8"
            >
              <span
                aria-hidden
                className="pointer-events-none absolute right-4 top-0 font-mono text-[6.5rem] font-semibold leading-none text-[color-mix(in_oklab,var(--site-accent)_15%,transparent)]"
              >
                {String(index + 1).padStart(2, '0')}
              </span>
              <span className={`${eyebrowClass} relative`}>
                {String(index + 1).padStart(2, '0')}
              </span>
              <h3 className="relative mt-12 text-xl font-semibold">{title}</h3>
              <p className="relative mt-3 text-sm leading-relaxed text-muted-foreground">
                {body}
              </p>
              {snippet ? (
                <code className="relative mt-6 block overflow-hidden text-ellipsis whitespace-nowrap rounded-lg bg-fd-muted px-3 py-2.5 font-mono text-xs text-fd-foreground">
                  <span aria-hidden className="mr-2 text-[var(--site-accent)]">
                    $
                  </span>
                  {snippet}
                </code>
              ) : null}
            </Reveal>
          ))}
        </div>
      </section>

      {/* Capabilities */}
      <section
        className="capabilities-band"
        data-theme-section="dark"
        aria-labelledby="capabilities-heading"
      >
        <div className="capabilities-band__inner">
          <Reveal className="capabilities-band__header">
            <p className="capabilities-band__eyebrow">Capabilities</p>
            <h2 id="capabilities-heading" className="capabilities-band__title">
              {site.featuresTitle ?? 'Everything, from one binary'}
            </h2>
          </Reveal>

          <div className="capabilities-grid">
            {site.features.map(({ title, body }, index) => (
              <Reveal
                key={title}
                delay={index * 60}
                className="capabilities-grid__item"
              >
                <span aria-hidden className="capabilities-grid__number">
                  {String(index + 1).padStart(2, '0')}
                </span>
                <h3 className="capabilities-grid__title">{title}</h3>
                <p className="capabilities-grid__body">{body}</p>
              </Reveal>
            ))}
          </div>

          {site.compatible?.length ? (
            <Reveal as="p" className="capabilities-band__compatible">
              {site.compatible.map((item, index) => (
                <span key={item}>
                  {index > 0 ? (
                    <span
                      aria-hidden
                      className="capabilities-band__separator"
                    >
                      {' · '}
                    </span>
                  ) : null}
                  {item}
                </span>
              ))}
            </Reveal>
          ) : null}
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}

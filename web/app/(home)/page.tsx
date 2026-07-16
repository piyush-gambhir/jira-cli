import type { CSSProperties } from 'react';
import { ArrowRight } from 'lucide-react';
import { ActionButton } from '@/components/ui/action-button';
import { InstallCommand } from '@/components/install-command';
import { HomeHero } from '@/components/home-hero';
import { Reveal } from '@/components/reveal';
import { SiteFooter } from '@/components/site-footer';
import { site, type SiteStep } from '@/lib/site';

const eyebrowClass =
  'font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-[var(--site-accent)]';

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

  return (
    <main className="flex flex-1 flex-col">
      <HomeHero />

      {/* Stack marquee */}
      {site.compatible && site.compatible.length > 0 ? (
        <section aria-labelledby="compatible-heading" className="py-8 sm:py-10">
          <p
            id="compatible-heading"
            className={`${eyebrowClass} px-4 text-center`}
          >
            Speaks the language of your stack
          </p>
          <div className="marquee mt-7 border-y border-border/70 py-5">
            <div
              className="marquee-track"
              style={{ '--marquee-duration': '28s' } as CSSProperties}
            >
              {[false, true].map((duplicate) => (
                <div
                  key={duplicate ? 'duplicate' : 'primary'}
                  aria-hidden={duplicate || undefined}
                  className="flex shrink-0 items-center"
                >
                  {site.compatible?.map((item) => (
                    <span key={item} className="flex shrink-0 items-center">
                      <span className="px-7 font-mono text-xs font-medium uppercase tracking-[0.12em] text-fd-foreground/60 sm:px-10">
                        {item}
                      </span>
                      <span aria-hidden className="text-[var(--site-accent)]">
                        ·
                      </span>
                    </span>
                  ))}
                </div>
              ))}
            </div>
          </div>
        </section>
      ) : null}

      {/* Getting started */}
      <section className="mx-auto w-full max-w-6xl px-4 py-24 sm:py-32">
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
              className="relative overflow-hidden rounded-lg border border-border/70 bg-card/40 p-7 sm:p-8"
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

      {/* Features */}
      <section className="mx-auto w-full max-w-6xl px-4 py-24 sm:py-32">
        <Reveal className="mx-auto max-w-2xl text-center">
          <p className={eyebrowClass}>Capabilities</p>
          <h2 className="mt-4 text-balance font-display text-4xl font-[550] leading-[1.02] tracking-[-0.04em] sm:text-5xl">
            {site.featuresTitle ?? 'Everything, from one binary'}
          </h2>
          <p className="mt-5 text-lg text-muted-foreground">
            {site.featuresSubtitle ??
              'Built for humans at the keyboard and coding agents alike.'}
          </p>
        </Reveal>

        <div className="mt-14 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
          {site.features.map(({ icon: Icon, title, body }, index) => (
            <Reveal
              key={title}
              delay={index * 60}
              className="micro-hover-card group rounded-lg border border-border/70 bg-card/40 p-7 sm:p-8"
            >
              <div className="micro-hover-icon mb-8 flex size-11 items-center justify-center rounded-lg bg-fd-muted text-fd-foreground">
                <Icon className="size-5" />
              </div>
              <h3 className="text-lg font-semibold">{title}</h3>
              <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
                {body}
              </p>
            </Reveal>
          ))}
        </div>
      </section>

      {/* CTA band */}
      <section className="mx-auto w-full max-w-6xl px-4 pb-28 pt-16 sm:pb-32 sm:pt-20">
        <Reveal className="relative overflow-hidden rounded-[2rem] border border-border/60 bg-fd-muted/50 px-6 py-20 text-center sm:py-24">
          <div
            aria-hidden
            className="pointer-events-none absolute inset-x-0 top-0 -z-10 h-64"
            style={{
              background:
                'radial-gradient(60% 100% at 50% 0%, color-mix(in oklab, var(--site-accent) 20%, transparent), transparent)',
            }}
          />
          <p className={eyebrowClass}>Start building</p>
          <h2 className="mx-auto mt-4 max-w-xl text-balance font-display text-4xl font-[550] leading-[1.02] tracking-[-0.04em] sm:text-5xl">
            Ready in one command
          </h2>
          <p className="mx-auto mt-5 max-w-md text-lg text-muted-foreground">
            {site.ctaBody ??
              'Install the binary, authenticate, and start querying. No runtime, no dependencies.'}
          </p>
          <div className="mt-9 flex flex-col items-center gap-5">
            <InstallCommand command={site.installCommand} />
            <ActionButton href="/docs" aria-label="Read the docs">
              Read the docs
              <ArrowRight className="size-4" />
            </ActionButton>
          </div>
        </Reveal>
      </section>

      <SiteFooter />
    </main>
  );
}

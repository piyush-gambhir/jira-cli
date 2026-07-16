import type { CSSProperties } from 'react';
import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { InstallCommand } from '@/components/install-command';
import { HeroTerminal } from '@/components/hero-terminal';
import { Reveal } from '@/components/reveal';
import { SiteFooter } from '@/components/site-footer';
import { site, type SiteStep } from '@/lib/site';

const eyebrowClass =
  'font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-[var(--accent)]';

export default function HomePage() {
  const repoUrl = `https://github.com/${site.repo}`;
  const taglineWords = site.tagline.split(/\s+/);
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
      {/* Hero */}
      <section className="relative isolate overflow-hidden">
        <div aria-hidden className="pointer-events-none absolute inset-0 -z-10">
          <div
            className="absolute left-[18%] top-[48%] size-[34rem] rounded-full opacity-60 blur-[100px]"
            style={{
              background:
                'radial-gradient(circle, color-mix(in oklab, var(--accent) 35%, transparent), transparent 70%)',
            }}
          />
          <div
            className="absolute right-[7%] top-[58%] size-[24rem] rounded-full opacity-50 blur-[100px]"
            style={{
              background:
                'radial-gradient(circle, color-mix(in oklab, var(--foreground) 8%, transparent), transparent 72%)',
            }}
          />
        </div>

        <div className="mx-auto flex max-w-5xl flex-col items-center px-4 pb-24 pt-[clamp(10rem,24svh,15rem)] text-center sm:pb-32">
          <p className={`${eyebrowClass} fade-up`}>
            <span aria-hidden className="mr-2">
              ▸
            </span>
            {site.badge}
          </p>
          <h1 className="mt-6 max-w-4xl text-balance font-display text-[clamp(2.75rem,7vw,5.5rem)] font-[575] leading-[0.95] tracking-[-0.04em]">
            {taglineWords.map((word, index) => (
              <span key={`${word}-${index}`}>
                {index > 0 ? ' ' : null}
                <span className="word-mask">
                  <span
                    className="word-mask-reveal"
                    style={{ '--word-index': index } as CSSProperties}
                  >
                    {word}
                  </span>
                </span>
              </span>
            ))}
          </h1>
          <p className="fade-up fade-up-delay-1 mt-7 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground">
            {site.description}
          </p>

          <div className="fade-up fade-up-delay-2 mt-9 flex flex-wrap items-center justify-center gap-3">
            <Button size="lg" render={<Link href="/docs" />} rollLabel>
              Get started
              <ArrowRight className="size-4" />
            </Button>
            <Button
              size="lg"
              variant="ghost"
              render={<Link href={repoUrl} />}
            >
              View on GitHub
            </Button>
          </div>
          <InstallCommand
            command={site.installCommand}
            className="fade-up fade-up-delay-3 mt-5"
          />

          <HeroTerminal
            title={site.exampleTitle}
            command={site.example}
            className="fade-up fade-up-delay-3 mt-16 w-full max-w-3xl text-left"
          />
        </div>
      </section>

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
                      <span aria-hidden className="text-[var(--accent)]">
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
                className="pointer-events-none absolute right-4 top-0 font-mono text-[6.5rem] font-semibold leading-none text-[color-mix(in_oklab,var(--accent)_15%,transparent)]"
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
                  <span aria-hidden className="mr-2 text-[var(--accent)]">
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
                'radial-gradient(60% 100% at 50% 0%, color-mix(in oklab, var(--accent) 20%, transparent), transparent)',
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
            <Button render={<Link href="/docs" />} rollLabel>
              Read the docs
              <ArrowRight className="size-4" />
            </Button>
          </div>
        </Reveal>
      </section>

      <SiteFooter />
    </main>
  );
}

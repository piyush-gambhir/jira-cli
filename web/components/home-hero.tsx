'use client';

import { useRef } from 'react';
import { ArrowRight } from 'lucide-react';
import { HeroTerminal } from '@/components/hero-terminal';
import { InstallCommand } from '@/components/install-command';
import { ActionButton } from '@/components/ui/action-button';
import { gsap } from '@/lib/motion/gsap';
import { useGsap } from '@/lib/motion/useGsap';
import { site } from '@/lib/site';

const eyebrowClass =
  'font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-[var(--accent)]';

export function HomeHero() {
  const rootRef = useRef<HTMLElement>(null);
  const taglineWords = site.tagline.split(/\s+/);
  const repoUrl = `https://github.com/${site.repo}`;

  useGsap(
    () => {
      const root = rootRef.current;
      if (!root) return;

      const words = gsap.utils.toArray<HTMLElement>('[data-hero-word]', root);
      const eyebrow = root.querySelector<HTMLElement>('[data-hero-eyebrow]');
      const description = root.querySelector<HTMLElement>(
        '[data-hero-description]',
      );
      const actions = root.querySelector<HTMLElement>('[data-hero-actions]');
      const install = root.querySelector<HTMLElement>('[data-hero-install]');
      const terminal = root.querySelector<HTMLElement>('[data-hero-terminal]');

      gsap.set(words, {
        yPercent: 100,
        rotation: 10,
        transformOrigin: 'bottom left',
      });

      gsap
        .timeline({ defaults: { ease: 'expo.out' } })
        .to(
          words,
          {
            yPercent: 0,
            rotation: 0,
            duration: 1.2,
            stagger: 0.05,
          },
          0,
        )
        .from(eyebrow, { y: '1em', autoAlpha: 0, duration: 0.8 }, 0.1)
        .from(
          description,
          { y: '2em', autoAlpha: 0, duration: 1.2 },
          0.2,
        )
        .from(actions, { y: '1.5em', autoAlpha: 0, duration: 1 }, 0.3)
        .from(install, { y: '1.5em', autoAlpha: 0, duration: 1 }, 0.38)
        .from(
          terminal,
          { y: '2.5em', autoAlpha: 0, duration: 1.2 },
          0.45,
        );
    },
    [],
    rootRef,
  );

  return (
    <section ref={rootRef} className="relative isolate overflow-hidden">
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
        <p className={eyebrowClass} data-hero-eyebrow>
          <span aria-hidden className="mr-2">
            ▸
          </span>
          {site.badge}
        </p>
        <h1
          className="mt-6 max-w-4xl text-balance font-display text-[clamp(2.75rem,7vw,5.5rem)] font-[575] leading-[0.95] tracking-[-0.04em]"
          aria-label={site.tagline}
        >
          <span aria-hidden="true">
            {taglineWords.map((word, index) => (
              <span className="hero-word-mask" key={`${word}-${index}`}>
                <span data-hero-word>{word}</span>
              </span>
            ))}
          </span>
        </h1>
        <p
          className="mt-7 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground"
          data-hero-description
        >
          {site.description}
        </p>

        <div
          className="mt-9 flex flex-wrap items-center justify-center gap-3"
          data-hero-actions
        >
          <ActionButton href="/docs" aria-label="Get started">
            Get started
            <ArrowRight className="size-4" />
          </ActionButton>
          <ActionButton
            href={repoUrl}
            theme="neutral"
            aria-label="View on GitHub"
          >
            View on GitHub
          </ActionButton>
        </div>
        <div className="mt-5 w-full max-w-xl" data-hero-install>
          <InstallCommand command={site.installCommand} />
        </div>

        <div className="mt-16 w-full max-w-3xl text-left" data-hero-terminal>
          <HeroTerminal title={site.exampleTitle} command={site.example} />
        </div>
      </div>
    </section>
  );
}

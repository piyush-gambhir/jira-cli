'use client';

import { useRef } from 'react';
import { ArrowRight } from 'lucide-react';
import { HeroTerminal } from '@/components/hero-terminal';
import { InstallCommand } from '@/components/install-command';
import { ActionButton } from '@/components/ui/action-button';
import { gsap } from '@/lib/motion/gsap';
import { useGsap } from '@/lib/motion/useGsap';
import { site, type SiteStat } from '@/lib/site';

const eyebrowClass =
  'font-mono text-[11px] font-medium uppercase tracking-[0.14em] text-[var(--site-accent)]';

interface DescriptionPart {
  text: string;
  highlighted: boolean;
}

interface HomeHeroProps {
  descriptionParts: DescriptionPart[];
}

export function HomeHero({ descriptionParts }: HomeHeroProps) {
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
      const ring = root.querySelector<HTMLElement>('[data-hero-ring]');
      const ringSpinner = root.querySelector<HTMLElement>(
        '[data-hero-ring-spinner]',
      );

      gsap.set(words, {
        yPercent: 100,
        rotation: 10,
        transformOrigin: 'bottom left',
      });

      const timeline = gsap
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

      if (ring) {
        timeline.from(
          ring,
          {
            rotation: '-=45',
            autoAlpha: 0,
            duration: 2,
            ease: 'brand-default',
          },
          0,
        );
      }

      if (!ringSpinner || !('IntersectionObserver' in window)) return;

      const ringObserver = new IntersectionObserver(([entry]) => {
        ringSpinner.classList.toggle('is-paused', !entry?.isIntersecting);
      });

      ringObserver.observe(root);

      return () => ringObserver.disconnect();
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
              'radial-gradient(circle, color-mix(in oklab, var(--site-accent) 35%, transparent), transparent 70%)',
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
        <div className="hero-heading-wrap mt-6">
          <div aria-hidden="true" className="hero-ray-glow" />
          <div aria-hidden="true" className="hero-ray-ring-stage">
            <div className="hero-ray-ring" data-hero-ring>
              <div
                className="hero-ray-ring__spinner"
                data-hero-ring-spinner
              />
            </div>
          </div>
          <h1
            className="max-w-4xl text-balance font-display text-[clamp(2.75rem,7vw,5.5rem)] font-[575] leading-[0.95] tracking-[-0.04em]"
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
        </div>
        <p
          className="mt-7 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground"
          data-hero-description
        >
          {descriptionParts.map(({ text, highlighted }, index) =>
            highlighted ? (
              <span
                className="hero-description-highlight"
                key={`${text}-${index}`}
              >
                {text}
              </span>
            ) : (
              text
            ),
          )}
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

interface HomeStatsProps {
  stats?: SiteStat[];
}

export function HomeStats({ stats }: HomeStatsProps) {
  const rootRef = useRef<HTMLElement>(null);

  useGsap(
    () => {
      const root = rootRef.current;
      if (!root) return;

      const tracks = gsap.utils.toArray<HTMLElement>(
        '[data-stat-track]',
        root,
      );
      if (!tracks.length) return;

      const duration =
        Number.parseFloat(
          getComputedStyle(root).getPropertyValue('--duration-m'),
        ) || 0.6;
      gsap.set(tracks, { yPercent: 0 });

      const roll = () => {
        gsap.to(tracks, {
          yPercent: -50,
          duration,
          stagger: 0.12,
          ease: 'brand-default',
        });
      };

      if (!('IntersectionObserver' in window)) {
        roll();
        return;
      }

      const observer = new IntersectionObserver(
        ([entry]) => {
          if (!entry?.isIntersecting) return;
          observer.disconnect();
          roll();
        },
        { threshold: 0.25 },
      );

      observer.observe(root);

      return () => observer.disconnect();
    },
    [],
    rootRef,
  );

  if (!stats?.length) return null;

  return (
    <section ref={rootRef} className="home-stats" aria-label="Project statistics">
      <div className="home-stats__grid">
        {stats.map(({ value, label }) => (
          <div className="home-stat" key={`${value}-${label}`}>
            <span className="sr-only">{value}</span>
            <span aria-hidden="true" className="home-stat__window">
              <span className="home-stat__track" data-stat-track>
                <span>0</span>
                <span>{value}</span>
              </span>
            </span>
            <span className="home-stat__label">{label}</span>
          </div>
        ))}
      </div>
    </section>
  );
}

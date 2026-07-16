'use client';

import Link from 'next/link';
import { Menu, Moon, Search, Sun, X } from 'lucide-react';
import { useTheme } from 'next-themes';
import { useSearchContext } from 'fumadocs-ui/contexts/search';
import { type CSSProperties, useEffect, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { site } from '@/lib/site';

const mobileItemStyle = (index: number) =>
  ({ '--mobile-item-index': index } as CSSProperties);

export function FloatingHeader() {
  const { setTheme, resolvedTheme } = useTheme();
  const { setOpenSearch } = useSearchContext();
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const headerRef = useRef<HTMLElement>(null);
  const repoUrl = `https://github.com/${site.repo}`;

  useEffect(() => {
    const updateScrolled = () => setIsScrolled(window.scrollY > 8);

    updateScrolled();
    window.addEventListener('scroll', updateScrolled, { passive: true });

    return () => window.removeEventListener('scroll', updateScrolled);
  }, []);

  useEffect(() => {
    if (!isMenuOpen) return;

    const previousOverflow = document.body.style.overflow;
    const closeOnOutsideClick = (event: PointerEvent) => {
      if (!headerRef.current?.contains(event.target as Node)) {
        setIsMenuOpen(false);
      }
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setIsMenuOpen(false);
    };

    document.body.style.overflow = 'hidden';
    document.addEventListener('pointerdown', closeOnOutsideClick);
    document.addEventListener('keydown', closeOnEscape);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener('pointerdown', closeOnOutsideClick);
      document.removeEventListener('keydown', closeOnEscape);
    };
  }, [isMenuOpen]);

  useEffect(() => {
    const desktop = window.matchMedia('(min-width: 640px)');
    const closeOnDesktop = (event: MediaQueryListEvent) => {
      if (event.matches) setIsMenuOpen(false);
    };

    desktop.addEventListener('change', closeOnDesktop);
    return () => desktop.removeEventListener('change', closeOnDesktop);
  }, []);

  const iconButtonClass =
    'flex size-9 items-center justify-center rounded-full text-muted-foreground hover:bg-fd-muted hover:text-fd-foreground';

  const openSearch = () => {
    setIsMenuOpen(false);
    setOpenSearch(true);
  };

  const toggleTheme = () => {
    setTheme(resolvedTheme === 'dark' ? 'light' : 'dark');
  };

  return (
    <header
      ref={headerRef}
      className="fixed inset-x-0 top-3 z-50 px-3 sm:top-4 sm:px-4"
    >
      <div className="mx-auto max-w-4xl">
        <nav
          data-scrolled={isScrolled}
          aria-label="Primary navigation"
          className="floating-header-surface flex h-14 items-center justify-between gap-3 rounded-full border bg-background/80 py-2 pr-2 pl-5 shadow-lg shadow-foreground/5 backdrop-blur-xl supports-[backdrop-filter]:bg-background/65"
        >
          <Link
            href="/"
            aria-label={`${site.name} home`}
            className="flex min-w-0 items-center gap-2 font-mono text-sm font-medium tracking-[-0.03em]"
            onClick={() => setIsMenuOpen(false)}
          >
            <span aria-hidden className="shrink-0 text-[var(--accent)]">
              &gt;_
            </span>
            <span className="truncate">{site.binary}</span>
          </Link>

          <div className="hidden items-center gap-1 sm:flex">
            <Link
              href="/docs"
              className="nav-underline-slide mx-2 py-1.5 text-sm text-muted-foreground hover:text-fd-foreground"
            >
              Docs
            </Link>

            <button
              type="button"
              onClick={openSearch}
              aria-label="Search"
              className={iconButtonClass}
            >
              <Search className="size-4" />
            </button>

            <button
              type="button"
              onClick={toggleTheme}
              aria-label="Toggle theme"
              className={iconButtonClass}
            >
              <Sun className="hidden size-4 dark:block" />
              <Moon className="size-4 dark:hidden" />
            </button>

            <a
              href={repoUrl}
              target="_blank"
              rel="noreferrer"
              aria-label="GitHub"
              className={iconButtonClass}
            >
              <svg viewBox="0 0 24 24" fill="currentColor" className="size-4">
                <path d="M12 .5A11.5 11.5 0 0 0 .5 12a11.5 11.5 0 0 0 7.86 10.92c.58.1.79-.25.79-.56v-2c-3.2.7-3.88-1.37-3.88-1.37-.53-1.34-1.29-1.7-1.29-1.7-1.05-.72.08-.71.08-.71 1.16.08 1.77 1.2 1.77 1.2 1.03 1.77 2.7 1.26 3.36.96.1-.75.4-1.26.73-1.55-2.55-.29-5.24-1.28-5.24-5.69 0-1.26.45-2.29 1.19-3.1-.12-.29-.52-1.46.11-3.05 0 0 .97-.31 3.18 1.18a11 11 0 0 1 5.79 0c2.2-1.49 3.17-1.18 3.17-1.18.63 1.59.23 2.76.11 3.05.74.81 1.19 1.84 1.19 3.1 0 4.42-2.69 5.4-5.25 5.68.41.36.77 1.06.77 2.14v3.17c0 .31.21.67.8.56A11.5 11.5 0 0 0 23.5 12 11.5 11.5 0 0 0 12 .5Z" />
              </svg>
            </a>

            <Button
              size="sm"
              className="ml-1 h-9 rounded-full px-4"
              render={<Link href="/docs" />}
              rollLabel
            >
              Get started
            </Button>
          </div>

          <button
            type="button"
            aria-expanded={isMenuOpen}
            aria-controls="mobile-navigation"
            aria-label={isMenuOpen ? 'Close navigation menu' : 'Open navigation menu'}
            onClick={() => setIsMenuOpen((open) => !open)}
            className={`${iconButtonClass} sm:hidden`}
          >
            {isMenuOpen ? <X className="size-4" /> : <Menu className="size-4" />}
          </button>
        </nav>

        <div
          id="mobile-navigation"
          data-open={isMenuOpen}
          className="grid-rows-accordion sm:hidden"
        >
          <div className="grid-rows-accordion-content">
            <div
              aria-hidden={!isMenuOpen}
              inert={isMenuOpen ? undefined : true}
              className="mt-2 rounded-[1.75rem] border border-border bg-background/90 p-3 shadow-lg shadow-foreground/5 backdrop-blur-xl supports-[backdrop-filter]:bg-background/75"
            >
              <div className="grid gap-1">
                <Link
                  href="/docs"
                  onClick={() => setIsMenuOpen(false)}
                  className="floating-header-mobile-item rounded-2xl px-4 py-3 text-sm font-medium hover:bg-fd-muted"
                  style={mobileItemStyle(0)}
                >
                  Docs
                </Link>
                <a
                  href={repoUrl}
                  target="_blank"
                  rel="noreferrer"
                  onClick={() => setIsMenuOpen(false)}
                  className="floating-header-mobile-item rounded-2xl px-4 py-3 text-sm font-medium hover:bg-fd-muted"
                  style={mobileItemStyle(1)}
                >
                  GitHub
                </a>
                <Button
                  className="floating-header-mobile-item mt-1 h-10 w-full rounded-full"
                  render={<Link href="/docs" />}
                  onClick={() => setIsMenuOpen(false)}
                  style={mobileItemStyle(2)}
                  rollLabel
                >
                  Get started
                </Button>
              </div>

              <div className="mt-3 grid grid-cols-2 gap-2 border-t border-border pt-3">
                <button
                  type="button"
                  onClick={openSearch}
                  className="floating-header-mobile-item flex items-center gap-2 rounded-2xl px-4 py-3 text-sm text-muted-foreground hover:bg-fd-muted hover:text-fd-foreground"
                  style={mobileItemStyle(3)}
                >
                  <Search className="size-4" />
                  Search
                </button>
                <button
                  type="button"
                  onClick={toggleTheme}
                  className="floating-header-mobile-item flex items-center gap-2 rounded-2xl px-4 py-3 text-sm text-muted-foreground hover:bg-fd-muted hover:text-fd-foreground"
                  style={mobileItemStyle(4)}
                >
                  <Sun className="hidden size-4 dark:block" />
                  <Moon className="size-4 dark:hidden" />
                  Theme
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

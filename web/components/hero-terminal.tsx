'use client';

import { useEffect, useMemo, useState } from 'react';
import { cn } from '@/lib/utils';

interface PlaybackPosition {
  lineIndex: number;
  charIndex: number;
}

// Lightweight, dependency-free shell highlighter for the hero example.
function Line({ line, isTyping = false }: { line: string; isTyping?: boolean }) {
  if (line.trim() === '') return isTyping ? null : <span>{'\n'}</span>;

  // Comment line
  if (line.trimStart().startsWith('#')) {
    return <span className="text-fd-muted-foreground/70">{line}</span>;
  }

  const tokens = line.split(/(\s+)/);
  let seenBinary = false;

  return (
    <span>
      {tokens.map((tok, i) => {
        if (/^\s+$/.test(tok)) return <span key={i}>{tok}</span>;

        // first non-space token = the binary
        if (!seenBinary) {
          seenBinary = true;
          return (
            <span key={i} className="text-violet-500 dark:text-violet-400">
              {tok}
            </span>
          );
        }
        if (tok.startsWith('-')) {
          return (
            <span key={i} className="text-amber-600 dark:text-amber-400">
              {tok}
            </span>
          );
        }
        if (
          /^["'].*["']$/.test(tok) ||
          tok.startsWith('"') ||
          tok.startsWith("'")
        ) {
          return (
            <span key={i} className="text-emerald-600 dark:text-emerald-400">
              {tok}
            </span>
          );
        }
        return (
          <span key={i} className="text-fd-foreground/90">
            {tok}
          </span>
        );
      })}
    </span>
  );
}

export function HeroTerminal({
  title,
  command,
  className,
}: {
  title: string;
  command: string;
  className?: string;
}) {
  const lines = useMemo(() => command.split('\n'), [command]);
  // `null` is deliberately the initial state so SSR and no-JS clients receive
  // the complete, highlighted example before the hydrated replay begins.
  const [playback, setPlayback] = useState<PlaybackPosition | null>(null);

  useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

    setPlayback({ lineIndex: 0, charIndex: 0 });
  }, []);

  useEffect(() => {
    if (!playback) return;

    const currentLine = lines[playback.lineIndex] ?? '';
    const lineIsComplete = playback.charIndex >= currentLine.length;
    const isLastLine = playback.lineIndex >= lines.length - 1;
    const delay = lineIsComplete ? 500 : 30;
    const timer = window.setTimeout(() => {
      if (!lineIsComplete) {
        setPlayback((position) =>
          position
            ? { ...position, charIndex: position.charIndex + 1 }
            : null,
        );
        return;
      }

      if (isLastLine) {
        setPlayback(null);
        return;
      }

      setPlayback({
        lineIndex: playback.lineIndex + 1,
        charIndex: 0,
      });
    }, delay);

    return () => window.clearTimeout(timer);
  }, [lines, playback]);

  const visibleLines = playback
    ? lines.slice(0, playback.lineIndex + 1)
    : lines;

  return (
    <div
      className={cn(
        'overflow-hidden rounded-2xl bg-fd-card shadow-[0_24px_80px_-24px_rgba(0,0,0,0.25)]',
        className,
      )}
    >
      {/* titlebar */}
      <div className="flex items-center gap-2 bg-fd-muted/50 px-4 py-3">
        <span className="size-3 rounded-full bg-red-400/90" />
        <span className="size-3 rounded-full bg-amber-400/90" />
        <span className="size-3 rounded-full bg-emerald-400/90" />
        <span className="ml-3 text-xs font-medium text-fd-muted-foreground">
          {title}
        </span>
      </div>
      {/* body */}
      <pre className="overflow-x-auto px-5 py-4 text-left font-mono text-[13px] leading-relaxed sm:text-sm">
        <code>
          {visibleLines.map((line, i) => {
            const isTypingLine = playback?.lineIndex === i;
            const visibleLine = isTypingLine
              ? line.slice(0, playback?.charIndex ?? 0)
              : line;

            return (
              <span key={i} className="block">
                {!line.trimStart().startsWith('#') && line.trim() !== '' ? (
                  <span className="mr-2 select-none text-[var(--accent)]">
                    $
                  </span>
                ) : null}
                <Line line={visibleLine} isTyping={isTypingLine} />
                {isTypingLine ? (
                  <span
                    aria-hidden
                    className="terminal-caret text-[var(--accent)]"
                  >
                    ▍
                  </span>
                ) : null}
              </span>
            );
          })}
        </code>
      </pre>
    </div>
  );
}

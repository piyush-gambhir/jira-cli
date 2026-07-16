'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { gsap } from '@/lib/motion/gsap';
import { useGsap } from '@/lib/motion/useGsap';
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
    return <span className="terminal-token--comment">{line}</span>;
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
            <span key={i} className="terminal-token--command">
              {tok}
            </span>
          );
        }
        if (tok.startsWith('-')) {
          return (
            <span key={i} className="terminal-token--flag">
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
            <span key={i} className="terminal-token--string">
              {tok}
            </span>
          );
        }
        return (
          <span key={i} className="terminal-token--text">
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
  const rootRef = useRef<HTMLDivElement>(null);
  const terminalRef = useRef<HTMLDivElement>(null);
  const lines = useMemo(() => command.split('\n'), [command]);
  // `null` is deliberately the initial state so SSR and no-JS clients receive
  // the complete, highlighted example before the hydrated replay begins.
  const [playback, setPlayback] = useState<PlaybackPosition | null>(null);

  useGsap(
    () => {
      const root = rootRef.current;
      const terminal = terminalRef.current;
      if (!root || !terminal || window.matchMedia('(hover: none)').matches) {
        return;
      }

      const rotateX = gsap.quickTo(terminal, 'rotationX', {
        duration: 0.6,
        ease: 'brand-default',
      });
      const rotateY = gsap.quickTo(terminal, 'rotationY', {
        duration: 0.6,
        ease: 'brand-default',
      });

      const onPointerMove = (event: PointerEvent) => {
        const bounds = root.getBoundingClientRect();
        const x = ((event.clientX - bounds.left) / bounds.width - 0.5) * 2;
        const y = ((event.clientY - bounds.top) / bounds.height - 0.5) * 2;

        rotateX(gsap.utils.clamp(-2.5, 2.5, y * -2.5));
        rotateY(gsap.utils.clamp(-2.5, 2.5, x * 2.5));
      };
      const resetTilt = () => {
        rotateX(0);
        rotateY(0);
      };

      root.addEventListener('pointermove', onPointerMove);
      root.addEventListener('pointerleave', resetTilt);

      return () => {
        root.removeEventListener('pointermove', onPointerMove);
        root.removeEventListener('pointerleave', resetTilt);
      };
    },
    [],
    rootRef,
  );

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
    <div ref={rootRef} className={cn('hero-terminal-perspective', className)}>
      <div ref={terminalRef} className="hero-terminal">
        <div aria-hidden className="terminal-scanlines" />
        {/* titlebar */}
        <div className="hero-terminal__titlebar">
          <span className="size-3 rounded-full bg-red-400/90" />
          <span className="size-3 rounded-full bg-amber-400/90" />
          <span className="size-3 rounded-full bg-emerald-400/90" />
          <span className="hero-terminal__title">{title}</span>
        </div>
        {/* body */}
        <pre className="hero-terminal__body">
          <code>
            {visibleLines.map((line, i) => {
              const isTypingLine = playback?.lineIndex === i;
              const visibleLine = isTypingLine
                ? line.slice(0, playback?.charIndex ?? 0)
                : line;

              return (
                <span key={i} className="block">
                  {!line.trimStart().startsWith('#') && line.trim() !== '' ? (
                    <span className="terminal-prompt mr-2 select-none">$</span>
                  ) : null}
                  <Line line={visibleLine} isTyping={isTypingLine} />
                  {isTypingLine ? (
                    <span aria-hidden className="terminal-caret">
                      ▍
                    </span>
                  ) : null}
                </span>
              );
            })}
          </code>
        </pre>
      </div>
    </div>
  );
}

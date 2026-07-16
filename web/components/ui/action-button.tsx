import Link from 'next/link';
import type {
  AnchorHTMLAttributes,
  ButtonHTMLAttributes,
  ReactNode,
} from 'react';
import { cn } from '@/lib/utils';

export type ActionButtonTheme = 'accent' | 'dark' | 'neutral';
export type ActionButtonRadius = 'square' | 'pill';

type SharedActionButtonProps = {
  children: ReactNode;
  className?: string;
  theme?: ActionButtonTheme;
  radius?: ActionButtonRadius;
  icon?: ReactNode;
};

type LinkActionButtonProps = SharedActionButtonProps &
  Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'children' | 'className' | 'href'> & {
    href: string;
    type?: never;
  };

type NativeActionButtonProps = SharedActionButtonProps &
  Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children' | 'className'> & {
    href?: never;
  };

export type ActionButtonProps = LinkActionButtonProps | NativeActionButtonProps;

function ActionButtonInner({
  children,
  icon,
}: Pick<SharedActionButtonProps, 'children' | 'icon'>) {
  return (
    <>
      <span className="button-bg" aria-hidden="true" />
      <span className="button-label-wrap" aria-hidden="true">
        <span className="button-label is--primary">{children}</span>
        <span className="button-label is--secondary">{children}</span>
      </span>
      {icon ? (
        <span className="button-icon" aria-hidden="true">
          {icon}
        </span>
      ) : null}
    </>
  );
}

export function ActionButton(props: ActionButtonProps) {
  const {
    children,
    className,
    theme = 'accent',
    radius = 'square',
    icon,
    ...elementProps
  } = props;
  const classes = cn(
    'action-button',
    `is--${theme}`,
    `is--${radius}`,
    className,
  );
  const accessibleLabel = typeof children === 'string' ? children : undefined;

  if ('href' in props && props.href) {
    const { href, target, rel, onClick, ...linkProps } =
      elementProps as Omit<LinkActionButtonProps, keyof SharedActionButtonProps>;
    return (
      <Link
        {...linkProps}
        href={href}
        target={target}
        rel={rel}
        onClick={onClick}
        className={classes}
        aria-label={linkProps['aria-label'] ?? accessibleLabel}
      >
        <ActionButtonInner icon={icon}>{children}</ActionButtonInner>
      </Link>
    );
  }

  const { type = 'button', ...buttonProps } =
    elementProps as Omit<NativeActionButtonProps, keyof SharedActionButtonProps>;
  return (
    <button
      {...buttonProps}
      type={type}
      className={classes}
      aria-label={buttonProps['aria-label'] ?? accessibleLabel}
    >
      <ActionButtonInner icon={icon}>{children}</ActionButtonInner>
    </button>
  );
}

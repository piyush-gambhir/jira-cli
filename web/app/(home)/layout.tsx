import { FloatingHeader } from '@/components/floating-header';
import { UnderNavMarquee } from '@/components/under-nav-marquee';

export default function Layout({ children }: LayoutProps<'/'>) {
  return (
    <>
      <FloatingHeader />
      <UnderNavMarquee />
      {children}
    </>
  );
}

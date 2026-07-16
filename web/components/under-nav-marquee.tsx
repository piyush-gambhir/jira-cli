import Link from 'next/link';

const items = Array.from({ length: 5 }, (_, index) => (
  <span className="nav-marquee__item" key={index}>
    <span className="nav-marquee__mark" aria-hidden="true">&gt;_</span>
    Read the docs — get productive in five minutes
  </span>
));

export function UnderNavMarquee() {
  return (
    <div className="under-nav-bar">
      <div className="under-nav-bar__inner">
        <Link
          href="/docs"
          className="nav-marquee"
          aria-label="Read the docs — get productive in five minutes"
        >
          <span className="marquee-css">
            <span className="marquee-css__list">{items}</span>
            <span className="marquee-css__list" aria-hidden="true">{items}</span>
          </span>
        </Link>
      </div>
    </div>
  );
}

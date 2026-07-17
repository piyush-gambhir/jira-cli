import type { Metadata } from 'next';
import Link from 'next/link';
import { LegalPage } from '@/components/legal-page';
import { createPageMetadata } from '@/lib/metadata';

export const metadata: Metadata = createPageMetadata({
  title: 'Terms of Service',
  summary: 'Terms of Service covering the license, warranty, liability, and acceptable use of Jira CLI.',
  path: '/terms',
});

export default function TermsPage() {
  return (
    <LegalPage
      title="Terms of Service"
      intro={
        <>
          By installing or using Jira CLI (the <code>jira</code> command-line
          tool), you agree to these terms.
        </>
      }
    >
      <section>
        <h2>1. License</h2>
        <p>
          Jira CLI is free, open-source software distributed under the{' '}
          <strong>MIT License</strong>. The full license text is in the{' '}
          <a
            href="https://github.com/piyush-gambhir/jira-cli/blob/main/LICENSE"
            target="_blank"
            rel="noreferrer"
          >
            repository
          </a>
          . You may use, copy, modify, and distribute it under those terms.
        </p>
      </section>

      <section>
        <h2>2. No warranty</h2>
        <p>
          The software is provided{' '}
          <strong>&quot;as is&quot;, without warranty of any kind</strong>,
          express or implied, including but not limited to the warranties of
          merchantability, fitness for a particular purpose, and
          non-infringement. You use it at your own risk.
        </p>
      </section>

      <section>
        <h2>3. Limitation of liability</h2>
        <p>
          In no event shall the maintainer be liable for any claim, damages, or
          other liability, including data loss or service disruption, arising
          from the use of, or inability to use, the software.
        </p>
      </section>

      <section>
        <h2>4. Acceptable use</h2>
        <p>You are responsible for how you use the tool. You must:</p>
        <ul>
          <li>
            have authorization to access any Atlassian Jira instance you
            connect it to;
          </li>
          <li>
            comply with Atlassian Jira&apos;s terms of service and your
            organization&apos;s policies;
          </li>
          <li>not use the tool for unauthorized access, or for any unlawful purpose.</li>
        </ul>
      </section>

      <section>
        <h2>5. No affiliation</h2>
        <p>
          Jira CLI is an <strong>independent, unofficial</strong> tool. It is not
          affiliated with, endorsed by, or sponsored by Atlassian Jira or its
          vendor. All product names and trademarks are the property of their
          respective owners.
        </p>
      </section>

      <section>
        <h2>6. Third-party services</h2>
        <p>
          Your use of Atlassian Jira (and any other service you connect to) is
          governed by that provider&apos;s own terms and policies, for which the
          maintainer is not responsible.
        </p>
      </section>

      <section>
        <h2>7. Changes</h2>
        <p>
          These terms may be updated; changes are posted here with a new
          effective date. Continued use after a change constitutes acceptance.
        </p>
      </section>

      <section>
        <h2>8. Contact</h2>
        <p>
          <strong>Piyush Gambhir</strong> ·{' '}
          <a href="mailto:developer.piyushgambhir@gmail.com">
            developer.piyushgambhir@gmail.com
          </a>{' '}
          · <Link href="/contact">contact page</Link>
        </p>
      </section>
    </LegalPage>
  );
}

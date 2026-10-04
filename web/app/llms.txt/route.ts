import { source } from '@/lib/source';
import { llms } from 'fumadocs-core/source';
import { site } from '@/lib/site';
import { getOtherSuiteProjects } from '@/lib/suite';
import { siteUrl } from '@/lib/shared';

export const revalidate = false;

const agentPreamble =
  'Jira CLI is agent-ready and harness-agnostic. Claude Code, OpenAI Codex, Cursor, and any agent harness that can run shell commands can use structured JSON/YAML output, read-only mode, and no-input flags to automate Jira issues, JQL search, sprints, and boards.';

export async function GET() {
  // index() returns a Promise since fumadocs-core 16.15.17. Its links are
  // root-relative and miss the basePath, so make them absolute.
  const index = (await llms(source).index()).replace(
    /\]\((\/[^)]+)\)/g,
    (_match, path: string) => `](${siteUrl}${path})`,
  );
  const relatedProjects = getOtherSuiteProjects(site.repo)
    .map(({ href }) => `- ${href}`)
    .join('\n');

  return new Response(
    `${agentPreamble}\n\n${index}\n\n## Related independent CLI projects\n\n${relatedProjects}\n`,
    {
      headers: {
        'Content-Type': 'text/plain; charset=utf-8',
      },
    },
  );
}

import {
  Bot,
  GitBranch,
  KeyRound,
  ListChecks,
  Search,
  Zap,
  type LucideIcon,
} from 'lucide-react';

export interface Feature {
  icon: LucideIcon;
  title: string;
  body: string;
}

export interface SiteConfig {
  /** Display name, e.g. "Jira CLI" */
  name: string;
  /** The binary invoked in examples, e.g. "jira" */
  binary: string;
  /** GitHub "owner/repo" */
  repo: string;
  /** One-line hero heading */
  tagline: string;
  /** Hero sub-paragraph */
  description: string;
  /** Small pill above the heading */
  badge: string;
  /** One-line install command shown in the hero */
  installCommand: string;
  /** Feature cards */
  features: Feature[];
  /** Title above the code block */
  exampleTitle: string;
  /** Shell example rendered in the terminal card */
  example: string;
  /** Optional: tech / query languages this CLI speaks (logo strip) */
  compatible?: string[];
  /** Optional: features section heading (default: "Everything, from one binary") */
  featuresTitle?: string;
  /** Optional: features section subheading */
  featuresSubtitle?: string;
  /** Optional: CTA band body (default mentions installing the binary) */
  ctaBody?: string;
}

export const site: SiteConfig = {
  name: 'Jira CLI',
  binary: 'jira',
  repo: 'piyush-gambhir/jira-cli',
  tagline: 'Jira from your terminal',
  description:
    'A fast, scriptable CLI over the Jira REST API. Manage issues, JQL search, transitions, comments, boards, and sprints — built for humans and coding agents alike.',
  badge: 'Open-source · Cloud & Server/DC',
  installCommand:
    'curl -sSfL https://raw.githubusercontent.com/piyush-gambhir/jira-cli/main/install.sh | sh',
  features: [
    {
      icon: Search,
      title: 'JQL & issues',
      body: 'Search with raw JQL or convenience filters. Create, edit, assign, transition, and delete issues from the shell.',
    },
    {
      icon: KeyRound,
      title: 'Every auth method',
      body: 'OAuth 2.0 (3LO), Cloud API tokens, scoped tokens, Server/DC PATs, and basic auth — with multiple named profiles.',
    },
    {
      icon: Bot,
      title: 'Agent-friendly',
      body: '-o json|yaml on every read, --read-only safety mode, --no-input, and clean stdout/stderr separation.',
    },
    {
      icon: GitBranch,
      title: 'Agile built in',
      body: 'Boards, sprints, epics, backlogs, comments, worklogs, attachments, links, watchers, and votes.',
    },
    {
      icon: Zap,
      title: 'Fast & scriptable',
      body: 'A single static binary. Automatic 429 backoff, cursor pagination, and ADF-aware rich text.',
    },
    {
      icon: ListChecks,
      title: 'Cloud & Server/DC',
      body: 'One CLI for both deployments across macOS, Linux, and Windows (amd64 and arm64).',
    },
  ],
  exampleTitle: 'A five-line tour',
  example: `# Authenticate (API token — no app needed)
jira auth login --type api_token

# Find your open work as JSON
jira issue search "assignee = currentUser() AND statusCategory != Done" -o json

# Create, comment, and move an issue
jira issue create -p ABC --type Task --summary "Set up CI"
jira issue comment ABC-123 --body "On it"
jira issue transition ABC-123 "In Progress"`,
  compatible: [
    "JQL",
    "ADF",
    "OAuth 2.0",
    "REST API v2 / v3",
    "Cloud",
    "Server / DC",
  ],
};

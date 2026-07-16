export interface SuiteProject {
  name: string;
  href: string;
}

export const suite: readonly SuiteProject[] = [
  {
    name: 'jira-cli',
    href: 'https://github.com/piyush-gambhir/jira-cli',
  },
  {
    name: 'jenkins-cli',
    href: 'https://github.com/piyush-gambhir/jenkins-cli',
  },
  {
    name: 'es-cli',
    href: 'https://github.com/piyush-gambhir/es-cli',
  },
  {
    name: 'grafana-cli',
    href: 'https://github.com/piyush-gambhir/grafana-cli',
  },
  {
    name: 'cubeapm-cli',
    href: 'https://github.com/piyush-gambhir/cubeapm-cli',
  },
  {
    name: 'nginxpm-cli',
    href: 'https://github.com/piyush-gambhir/nginxpm-cli',
  },
  {
    name: 'reckon',
    href: 'https://github.com/piyush-gambhir/reckon',
  },
];

export function getOtherSuiteProjects(currentSite: string): SuiteProject[] {
  const currentName = currentSite.split('/').at(-1)?.replace(/\.git$/, '');

  return suite.filter(({ name }) => name !== currentName);
}

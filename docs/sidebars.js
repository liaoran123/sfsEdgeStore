/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  tutorialSidebar: [
    {
      type: 'category',
      label: 'Get Started',
      link: { type: 'doc', id: 'get-started/index' },
      items: [
        'get-started/introduction',
        'get-started/quick-start',
        'get-started/installation',
        'get-started/first-deployment',
      ],
    },
    {
      type: 'category',
      label: 'How-To Guides',
      link: { type: 'doc', id: 'how-to/index' },
      items: [
        'how-to/configure-mqtt',
        'how-to/manage-devices',
        'how-to/setup-tls',
        'how-to/backup-restore',
        'how-to/upgrade',
        'how-to/troubleshoot',
        'how-to/configure-monitoring',
        'how-to/setup-alerts',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      link: { type: 'doc', id: 'reference/index' },
      items: [
        {
          type: 'category',
          label: 'API Reference',
          link: { type: 'doc', id: 'reference/api/index' },
          items: [
            'reference/api/overview',
            'reference/api/endpoints',
            'reference/api/authentication',
          ],
        },
        'reference/configuration',
        'reference/cli-commands',
        'reference/metrics',
        {
          type: 'category',
          label: 'Architecture',
          link: { type: 'doc', id: 'reference/architecture/index' },
          items: [
            'reference/architecture/overview',
            'reference/architecture/components',
            'reference/architecture/data-flow',
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Concepts',
      link: { type: 'doc', id: 'concepts/index' },
      items: [
        'concepts/edge-computing',
        'concepts/edgex-integration',
        'concepts/database-design',
        'concepts/security-model',
        'concepts/licensing',
      ],
    },
    {
      type: 'category',
      label: 'Support',
      link: { type: 'doc', id: 'support/index' },
      items: [
        'support/faq',
        'support/troubleshooting',
        'support/changelog',
        'support/roadmap',
        'support/contact',
      ],
    },
  ],
};

module.exports = sidebars;

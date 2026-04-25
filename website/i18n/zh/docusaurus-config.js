module.exports = {
  title: 'sfsEdgeStore',
  tagline: '轻量级工业物联网边缘数据存储适配器',
  favicon: 'img/favicon.ico',

  url: 'https://your-domain.com',
  baseUrl: '/',

  organizationName: 'your-org',
  projectName: 'sfsEdgeStore',

  trailingSlash: false,

  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'zh'],
    localeConfigs: {
      en: {
        label: 'English',
        direction: 'ltr',
        htmlLang: 'en-US',
      },
      zh: {
        label: '中文',
        direction: 'ltr',
        htmlLang: 'zh-CN',
      },
    },
  },

  presets: [
    [
      'classic',
      ({
        docs: {
          sidebarPath: './sidebars.js',
          editUrl: 'https://github.com/your-org/sfsEdgeStore/tree/main/website/',
          showLastUpdateTime: true,
          showLastUpdateAuthor: true,
          lastVersion: 'current',
          versions: {
            current: {
              label: 'Latest',
              path: '',
            },
          },
        },
        blog: {
          showReadingTime: true,
          editUrl: 'https://github.com/your-org/sfsEdgeStore/tree/main/website/',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themeConfig: ({
    navbar: {
      title: 'sfsEdgeStore',
      logo: {
        alt: 'sfsEdgeStore Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: '文档',
        },
        {
          to: '/api',
          label: 'API 参考',
          position: 'left',
        },
        {
          to: '/pricing',
          label: '定价',
          position: 'left',
        },
        {
          type: 'docsVersionDropdown',
          position: 'right',
        },
        {
          type: 'localeDropdown',
          position: 'right',
        },
        {
          href: 'https://github.com/your-org/sfsEdgeStore',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },

    footer: {
      style: 'dark',
      links: [
        {
          title: '文档',
          items: [
            {
              label: '快速开始',
              to: '/docs/get-started/quick-start',
            },
            {
              label: '安装指南',
              to: '/docs/get-started/installation',
            },
            {
              label: '配置参考',
              to: '/docs/reference/configuration',
            },
          ],
        },
        {
          title: '社区',
          items: [
            {
              label: 'GitHub Issues',
              href: 'https://github.com/your-org/sfsEdgeStore/issues',
            },
            {
              label: '讨论区',
              href: 'https://github.com/your-org/sfsEdgeStore/discussions',
            },
          ],
        },
        {
          title: '更多',
          items: [
            {
              label: '博客',
              to: '/blog',
            },
            {
              label: '变更日志',
              to: '/docs/support/changelog',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/your-org/sfsEdgeStore',
            },
          ],
        },
      ],
      copyright: `版权所有 © ${new Date().getFullYear()} sfsEdgeStore。使用 Docusaurus 构建。`,
    },

    prism: {
      theme: require('prism-react-renderer/themes/github'),
      darkTheme: require('prism-react-renderer/themes/dracula'),
      additionalLanguages: ['bash', 'json', 'go', 'yaml', 'docker'],
    },

    colorMode: {
      defaultMode: 'light',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
  }),

  markdown: {
    mermaid: true,
  },
  themes: ['@docusaurus/theme-mermaid'],
};

module.exports = config;

// @ts-check
// `@type` JSDoc annotations allow editor autocompletion and type checking
// (when paired with `@ts-check`).

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'sfsEdgeStore',
  tagline: 'Lightweight Industrial IoT Edge Data Storage Adapter',
  favicon: 'img/favicon.ico',

  // Set the production URL for the site
  url: 'https://your-domain.com',
  baseUrl: '/',

  // GitHub pages deployment config
  organizationName: 'your-org',
  projectName: 'sfsEdgeStore',

  // Enable trailing slashes
  trailingSlash: false,

  // Internationalization
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

  // Theme presets and plugins
  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
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

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Replace with your project's social card
      // image: 'img/social-card.png',

      // Navigation bar
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
            label: 'Documentation',
          },
          {
            to: '/api',
            label: 'API Reference',
            position: 'left',
          },
          {
            to: '/pricing',
            label: 'Pricing',
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

      // Footer
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Quick Start',
                to: '/docs/get-started/quick-start',
              },
              {
                label: 'Installation',
                to: '/docs/get-started/installation',
              },
              {
                label: 'Configuration',
                to: '/docs/reference/configuration',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'GitHub Issues',
                href: 'https://github.com/your-org/sfsEdgeStore/issues',
              },
              {
                label: 'Discussions',
                href: 'https://github.com/your-org/sfsEdgeStore/discussions',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'Blog',
                to: '/blog',
              },
              {
                label: 'Changelog',
                to: '/docs/support/changelog',
              },
              {
                label: 'GitHub',
                href: 'https://github.com/your-org/sfsEdgeStore',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} sfsEdgeStore. Built with Docusaurus.`,
      },

      // Prism syntax highlighting
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ['bash', 'json', 'go', 'yaml', 'docker'],
      },

      // Color mode (light/dark)
      colorMode: {
        defaultMode: 'light',
        disableSwitch: false,
        respectPrefersColorScheme: true,
      },
    }),

  // Markdown plugins configuration
  markdown: {
    mermaid: true,
  },
  themes: ['@docusaurus/theme-mermaid'],
};

module.exports = config;

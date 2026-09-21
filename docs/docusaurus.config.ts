import type {PrismTheme} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const nordPrismDarkTheme: PrismTheme = {
  plain: {
    color: '#d8dee9',
    backgroundColor: '#2e3440',
  },
  styles: [
    {types: ['comment', 'prolog', 'doctype', 'cdata'], style: {color: '#4c566a', fontStyle: 'italic'}},
    {types: ['punctuation'], style: {color: '#d8dee9'}},
    {types: ['property', 'tag', 'boolean', 'constant', 'symbol', 'deleted'], style: {color: '#91a1c1'}},
    {types: ['number'], style: {color: '#b48ead'}},
    {types: ['selector', 'attr-name', 'string', 'char', 'builtin', 'inserted', 'attr-value'], style: {color: '#a3be8c'}},
    {types: ['operator', 'entity', 'url', 'atrule', 'keyword'], style: {color: '#91a1c1'}},
    {types: ['function', 'class-name'], style: {color: '#88c0d0'}},
    {types: ['regex', 'important', 'variable'], style: {color: '#ebcb8b'}},
    {types: ['important', 'bold'], style: {fontWeight: 'bold'}},
    {types: ['italic'], style: {fontStyle: 'italic'}},
  ],
};

const nordPrismLightTheme: PrismTheme = {
  plain: {
    color: '#2e3440',
    backgroundColor: '#e5e9f0',
  },
  styles: [
    {types: ['comment', 'prolog', 'doctype', 'cdata'], style: {color: '#434c5e', fontStyle: 'italic'}},
    {types: ['punctuation'], style: {color: '#2e3440'}},
    {types: ['property', 'tag', 'boolean', 'constant', 'symbol', 'deleted'], style: {color: '#5e81ac'}},
    {types: ['number'], style: {color: '#b48ead'}},
    {types: ['selector', 'attr-name', 'string', 'char', 'builtin', 'inserted', 'attr-value'], style: {color: '#4c8a5e'}},
    {types: ['operator', 'entity', 'url', 'atrule', 'keyword'], style: {color: '#5e81ac'}},
    {types: ['function', 'class-name'], style: {color: '#3b7d8f'}},
    {types: ['regex', 'important', 'variable'], style: {color: '#a86f1f'}},
    {types: ['important', 'bold'], style: {fontWeight: 'bold'}},
    {types: ['italic'], style: {fontStyle: 'italic'}},
  ],
};

const config: Config = {
  title: 'Inundated',
  tagline: 'A personal time and task management system.',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://your-docusaurus-site.example.com',
  baseUrl: '/',

  organizationName: 'LarssonOliver',
  projectName: 'inundated',

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/LarssonOliver/inundated/tree/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Inundated',
      logo: {
        alt: 'Inundated',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          href: 'https://github.com/LarssonOliver/inundated',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      copyright: `© ${new Date().getFullYear()} Oliver Larsson. Built with Docusaurus.`,
    },
    prism: {
      theme: nordPrismLightTheme,
      darkTheme: nordPrismDarkTheme,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;

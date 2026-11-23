import path from 'node:path';
import type {Config, Plugin} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

function yamlSourcePlugin(): Plugin {
  return {
    name: 'yaml-source',
    configureWebpack() {
      return {
        module: {
          rules: [{test: /\.ya?ml$/, type: 'asset/source'}],
        },
      };
    },
  };
}

const config: Config = {
  title: 'Telemetry',
  tagline: 'Documentation and capability specs',
  url: 'http://localhost',
  baseUrl: '/',
  onBrokenLinks: 'warn',
  staticDirectories: [path.resolve(__dirname, 'static')],
  clientModules: [path.resolve(__dirname, 'src/client/d3.ts')],
  markdown: {
    mermaid: true,
    hooks: {onBrokenMarkdownLinks: 'warn'},
  },
  themes: ['@docusaurus/theme-mermaid'],
  plugins: [
    yamlSourcePlugin,
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'docs',
        path: '..',
        routeBasePath: 'docs',
        sidebarPath: './sidebars/docs.ts',
        include: ['**/*.md', '**/*.mdx'],
        exclude: ['**/website/**', '**/node_modules/**', '**/build/**', '**/.docusaurus/**'],
      },
    ],
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'specs',
        path: '../../openspec/specs',
        routeBasePath: 'specs',
        sidebarPath: './sidebars/specs.ts',
        include: ['**/*.md', '**/*.mdx'],
      },
    ],
  ],
  presets: [
    [
      'classic',
      {
        docs: false,
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],
  themeConfig: {
    navbar: {
      title: 'Telemetry',
      items: [
        {type: 'doc', docId: 'architecture', docsPluginId: 'docs', label: 'Docs', position: 'left'},
        {type: 'doc', docId: 'api', docsPluginId: 'docs', label: 'API', position: 'left'},
        {
          type: 'doc',
          docId: 'aws-deployment/overview',
          docsPluginId: 'docs',
          label: 'AWS',
          position: 'left',
        },
        {
          type: 'doc',
          docId: 'satellite-telemetry/spec',
          label: 'Specs',
          position: 'left',
          docsPluginId: 'specs',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {label: 'Architecture', to: '/docs/architecture'},
            {label: 'API reference', to: '/docs/api'},
            {label: 'AWS deployment', to: '/docs/aws-deployment/overview'},
          ],
        },
        {
          title: 'Reference',
          items: [
            {label: 'Capability specs', to: '/specs/satellite-telemetry/spec'},
            {label: 'Decision records', to: '/docs/adr/'},
            {
              label: 'OpenAPI source',
              href: 'https://github.com/mfloriach/telemetry-service/blob/main/docs/openapi.yaml',
            },
          ],
        },
      ],
      copyright: `Telemetry documentation · Built ${new Date().getFullYear()}`,
    },
    colorMode: {
      defaultMode: 'light',
      respectPrefersColorScheme: true,
      disableSwitch: false,
    },
    mermaid: {
      options: {
        flowchart: {curve: 'basis', htmlLabels: true, useMaxWidth: true},
      },
    },
  },
};

export default config;

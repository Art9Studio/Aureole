const {description} = require('../../package')

module.exports = {
    /**
     * Ref：https://v1.vuepress.vuejs.org/config/#title
     */
    title: 'Документация Aureole',
    /**
     * Ref：https://v1.vuepress.vuejs.org/config/#description
     */
    description: description,

    /**
     * Extra tags to be injected to the page HTML `<head>`
     *
     * ref：https://v1.vuepress.vuejs.org/config/#head
     */
    head: [
        ['meta', {name: 'theme-color', content: '#3eaf7c'}],
        ['meta', {name: 'apple-mobile-web-app-capable', content: 'yes'}],
        ['meta', {name: 'apple-mobile-web-app-status-bar-style', content: 'black'}]
    ],

    /**
     * Theme configuration, here is the default theme configuration for VuePress.
     *
     * ref：https://v1.vuepress.vuejs.org/theme/default-theme-config.html
     */
    themeConfig: {
        repo: 'https://github.com/Art9Studio/Aureole',
        editLinks: false,
        docsDir: 'docs',
        editLinkText: '',
        lastUpdated: false,
        searchMaxSuggestions: 10,
        sidebarDepth: 2,
        collapsable: true,
        nav: [
            {
                text: 'Config',
                link: '/config/project/'
            },
            {
                text: 'VuePress',
                link: 'https://v1.vuepress.vuejs.org'
            }
        ],
        sidebar: {
            title: 'Aureole Config',
            '/config/': [
                'project',
                'identity',
                'collection',
                'authn',
                'authz',
                'storage',
                'hasher',
                'crypto_key',
                'sender'
            ],
        }
    },

    /**
     * Apply plugins，ref：https://v1.vuepress.vuejs.org/zh/plugin/
     */
    plugins: [
        '@vuepress/plugin-back-to-top',
        '@vuepress/plugin-medium-zoom',
    ]
}

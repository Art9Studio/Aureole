module.exports = {
    base: "/Aureole/",
    title: 'Документация Aureole',
    description: "Самый гибкий и современный сервер аутентификации с открытым исходным кодом.",

    head: [
        ['meta', {name: 'theme-color', content: '#3eaf7c'}],
        ['meta', {name: 'apple-mobile-web-app-capable', content: 'yes'}],
        ['meta', {name: 'apple-mobile-web-app-status-bar-style', content: 'black'}],
        ['link', {rel: 'icon', href: 'favicon.ico'}]
    ],

    themeConfig: {
        repo: 'https://github.com/Art9Studio/Aureole',
        editLink: false,
        docsDir: 'docs',
        lastUpdated: false,
        searchMaxSuggestions: 10,
        sidebarDepth: 2,
        collapsable: true,
        navbar: [
            {
                text: 'Конфиг',
                link: '/config/project/'
            },
            {
                text: 'Render',
                link: '/render/'
            }
        ],
        sidebar: {
            '/render/': [
                {
                    text: 'Render',
                    children: ['/render/Readme.md']
                }
            ],
            '/config/': [
                '/config/project.md',
                '/config/authn.md',
                '/config/authz.md',
                '/config/2fa.md',
                '/config/id_manager.md',
                '/config/storage.md',
                '/config/crypto_storage.md',
                '/config/crypto_key.md',
                '/config/sender.md',
                '/config/admin_plugin.md'
            ],
        }
    },

    plugins: [
        '@vuepress/plugin-back-to-top',
        '@vuepress/plugin-medium-zoom',
        '@vuepress/plugin-search',
    ]
}

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: false },
  modules: [
    '@nuxtjs/tailwindcss'
  ],
  runtimeConfig: {
    // Server-only: internal gateway URL for SSR fetches (never exposed to browsers).
    apiBaseInternal: process.env.NUXT_API_BASE_INTERNAL || 'http://localhost:8080/api/v1',
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080/api/v1',
    },
  },
  nitro: {
    preset: process.env.NITRO_PRESET || 'cloudflare-module'
  },
  app: {
    head: {
      title: 'Treewiki - Collaborative Knowledge Platform',
      htmlAttrs: { lang: 'en' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { name: 'description', content: 'Treewiki is a global collaborative knowledge platform organizing human understanding into topic trees, revision ledgers, and neutral encyclopedic references. Free under CC BY-SA 4.0.' },
        { name: 'theme-color', content: '#0F766E' },
        { property: 'og:site_name', content: 'Treewiki' },
        { property: 'og:type', content: 'website' },
        { property: 'og:title', content: 'Treewiki - Collaborative Knowledge Platform' },
        { property: 'og:description', content: 'Structured knowledge, topic trees, and immutable revision ledgers. Free under CC BY-SA 4.0.' },
        { property: 'og:image', content: '/logos/treewiki.svg' },
        { name: 'twitter:card', content: 'summary' },
        { name: 'twitter:title', content: 'Treewiki - Collaborative Knowledge Platform' },
        { name: 'twitter:description', content: 'Structured knowledge, topic trees, and immutable revision ledgers. Free under CC BY-SA 4.0.' },
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'canonical', href: process.env.NUXT_PUBLIC_SITE_URL || 'http://localhost:3000' },
      ],
    }
  }
})

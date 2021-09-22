<template>
  <v-fade-transition mode="out-in">
    <router-view />
  </v-fade-transition>
</template>

<script>
  // Styles
  import '@/styles/overrides.sass'
  import WebsocketService from '@/services/websocket.service'

  export default {
    name: 'App',
    metaInfo: {
      title: 'dd-opcda',
      titleTemplate: '%s | cyops-se admin',
      htmlAttrs: { lang: 'en' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      ],
    },

    created () {
      WebsocketService.connect()
    },
  }
</script>

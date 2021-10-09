<template>
  <v-card
    v-bind="$attrs"
    class="v-card--material mt-4"
  >
    <v-card-title class="align-start">
      <v-sheet
        :color="copy.status === 1 ? 'success' : 'error'"
        width="100%"
        class="overflow-hidden mt-n9 transition-swing v-card--material__sheet"
        elevation="6"
        max-width="100%"
        rounded
      >
        <v-theme-provider dark>
          <v-row align="center">
            <v-col cols="2">
              <div class="pa-5">
                <v-icon large>
                  mdi-tag-multiple
                </v-icon>
              </div>
            </v-col>
            <v-col cols="6">
              <div class="pa-5 white--text">
                <span class="text-h3 text-no-wrap">
                  {{ copy.name }}
                </span>
                <span class="text-h4 text-no-wrap">
                  {{ copy.status == 1 ? 'RUNNING' : 'STOPPED' }}
                </span>
                <div>Send count: {{ copy.counter }}</div>
              </div>
            </v-col>
            <v-col cols="2">
              <div class="text-right">
                <v-btn @click="startStop">
                  <div v-html="copy.status == 1 ? 'STOP' : 'START'" />
                </v-btn>
              </div>
            </v-col>
          </v-row>
        </v-theme-provider>
      </v-sheet>

      <div class="pl-3 text-h4 v-card--material__title">
        <div class="text-subtitle-1 mb-n4 mt-4">
          <template>
            {{ copy.description }}
          </template>
        </div>
      </div>
    </v-card-title>

    <slot />

    <template>
      <v-divider class="mt-2 mx-4" />

      <v-card-actions class="px-4 text-caption grey--text">
        <v-icon
          class="mr-1"
          small
        >
          mdi-clock-outline
        </v-icon>

        <span
          class="text-caption grey--text font-weight-light"
          v-text="'Last run: ' + copy.lastrun.replace('T', ' ').substring(0, 19)"
        />
      </v-card-actions>
    </template>
  </v-card>
</template>

<script>
  import ApiService from '@/services/api.service'
  import WebsocketService from '@/services/websocket.service'
  export default {
    name: 'MaterialGroupCard',

    inheritAttrs: false,

    props: {
      group: {
        type: Object,
        default: () => ({}),
      },
      eventHandlers: {
        type: Array,
        default: () => ([]),
      },
    },

    data: () => ({
      copy: {},
    }),

    watch: {
      $route (to, from) {
        console.log('route change: ', to, from)
      },
    },

    created () {
      this.copy = Object.assign({}, this.group)
      var t = this
      WebsocketService.topic('data.group', function (topic, group) {
        if (t.copy.ID === group.ID) t.copy = group
      })
    },

    methods: {
      startStop () {
        console.log('START STOP: ' + this.group.name)
        var action = this.copy.status === 1 ? 'stop' : 'start'
        ApiService.get('opc/group/' + action + '/' + this.group.ID)
          .then(response => {
            this.$notification.success('Collection of group tags ' + (this.copy.status === 1 ? 'stopped' : 'started'))
            this.copy.status = this.copy.status === 1 ? 0 : 1
            if (this.copy.status === 1) {
            } else {
              console.log('clearing timer for: ' + this.group.name)
            }
          }).catch(response => {
            console.log('ERROR response: ' + response.message)
            this.$notification.error('Failed to start collection of group tags: ' + response.message)
          })
      },

      refresh () {
        ApiService.get('opc/group/' + this.group.ID)
          .then(response => {
            this.copy = response.data
          }).catch(response => {
            console.log('ERROR response (refresh): ' + response.message)
          })
      },
    },
  }
</script>

<style lang="sass">
.group-button
  font-size: .875rem !important
  margin-left: auto
  text-align: right
</style>

<template>
  <v-container
    id="dashboard-view"
    fluid
    tag="section"
  >
    <v-row>
      <v-col cols="12">
        <v-row>
          <v-col
            v-for="(group, i) in groups"
            :key="`group-${i}`"
            cols="12"
            md="6"
            lg="4"
          >
            <material-group-card :group="group" />
          </v-col>
        </v-row>
      </v-col>
      <error-logs-tables-view />
    </v-row>
  </v-container>
</template>

<script>
  // Utilities
  import ErrorLogsTablesView from './ErrorLogs'
  import ApiService from '@/services/api.service'

  export default {
    name: 'DashboardView',

    components: {
      ErrorLogsTablesView,
    },

    data: () => ({
      tabs: 0,
      tags: [],
      groups: [],
      servers: [],
    }),

    computed: {
    },

    watch: {
      $route (to, from) {
        console.log('route change: ', to, from)
      },
    },

    created () {
      ApiService.get('data/opc_groups')
        .then(response => {
          this.groups = response.data
          // this.charts = []
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })
    },
  }
</script>

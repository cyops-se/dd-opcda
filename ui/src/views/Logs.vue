<template>
  <v-container
    id="logs-view"
    fluid
    tag="section"
  >
    <v-card>
      <v-card-title class="text-h4">
        System logs
        <v-spacer />
        <v-text-field
          v-model="search"
          append-icon="mdi-magnify"
          label="Search"
          single-line
          hide-details
        />
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="items"
        :search="search"
        :loading="loading"
        loading-text="Loading... Please wait"
        sort-by="time"
        :sort-desc="sortDesc"
      />
    </v-card>
  </v-container>
</template>

<script>
  import ApiService from '@/services/api.service'
  export default {
    name: 'LogsTablesView',

    data: () => ({
      search: '',
      loading: false,
      headers: [
        {
          text: 'Time',
          align: 'start',
          filterable: true,
          value: 'time',
          width: 250,
        },
        { text: 'Category', value: 'category', width: 150 },
        { text: 'Title', value: 'title', width: 150 },
        { text: 'Description', value: 'description' },
      ],
      items: [],
      sortDesc: true,
    }),

    created () {
      ApiService.get('data/logs')
        .then(response => {
          for (const i of response.data) {
            i.time = i.time.replace('T', ' ').replace('Z', '').substring(0, 19)
          }
          this.items = response.data
          this.loading = false
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })
    },
  }
</script>

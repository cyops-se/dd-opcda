<template>
  <v-data-table
    :headers="headers"
    :items="items"
    class="elevation-1"
  >
    <template v-slot:top>
      <v-toolbar
        flat
      >
        <v-toolbar-title>Tags</v-toolbar-title>
        <v-divider
          class="mx-4"
          inset
          vertical
        />
        <v-spacer />
        <v-dialog
          v-model="dialog"
          max-width="500px"
        >
          <v-card>
            <v-card-title>
              <span class="text-h5">Tag</span>
            </v-card-title>

            <v-card-text>
              <v-container>
                <v-row>
                  <v-col
                    cols="12"
                  >
                    <v-text-field
                      v-model="editedItem.name"
                      label="Name"
                      readonly
                    />
                  </v-col>
                </v-row>
                <v-row>
                  <v-col
                    cols="12"
                  >
                    <v-combobox
                      v-model="editedItem.group"
                      :items="availableGroups"
                      item-text="name"
                      label="Group"
                      outlined
                    />
                  </v-col>
                </v-row>
                <v-row>
                  <v-col
                    cols="12"
                  >
                    <v-textarea
                      v-model="editedItem.description"
                      label="Description"
                      outlined
                    />
                  </v-col>
                </v-row>
              </v-container>
            </v-card-text>

            <v-card-actions>
              <v-spacer />
              <v-btn
                color="blue darken-1"
                text
                @click="close"
              >
                Cancel
              </v-btn>
              <v-btn
                color="blue darken-1"
                text
                @click="save"
              >
                Save
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </v-toolbar>
    </template>
    <template v-slot:item.actions="{ item }">
      <v-icon
        class="mr-2"
        @click="editItem(item)"
      >
        mdi-pencil
      </v-icon>
      <v-icon
        @click="deleteItem(item)"
      >
        mdi-delete
      </v-icon>
    </template>
  </v-data-table>
</template>

<script>
  import ApiService from '@/services/api.service'
  export default {
    name: 'TagTableView',

    data: () => ({
      snackbar: false,
      snackbarMessage: '',
      snackbarType: 'warning',
      dialog: false,
      dialogDelete: false,
      search: '',
      loading: false,
      headers: [
        {
          text: 'ID',
          align: 'start',
          filterable: false,
          value: 'ID',
          width: 75,
        },
        { text: 'Name', value: 'name', width: '10%' },
        { text: 'Description', value: 'description', width: '40%' },
        { text: 'Group', value: 'group.name', width: '40%' },
        { text: 'Actions', value: 'actions', width: 1, sortable: false },
      ],
      items: [],
      editedIndex: -1,
      editedItem: {
        fullname: '',
        email: '',
      },
      defaultItem: {
        fullname: '',
        email: '',
      },
      availableGroups: [],
      groups: [],
      groupsTable: {},
    }),

    created () {
      this.loading = true
      ApiService.get('data/opc_tags')
        .then(response => {
          this.items = response.data
          this.loading = false
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })
      ApiService.get('data/opc_groups')
        .then(response => {
          this.groups = response.data
          this.availableGroups = this.groups
        }).catch(response => {
          console.log('ERROR response: ' + JSON.stringify(response))
        })
    },

    methods: {
      initialize () {},

      editItem (item) {
        this.editedIndex = this.items.indexOf(item)
        this.editedItem = Object.assign({}, item)
        this.editedItem.groupname = item.group.name
        this.dialog = true
      },

      deleteItem (item) {
        console.log('deleting item: ' + JSON.stringify(item))
        ApiService.delete('data/opc_tags/' + item.ID)
          .then(response => {
            for (var i = 0; i < this.items.length; i++) {
              if (this.items[i].ID === item.ID) this.items.splice(i, 1)
            }
            this.$notification.success('Tag deleted')
          }).catch(response => {
            console.log('ERROR response: ' + response.message)
          })
      },

      close () {
        this.dialog = false
        this.$nextTick(() => {
          this.editedItem = Object.assign({}, this.defaultItem)
          this.editedIndex = -1
        })
      },

      save () {
        if (this.editedIndex > -1) {
          console.log('edited item: ' + JSON.stringify(this.editedItem))
          Object.assign(this.items[this.editedIndex], this.editedItem)
          ApiService.put('data/opc_tags', this.editedItem)
            .then(response => {
              // this.$notification.success('Tag ' + response.data.fullname + ' successfully updated!')
            }).catch(response => {
              this.$notification.error('Failed to update tag!' + response)
            })
        } else {
          this.items.push(this.editedItem)
          ApiService.post('data/opc_tags', this.editedItem)
            .then(response => {
              // this.successMessage('Tag ' + response.data.fullname + ' successfully added!')
            }).catch(response => {
              this.failureMessage('Failed to add tag!' + response)
            })
        }
        this.close()
      },
    },
  }
</script>

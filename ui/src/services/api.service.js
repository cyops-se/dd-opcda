import Vue from 'vue'
import axios from 'axios'
import VueAxios from 'vue-axios'
import JwtService from '@/services/jwt.service'

/**
 * Service to call HTTP request via Axios
 */
const ApiService = {
  init () {
    Vue.use(VueAxios, axios)

    if (process.env.NODE_ENV === 'production') {
      Vue.axios.defaults.baseURL = 'http://localhost:3000'
    } else {
      Vue.axios.defaults.baseURL = 'http://localhost:8080'
    }
  },

  /**
   * Set the default HTTP request headers
   */
  setHeader () {
    Vue.axios.defaults.headers.common.Authorization = `Bearer ${JwtService.getToken()}`
  },

  query (resource, params) {
    resource = 'api/' + resource
    return Vue.axios.get(resource, params).catch(error => {
      throw new Error(`[KT] ApiService ${error}`)
    })
  },

  /**
   * Send the GET HTTP request
   * @param resource
   * @param slug
   * @returns {*}
   */
  get (resource) {
    ApiService.setHeader()
    resource = 'api/' + resource
    return Vue.axios.get(`${resource}`).catch(error => {
      throw new Error(`[KT] ApiService ${error}`)
    })
  },

  /**
   * Set the POST HTTP request
   * @param resource
   * @param params
   * @returns {*}
   */
  post (resource, params) {
    ApiService.setHeader()
    resource = 'api/' + resource
    return Vue.axios.post(`${resource}`, params)
  },

  /**
   * Set the POST HTTP request for upload
   * @param resource
   * @param files
   * @returns {*}
   */
  upload (resource, files) {
    ApiService.setHeader()
    resource = 'api/' + resource
    var formData = new FormData()
    for (var i = 0; i < files.length; i++) { formData.append('file', files[i]) }
    return Vue.axios.post(`${resource}`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data; charset=utf-8',
      },
    })
  },

  /**
   * Send the UPDATE HTTP request
   * @param resource
   * @param slug
   * @param params
   * @returns {IDBRequest<IDBValidKey> | Promise<void>}
   */
  update (resource, slug, params) {
    ApiService.setHeader()
    resource = 'api/' + resource
    return Vue.axios.put(`${resource}/${slug}`, params)
  },

  /**
   * Send the PUT HTTP request
   * @param resource
   * @param params
   * @returns {IDBRequest<IDBValidKey> | Promise<void>}
   */
  put (resource, params) {
    ApiService.setHeader()
    resource = 'api/' + resource
    return Vue.axios.put(`${resource}`, params)
  },

  /**
   * Send the DELETE HTTP request
   * @param resource
   * @returns {*}
   */
  delete (resource) {
    ApiService.setHeader()
    resource = 'api/' + resource
    return Vue.axios.delete(resource)
    // .catch(error => {
    //   throw new Error(`[RWV] ApiService ${error}`)
    // })
  },
}

export default ApiService

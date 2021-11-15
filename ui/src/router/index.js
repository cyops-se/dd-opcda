// Imports
import Vue from 'vue'
import Router from 'vue-router'
// import { trailingSlash } from '@/util/helpers'
import {
  layout,
  route,
} from '@/util/routes'

Vue.use(Router)

const router = new Router({
  mode: 'history',
  // base: '/admin/', // process.env.BASE_URL,
  base: '/ui', // process.env.BASE_URL,
  scrollBehavior: (to, from, savedPosition) => {
    if (to.hash) return { selector: to.hash }
    if (savedPosition) return savedPosition

    return { x: 0, y: 0 }
  },
  routes: [
    layout('Default', [
      route('Dashboard', null, '/'),

      // Pages
      route('ServerTable', null, 'pages/servers'),
      route('TagBrowser', null, 'pages/browse/:serverid'),
      route('GroupTable', null, 'pages/groups'),
      route('DiodeProxyTable', null, 'pages/diodeproxies'),
      route('TagTable', null, 'pages/tags'),
      route('Cache', null, 'pages/cache'),
      route('SystemSettings', null, 'pages/systemsettings'),

      // Tables
      route('Logs', null, 'tables/logs'),
      route('Users Table', null, 'tables/users'),
    ]),
    layout('Login', [

      // Pages
      route('Login', null, 'auth/login'),
    ]),
    layout('Logout', [

      // Pages
      route('Logout', null, 'auth/logout'),
    ]),
    layout('Register', [

      // Pages
      route('Register', null, 'auth/register'),
    ]),
  ],
})

router.beforeEach((to, from, next) => {
  return next() // to.path.endsWith('/') ? next() : next(trailingSlash(to.path))
})

export default router

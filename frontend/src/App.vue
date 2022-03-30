<template>
  <v-app dark>
    <v-toolbar class="toolbar" dark>
      <v-app-bar-nav-icon>
        <v-img class="logo" src="./assets/3fold_icon.png" />
      </v-app-bar-nav-icon>
      <v-toolbar-title>ThreeFold Capacity Explorer</v-toolbar-title>
    </v-toolbar>
    <v-toolbar class="toolbar" color="red">
      <v-toolbar-title class="white--text flex">Make sure to migrate your nodes to <a target="_blank" href="https://forum.threefold.io/t/farming-migration-grid-v2-v3/2143" class="text-white">TF Grid v3</a> before May 1st 2022 as this tool will no longer be supported after that date.</v-toolbar-title>
    </v-toolbar>

    <v-content class="content">
      <v-col>
        <v-row class="pa-4 mx-1">
          <h1 class="headline pt-0 pb-1 text-uppercase">
            <span>TF</span>
            <span class="font-weight-light">explorer</span>
            <span class="title font-weight-light">- {{$route.meta.displayName}}</span>
          </h1>
          <v-progress-circular
            class="refresh"
            v-if="nodesLoading || farmsLoading || gatewaysLoading"
            indeterminate
            color="primary"
          ></v-progress-circular>
          <v-btn class="refresh" icon v-else @click="refresh">
            <v-icon
              big
              color="primary"
              left
            >
              fas fa-sync-alt
            </v-icon>
          </v-btn>
        </v-row>
        <router-view></router-view>
      </v-col>
    </v-content>
    <v-bottom-navigation
      v-if="$vuetify.breakpoint.mdAndDown"
      grow
      dark
      class="primary topround"
      app
      fixed
      shift
      :value="$route.name"
    >
      <v-btn
        :value="route.name"
        icon
        v-for="(route, i) in routes"
        :key="i"
        @click="$router.push(route)"
      >
        <span>{{route.meta.displayName}}</span>
        <v-icon>{{route.meta.icon}}</v-icon>
      </v-btn>
    </v-bottom-navigation>
  </v-app>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'

export default {
  name: 'App',
  components: {},
  data: () => ({
    showDialog: false,
    dilogTitle: 'title',
    dialogBody: '',
    dialogActions: [],
    dialogImage: null,
    block: null,
    showBadge: true,
    menu: false,
    start: undefined,
    refreshInterval: undefined
  }),
  computed: {
    routes () {
      return this.$router.options.routes
    },
    ...mapGetters([
      'nodesLoading',
      'farmsLoading',
      'gatewaysLoading'
    ])
  },
  mounted () {
    // keep track when the user opened this app
    this.start = new Date()

    const _this = this
    // refresh every 10 minutes
    this.refreshInterval = setInterval(function () {
      _this.refresh()
    }, 60000)

    // if user loses focus, clear the refreshing interval
    // we don't refresh data if the page is not focused.
    window.onblur = () => {
      clearInterval(this.refreshInterval)
    }

    // instead of refreshing every 10 minutes in the background
    // we do following:
    // when the user enters the page again and 10 minutes have passed since the first visit
    // refresh all data. Start the refresh interval again (since we assume the user is going to stay on this page)
    // if the user decides to leave this page again this interval will be cleared again
    window.onfocus = () => {
      const now = new Date()
      let elapsedTime = now - this.start
      // strip the ms
      elapsedTime /= 1000
      const seconds = Math.round(elapsedTime)

      // if 10 minutes are passed since last focus, refresh data.
      if (seconds >= 600) {
        this.start = new Date()
        this.refresh()
        this.refreshInterval = setInterval(function () {
          _this.refresh()
        }, 60000)
      }
    }
  },
  methods: {
    ...mapActions({
      refresh: 'refreshData'
    })
  }
}
</script>

<style lang="scss">
.content {
  background: #fafafa !important;
}
.topround {
  border-radius: 10px 10px 0 0 !important;
}
.rounded {
  border-radius: 0 10px 10px 0 !important;
}
.v-menu__content,
.v-card {
  border-radius: 10px !important;
}
.v-card__title {
  font-size: 18px !important;
}
.spinner {
  margin-left: 20px;
}
.refresh {
  position: absolute !important;
  right: 25px !important;
  top: 30px !important;
  font-size: 30px !important;
}
.actions {
  height: 60px;
}
.logo {
  width: 40px;
  height: 40px;
  margin-left: 10px;
  margin-bottom: 3px;
}
.toolbar {
  background-color: #2d4052 !important;
  max-height: 60px;
}
</style>

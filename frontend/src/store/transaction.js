import tfService from '../services/tfService'
import lodash from 'lodash'

const getDefaultState = () => {
  return {
    user: {},
    registeredNodes: [],
    nodePage: 1,
    nodesLoading: true,
    farmPage: 1,
    farmsLoading: true,
    gatewayPage: 1,
    gatewaysLoading: true,
    nodes: undefined,
    registeredFarms: [],
    registeredGateways: [],
    farms: [],
    gatewaySpecs: {
      amountRegisteredGateways: 0,
      onlineGateways: 0
    },
    nodeSpecs: {
      amountregisteredNodes: 0,
      amountregisteredFarms: 0,
      countries: 0,
      onlinenodes: 0,
      cru: 0,
      mru: 0,
      sru: 0,
      hru: 0,
      network: 0,
      volume: 0,
      container: 0,
      zdb_namespace: 0,
      k8s_vm: 0
    }
  }
}

export default ({
  state: getDefaultState(),
  actions: {
    getName: async context => {
      var response = await tfService.getName()
      return response.data.name
    },
    getUser: async context => {
      var name = await context.dispatch('getName')
      var response = await tfService.getUser(name)
      context.commit('setUser', response.data)
    },
    getRegisteredNodes (context, params) {
      let page = context.state.nodePage
      if (!page) return

      const getNodes = (page) => {
        return tfService.getNodes(undefined, params.size, page)
      }

      getNodes(page).then(response => {
        context.commit('setRegisteredNodes', response)
        const pages = parseInt(response.headers.pages, 10)
        const promises = []
        for (var index = context.state.nodePage; index <= pages; index++) {
          promises.push(getNodes(index))
        }
        Promise.all(promises)
          .then(res => {
            res.map(response => {
              context.commit('setRegisteredNodes', response)
            })
          })
          .finally(() => {
            context.commit('setTotalSpecs', context.state.registeredNodes)
            context.commit('setNodesLoading', false)
          })
      })
    },
    getRegisteredFarms (context, params) {
      let page = context.state.nodePage
      if (!page) return

      const getFarms = (page) => {
        return tfService.registeredfarms(params.size, page)
      }

      getFarms(page).then(response => {
        context.commit('setRegisteredFarms', response)
        const pages = parseInt(response.headers.pages, 10)
        const promises = []
        for (var index = context.state.farmPage; index <= pages; index++) {
          promises.push(getFarms(index))
        }
        Promise.all(promises)
          .then(res => {
            res.map(response => {
              context.commit('setRegisteredFarms', response)
            })
          })
          .finally(() => {
            context.commit('setAmountOfFarms', context.state.registeredFarms)
            context.commit('setFarmsLoading', false)
          })
      })
    },
    getRegisteredGateways (context, params) {
      let page = params.page || context.state.gatewayPage
      if (!page) return

      const getGateways = (page) => {
        return tfService.getGateways(params.size, page)
      }

      getGateways(page).then(response => {
        context.commit('setRegisteredGateways', response)
        const pages = parseInt(response.headers.pages, 10)
        const promises = []
        for (var index = context.state.gatewayPage; index <= pages; index++) {
          promises.push(getGateways(index))
        }
        Promise.all(promises)
          .then(res => {
            res.map(response => {
              context.commit('setRegisteredGateways', response)
            })
          })
          .finally(() => {
            context.commit('setGatewaySpecs', context.state.registeredGateways)
            context.commit('setGatewaysLoading', false)
          })
      })
    },
    getFarms: context => {
      tfService.getFarms(context.getters.user.id).then(response => {
        context.commit('setFarms', response.data)
      })
    },
    resetNodes: context => {
      context.commit('setNodes', undefined)
    },
    resetState: context => {
      context.commit('resetState')
    },
    refreshData: ({ dispatch }) => {
      // reset the vuex store
      dispatch('resetState')

      // load 500 nodes at a time
      dispatch('getRegisteredNodes', { size: 500, page: 1 })
      dispatch('getRegisteredFarms', { size: 500, page: 1 })
      dispatch('getRegisteredGateways', { size: 500, page: 1 })
    }
  },
  mutations: {
    setRegisteredNodes (state, response) {
      state.registeredNodes = state.registeredNodes.concat(response.data)
      state.nodePage += 1
    },
    setNodesLoading (state, loading) {
      state.nodesLoading = loading
    },
    setRegisteredFarms (state, response) {
      // more pages to load, concat data and increase page number
      state.registeredFarms = state.registeredFarms.concat(response.data)
      state.farmPage += 1
    },
    setFarmsLoading (state, loading) {
      state.farmsLoading = loading
    },
    setRegisteredGateways (state, response) {
      // more pages to load, concat data and increase page number
      state.registeredGateways = state.registeredGateways.concat(response.data)
      state.gatewayPage += 1
    },
    setGatewaysLoading (state, loading) {
      state.gatewaysLoading = loading
    },
    setFarms (state, value) {
      state.farms = value
    },
    setNodes (state, value) {
      state.nodes = value
    },
    setUser: (state, user) => {
      state.user = user
    },
    setAmountOfFarms (state, value) {
      if (value.length === 0) {
        return
      }
      state.nodeSpecs.amountregisteredFarms += value.length
    },
    setTotalSpecs (state, nodes) {
      if (nodes.length === 0) {
        return
      }

      var onlineNodes = nodes.filter(online)

      state.nodeSpecs.amountregisteredNodes = nodes.length
      state.nodeSpecs.onlinenodes = onlineNodes.length
      state.nodeSpecs.countries = lodash.uniqBy(nodes, node => node.location.country).length
      state.nodeSpecs.cru = lodash.sumBy(onlineNodes, node => node.total_resources.cru)
      state.nodeSpecs.mru = lodash.sumBy(onlineNodes, node => node.total_resources.mru)
      state.nodeSpecs.sru = lodash.sumBy(onlineNodes, node => node.total_resources.sru)
      state.nodeSpecs.hru = lodash.sumBy(onlineNodes, node => node.total_resources.hru)
      state.nodeSpecs.network = lodash.sumBy(onlineNodes, node => node.workloads.network)
      state.nodeSpecs.volume = lodash.sumBy(onlineNodes, node => node.workloads.volume)
      state.nodeSpecs.container = lodash.sumBy(onlineNodes, node => node.workloads.container)
      state.nodeSpecs.zdb_namespace = lodash.sumBy(onlineNodes, node => node.workloads.zdb_namespace)
      state.nodeSpecs.k8s_vm = lodash.sumBy(onlineNodes, node => node.workloads.k8s_vm)
    },
    setGatewaySpecs (state, gateways) {
      if (gateways.length === 0) {
        return
      }
      var onlineGateways = gateways.filter(online)
      state.gatewaySpecs.amountRegisteredGateways += gateways.length
      state.gatewaySpecs.onlineGateways += onlineGateways.length
    },
    resetState (state) {
      // Merge rather than replace so we don't lose observers
      // https://github.com/vuejs/vuex/issues/1118
      Object.assign(state, getDefaultState())
    }
  },
  getters: {
    user: state => state.user,
    registeredNodes: state => state.registeredNodes,
    nodes: state => state.nodes,
    registeredFarms: state => state.registeredFarms,
    registeredGateways: state => state.registeredGateways,
    farms: state => state.farms,
    nodeSpecs: state => state.nodeSpecs,
    gatewaySpecs: state => state.gatewaySpecs,
    nodesLoading: state => state.nodesLoading,
    farmsLoading: state => state.farmsLoading,
    gatewaysLoading: state => state.gatewaysLoading
  }
})

function online (node) {
  const timestamp = new Date().getTime() / 1000
  const minutes = (timestamp - node.updated) / 60
  return minutes < 20
}

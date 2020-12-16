import miniGraph from '../../components/minigraph'
import capacityMap from '../../components/capacitymap'
import nodesTable from '../../components/nodestable'
import gatewaysTable from '../../components/gatewaystable'

import scrollablecard from '../../components/scrollablecard'
import { mapGetters, mapActions } from 'vuex'

export default {
  name: 'capacity',
  components: { miniGraph, capacityMap, nodesTable, scrollablecard, gatewaysTable },
  props: [],
  data () {
    return {
      showDialog: false,
      dilogTitle: 'title',
      dialogBody: '',
      dialogActions: [],
      dialogImage: null,
      block: null,
      showBadge: true,
      menu: false,
      selectedNode: '',
      selectedGateway: ''
    }
  },
  computed: {
    ...mapGetters([
      'nodeSpecs',
      'registeredNodes',
      'registeredGateways',
      'gatewaySpecs',
      'prices'
    ]),
    SuPrice: function () {
      return `$ ${this.prices.SuPriceDollarMonth}`
    },
    CuPrice: function () {
      return `$ ${this.prices.CuPriceDollarMonth}`
    },
    TftPrice: function () {
      return `$ ${this.prices.TftPriceMill / 1000}`
    },
    IP4Price: function () {
      return `$ ${this.prices.IP4uPriceDollarMonth}`
    }
  },
  mounted () {
    this.getPrices()
    this.refresh()
  },

  methods: {
    ...mapActions({
      refresh: 'refreshData',
      getPrices: 'getPrices'
    }),
    changeSelectedNode (data) {
      this.selectedNode = data
    },
    changeSelectedGateway (data) {
      this.selectedGateway = data
    }
  }
}

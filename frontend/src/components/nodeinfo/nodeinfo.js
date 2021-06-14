export default {
  name: 'nodeinfo',
  props: ['node'],
  data () {
    return {
      freeIcon: this.node.freeToUse === true ? { icon: 'fa-check', color: 'green' } : { icon: 'fa-times', color: 'red' },
      isV3Icon: this.node.isV3 === true ? { icon: 'fa-check', color: 'green' } : { icon: 'fa-times', color: 'red' }
    }
  },
  mounted () {
    console.log(this.node)
  },
  methods: {
    getPercentage (type) {
      const reservedResources = this.node.reservedResources[type]
      const totalResources = this.node.totalResources[type]
      if (reservedResources === 0 && totalResources === 0) return 0
      return (reservedResources / totalResources) * 100
    },
    openLink () {
      window.open('https://threefold.io/info/threefold#/threefold__farming_logic3', '_blank')
    }
  }
}

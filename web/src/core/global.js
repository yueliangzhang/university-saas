import config from './config'
import { h } from 'vue'

// 统一导入el-icon图标
import * as ElIconModules from '@element-plus/icons-vue'
import svgIcon from '@/components/svgIcon/svgIcon.vue'
// 导入转换图标名称的函数

const createIconComponent = (name) => ({
  name: 'SvgIcon',
  render() {
    return h(svgIcon, {
      localIcon: name
    })
  }
})

const registerIcons = (app) => {
  const iconModules = import.meta.glob(
    ['/src/assets/icons/**/*.svg', '/src/plugin/**/assets/icons/**/*.svg'],
    { eager: true, query: '?raw', import: 'default' }
  )
  const namespace = 'http://www.w3.org/2000/svg'
  const sprite = document.createElementNS(namespace, 'svg')
  sprite.setAttribute('aria-hidden', 'true')
  sprite.style.display = 'none'
  const parser = new DOMParser()
  let allKeys = []
  for (const [path, markup] of Object.entries(iconModules)) {
    let pluginName = ''
    if (path.startsWith('/src/plugin/')) {
      pluginName = `${path.split('/')[3]}-`
    }
    const iconName = path
      .split('/')
      .pop()
      .replace(/\.svg$/, '')
    // 如果iconName带空格则不加入到图标库中并且提示名称不合法
    if (iconName.indexOf(' ') !== -1) {
      console.error(`icon ${iconName}.svg includes whitespace in ${path}`)
      continue
    }
    const key = `${pluginName}${iconName}`
    const svg = parser.parseFromString(markup, 'image/svg+xml').documentElement
    if (svg.localName !== 'svg') {
      console.error(`Invalid local SVG: ${path}`)
      continue
    }
    const symbol = document.createElementNS(namespace, 'symbol')
    symbol.id = key
    for (const attribute of svg.attributes) {
      if (['id', 'width', 'height', 'xmlns', 'xmlns:xlink'].includes(attribute.name)) continue
      symbol.setAttributeNS(attribute.namespaceURI, attribute.name, attribute.value)
    }
    symbol.append(...Array.from(svg.childNodes, node => document.importNode(node, true)))
    sprite.append(symbol)
    const iconComponent = createIconComponent(key)
    config.logs.push({
      key: key,
      label: key
    })
    app.component(key, iconComponent)

    // 开发模式下列出所有 svg 图标，方便开发者直接查找复制使用
    allKeys.push(key)
  }
  document.body.prepend(sprite)

  import.meta.env.MODE == 'development' &&
    console.log(`所有可用的本地图标: ${allKeys.join(', ')}`)
}

export const register = (app) => {
  // 统一注册el-icon图标
  for (const iconName in ElIconModules) {
    app.component(iconName, ElIconModules[iconName])
  }
  app.component('SvgIcon', svgIcon)
  registerIcons(app)
  app.config.globalProperties.$GIN_VUE_ADMIN = config
}

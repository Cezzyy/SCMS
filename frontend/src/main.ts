import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { 
  faUser, 
  faLock, 
  faBars, 
  faTimes, 
  faHome, 
  faChartLine, 
  faClipboardList, 
  faUsers, 
  faCog, 
  faSignOutAlt,
  faBox,
  faFileInvoiceDollar,
  faPlus,
  faAddressBook,
  faSearch,
  faEdit,
  faSort,
  faChevronLeft,
  faChevronRight 
} from '@fortawesome/free-solid-svg-icons'

import App from './App.vue'
import router from './router'

import './assets/main.css'

/* add icons to the library */
library.add(
  faUser, 
  faLock, 
  faBars, 
  faTimes, 
  faHome, 
  faChartLine, 
  faClipboardList, 
  faUsers, 
  faCog, 
  faSignOutAlt,
  faBox,
  faFileInvoiceDollar,
  faPlus,
  faAddressBook,
  faSearch,
  faEdit,
  faSort,
  faChevronLeft,
  faChevronRight
)

const app = createApp(App)

app.component('font-awesome-icon', FontAwesomeIcon)
app.use(createPinia())
app.use(router)

app.mount('#app')

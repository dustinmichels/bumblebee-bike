import { createApp } from 'vue'
import App from './App.vue'
import MapTest from './components/MapTest.vue'

const root = window.location.pathname === '/map' ? MapTest : App
createApp(root).mount('#app')

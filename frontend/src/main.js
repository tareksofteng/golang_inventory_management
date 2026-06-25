import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import './style.css'

// Apply saved theme before mount to avoid a flash of the wrong colour.
if (localStorage.getItem('theme') === 'dark') {
  document.documentElement.classList.add('dark')
}

createApp(App).use(createPinia()).use(router).mount('#app')

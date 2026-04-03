<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import NavBar from '@/components/NavBar.vue'

const auth = useAuthStore()
const theme = useThemeStore()

onMounted(() => {
  auth.init()
  theme.init()
})
</script>

<template>
  <div class="app">
    <NavBar v-if="auth.isAuthenticated" />
    <main :class="{ 'with-nav': auth.isAuthenticated }">
      <RouterView />
    </main>
  </div>
</template>

<style>
.app {
  min-height: 100vh;
}

main {
  padding: 0;
}

main.with-nav {
  margin-left: 250px;
  padding: 0;
}

@media (max-width: 1023px) {
  main.with-nav {
    margin-left: 0;
    padding-top: 60px;
  }
}
</style>

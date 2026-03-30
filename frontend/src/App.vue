<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import NavBar from '@/components/NavBar.vue'

const auth = useAuthStore()

onMounted(() => {
  auth.init()
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
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
  background: #f5f7fa;
  color: #333;
}

.app {
  min-height: 100vh;
}

main {
  padding: 0;
}

main.with-nav {
  margin-left: 250px;
  padding: 2rem;
}
</style>

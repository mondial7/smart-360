<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div class="my-feedback-page">
    <header class="page-header">
      <h1>My Feedback</h1>
      <p>View your consolidated feedback and development insights</p>
    </header>

    <div class="feedback-list">
      <div class="empty-state">
        <p>No feedback has been shared with you yet.</p>
        <p class="subtext">When an administrator shares consolidated feedback, it will appear here.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.my-feedback-page {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}

.page-header p {
  color: #666;
}

.empty-state {
  background: white;
  padding: 3rem;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
  text-align: center;
}

.empty-state p {
  color: #666;
  margin-bottom: 0.5rem;
}

.subtext {
  font-size: 0.9rem;
  color: #999;
}
</style>

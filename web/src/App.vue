<script setup lang="ts">
import { ref, onMounted } from 'vue'

const health = ref<string | null>(null)

onMounted(async () => {
  try {
    const res = await fetch('/api/health')
    const data = await res.json() as { status: string }
    health.value = data.status
  } catch {
    health.value = 'unreachable'
  }
})
</script>

<template>
  <main>
    <h1>Bumblebee Bike</h1>
    <p>API status: <code>{{ health ?? '…' }}</code></p>
  </main>
</template>

<style>
*, *::before, *::after { box-sizing: border-box; }

body {
  margin: 0;
  font-family: system-ui, sans-serif;
  background: #0f0f0f;
  color: #e0e0e0;
  min-height: 100dvh;
  display: grid;
  place-items: center;
}

main { text-align: center; }

h1 { font-size: 2rem; margin-bottom: 1rem; }

code {
  background: #1e1e1e;
  padding: 0.15em 0.4em;
  border-radius: 4px;
  font-size: 0.9em;
}
</style>

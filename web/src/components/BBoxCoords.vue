<script setup lang="ts">
defineProps<{
  bbox: [number, number, number, number]; // [minLng, minLat, maxLng, maxLat]
}>();
</script>

<template>
  <div class="bbox-coords">
    <div class="coords-title">Bounds</div>
    <div class="coords-grid">
      <div class="coords-col">
        <div class="coords-row">
          <span class="coord-label">W:</span>
          <span class="coord-value">{{ bbox[0].toFixed(4) }}°</span>
        </div>
        <div class="coords-row">
          <span class="coord-label">E:</span>
          <span class="coord-value">{{ bbox[2].toFixed(4) }}°</span>
        </div>
      </div>
      <div class="coords-col">
        <div class="coords-row">
          <span class="coord-label">S:</span>
          <span class="coord-value">{{ bbox[1].toFixed(4) }}°</span>
        </div>
        <div class="coords-row">
          <span class="coord-label">N:</span>
          <span class="coord-value">{{ bbox[3].toFixed(4) }}°</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.bbox-coords {
  margin-top: 12px;
  padding: 8px 10px;
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 4px;
}

.coords-title {
  font-size: 0.75rem;
  font-weight: 600;
  color: #888;
  margin-bottom: 6px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.coords-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.coords-col {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.coords-row {
  display: flex;
  justify-content: space-between;
  font-family: monospace;
  font-size: 0.8rem;
}

.coord-label {
  color: #aaa;
  margin-right: 4px;
}

.coord-value {
  color: #ff9900;
  font-weight: 600;
}
</style>

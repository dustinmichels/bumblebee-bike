<script setup lang="ts">
import { ref, defineAsyncComponent } from "vue";
import SearchCity from "./SearchCity.vue";
import BBoxCoords from "./BBoxCoords.vue";

const MapView = defineAsyncComponent(() => import("./MapView.vue"));

// Default: Boston, MA
const cityName = ref("Boston, MA, USA");
const bbox = ref<[number, number, number, number]>([-71.1912, 42.2279, -70.9227, 42.3969]);
const center = ref<[number, number]>([-71.0589, 42.3601]);

const handleSelectCity = (payload: {
  name: string;
  bbox: [number, number, number, number];
  lat: number;
  lon: number;
}) => {
  cityName.value = payload.name;
  bbox.value = payload.bbox;
  center.value = [payload.lon, payload.lat];
};
</script>

<template>
  <div class="map-test-layout">
    <header class="map-test-header">
      <strong>Map Test</strong>
      <a href="/" class="back-link">← Back to app</a>
    </header>

    <div class="card-group">
      <div class="card flow-card">
        <h2>Select Bounding Box</h2>
        <p>Search for a city and adjust the bounding box corners to frame the area of interest.</p>

        <div class="control-panel">
          <label class="control-label">Search for a City</label>
          <SearchCity @select-city="handleSelectCity" />
        </div>

        <div class="city-box">
          <h4>📍 Current Area</h4>
          <div class="city-name">{{ cityName }}</div>
          <BBoxCoords :bbox="bbox" />
        </div>
      </div>

      <div class="map-container-wrapper">
        <MapView v-model:bbox="bbox" :center="center" :showBBox="true" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.map-test-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg, #0d0d0d);
}

.map-test-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.5rem;
  background: var(--surface, #1a1a1a);
  border-bottom: 1px solid var(--border, #333);
  color: var(--text, #eee);
}

.back-link {
  color: var(--accent, #ff9900);
  text-decoration: none;
  font-size: 0.9rem;
}

.back-link:hover {
  text-decoration: underline;
}

/* Reuse App.vue card-group / flow-card layout */
.card-group {
  flex: 1;
  display: flex;
  gap: 1rem;
  padding: 1rem;
  min-height: 0;
}

.card {
  background: var(--surface, #1a1a1a);
  border: 1px solid var(--border, #333);
  border-radius: 8px;
  padding: 1.5rem;
  color: var(--text, #eee);
}

.flow-card {
  width: 320px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.map-container-wrapper {
  flex: 1;
  min-height: 400px;
  border-radius: 8px;
  overflow: hidden;
}

.control-panel {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.control-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted, #aaa);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.city-box {
  background: var(--bg, #0d0d0d);
  border: 1px solid var(--border, #333);
  border-radius: 6px;
  padding: 0.75rem 1rem;
}

.city-box h4 {
  margin: 0 0 0.25rem;
  font-size: 0.8rem;
  color: var(--text-muted, #aaa);
}

.city-name {
  font-size: 0.95rem;
  font-weight: 500;
}
</style>

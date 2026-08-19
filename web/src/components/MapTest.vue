<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref } from "vue";
import {
  EMPTY_FEATURE_COLLECTION,
  bboxCenter,
  getFeatureCollectionBounds,
  type BBox,
  type GeoJSONFeatureCollection,
  type LngLat,
  type RouteLayer,
} from "../lib/activity";

const MapView = defineAsyncComponent(() => import("./MapView.vue"));

const bbox = ref<BBox>([-71.1912, 42.2279, -70.9227, 42.3969]);
const center = ref<LngLat>([-71.0589, 42.3601]);
const activitiesGeoJSON = ref<GeoJSONFeatureCollection>(EMPTY_FEATURE_COLLECTION);
const activityCount = ref<number | null>(null);
const isLoading = ref(true);
const loadError = ref<string | null>(null);
const previewRoute = computed<RouteLayer[]>(() => [
  {
    id: "map-test",
    label: "Preview",
    color: "#ff8c00",
    data: activitiesGeoJSON.value,
  },
]);

const loadMap = async () => {
  isLoading.value = true;
  loadError.value = null;

  try {
    const res = await fetch("/api/map-test");
    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || `Server returned status ${res.status}`);
    }

    const geoJSON = (await res.json()) as GeoJSONFeatureCollection;
    activitiesGeoJSON.value = geoJSON;
    activityCount.value = geoJSON.features.length;

    const bounds = getFeatureCollectionBounds(geoJSON);
    if (bounds) {
      bbox.value = bounds;
      center.value = bboxCenter(bounds);
    }
  } catch (error) {
    console.error(error);
    loadError.value = error instanceof Error ? error.message : "Unable to load map test data.";
  } finally {
    isLoading.value = false;
  }
};

onMounted(() => {
  void loadMap();
});
</script>

<template>
  <div class="map-test-layout">
    <header class="map-test-header">
      <strong>Map Test</strong>
      <a href="/" class="back-link">← Back to map tools</a>
    </header>

    <div class="card-group">
      <div class="card flow-card">
        <h2>Direct dataset preview</h2>
        <p>
          Loads <code>data/activities.parquet</code> directly and renders every activity geometry as
          map routes.
        </p>

        <div v-if="isLoading" class="status-card">
          <strong>Loading map data…</strong>
          <p>Reading the parquet dataset and preparing the preview.</p>
        </div>

        <div v-else-if="loadError" class="error-banner">⚠️ {{ loadError }}</div>

        <div v-else class="status-card">
          <strong>{{ activityCount }} activities loaded</strong>
          <p>The viewport is fitted to the full dataset.</p>
          <div class="summary-block">
            <span class="summary-label">Bounding box</span>
            <code class="summary-code"
              >{{ bbox[0].toFixed(4) }}, {{ bbox[1].toFixed(4) }}, {{ bbox[2].toFixed(4) }},
              {{ bbox[3].toFixed(4) }}</code
            >
          </div>
        </div>
      </div>

      <div class="map-container-wrapper">
        <MapView v-model:bbox="bbox" :center="center" :routes="previewRoute" :show-b-box="false" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.map-test-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #121212;
  color: #e0e0e0;
}

.map-test-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 20px 24px;
  border-bottom: 1px solid #222;
  background: #1a1a1a;
}

.back-link {
  color: #ff9900;
  text-decoration: none;
  font-size: 0.95rem;
  font-weight: 600;
}

.back-link:hover {
  text-decoration: underline;
}

.card-group {
  flex: 1;
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 20px;
  padding: 24px;
}

.card {
  background: #1a1a1a;
  border: 1px solid #2d2d2d;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.flow-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.flow-card h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #fff;
}

.flow-card p {
  margin: 0;
  line-height: 1.5;
  color: #b0b0b0;
}

.status-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px;
  border-radius: 8px;
  background: #242424;
  border: 1px solid #333;
}

.summary-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 8px;
}

.summary-label {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #888;
}

.summary-code {
  display: block;
  font-family: monospace;
  color: #ff9900;
}

.error-banner {
  background: rgba(220, 50, 50, 0.15);
  border: 1px solid #d32f2f;
  color: #ef5350;
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 14px;
  text-align: left;
}

.map-container-wrapper {
  min-height: 550px;
}

@media (max-width: 900px) {
  .card-group {
    grid-template-columns: 1fr;
  }
}
</style>

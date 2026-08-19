<script setup lang="ts">
import { defineAsyncComponent, onMounted, ref } from "vue";

const MapView = defineAsyncComponent(() => import("./MapView.vue"));

type Geometry =
  | { type: "LineString"; coordinates: [number, number][] }
  | { type: "MultiLineString"; coordinates: [number, number][][] }
  | { type: "Point"; coordinates: [number, number] }
  | { type: "MultiPoint"; coordinates: [number, number][] }
  | { type: string; coordinates?: unknown };

type GeoJSONFeature = {
  geometry?: Geometry | null;
  properties?: Record<string, unknown>;
};

type GeoJSONFeatureCollection = {
  type: "FeatureCollection";
  features: GeoJSONFeature[];
};

const bbox = ref<[number, number, number, number]>([-71.1912, 42.2279, -70.9227, 42.3969]);
const center = ref<[number, number]>([-71.0589, 42.3601]);
const activitiesGeoJSON = ref<GeoJSONFeatureCollection | null>(null);
const activityCount = ref<number | null>(null);
const isLoading = ref(true);
const loadError = ref<string | null>(null);

const includeCoordinate = (
  coords: [number, number],
  bounds: { minLng: number; minLat: number; maxLng: number; maxLat: number },
) => {
  bounds.minLng = Math.min(bounds.minLng, coords[0]);
  bounds.minLat = Math.min(bounds.minLat, coords[1]);
  bounds.maxLng = Math.max(bounds.maxLng, coords[0]);
  bounds.maxLat = Math.max(bounds.maxLat, coords[1]);
};

const updateViewport = (geoJSON: GeoJSONFeatureCollection) => {
  const bounds = {
    minLng: Number.POSITIVE_INFINITY,
    minLat: Number.POSITIVE_INFINITY,
    maxLng: Number.NEGATIVE_INFINITY,
    maxLat: Number.NEGATIVE_INFINITY,
  };

  for (const feature of geoJSON.features) {
    const geometry = feature.geometry;
    if (!geometry) {
      continue;
    }

    if (geometry.type === "LineString") {
      for (const coords of geometry.coordinates) {
        includeCoordinate(coords, bounds);
      }
      continue;
    }

    if (geometry.type === "MultiLineString") {
      for (const line of geometry.coordinates) {
        for (const coords of line) {
          includeCoordinate(coords, bounds);
        }
      }
      continue;
    }

    if (geometry.type === "Point") {
      includeCoordinate(geometry.coordinates, bounds);
      continue;
    }

    if (geometry.type === "MultiPoint") {
      for (const coords of geometry.coordinates) {
        includeCoordinate(coords, bounds);
      }
    }
  }

  if (!Number.isFinite(bounds.minLng)) {
    return;
  }

  bbox.value = [bounds.minLng, bounds.minLat, bounds.maxLng, bounds.maxLat];
  center.value = [
    (bounds.minLng + bounds.maxLng) / 2,
    (bounds.minLat + bounds.maxLat) / 2,
  ];
};

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
    updateViewport(geoJSON);
  } catch (err: any) {
    console.error(err);
    loadError.value = err.message || "Unable to load map test data.";
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
      <a href="/" class="back-link">← Back to app</a>
    </header>

    <div class="card-group">
      <div class="card flow-card">
        <h2>Direct dataset preview</h2>
        <p>
          Loads <code>data/activities.parquet</code> directly and renders every activity geometry as
          map points.
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
        <MapView
          v-if="activitiesGeoJSON"
          v-model:bbox="bbox"
          :center="center"
          :activitiesGeoJSON="activitiesGeoJSON"
          :showBBox="false"
        />
        <div v-else class="map-placeholder">
          <span v-if="isLoading">Preparing preview…</span>
          <span v-else>Map preview unavailable.</span>
        </div>
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
  background: #111;
  padding: 8px 12px;
  border-radius: 4px;
  color: #ff9900;
}

.error-banner {
  background: rgba(220, 50, 50, 0.15);
  border: 1px solid #d32f2f;
  color: #ef5350;
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 14px;
}

.map-container-wrapper {
  min-height: 550px;
}

.map-placeholder {
  width: 100%;
  height: 550px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
  border: 1px solid #2d2d2d;
  background: #1a1a1a;
  color: #888;
}

@media (max-width: 900px) {
  .card-group {
    grid-template-columns: 1fr;
  }
}
</style>

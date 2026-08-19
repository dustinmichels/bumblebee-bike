<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import SearchCity from './components/SearchCity.vue'
import MapView from './components/MapView.vue'

const health = ref<string | null>(null)

// Default location: Amsterdam
const cityName = ref('Amsterdam, Netherlands')
const bbox = ref<[number, number, number, number]>([4.728797, 52.278174, 5.079162, 52.431064])
const center = ref<[number, number]>([4.9041, 52.3676])

onMounted(async () => {
  try {
    const res = await fetch('/api/health')
    const data = await res.json() as { status: string }
    health.value = data.status
  } catch {
    health.value = 'unreachable'
  }
})

const handleSelectCity = (payload: {
  name: string
  bbox: [number, number, number, number]
  lat: number
  lon: number
}) => {
  cityName.value = payload.name
  bbox.value = payload.bbox
  center.value = [payload.lon, payload.lat]
}

// Formatting coordinates for output
const formattedGeoJSON = computed(() => {
  const [minLng, minLat, maxLng, maxLat] = bbox.value
  return `[${minLng.toFixed(6)}, ${minLat.toFixed(6)}, ${maxLng.toFixed(6)}, ${maxLat.toFixed(6)}]`
})

const formattedOSM = computed(() => {
  const [minLng, minLat, maxLng, maxLat] = bbox.value
  return `${minLat.toFixed(6)},${minLng.toFixed(6)},${maxLat.toFixed(6)},${maxLng.toFixed(6)}`
})

const formattedCSV = computed(() => {
  const [minLng, minLat, maxLng, maxLat] = bbox.value
  return `${minLng.toFixed(6)}, ${minLat.toFixed(6)}, ${maxLng.toFixed(6)}, ${maxLat.toFixed(6)}`
})

const copyStatus = ref('')
const activeCopyType = ref('')

const copyToClipboard = (text: string, type: string) => {
  navigator.clipboard.writeText(text).then(() => {
    activeCopyType.value = type
    copyStatus.value = 'Copied!'
    setTimeout(() => {
      copyStatus.value = ''
      activeCopyType.value = ''
    }, 2000)
  }).catch((err) => {
    console.error('Failed to copy: ', err)
    copyStatus.value = 'Failed to copy'
  })
}
</script>

<template>
  <div class="app-layout">
    <header class="app-header">
      <div class="header-main">
        <div class="logo">🐝</div>
        <div>
          <h1>Bumblebee Bike</h1>
          <p class="tagline">Select a city and drag the bounding box corners to export coordinates.</p>
        </div>
      </div>
      <div class="api-badge" :class="health">
        API status: <code>{{ health ?? '…' }}</code>
      </div>
    </header>

    <main class="app-content">
      <section class="control-panel">
        <label class="control-label">Search for a City</label>
        <SearchCity @select-city="handleSelectCity" />
      </section>

      <section class="map-section">
        <MapView v-model:bbox="bbox" :center="center" />
      </section>

      <section class="details-section">
        <div class="card city-info">
          <h3>📍 Selected Location</h3>
          <p class="city-name">{{ cityName }}</p>
          <div class="grid-coords">
            <div class="coord-box">
              <span class="coord-label">Min Lng (West)</span>
              <span class="coord-val">{{ bbox[0].toFixed(6) }}</span>
            </div>
            <div class="coord-box">
              <span class="coord-label">Min Lat (South)</span>
              <span class="coord-val">{{ bbox[1].toFixed(6) }}</span>
            </div>
            <div class="coord-box">
              <span class="coord-label">Max Lng (East)</span>
              <span class="coord-val">{{ bbox[2].toFixed(6) }}</span>
            </div>
            <div class="coord-box">
              <span class="coord-label">Max Lat (North)</span>
              <span class="coord-val">{{ bbox[3].toFixed(6) }}</span>
            </div>
          </div>
        </div>

        <div class="card export-formats">
          <h3>💾 Export Formats</h3>
          
          <div class="format-row">
            <div class="format-info">
              <span class="format-name">GeoJSON / Mapbox</span>
              <span class="format-desc"><code>[minLng, minLat, maxLng, maxLat]</code></span>
            </div>
            <div class="format-action">
              <input readonly :value="formattedGeoJSON" class="format-input" />
              <button 
                @click="copyToClipboard(formattedGeoJSON, 'geojson')" 
                class="copy-btn"
              >
                {{ activeCopyType === 'geojson' ? copyStatus : 'Copy' }}
              </button>
            </div>
          </div>

          <div class="format-row">
            <div class="format-info">
              <span class="format-name">OSM Overpass / Nominatim</span>
              <span class="format-desc"><code>minLat,minLng,maxLat,maxLng</code></span>
            </div>
            <div class="format-action">
              <input readonly :value="formattedOSM" class="format-input" />
              <button 
                @click="copyToClipboard(formattedOSM, 'osm')" 
                class="copy-btn"
              >
                {{ activeCopyType === 'osm' ? copyStatus : 'Copy' }}
              </button>
            </div>
          </div>

          <div class="format-row">
            <div class="format-info">
              <span class="format-name">Standard CSV</span>
              <span class="format-desc"><code>minLng, minLat, maxLng, maxLat</code></span>
            </div>
            <div class="format-action">
              <input readonly :value="formattedCSV" class="format-input" />
              <button 
                @click="copyToClipboard(formattedCSV, 'csv')" 
                class="copy-btn"
              >
                {{ activeCopyType === 'csv' ? copyStatus : 'Copy' }}
              </button>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<style>
*, *::before, *::after { box-sizing: border-box; }

body {
  margin: 0;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  background: #121212;
  color: #e0e0e0;
  min-height: 100vh;
}

code {
  background: #1e1e1e;
  padding: 0.15em 0.4em;
  border-radius: 4px;
  font-size: 0.9em;
  color: #ff9900;
}

.app-layout {
  max-width: 1000px;
  width: 100%;
  margin: 0 auto;
  padding: 24px 16px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #222;
  padding-bottom: 16px;
  flex-wrap: wrap;
  gap: 16px;
}

.header-main {
  display: flex;
  align-items: center;
  gap: 16px;
}

.logo {
  font-size: 2.5rem;
}

.app-header h1 {
  font-size: 1.8rem;
  margin: 0 0 4px 0;
  font-weight: 700;
  color: #fff;
}

.tagline {
  margin: 0;
  font-size: 0.95rem;
  color: #a0a0a0;
}

.api-badge {
  font-size: 13px;
  padding: 6px 12px;
  border-radius: 20px;
  background: #1a1a1a;
  border: 1px solid #333;
  color: #aaa;
}

.api-badge.ok code {
  color: #44bb44;
}

.api-badge.unreachable code {
  color: #ff4444;
}

.app-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.control-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.control-label {
  font-size: 14px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #ff9900;
}

.map-section {
  width: 100%;
}

.details-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

@media (max-width: 768px) {
  .details-section {
    grid-template-columns: 1fr;
  }
}

.card {
  background: #1a1a1a;
  border: 1px solid #2d2d2d;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.card h3 {
  margin-top: 0;
  margin-bottom: 16px;
  font-size: 15px;
  font-weight: 600;
  text-transform: uppercase;
  color: #ff9900;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #2d2d2d;
  padding-bottom: 10px;
}

.city-name {
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 16px 0;
  color: #fff;
}

.grid-coords {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.coord-box {
  background: #242424;
  border-radius: 6px;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  border: 1px solid #333;
}

.coord-label {
  font-size: 11px;
  color: #888;
  text-transform: uppercase;
}

.coord-val {
  font-family: monospace;
  font-size: 14px;
  color: #fff;
}

.format-row {
  margin-bottom: 16px;
}

.format-row:last-child {
  margin-bottom: 0;
}

.format-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.format-name {
  font-size: 13px;
  font-weight: 600;
  color: #e0e0e0;
}

.format-desc code {
  font-size: 11px;
  padding: 2px 4px;
  background: #111;
}

.format-action {
  display: flex;
  gap: 8px;
}

.format-input {
  flex-grow: 1;
  background: #242424;
  border: 1px solid #333;
  border-radius: 6px;
  padding: 8px 12px;
  color: #aaa;
  font-family: monospace;
  font-size: 13px;
  text-overflow: ellipsis;
}

.format-input:focus {
  outline: none;
  border-color: #ff9900;
}

.copy-btn {
  background: #ff9900;
  color: #000;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
  min-width: 75px;
  text-align: center;
}

.copy-btn:hover {
  background: #ffaa33;
}
</style>


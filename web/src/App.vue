<script setup lang="ts">
import { ref, onMounted, watch, defineAsyncComponent } from "vue";
import SearchCity from "./components/SearchCity.vue";
import BBoxCoords from "./components/BBoxCoords.vue";

const MapView = defineAsyncComponent(() => import("./components/MapView.vue"));

const health = ref<string | null>(null);
const testMode = ref(false);
const hasDefaultZip = ref(false);
const hasDefaultParquet = ref(false);
const selectedSource = ref<"parquet" | "zip" | "upload">("upload");

// Default location: Boston, MA, USA
const cityName = ref("Boston, MA, USA");
const bbox = ref<[number, number, number, number]>([-71.1912, 42.2279, -70.9227, 42.3969]);
const center = ref<[number, number]>([-71.0589, 42.3601]);

onMounted(async () => {
  try {
    const res = await fetch("/api/health");
    const data = (await res.json()) as {
      status: string;
      testMode?: boolean;
      hasDefaultZip?: boolean;
      hasDefaultParquet?: boolean;
    };
    health.value = data.status;
    if (data.testMode) {
      testMode.value = true;
      hasDefaultZip.value = !!data.hasDefaultZip;
      hasDefaultParquet.value = !!data.hasDefaultParquet;

      // Default bounding box for test mode: -71.1637, 42.2951, -71.0068, 42.3969
      bbox.value = [-71.1637, 42.2951, -71.0068, 42.3969];
      center.value = [-71.08525, 42.346];
      cityName.value = "Boston, MA (Test Area)";

      // Select default source
      if (hasDefaultParquet.value) {
        selectedSource.value = "parquet";
      } else if (hasDefaultZip.value) {
        selectedSource.value = "zip";
      } else {
        selectedSource.value = "upload";
      }
    }
  } catch {
    health.value = "unreachable";
  }
});

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

const currentStep = ref(0);
const steps = [
  { number: 1, label: "Export" },
  { number: 2, label: "Upload" },
  { number: 3, label: "Area" },
  { number: 4, label: "Process" },
  { number: 5, label: "Map" },
];

// Zip Upload state
const selectedFile = ref<File | null>(null);
const isDraggingFile = ref(false);
const isUploading = ref(false);
const uploadError = ref<string | null>(null);
const uploadSuccess = ref(false);
const sessionId = ref<string | null>(null);
const totalCount = ref<number | null>(null);
const parsedCount = ref<number | null>(null);
const rideCount = ref<number | null>(null);

// Filtering state
const isFiltering = ref(false);
const filterError = ref<string | null>(null);
const activitiesCount = ref<number | null>(null);
const activitiesGeoJSON = ref<any>(null);

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    selectedFile.value = target.files[0];
    uploadError.value = null;
    uploadSuccess.value = false;
  }
};

const handleDrop = (event: DragEvent) => {
  isDraggingFile.value = false;
  if (event.dataTransfer && event.dataTransfer.files.length > 0) {
    const file = event.dataTransfer.files[0];
    if (file.name.endsWith(".zip")) {
      selectedFile.value = file;
      uploadError.value = null;
      uploadSuccess.value = false;
    } else {
      uploadError.value = "Please select a .zip archive.";
    }
  }
};

const formatSize = (bytes: number) => {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

const submitZip = async () => {
  if (selectedSource.value === "upload" && !selectedFile.value) return;
  isUploading.value = true;
  uploadError.value = null;
  uploadSuccess.value = false;
  totalCount.value = null;
  parsedCount.value = null;
  rideCount.value = null;

  try {
    let res: Response;
    if (selectedSource.value === "parquet") {
      res = await fetch("/api/upload?useDefaultParquet=true", {
        method: "POST",
      });
    } else if (selectedSource.value === "zip") {
      res = await fetch("/api/upload?useDefaultZip=true", {
        method: "POST",
      });
    } else {
      const formData = new FormData();
      formData.append("file", selectedFile.value!);
      res = await fetch("/api/upload", {
        method: "POST",
        body: formData,
      });
    }

    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || `Server returned status ${res.status}`);
    }

    const data = (await res.json()) as {
      sessionId: string;
      total: number;
      parsed: number;
      rideCount: number;
    };
    sessionId.value = data.sessionId;
    totalCount.value = data.total;
    parsedCount.value = data.parsed;
    rideCount.value = data.rideCount;
    uploadSuccess.value = true;
  } catch (err: any) {
    console.error(err);
    uploadError.value = err.message || "An error occurred during upload.";
  } finally {
    isUploading.value = false;
  }
};

const filterActivities = async () => {
  if (!sessionId.value) return;
  isFiltering.value = true;
  filterError.value = null;
  activitiesCount.value = null;
  activitiesGeoJSON.value = null;

  try {
    const res = await fetch("/api/filter", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        sessionId: sessionId.value,
        bbox: bbox.value,
      }),
    });

    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || `Server returned status ${res.status}`);
    }

    const geoJSON = await res.json();
    activitiesGeoJSON.value = geoJSON;
    activitiesCount.value = geoJSON.features ? geoJSON.features.length : 0;
  } catch (err: any) {
    console.error(err);
    filterError.value = err.message || "An error occurred while filtering activities.";
  } finally {
    isFiltering.value = false;
  }
};

watch(currentStep, (newStep) => {
  if (newStep === 2) {
    if (testMode.value && selectedSource.value === "parquet" && !uploadSuccess.value) {
      submitZip();
    }
  }
  if (newStep === 4) {
    filterActivities();
  }
});

watch(selectedSource, (newSource) => {
  // Reset upload state when source changes
  uploadSuccess.value = false;
  sessionId.value = null;
  totalCount.value = null;
  parsedCount.value = null;
  rideCount.value = null;
  uploadError.value = null;

  // If changing back to parquet in Step 2, auto-submit
  if (newSource === "parquet" && currentStep.value === 2) {
    submitZip();
  }
});

const resetFlow = () => {
  currentStep.value = 0;
  selectedFile.value = null;
  uploadSuccess.value = false;
  sessionId.value = null;
  totalCount.value = null;
  parsedCount.value = null;
  rideCount.value = null;
  activitiesCount.value = null;
  activitiesGeoJSON.value = null;
  uploadError.value = null;
  filterError.value = null;
  if (testMode.value) {
    if (hasDefaultParquet.value) {
      selectedSource.value = "parquet";
    } else if (hasDefaultZip.value) {
      selectedSource.value = "zip";
    } else {
      selectedSource.value = "upload";
    }
  } else {
    selectedSource.value = "upload";
  }
};
</script>

<template>
  <div class="app-layout">
    <header class="app-header">
      <div class="header-main">
        <div class="logo">🐝</div>
        <div>
          <h1>Bumblebee Bike</h1>
          <p class="tagline">Create a bumblebee map using the bulk export of your Strava data!</p>
        </div>
      </div>
      <div class="api-badge" :class="health">
        API status: <code>{{ health ?? "…" }}</code>
      </div>
    </header>

    <main class="app-content">
      <!-- Status Stepper (if currentStep > 0) -->
      <div v-if="currentStep > 0" class="stepper">
        <div
          v-for="step in steps"
          :key="step.number"
          class="step-item"
          :class="{
            active: currentStep === step.number,
            completed: currentStep > step.number,
          }"
        >
          <div class="step-circle">{{ step.number }}</div>
          <span class="step-label">{{ step.label }}</span>
        </div>
      </div>

      <!-- Step 0: Homepage -->
      <div v-if="currentStep === 0" class="card hero-card text-center">
        <h2>Welcome to Bumblebee Bike!</h2>
        <p class="lead-text">
          Create a beautiful, custom map of all your cycling activities using a bulk export of your
          Strava data.
        </p>
        <div class="start-actions">
          <button @click="currentStep = 1" class="btn btn-primary btn-large">Begin</button>
        </div>
      </div>

      <!-- Step 1: Download Bulk Data -->
      <div v-else-if="currentStep === 1" class="card flow-card">
        <h2>Step 1: Download Bulk Data from Strava</h2>
        <div class="step-instructions">
          <p>
            To create your bumblebee map, you need to request a bulk export of your account data
            from Strava. Follow these steps:
          </p>
          <ol>
            <li>
              Log into
              <a href="https://www.strava.com" target="_blank" class="link">Strava.com</a> on your
              computer.
            </li>
            <li>
              Go to your settings page (hover over your profile picture →
              <strong>Settings</strong>).
            </li>
            <li>Select <strong>My Account</strong> from the menu on the left.</li>
            <li>
              Scroll down to <strong>Download or Delete Your Account</strong> and click
              <strong>Get Started</strong>.
            </li>
            <li>
              Click <strong>Request Your Archive</strong> (make sure not to request account
              deletion!).
            </li>
            <li>
              Strava will email you a link to download your ZIP file, which may take a few minutes
              or hours depending on account size.
            </li>
          </ol>
          <p>
            For more detailed information, see the official
            <a
              href="https://support.strava.com/en-us/articles/15401919-exporting-your-data-and-bulk-export"
              target="_blank"
              class="link"
              >Strava Bulk Export Guide</a
            >.
          </p>
        </div>
        <div class="card-actions">
          <button @click="currentStep = 0" class="btn btn-secondary">Back</button>
          <button @click="currentStep = 2" class="btn btn-primary">Next</button>
        </div>
      </div>

      <!-- Step 2: Upload ZIP -->
      <div v-else-if="currentStep === 2" class="card flow-card">
        <h2>Step 2: Upload bulk export ZIP</h2>
        <p>
          Select the downloaded Strava export <code>.zip</code> archive file to parse it into a
          GeoParquet file.
        </p>

        <!-- Test Mode Source Selector -->
        <div v-if="testMode" class="test-mode-selector">
          <div class="test-mode-badge">🛠️ Test Mode Active</div>
          <p class="test-mode-intro">Select your data source for testing:</p>
          <div class="test-options">
            <label
              class="test-option-card"
              :class="{ selected: selectedSource === 'parquet', disabled: !hasDefaultParquet }"
            >
              <input
                type="radio"
                name="dataSource"
                value="parquet"
                v-model="selectedSource"
                :disabled="!hasDefaultParquet"
              />
              <div class="option-details">
                <span class="option-title">Use pre-generated activities.parquet (Instant)</span>
                <span class="option-desc" v-if="hasDefaultParquet"
                  >Uses the pre-parsed dataset in <code>data/activities.parquet</code>.</span
                >
                <span class="option-desc error-text" v-else
                  >File <code>data/activities.parquet</code> not found on server.</span
                >
              </div>
            </label>

            <label
              class="test-option-card"
              :class="{ selected: selectedSource === 'zip', disabled: !hasDefaultZip }"
            >
              <input
                type="radio"
                name="dataSource"
                value="zip"
                v-model="selectedSource"
                :disabled="!hasDefaultZip"
              />
              <div class="option-details">
                <span class="option-title">Use default strava_export.zip</span>
                <span class="option-desc" v-if="hasDefaultZip"
                  >Parses the zip dataset in <code>data/strava_export.zip</code>.</span
                >
                <span class="option-desc error-text" v-else
                  >File <code>data/strava_export.zip</code> not found on server.</span
                >
              </div>
            </label>

            <label class="test-option-card" :class="{ selected: selectedSource === 'upload' }">
              <input type="radio" name="dataSource" value="upload" v-model="selectedSource" />
              <div class="option-details">
                <span class="option-title">Upload a custom ZIP</span>
                <span class="option-desc"
                  >Manually choose or drag a ZIP file from your computer.</span
                >
              </div>
            </label>
          </div>
        </div>

        <!-- Custom ZIP Upload Zone -->
        <div
          v-if="selectedSource === 'upload'"
          class="upload-zone"
          :class="{ dragging: isDraggingFile }"
          @dragover.prevent="isDraggingFile = true"
          @dragleave.prevent="isDraggingFile = false"
          @drop.prevent="handleDrop"
        >
          <input
            type="file"
            id="zip-upload"
            accept=".zip"
            @change="handleFileChange"
            class="file-input"
          />
          <label for="zip-upload" class="upload-label">
            <span class="upload-icon">📦</span>
            <span v-if="selectedFile" class="file-name"
              >{{ selectedFile.name }} ({{ formatSize(selectedFile.size) }})</span
            >
            <span v-else>Click to choose file or drag it here</span>
          </label>
        </div>

        <!-- Default ZIP / Parquet Selected Info -->
        <div v-else class="test-selected-info">
          <div v-if="selectedSource === 'parquet'" class="info-details">
            <span class="info-icon">⚡</span>
            <div>
              <strong>Pre-generated dataset selected:</strong> <code>data/activities.parquet</code>
              <p class="sub-desc" v-if="uploadSuccess">
                Successfully loaded into test session: <code>{{ sessionId }}</code
                >.
              </p>
              <p class="sub-desc" v-else-if="isUploading">Initializing test session...</p>
              <p class="sub-desc" v-else>Ready to load.</p>
            </div>
          </div>
          <div v-else-if="selectedSource === 'zip'" class="info-details">
            <span class="info-icon">📦</span>
            <div>
              <strong>Default zip archive selected:</strong> <code>data/strava_export.zip</code>
              <p class="sub-desc" v-if="uploadSuccess">
                Successfully parsed into test session: <code>{{ sessionId }}</code
                >.
              </p>
              <p class="sub-desc" v-else-if="isUploading">
                Parsing default zip archive (this will take a few seconds)...
              </p>
              <p class="sub-desc" v-else>Click <strong>Submit</strong> to parse the archive.</p>
            </div>
          </div>
        </div>

        <div v-if="uploadError" class="error-banner">⚠️ {{ uploadError }}</div>

        <div v-if="isUploading" class="progress-container">
          <div class="progress-spinner"></div>
          <span
            >Processing zip archive, parsing activities, and writing GeoParquet... This may take a
            moment.</span
          >
        </div>

        <div v-if="uploadSuccess" class="success-banner">
          <div>✅ File successfully parsed into GeoParquet! Ready to continue.</div>
          <div style="margin-top: 8px; font-weight: 500">
            Succesfully parsed {{ parsedCount }} / {{ totalCount }} activities. {{ rideCount }} are
            type = Ride.
          </div>
        </div>

        <div class="card-actions">
          <button @click="currentStep = 1" class="btn btn-secondary" :disabled="isUploading">
            Back
          </button>
          <button
            @click="submitZip"
            class="btn btn-primary"
            :disabled="
              (selectedSource === 'upload' && !selectedFile) || isUploading || uploadSuccess
            "
          >
            Submit
          </button>
          <button @click="currentStep = 3" class="btn btn-primary" :disabled="!uploadSuccess">
            Next
          </button>
        </div>
      </div>

      <!-- Step 3: Select Area -->
      <div v-else-if="currentStep === 3" class="card-group">
        <div class="card flow-card">
          <h2>Step 3: Select Bounding Box</h2>
          <p>
            Search for a city and adjust the bounding box corners to frame the area of interest.
          </p>

          <div class="control-panel">
            <label class="control-label">Search for a City</label>
            <SearchCity @select-city="handleSelectCity" />
          </div>

          <div class="city-box">
            <h4>📍 Current Area</h4>
            <div class="city-name">{{ cityName }}</div>
            <BBoxCoords :bbox="bbox" />
          </div>

          <div class="card-actions mt-auto">
            <button @click="currentStep = 2" class="btn btn-secondary">Back</button>
            <button @click="currentStep = 4" class="btn btn-primary">Next</button>
          </div>
        </div>

        <div class="map-container-wrapper">
          <MapView v-model:bbox="bbox" :center="center" :showBBox="true" />
        </div>
      </div>

      <!-- Step 4: Filter Activities (Process) -->
      <div v-else-if="currentStep === 4" class="card flow-card text-center">
        <h2>Step 4: Identifying Cycling Activities</h2>
        <p>
          Filtering activities from your GeoParquet file that intersect with the bounding box,
          restricting to cycling rides only.
        </p>

        <div v-if="isFiltering" class="processing-indicator">
          <div class="processing-ring"></div>
          <h3>Filtering with DuckDB Spatial Extension...</h3>
          <p>Running: <code>ST_Intersects(geometry, ST_MakeEnvelope(...))</code></p>
        </div>

        <div v-else-if="filterError" class="error-banner">
          ⚠️ {{ filterError }}
          <div class="mt-4">
            <button @click="filterActivities" class="btn btn-primary">Retry</button>
          </div>
        </div>

        <div v-else-if="activitiesCount !== null" class="success-banner">
          <h3>✅ Query complete!</h3>
          <p class="lead-text">
            Found <strong>{{ activitiesCount }}</strong> ride activities inside the bounding box.
          </p>
        </div>

        <div class="card-actions">
          <button @click="currentStep = 3" class="btn btn-secondary" :disabled="isFiltering">
            Back
          </button>
          <button
            @click="currentStep = 5"
            class="btn btn-primary"
            :disabled="activitiesCount === null || isFiltering"
          >
            Next
          </button>
        </div>
      </div>

      <!-- Step 5: Show Map -->
      <div v-else-if="currentStep === 5" class="card-group">
        <div class="card flow-card final-card">
          <h2>Your Bumblebee Map 🐝</h2>
          <p>
            Showing all <strong>{{ activitiesCount }}</strong> rides intersecting the bounding box
            in bright orange on a dark background.
          </p>

          <div class="export-summary">
            <h4>Location:</h4>
            <p>{{ cityName }}</p>
            <h4>Bounding Box:</h4>
            <code class="block"
              >{{ bbox[0].toFixed(4) }}, {{ bbox[1].toFixed(4) }}, {{ bbox[2].toFixed(4) }},
              {{ bbox[3].toFixed(4) }}</code
            >
          </div>

          <div class="card-actions mt-auto">
            <button @click="currentStep = 4" class="btn btn-secondary">Back</button>
            <button @click="resetFlow" class="btn btn-secondary">Start Over</button>
          </div>
        </div>

        <div class="map-container-wrapper">
          <MapView
            v-model:bbox="bbox"
            :center="center"
            :activitiesGeoJSON="activitiesGeoJSON"
            :showBBox="false"
          />
        </div>
      </div>
    </main>
  </div>
</template>
<style>
*,
*::before,
*::after {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family:
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    Roboto,
    Oxygen,
    Ubuntu,
    Cantarell,
    sans-serif;
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
  max-width: 1100px;
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
  gap: 20px;
}

.control-panel {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.control-label {
  font-size: 13px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #ff9900;
}

/* Stepper Styles */
.stepper {
  display: flex;
  justify-content: space-between;
  background: #1a1a1a;
  border: 1px solid #2d2d2d;
  padding: 16px 24px;
  border-radius: 12px;
  margin-bottom: 8px;
  gap: 8px;
  flex-wrap: wrap;
}
.step-item {
  display: flex;
  align-items: center;
  gap: 8px;
  opacity: 0.4;
  transition: opacity 0.2s;
}
.step-item.active {
  opacity: 1;
  color: #ff9900;
}
.step-item.completed {
  opacity: 0.8;
  color: #44bb44;
}
.step-circle {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: #333;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
}
.step-item.active .step-circle {
  background: #ff9900;
  color: #000;
}
.step-item.completed .step-circle {
  background: #44bb44;
  color: #000;
}
.step-label {
  font-size: 13px;
  font-weight: 600;
}
@media (max-width: 600px) {
  .step-label {
    display: none;
  }
}

/* Button Styles */
.btn {
  padding: 10px 20px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition:
    background 0.15s,
    border-color 0.15s,
    opacity 0.15s;
  border: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.btn-primary {
  background: #ff9900;
  color: #000;
}
.btn-primary:hover:not(:disabled) {
  background: #ffaa33;
}
.btn-secondary {
  background: #2a2a2a;
  color: #fff;
  border: 1px solid #444;
}
.btn-secondary:hover:not(:disabled) {
  background: #3a3a3a;
  border-color: #555;
}
.btn-large {
  padding: 14px 28px;
  font-size: 16px;
}

/* Card Styles */
.card {
  background: #1a1a1a;
  border: 1px solid #2d2d2d;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}
.card h2 {
  margin-top: 0;
  margin-bottom: 12px;
  font-size: 1.5rem;
  color: #fff;
}
.card p {
  margin-top: 0;
  color: #b0b0b0;
  line-height: 1.5;
}

.card-group {
  display: grid;
  grid-template-columns: 380px 1fr;
  gap: 20px;
  min-height: 550px;
}
@media (max-width: 900px) {
  .card-group {
    grid-template-columns: 1fr;
  }
}
.map-container-wrapper {
  width: 100%;
  height: 550px;
}
.card-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}
.mt-auto {
  margin-top: auto;
}
.text-center {
  text-align: center;
}
.hero-card {
  padding: 60px 40px;
}
.hero-card h2 {
  font-size: 2.2rem;
  margin-bottom: 16px;
  color: #fff;
}
.lead-text {
  font-size: 1.15rem;
  color: #aaa;
  max-width: 600px;
  margin: 0 auto 30px auto;
  line-height: 1.6;
}

/* Upload Zone */
.upload-zone {
  border: 2px dashed #444;
  border-radius: 8px;
  padding: 50px 20px;
  text-align: center;
  cursor: pointer;
  background: #1e1e1e;
  transition:
    border-color 0.15s,
    background 0.15s;
  position: relative;
  margin-top: 16px;
}
.upload-zone:hover,
.upload-zone.dragging {
  border-color: #ff9900;
  background: #222;
}
.file-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
  width: 100%;
}
.upload-label {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: #aaa;
  font-size: 14px;
  cursor: pointer;
}
.upload-icon {
  font-size: 32px;
}
.file-name {
  color: #ff9900;
  font-weight: 600;
}

/* Status / Banner Styles */
.error-banner {
  background: rgba(220, 50, 50, 0.15);
  border: 1px solid #d32f2f;
  color: #ef5350;
  padding: 12px 16px;
  border-radius: 6px;
  margin-top: 16px;
  font-size: 14px;
  text-align: left;
}
.success-banner {
  background: rgba(50, 200, 50, 0.1);
  border: 1px solid #388e3c;
  color: #81c784;
  padding: 16px;
  border-radius: 8px;
  margin-top: 16px;
  font-size: 14px;
  text-align: left;
}
.progress-container {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 16px;
  background: #1a1a1a;
  border: 1px solid #333;
  padding: 16px;
  border-radius: 8px;
  color: #aaa;
  font-size: 13px;
}
.progress-spinner {
  width: 20px;
  height: 20px;
  border: 3px solid #333;
  border-top-color: #ff9900;
  border-radius: 50%;
  animation: spin-spin 1s linear infinite;
  flex-shrink: 0;
}

@keyframes spin-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* Processing indicator (duckdb) */
.processing-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 40px 20px;
}
.processing-ring {
  width: 60px;
  height: 60px;
  border: 5px solid #2a2a2a;
  border-top-color: #ff9900;
  border-radius: 50%;
  animation: spin-spin 1s linear infinite;
  margin-bottom: 20px;
}

.step-instructions ol {
  padding-left: 20px;
  line-height: 1.6;
  color: #b0b0b0;
}
.step-instructions li {
  margin-bottom: 8px;
}
.link {
  color: #ff9900;
  text-decoration: none;
  font-weight: 600;
}
.link:hover {
  text-decoration: underline;
}
.city-box {
  background: #242424;
  border: 1px solid #333;
  padding: 12px 16px;
  border-radius: 6px;
  margin-top: 20px;
}
.city-box h4 {
  margin: 0 0 6px 0;
  color: #888;
  font-size: 11px;
  text-transform: uppercase;
}
.city-name {
  font-size: 16px;
  font-weight: 600;
  color: #fff;
}
.export-summary {
  margin: 20px 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.export-summary h4 {
  margin: 0;
  color: #888;
  font-size: 11px;
  text-transform: uppercase;
}
.export-summary p {
  margin: 0;
  color: #fff;
  font-size: 15px;
}
.block {
  display: block;
  font-family: monospace;
  background: #111;
  padding: 8px 12px;
  border-radius: 4px;
}
.flow-card {
  display: flex;
  flex-direction: column;
}
.final-card {
  display: flex;
  flex-direction: column;
}
.test-mode-selector {
  background: rgba(245, 158, 11, 0.05);
  border: 1px solid rgba(245, 158, 11, 0.2);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 20px;
}
.test-mode-badge {
  display: inline-block;
  background: #f59e0b;
  color: #000;
  font-weight: bold;
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.test-mode-intro {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #ccc;
  text-align: left;
}
.test-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
  text-align: left;
}
.test-option-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.test-option-card:hover:not(.disabled) {
  border-color: #555;
  background: #252525;
}
.test-option-card.selected {
  border-color: #f59e0b;
  background: rgba(245, 158, 11, 0.08);
}
.test-option-card.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.test-option-card input[type="radio"] {
  margin-top: 4px;
}
.option-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.option-title {
  font-weight: 600;
  color: #fff;
  font-size: 14px;
}
.option-desc {
  font-size: 12px;
  color: #aaa;
}
.option-desc code {
  background: rgba(255, 255, 255, 0.1);
  padding: 1px 4px;
  border-radius: 3px;
  font-family: monospace;
}
.error-text {
  color: #ef4444;
}
.test-selected-info {
  background: #1a1a1a;
  border: 1px dashed #444;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 20px;
  text-align: left;
}
.info-details {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  color: #ddd;
}
.info-icon {
  font-size: 24px;
}
.sub-desc {
  margin: 4px 0 0 0;
  font-size: 13px;
  color: #aaa;
}
</style>

<script setup lang="ts">
import { computed, defineAsyncComponent, ref, watch } from "vue";
import AreaSelectionCard from "./AreaSelectionCard.vue";
import DatasetUploadCard from "./DatasetUploadCard.vue";
import FlowStepper from "./FlowStepper.vue";
import { useActivityDataset } from "../../composables/useActivityDataset";
import {
  DEFAULT_BOSTON_BBOX,
  DEFAULT_BOSTON_CENTER,
  type BBox,
  type LngLat,
  type RouteLayer,
  type SelectedCity,
} from "../../lib/activity";

const MapView = defineAsyncComponent(() => import("../MapView.vue"));

const steps = [
  { number: 1, label: "Export" },
  { number: 2, label: "Upload" },
  { number: 3, label: "Area" },
  { number: 4, label: "Process" },
  { number: 5, label: "Map" },
];

const currentStep = ref(1);
const cityName = ref("Boston, MA, USA");
const bbox = ref<BBox>([...DEFAULT_BOSTON_BBOX]);
const center = ref<LngLat>([...DEFAULT_BOSTON_CENTER]);
const dataset = useActivityDataset();

const lightningRoutes = computed<RouteLayer[]>(() => [
  {
    id: "lightning-map",
    label: "Lightning Map",
    color: "#ff8c00",
    data: dataset.activitiesGeoJSON.value,
  },
]);

const handleSelectCity = (payload: SelectedCity) => {
  cityName.value = payload.name;
  bbox.value = payload.bbox;
  center.value = [payload.lon, payload.lat];
};

watch(currentStep, (step) => {
  if (step === 4 && dataset.readyToFilter.value) {
    void dataset.filterActivities(bbox.value);
  }
});

const resetFlow = () => {
  currentStep.value = 1;
  cityName.value = "Boston, MA, USA";
  bbox.value = [...DEFAULT_BOSTON_BBOX];
  center.value = [...DEFAULT_BOSTON_CENTER];
  dataset.reset();
};
</script>

<template>
  <section class="flow-layout">
    <FlowStepper :current-step="currentStep" :steps="steps" />

    <div v-if="currentStep === 1" class="card flow-card">
      <h2>Step 1: Download your Strava bulk export</h2>
      <div class="step-instructions">
        <p>
          Lightning Map starts from the same archive flow you already had. Request your Strava bulk
          export, then bring the downloaded ZIP back here.
        </p>
        <ol>
          <li>
            Log into <a href="https://www.strava.com" target="_blank" class="link">Strava</a>
            on your computer.
          </li>
          <li>Open <strong>Settings</strong>, then choose <strong>My Account</strong>.</li>
          <li>
            Scroll to <strong>Download or Delete Your Account</strong> and choose
            <strong>Get Started</strong>.
          </li>
          <li>Click <strong>Request Your Archive</strong>.</li>
          <li>Download the ZIP link from the email Strava sends you.</li>
        </ol>
        <p>
          Need the original guide?
          <a
            href="https://support.strava.com/en-us/articles/15401919-exporting-your-data-and-bulk-export"
            target="_blank"
            class="link"
            >Read Strava's bulk export instructions</a
          >.
        </p>
      </div>
      <div class="card-actions">
        <button class="btn btn-primary" @click="currentStep = 2">Next</button>
      </div>
    </div>

    <div v-else-if="currentStep === 2" class="flow-layout">
      <DatasetUploadCard
        title="Step 2: Upload bulk export ZIP"
        description="Select the downloaded Strava export archive. The server parses the activities into GeoParquet before the map query runs."
        :selected-file="dataset.selectedFile.value"
        :upload-error="dataset.uploadError.value"
        :is-uploading="dataset.isUploading.value"
        :upload-success="dataset.uploadSuccess.value"
        :total-count="dataset.totalCount.value"
        :parsed-count="dataset.parsedCount.value"
        :ride-count="dataset.rideCount.value"
        @select-file="dataset.setSelectedFile"
        @upload="dataset.submitZip"
      />
      <div class="card-actions">
        <button class="btn btn-secondary" :disabled="dataset.isUploading.value" @click="currentStep = 1">
          Back
        </button>
        <button class="btn btn-primary" :disabled="!dataset.uploadSuccess.value" @click="currentStep = 3">
          Next
        </button>
      </div>
    </div>

    <AreaSelectionCard
      v-else-if="currentStep === 3"
      title="Step 3: Pick the map area"
      description="Search for a city and drag the corners to frame the routes you want in the final map."
      :city-name="cityName"
      :bbox="bbox"
      @back="currentStep = 2"
      @next="currentStep = 4"
      @select-city="handleSelectCity"
    >
      <MapView v-model:bbox="bbox" :center="center" :show-b-box="true" :routes="[]" />
    </AreaSelectionCard>

    <div v-else-if="currentStep === 4" class="card flow-card text-center">
      <h2>Step 4: Build the lightning map</h2>
      <p>
        Filtering the uploaded activities against the selected bounding box and keeping only rides.
      </p>

      <div v-if="dataset.isFiltering.value" class="processing-indicator">
        <div class="processing-ring"></div>
        <h3>Running the route filter…</h3>
        <p>Querying the uploaded GeoParquet data inside your selected area.</p>
      </div>

      <div v-else-if="dataset.filterError.value" class="error-banner">
        ⚠️ {{ dataset.filterError.value }}
        <div class="mt-4">
          <button class="btn btn-primary" @click="dataset.filterActivities(bbox)">Retry</button>
        </div>
      </div>

      <div v-else-if="dataset.activitiesCount.value !== null" class="success-banner centered-banner">
        <h3>Routes ready</h3>
        <p class="lead-text compact-lead">
          Found <strong>{{ dataset.activitiesCount.value }}</strong> ride activities in
          <strong>{{ cityName }}</strong>.
        </p>
      </div>

      <div class="card-actions">
        <button class="btn btn-secondary" :disabled="dataset.isFiltering.value" @click="currentStep = 3">
          Back
        </button>
        <button
          class="btn btn-primary"
          :disabled="dataset.activitiesCount.value === null || dataset.isFiltering.value"
          @click="currentStep = 5"
        >
          Next
        </button>
      </div>
    </div>

    <div v-else class="card-group">
      <section class="card flow-card final-card">
        <h2>Lightning Map</h2>
        <p>
          Showing <strong>{{ dataset.activitiesCount.value }}</strong> rides intersecting
          <strong>{{ cityName }}</strong> in a single orange route layer.
        </p>

        <div class="export-summary">
          <h4>Location</h4>
          <p>{{ cityName }}</p>
          <h4>Bounding Box</h4>
          <code class="block"
            >{{ bbox[0].toFixed(4) }}, {{ bbox[1].toFixed(4) }}, {{ bbox[2].toFixed(4) }},
            {{ bbox[3].toFixed(4) }}</code
          >
        </div>

        <div class="card-actions mt-auto">
          <button class="btn btn-secondary" @click="currentStep = 4">Back</button>
          <button class="btn btn-secondary" @click="resetFlow">Start Over</button>
        </div>
      </section>

      <div class="map-container-wrapper">
        <MapView v-model:bbox="bbox" :center="center" :show-b-box="false" :routes="lightningRoutes" />
      </div>
    </div>
  </section>
</template>

<style scoped>
.flow-layout {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.centered-banner {
  text-align: center;
}

.compact-lead {
  margin-bottom: 0;
}
</style>

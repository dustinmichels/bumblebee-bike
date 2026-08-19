<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, ref, watch } from "vue";
import AreaSelectionCard from "./AreaSelectionCard.vue";
import DatasetUploadCard from "./DatasetUploadCard.vue";
import FlowStepper from "./FlowStepper.vue";
import UploadedDatasetList from "../uploads/UploadedDatasetList.vue";
import { useActivityDataset } from "../../composables/useActivityDataset";
import { useUploadedDatasets } from "../../composables/useUploadedDatasets";
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
  { number: 1, label: "Upload" },
  { number: 2, label: "Area" },
  { number: 3, label: "Process" },
  { number: 4, label: "Map" },
];

const currentStep = ref(1);
const cityName = ref("Boston, MA, USA");
const bbox = ref<BBox>([...DEFAULT_BOSTON_BBOX]);
const center = ref<LngLat>([...DEFAULT_BOSTON_CENTER]);
const personOneColor = ref("#ff8c00");
const personTwoColor = ref("#2563eb");
const personOne = useActivityDataset();
const personTwo = useActivityDataset();
const uploadLibrary = useUploadedDatasets();

const compareRoutes = computed<RouteLayer[]>(() => [
  {
    id: "person-one",
    label: "Person 1",
    color: personOneColor.value,
    data: personOne.activitiesGeoJSON.value,
  },
  {
    id: "person-two",
    label: "Person 2",
    color: personTwoColor.value,
    data: personTwo.activitiesGeoJSON.value,
  },
]);

const totalComparedActivities = computed(
  () => (personOne.activitiesCount.value ?? 0) + (personTwo.activitiesCount.value ?? 0),
);

const hasBothUploads = computed(
  () => personOne.uploadSuccess.value && personTwo.uploadSuccess.value,
);
const isFiltering = computed(() => personOne.isFiltering.value || personTwo.isFiltering.value);
const hasResults = computed(
  () => personOne.activitiesCount.value !== null && personTwo.activitiesCount.value !== null,
);
const filterErrors = computed(() =>
  [personOne.filterError.value, personTwo.filterError.value].filter((message): message is string =>
    Boolean(message),
  ),
);

const handleSelectCity = (payload: SelectedCity) => {
  cityName.value = payload.name;
  bbox.value = payload.bbox;
  center.value = [payload.lon, payload.lat];
};

const submitPersonOneArchive = async () => {
  const upload = await personOne.submitZip();
  if (upload) {
    await uploadLibrary.loadUploads();
  }
};

const submitPersonTwoArchive = async () => {
  const upload = await personTwo.submitZip();
  if (upload) {
    await uploadLibrary.loadUploads();
  }
};

const runCompare = async () => {
  await Promise.all([
    personOne.filterActivities(bbox.value),
    personTwo.filterActivities(bbox.value),
  ]);
};

watch(currentStep, (step) => {
  if (step === 3 && hasBothUploads.value) {
    void runCompare();
  }
});

onMounted(() => {
  void uploadLibrary.loadUploads();
});

const resetFlow = () => {
  currentStep.value = 1;
  cityName.value = "Boston, MA, USA";
  bbox.value = [...DEFAULT_BOSTON_BBOX];
  center.value = [...DEFAULT_BOSTON_CENTER];
  personOneColor.value = "#ff8c00";
  personTwoColor.value = "#2563eb";
  personOne.reset();
  personTwo.reset();
};
</script>

<template>
  <section class="flow-layout">
    <FlowStepper :current-step="currentStep" :steps="steps" />

    <div v-if="currentStep === 1" class="flow-layout">
      <div class="compare-upload-grid">
        <DatasetUploadCard
          title="Person 1"
          description="Reuse a saved GeoParquet file or upload the first Strava bulk export ZIP."
          :selected-file="personOne.selectedFile.value"
          :upload-error="personOne.uploadError.value"
          :is-uploading="personOne.isUploading.value"
          :upload-success="personOne.uploadSuccess.value"
          :total-count="personOne.totalCount.value"
          :parsed-count="personOne.parsedCount.value"
          :ride-count="personOne.rideCount.value"
          :active-dataset-name="personOne.activeDataset.value?.displayName ?? null"
          :using-existing-dataset="personOne.usingExistingDataset.value"
          :color="personOneColor"
          color-label="Route color"
          :show-color-picker="true"
          @select-file="personOne.setSelectedFile"
          @upload="submitPersonOneArchive"
          @update-color="personOneColor = $event"
        >
          <template #sourceSelection>
            <UploadedDatasetList
              v-if="uploadLibrary.uploads.value.length"
              title="Saved uploads"
              description="Select a GeoParquet file you already processed, or upload a fresh archive below."
              :uploads="uploadLibrary.uploads.value"
              :selected-dataset-id="personOne.activeDataset.value?.datasetId ?? null"
              :selectable="true"
              action-label="Use saved upload"
              @select="personOne.useExistingDataset"
            />
          </template>
        </DatasetUploadCard>
        <DatasetUploadCard
          title="Person 2"
          description="Reuse another saved GeoParquet file or upload the second Strava bulk export ZIP."
          :selected-file="personTwo.selectedFile.value"
          :upload-error="personTwo.uploadError.value"
          :is-uploading="personTwo.isUploading.value"
          :upload-success="personTwo.uploadSuccess.value"
          :total-count="personTwo.totalCount.value"
          :parsed-count="personTwo.parsedCount.value"
          :ride-count="personTwo.rideCount.value"
          :active-dataset-name="personTwo.activeDataset.value?.displayName ?? null"
          :using-existing-dataset="personTwo.usingExistingDataset.value"
          :color="personTwoColor"
          color-label="Route color"
          :show-color-picker="true"
          @select-file="personTwo.setSelectedFile"
          @upload="submitPersonTwoArchive"
          @update-color="personTwoColor = $event"
        >
          <template #sourceSelection>
            <UploadedDatasetList
              v-if="uploadLibrary.uploads.value.length"
              title="Saved uploads"
              description="Pick an existing GeoParquet file, or upload another ZIP below."
              :uploads="uploadLibrary.uploads.value"
              :selected-dataset-id="personTwo.activeDataset.value?.datasetId ?? null"
              :selectable="true"
              :show-manage-link="true"
              action-label="Use saved upload"
              @select="personTwo.useExistingDataset"
            />
          </template>
        </DatasetUploadCard>
      </div>

      <div v-if="uploadLibrary.error.value" class="error-banner">
        ⚠️ {{ uploadLibrary.error.value }}
      </div>

      <div class="card compare-hint">
        <h2>Compare two people on one map</h2>
        <p>
          Choose both uploads first. After that you'll select one shared area and overlay both ride
          collections on the same basemap.
        </p>
      </div>

      <div class="card-actions">
        <button class="btn btn-primary" :disabled="!hasBothUploads" @click="currentStep = 2">
          Next
        </button>
      </div>
    </div>

    <AreaSelectionCard
      v-else-if="currentStep === 2"
      title="Step 2: Pick one shared area"
      description="Search for the place you want to compare, then drag the bounding box to frame both riders together."
      :city-name="cityName"
      :bbox="bbox"
      @back="currentStep = 1"
      @next="currentStep = 3"
      @select-city="handleSelectCity"
    >
      <MapView v-model:bbox="bbox" :center="center" :show-b-box="true" :routes="[]" />
    </AreaSelectionCard>

    <div v-else-if="currentStep === 3" class="card flow-card text-center">
      <h2>Step 3: Build the comparison map</h2>
      <p>Filtering both uploaded datasets against the same bounding box.</p>

      <div v-if="isFiltering" class="processing-indicator">
        <div class="processing-ring"></div>
        <h3>Processing both riders…</h3>
        <p>Running the route query for Person 1 and Person 2 at the same time.</p>
      </div>

      <div v-else-if="filterErrors.length" class="error-stack">
        <div v-for="message in filterErrors" :key="message" class="error-banner">
          ⚠️ {{ message }}
        </div>
        <div class="mt-4 retry-actions">
          <button class="btn btn-primary" @click="runCompare">Retry both</button>
        </div>
      </div>

      <div v-else-if="hasResults" class="success-banner centered-banner">
        <h3>Comparison ready</h3>
        <p class="lead-text compact-lead">
          Found <strong>{{ totalComparedActivities }}</strong> total rides inside
          <strong>{{ cityName }}</strong
          >.
        </p>
      </div>

      <div class="card-actions">
        <button class="btn btn-secondary" :disabled="isFiltering" @click="currentStep = 2">
          Back
        </button>
        <button
          class="btn btn-primary"
          :disabled="!hasResults || isFiltering"
          @click="currentStep = 4"
        >
          Next
        </button>
      </div>
    </div>

    <div v-else class="card-group compare-results-layout">
      <section class="card flow-card final-card">
        <h2>Compare</h2>
        <p>
          Overlaying <strong>{{ totalComparedActivities }}</strong> rides from both uploaded exports
          inside <strong>{{ cityName }}</strong
          >.
        </p>

        <div class="compare-legend">
          <div class="legend-row">
            <span class="legend-swatch" :style="{ backgroundColor: personOneColor }"></span>
            <div class="legend-copy">
              <strong>Person 1</strong>
              <span>{{ personOne.activitiesCount.value }} rides</span>
            </div>
            <input
              v-model="personOneColor"
              type="color"
              class="legend-picker"
              aria-label="Person 1 color"
            />
          </div>
          <div class="legend-row">
            <span class="legend-swatch" :style="{ backgroundColor: personTwoColor }"></span>
            <div class="legend-copy">
              <strong>Person 2</strong>
              <span>{{ personTwo.activitiesCount.value }} rides</span>
            </div>
            <input
              v-model="personTwoColor"
              type="color"
              class="legend-picker"
              aria-label="Person 2 color"
            />
          </div>
        </div>

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
          <button class="btn btn-secondary" @click="currentStep = 3">Back</button>
          <button class="btn btn-secondary" @click="resetFlow">Start Over</button>
        </div>
      </section>

      <div class="map-container-wrapper">
        <MapView v-model:bbox="bbox" :center="center" :show-b-box="false" :routes="compareRoutes" />
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

.compare-upload-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

.compare-hint {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.compare-hint h2 {
  margin: 0;
}

.compare-hint p {
  margin: 0;
}

.error-stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.retry-actions {
  display: flex;
  justify-content: center;
}

.compare-results-layout {
  align-items: stretch;
}

.compare-legend {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 20px;
}

.legend-row {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 12px;
  align-items: center;
  padding: 12px 14px;
  border-radius: 10px;
  background: #242424;
  border: 1px solid #333;
}

.legend-swatch {
  width: 16px;
  height: 16px;
  border-radius: 999px;
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.08);
}

.legend-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
  color: #d0d0d0;
}

.legend-copy strong {
  color: #fff;
}

.legend-copy span {
  font-size: 0.9rem;
  color: #a0a0a0;
}

.legend-picker {
  width: 46px;
  height: 34px;
  border: 1px solid #444;
  border-radius: 8px;
  background: #111;
  padding: 4px;
}

.centered-banner {
  text-align: center;
}

.compact-lead {
  margin-bottom: 0;
}

@media (max-width: 900px) {
  .compare-upload-grid {
    grid-template-columns: 1fr;
  }
}
</style>

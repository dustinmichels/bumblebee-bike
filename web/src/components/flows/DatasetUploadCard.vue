<script setup lang="ts">
import { ref, useSlots } from "vue";
import { formatFileSize } from "../../lib/activity";

const props = withDefaults(
  defineProps<{
    title: string;
    description: string;
    selectedFile: File | null;
    uploadError: string | null;
    isUploading: boolean;
    uploadSuccess: boolean;
    totalCount: number | null;
    parsedCount: number | null;
    rideCount: number | null;
    activeDatasetName?: string | null;
    usingExistingDataset?: boolean;
    color?: string;
    colorLabel?: string;
    showColorPicker?: boolean;
    uploadLabel?: string;
  }>(),
  {
    activeDatasetName: null,
    usingExistingDataset: false,
    color: "#ff9900",
    colorLabel: "Route color",
    showColorPicker: false,
    uploadLabel: "Upload archive",
  },
);

const emit = defineEmits<{
  selectFile: [file: File | null];
  upload: [];
  updateColor: [color: string];
}>();

const slots = useSlots();
const isDragging = ref(false);

const fileListChanged = (event: Event) => {
  const target = event.target as HTMLInputElement;
  emit("selectFile", target.files?.[0] ?? null);
};

const droppedFile = (event: DragEvent) => {
  isDragging.value = false;
  emit("selectFile", event.dataTransfer?.files?.[0] ?? null);
};
</script>

<template>
  <section class="card upload-card">
    <div class="upload-card-head">
      <div>
        <h3>{{ title }}</h3>
        <p>{{ description }}</p>
      </div>
      <label v-if="showColorPicker" class="color-control">
        <span>{{ colorLabel }}</span>
        <input
          :value="color"
          type="color"
          @input="emit('updateColor', ($event.target as HTMLInputElement).value)"
        />
      </label>
    </div>

    <div v-if="slots.sourceSelection" class="source-selection">
      <slot name="sourceSelection" />
    </div>

    <div
      class="upload-zone"
      :class="{ dragging: isDragging }"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="droppedFile"
    >
      <input type="file" accept=".zip" class="file-input" @change="fileListChanged" />
      <label class="upload-label">
        <span class="upload-icon">📦</span>
        <span v-if="selectedFile" class="file-name"
          >{{ selectedFile.name }} ({{ formatFileSize(selectedFile.size) }})</span
        >
        <span v-else>Click to choose file or drag it here</span>
      </label>
    </div>

    <div v-if="uploadError" class="error-banner">⚠️ {{ uploadError }}</div>

    <div v-if="isUploading" class="progress-container">
      <div class="progress-spinner"></div>
      <span>Processing zip archive, parsing activities, and writing GeoParquet...</span>
    </div>

    <div v-if="uploadSuccess" class="success-banner upload-success">
      <strong>{{ usingExistingDataset ? "Saved upload selected." : "Archive ready." }}</strong>
      <span v-if="usingExistingDataset">
        Using {{ activeDatasetName ?? "the selected GeoParquet file" }} from your local upload
        library.
      </span>
      <span v-else>
        Successfully parsed {{ parsedCount }} / {{ totalCount }} activities. {{ rideCount }} are
        rides.
      </span>
    </div>

    <button
      class="btn btn-primary"
      :disabled="!selectedFile || isUploading || uploadSuccess"
      @click="emit('upload')"
    >
      {{ uploadLabel }}
    </button>
  </section>
</template>

<style scoped>
.upload-card {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.upload-card-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.upload-card-head h3 {
  margin: 0 0 8px;
  color: #fff;
}

.upload-card-head p {
  margin: 0;
  color: #b0b0b0;
  line-height: 1.5;
}

.source-selection {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.color-control {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #888;
}

.color-control input {
  width: 52px;
  height: 36px;
  border: 1px solid #444;
  border-radius: 8px;
  background: #111;
  padding: 4px;
}

.upload-success {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

@media (max-width: 700px) {
  .upload-card-head {
    flex-direction: column;
  }
}
</style>

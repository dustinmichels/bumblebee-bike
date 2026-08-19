<script setup lang="ts">
import { computed, ref } from "vue";
import { formatFileSize } from "../../lib/activity";

const props = defineProps<{
  selectedFiles: File[];
  isUploading: boolean;
  uploadError: string | null;
  uploadSummary: string | null;
}>();

const emit = defineEmits<{
  selectFiles: [files: File[]];
  upload: [];
}>();

const isDragging = ref(false);

const selectedFilesLabel = computed(() => {
  if (!props.selectedFiles.length) {
    return "Click to choose one or more ZIP files, or drag them here.";
  }

  if (props.selectedFiles.length === 1) {
    const [file] = props.selectedFiles;
    return `${file.name} (${formatFileSize(file.size)})`;
  }

  return `${props.selectedFiles.length} ZIP files selected`;
});

const selectFiles = (files: FileList | null) => {
  emit("selectFiles", files ? Array.from(files) : []);
};

const fileListChanged = (event: Event) => {
  const target = event.target as HTMLInputElement;
  selectFiles(target.files);
};

const droppedFiles = (event: DragEvent) => {
  isDragging.value = false;
  selectFiles(event.dataTransfer?.files ?? null);
};
</script>

<template>
  <section class="card upload-card">
    <div class="upload-card-head">
      <div>
        <h2>Upload new Strava exports</h2>
        <p>
          Choose one or more bulk export ZIP files. Each upload gets processed into GeoParquet and
          saved locally in your upload library.
        </p>
      </div>
    </div>

    <div
      class="upload-zone"
      :class="{ dragging: isDragging }"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="droppedFiles"
    >
      <input type="file" accept=".zip" multiple class="file-input" @change="fileListChanged" />
      <label class="upload-label">
        <span class="upload-icon">📦</span>
        <span class="file-name">{{ selectedFilesLabel }}</span>
      </label>
    </div>

    <ul v-if="selectedFiles.length > 1" class="selected-file-list">
      <li v-for="file in selectedFiles" :key="`${file.name}-${file.size}`">
        <span>{{ file.name }}</span>
        <span>{{ formatFileSize(file.size) }}</span>
      </li>
    </ul>

    <div v-if="uploadError" class="error-banner">⚠️ {{ uploadError }}</div>

    <div v-if="isUploading" class="progress-container">
      <div class="progress-spinner"></div>
      <span>Processing ZIP archives, parsing activities, and writing GeoParquet files...</span>
    </div>

    <div v-if="uploadSummary" class="success-banner upload-success">
      <strong>Upload complete.</strong>
      <span>{{ uploadSummary }}</span>
    </div>

    <button
      class="btn btn-primary"
      :disabled="!selectedFiles.length || isUploading"
      @click="emit('upload')"
    >
      Upload selected archives
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

.upload-card-head h2 {
  margin: 0 0 8px;
  color: #fff;
}

.upload-card-head p {
  margin: 0;
  color: #b0b0b0;
  line-height: 1.5;
}

.selected-file-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.selected-file-list li {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 10px;
  background: #242424;
  border: 1px solid #333;
  color: #d0d0d0;
}

.upload-success {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
</style>

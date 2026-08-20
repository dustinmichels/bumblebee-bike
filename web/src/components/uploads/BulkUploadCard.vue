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
    return "Choose one or more ZIPs, or drop them here.";
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
      <div class="upload-card-copy">
        <h2>Process Strava ZIPs</h2>
        <p>Select one or more bulk export ZIPs and save them as reusable GeoParquet datasets.</p>
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
      <span>Processing ZIPs and writing saved GeoParquet datasets…</span>
    </div>

    <div v-if="uploadSummary" class="success-banner upload-success">
      <strong>Done.</strong>
      <span>{{ uploadSummary }}</span>
    </div>

    <button
      class="btn btn-primary upload-action"
      :disabled="!selectedFiles.length || isUploading"
      @click="emit('upload')"
    >
      Process selected ZIPs
    </button>
  </section>
</template>

<style scoped>
.upload-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.upload-card-head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: flex-start;
}

.upload-card-copy {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.upload-card-copy h2,
.upload-card-copy p {
  margin: 0;
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
  padding: 10px 12px;
  border-radius: 10px;
  background: #202020;
  border: 1px solid #333;
  color: #d0d0d0;
  font-size: 0.92rem;
}

.upload-success {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.upload-action {
  align-self: flex-start;
}

@media (max-width: 700px) {
  .upload-action {
    width: 100%;
  }
}
</style>

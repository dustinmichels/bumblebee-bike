import { computed, ref } from "vue";
import type { BBox, GeoJSONFeatureCollection, UploadSummary } from "../lib/activity";

const ZIP_ERROR_MESSAGE = "Please select a .zip archive.";

export function useActivityDataset() {
  const selectedFile = ref<File | null>(null);
  const isUploading = ref(false);
  const uploadError = ref<string | null>(null);
  const uploadSuccess = ref(false);
  const sessionId = ref<string | null>(null);
  const totalCount = ref<number | null>(null);
  const parsedCount = ref<number | null>(null);
  const rideCount = ref<number | null>(null);

  const isFiltering = ref(false);
  const filterError = ref<string | null>(null);
  const activitiesCount = ref<number | null>(null);
  const activitiesGeoJSON = ref<GeoJSONFeatureCollection | null>(null);

  const clearFilterState = () => {
    filterError.value = null;
    activitiesCount.value = null;
    activitiesGeoJSON.value = null;
  };

  const clearUploadState = () => {
    uploadSuccess.value = false;
    sessionId.value = null;
    totalCount.value = null;
    parsedCount.value = null;
    rideCount.value = null;
    clearFilterState();
  };

  const setSelectedFile = (file: File | null) => {
    if (!file) {
      selectedFile.value = null;
      uploadError.value = null;
      clearUploadState();
      return;
    }

    if (!file.name.toLowerCase().endsWith(".zip")) {
      uploadError.value = ZIP_ERROR_MESSAGE;
      return;
    }

    selectedFile.value = file;
    uploadError.value = null;
    clearUploadState();
  };

  const submitZip = async () => {
    if (!selectedFile.value) {
      return;
    }

    isUploading.value = true;
    uploadError.value = null;
    clearUploadState();

    try {
      const formData = new FormData();
      formData.append("file", selectedFile.value);

      const res = await fetch("/api/upload", {
        method: "POST",
        body: formData,
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || `Server returned status ${res.status}`);
      }

      const data = (await res.json()) as UploadSummary;
      sessionId.value = data.sessionId;
      totalCount.value = data.total;
      parsedCount.value = data.parsed;
      rideCount.value = data.rideCount;
      uploadSuccess.value = true;
    } catch (error) {
      console.error(error);
      uploadError.value = error instanceof Error ? error.message : "An error occurred during upload.";
    } finally {
      isUploading.value = false;
    }
  };

  const filterActivities = async (bbox: BBox) => {
    if (!sessionId.value) {
      return;
    }

    isFiltering.value = true;
    clearFilterState();

    try {
      const res = await fetch("/api/filter", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          sessionId: sessionId.value,
          bbox,
        }),
      });

      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || `Server returned status ${res.status}`);
      }

      const geoJSON = (await res.json()) as GeoJSONFeatureCollection;
      activitiesGeoJSON.value = geoJSON;
      activitiesCount.value = geoJSON.features.length;
    } catch (error) {
      console.error(error);
      filterError.value =
        error instanceof Error ? error.message : "An error occurred while filtering activities.";
    } finally {
      isFiltering.value = false;
    }
  };

  const reset = () => {
    selectedFile.value = null;
    uploadError.value = null;
    isUploading.value = false;
    isFiltering.value = false;
    clearUploadState();
  };

  return {
    selectedFile,
    isUploading,
    uploadError,
    uploadSuccess,
    sessionId,
    totalCount,
    parsedCount,
    rideCount,
    isFiltering,
    filterError,
    activitiesCount,
    activitiesGeoJSON,
    readyToFilter: computed(() => uploadSuccess.value && sessionId.value !== null),
    setSelectedFile,
    submitZip,
    filterActivities,
    reset,
  };
}

<script setup lang="ts">
import SearchCity from "../SearchCity.vue";
import BBoxCoords from "../BBoxCoords.vue";
import type { BBox, SelectedCity } from "../../lib/activity";

const props = withDefaults(
  defineProps<{
    title: string;
    description: string;
    cityName: string;
    bbox: BBox;
    backLabel?: string;
    nextLabel?: string;
    nextDisabled?: boolean;
  }>(),
  {
    backLabel: "Back",
    nextLabel: "Next",
    nextDisabled: false,
  },
);

const emit = defineEmits<{
  back: [];
  next: [];
  selectCity: [payload: SelectedCity];
}>();
</script>

<template>
  <div class="card-group">
    <section class="card flow-card">
      <h2>{{ props.title }}</h2>
      <p>{{ props.description }}</p>

      <div class="control-panel">
        <label class="control-label">Search for a City</label>
        <SearchCity @select-city="emit('selectCity', $event)" />
      </div>

      <div class="city-box">
        <h4>Current Area</h4>
        <div class="city-name">{{ props.cityName }}</div>
        <BBoxCoords :bbox="props.bbox" />
      </div>

      <div class="card-actions mt-auto">
        <button class="btn btn-secondary" @click="emit('back')">{{ props.backLabel }}</button>
        <button class="btn btn-primary" :disabled="props.nextDisabled" @click="emit('next')">
          {{ props.nextLabel }}
        </button>
      </div>
    </section>

    <div class="map-container-wrapper">
      <slot />
    </div>
  </div>
</template>

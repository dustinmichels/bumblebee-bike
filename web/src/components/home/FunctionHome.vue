<script setup lang="ts">
type ToolKey = "lightning-map" | "compare" | "upload";

const emit = defineEmits<{
  selectTool: [tool: ToolKey];
}>();

const tools: Array<{
  key: ToolKey;
  eyebrow: string;
  title: string;
  summary: string;
  bullets: string[];
  action: string;
}> = [
  {
    key: "lightning-map",
    eyebrow: "Single export",
    title: "Lightning Map",
    summary: "Run the existing Strava export flow and turn one rider's archive into a route map.",
    bullets: [
      "Upload one Strava bulk export ZIP",
      "Pick the area on the map",
      "Render every matching ride in one pass",
    ],
    action: "Open Lightning Map",
  },
  {
    key: "compare",
    eyebrow: "Two riders",
    title: "Compare",
    summary: "Upload two exports, pick a color for each person, and view both route sets together.",
    bullets: [
      "Upload two Strava bulk export ZIP files",
      "Assign default orange and blue route colors",
      "Overlay both ride collections on one map",
    ],
    action: "Open Compare",
  },
  {
    key: "upload",
    eyebrow: "Local files",
    title: "Uploads",
    summary:
      "Process one or more Strava exports into GeoParquet, then rename, delete, and reuse them later.",
    bullets: [
      "Upload multiple ZIP files in one pass",
      "Browse saved GeoParquet files from your local library",
      "Reuse existing data without reprocessing",
    ],
    action: "Manage Uploads",
  },
];
</script>

<template>
  <section class="home-layout">
    <div class="card hero-card home-hero">
      <span class="hero-kicker">Route workflows</span>
      <h2>Choose a map tool</h2>
      <p class="lead-text home-copy">
        Map Tools turns Strava bulk exports into focused ride maps. Start with one rider, compare
        two people, or manage the GeoParquet files you already saved locally.
      </p>
    </div>

    <div class="tool-grid">
      <article v-for="tool in tools" :key="tool.key" class="card tool-card">
        <div class="tool-head">
          <span class="tool-eyebrow">{{ tool.eyebrow }}</span>
          <h3>{{ tool.title }}</h3>
        </div>
        <p class="tool-summary">{{ tool.summary }}</p>
        <ul class="tool-list">
          <li v-for="bullet in tool.bullets" :key="bullet">{{ bullet }}</li>
        </ul>
        <button class="btn btn-primary" @click="emit('selectTool', tool.key)">
          {{ tool.action }}
        </button>
      </article>
    </div>
  </section>
</template>

<style scoped>
.home-layout {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.home-hero {
  display: flex;
  flex-direction: column;
  gap: 12px;
  text-align: left;
}

.hero-kicker,
.tool-eyebrow {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: #ff9900;
}

.home-hero h2,
.tool-card h3 {
  margin: 0;
  color: #fff;
}

.home-copy {
  max-width: 720px;
  margin: 0;
}

.tool-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
}

.tool-card {
  display: flex;
  flex-direction: column;
  gap: 18px;
  min-height: 100%;
}

.tool-head {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.tool-summary {
  margin: 0;
  color: #b0b0b0;
  line-height: 1.6;
}

.tool-list {
  margin: 0;
  padding-left: 18px;
  color: #d0d0d0;
  line-height: 1.6;
  flex: 1;
}

.tool-list li + li {
  margin-top: 8px;
}
</style>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import CompareFlow from "./components/flows/CompareFlow.vue";
import LightningMapFlow from "./components/flows/LightningMapFlow.vue";
import FunctionHome from "./components/home/FunctionHome.vue";

type ToolKey = "home" | "lightning-map" | "compare";

const activeTool = ref<ToolKey>("home");
const health = ref<string | null>(null);

onMounted(async () => {
  try {
    const res = await fetch("/api/health");
    const data = (await res.json()) as { status: string };
    health.value = data.status;
  } catch {
    health.value = "unreachable";
  }
});
</script>

<template>
  <div class="app-layout">
    <header class="app-header">
      <div class="header-main">
        <div class="logo">🗺️</div>
        <div>
          <h1>Map Tools</h1>
          <p class="tagline">Build ride maps from Strava bulk exports.</p>
        </div>
      </div>

      <div class="header-side">
        <button v-if="activeTool !== 'home'" class="btn btn-secondary" @click="activeTool = 'home'">
          All Tools
        </button>
        <div class="api-badge" :class="health">API status: <code>{{ health ?? "…" }}</code></div>
      </div>
    </header>

    <main class="app-content">
      <FunctionHome v-if="activeTool === 'home'" @select-tool="activeTool = $event" />
      <LightningMapFlow v-else-if="activeTool === 'lightning-map'" />
      <CompareFlow v-else />
    </main>
  </div>
</template>

<style>
*,
*::before,
*::after {
  box-sizing: border-box;
}

:root {
  color-scheme: dark;
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

button,
input,
select,
textarea {
  font: inherit;
}

code {
  background: #1e1e1e;
  padding: 0.15em 0.4em;
  border-radius: 4px;
  font-size: 0.9em;
  color: #ff9900;
}

.app-layout {
  max-width: 1240px;
  width: 100%;
  margin: 0 auto;
  padding: 24px 16px 40px;
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

.header-side {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.logo {
  font-size: 2.25rem;
}

.app-header h1 {
  font-size: 1.9rem;
  margin: 0 0 4px;
  font-weight: 700;
  color: #fff;
}

.tagline {
  margin: 0;
  font-size: 0.98rem;
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

.stepper {
  display: flex;
  justify-content: space-between;
  background: #1a1a1a;
  border: 1px solid #2d2d2d;
  padding: 16px 24px;
  border-radius: 12px;
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
  font-weight: 700;
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

.btn {
  padding: 10px 20px;
  border-radius: 8px;
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

@media (max-width: 980px) {
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
  flex-wrap: wrap;
}

.mt-auto {
  margin-top: auto;
}

.mt-4 {
  margin-top: 16px;
}

.text-center {
  text-align: center;
}

.hero-card {
  padding: 36px;
}

.lead-text {
  font-size: 1.05rem;
  color: #aaa;
  line-height: 1.6;
}

.upload-zone {
  border: 2px dashed #444;
  border-radius: 8px;
  padding: 40px 20px;
  text-align: center;
  cursor: pointer;
  background: #1e1e1e;
  transition:
    border-color 0.15s,
    background 0.15s;
  position: relative;
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

.error-banner {
  background: rgba(220, 50, 50, 0.15);
  border: 1px solid #d32f2f;
  color: #ef5350;
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 14px;
  text-align: left;
}

.success-banner {
  background: rgba(50, 200, 50, 0.1);
  border: 1px solid #388e3c;
  color: #81c784;
  padding: 16px;
  border-radius: 8px;
  font-size: 14px;
  text-align: left;
}

.progress-container {
  display: flex;
  align-items: center;
  gap: 16px;
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

.step-instructions li + li {
  margin-top: 8px;
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
  margin: 0 0 6px;
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

.flow-card,
.final-card {
  display: flex;
  flex-direction: column;
}
</style>

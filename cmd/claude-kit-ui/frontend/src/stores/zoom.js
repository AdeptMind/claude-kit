import { writable } from 'svelte/store'

const STORAGE_KEY = 'ck-zoom-level'
const MIN_ZOOM = 0.6
const MAX_ZOOM = 1.6
const STEP = 0.1
const DEFAULT_ZOOM = 1.0

function loadZoom() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const val = parseFloat(raw)
      if (val >= MIN_ZOOM && val <= MAX_ZOOM) return val
    }
  } catch {}
  return DEFAULT_ZOOM
}

export const zoomLevel = writable(loadZoom())

zoomLevel.subscribe(val => {
  try { localStorage.setItem(STORAGE_KEY, String(val)) } catch {}
})

export function zoomIn() {
  zoomLevel.update(z => Math.min(MAX_ZOOM, Math.round((z + STEP) * 10) / 10))
}

export function zoomOut() {
  zoomLevel.update(z => Math.max(MIN_ZOOM, Math.round((z - STEP) * 10) / 10))
}

export function zoomReset() {
  zoomLevel.set(DEFAULT_ZOOM)
}

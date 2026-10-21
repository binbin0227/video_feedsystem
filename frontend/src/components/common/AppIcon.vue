<script setup>
defineProps({ name: { type: String, default: 'play' }, size: { type: [Number, String], default: 24 } })
const paths = {
  home: 'm3 10 9-7 9 7v10a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1Z',
  fire: 'M12 3c2 5-2 6 1 9 2-1 3-3 3-5 4 4 6 8 3 12-3 4-11 4-14-1C2 13 6 9 8 7c0 3 1 4 2 4 2-2 2-5 2-8Z',
  users: 'M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2M16 3a4 4 0 0 1 0 8m6 10v-2a4 4 0 0 0-3-3.87M13 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0Z',
  user: 'M20 21v-2a7 7 0 0 0-14 0v2M17 7a4 4 0 1 1-8 0 4 4 0 0 1 8 0Z',
  search: 'm21 21-5-5M18 10a8 8 0 1 1-16 0 8 8 0 0 1 16 0Z',
  plus: 'M12 5v14M5 12h14',
  heart: 'M20.8 4.6a5.5 5.5 0 0 0-7.8 0L12 5.7l-1.1-1.1a5.5 5.5 0 0 0-7.8 7.8L12 21l8.8-8.6a5.5 5.5 0 0 0 0-7.8Z',
  comment: 'M21 11.5a8.5 8.5 0 0 1-8.5 8.5 10 10 0 0 1-4-.8L3 21l1.8-5.5a10 10 0 0 1-.8-4A8.5 8.5 0 0 1 12.5 3H13a8.5 8.5 0 0 1 8 8v.5Z',
  arrow: 'M7 17 17 7M7 7h10v10',
  up: 'm6 15 6-6 6 6',
  down: 'm6 9 6 6 6-6',
  volume: 'm11 5-6 4H2v6h3l6 4ZM15 8a6 6 0 0 1 0 8M18 5a10 10 0 0 1 0 14',
  muted: 'm11 5-6 4H2v6h3l6 4ZM17 9l5 6m0-6-5 6',
  play: 'm8 4 13 8-13 8Z',
  pause: 'M8 4v16M16 4v16',
  fullscreen: 'M8 3H3v5m13-5h5v5M3 16v5h5m13-5v5h-5',
  close: 'm6 6 12 12M6 18 18 6',
  logout: 'M9 21H4a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1h5m5 5 5 4-5 4m-6-4h12',
}
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path :d="paths[name] || paths.play" /></svg>
</template>

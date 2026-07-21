<script setup>
defineProps({
  eyebrow: {
    type: String,
    default: 'FrameFlow',
  },
  title: {
    type: String,
    required: true,
  },
  description: {
    type: String,
    required: true,
  },
})
</script>

<template>
  <section class="placeholder-page">
    <div class="placeholder-page__copy">
      <p class="eyebrow">{{ eyebrow }}</p>
      <h1>{{ title }}</h1>
      <p class="placeholder-page__description">{{ description }}</p>
      <div class="status-line">
        <span class="status-pill">页面骨架已就绪</span>
        <span>业务接口将在后续步骤接入</span>
      </div>
    </div>

    <div class="placeholder-visual" aria-hidden="true">
      <div class="floating-card floating-card--left">Vue 3 <strong>READY</strong></div>
      <div class="video-frame">
        <div class="video-frame__top">
          <span>FRAME / 001</span>
          <span>00:15</span>
        </div>
        <div class="video-frame__play"></div>
        <div class="video-frame__meta">
          <div class="video-frame__line"></div>
          <div class="video-frame__line"></div>
          <div class="video-frame__line"></div>
        </div>
      </div>
      <div class="floating-card floating-card--right">Hertz API <strong>CONNECTED NEXT</strong></div>
    </div>
  </section>
</template>

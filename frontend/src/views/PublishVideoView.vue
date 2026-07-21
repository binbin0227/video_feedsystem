<script setup>
import { computed, onBeforeUnmount, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { publishVideo, uploadCoverFile, uploadVideoFile } from '../api/video'

const MAX_VIDEO_SIZE = 200 * 1024 * 1024
const MAX_COVER_SIZE = 10 * 1024 * 1024
const CAPTURE_COVER_WIDTH = 1280
const CAPTURE_COVER_HEIGHT = 720

const route = useRoute()
const router = useRouter()
const form = reactive({
  title: '',
  description: '',
})
const previewVideo = ref(null)
const videoFile = ref(null)
const coverFile = ref(null)
const videoPreviewUrl = ref('')
const coverPreviewUrl = ref('')
const coverSource = ref('')
const frameCaptureReady = ref(false)
const capturingFrame = ref(false)
const capturedFrameTime = ref(null)
const uploadedVideoUrl = ref('')
const uploadedCoverUrl = ref('')
const videoProgress = ref(0)
const coverProgress = ref(0)
const stage = ref('idle')
const errorMessage = ref('')
const publishedVideoId = ref('')

const submitting = computed(() => stage.value !== 'idle' || Boolean(publishedVideoId.value))
const publishedDetailHref = computed(() => publishedVideoId.value ? router.resolve({ name: 'video-detail', params: { videoId: publishedVideoId.value } }).href : '')
const stageLabel = computed(() => {
  const labels = {
    'upload-video': '正在上传视频…',
    'upload-cover': '正在上传封面…',
    publishing: '正在发布视频…',
    published: '已经发布',
  }

  return labels[stage.value] || '上传并发布'
})

function preparePlayer(event) {
  // 桌面浏览器使用 30% 初始音量；手机浏览器保留原生音量并交给系统媒体音量控制。
  if (!window.matchMedia('(hover: none), (pointer: coarse)').matches) {
    event.currentTarget.volume = 0.3
  }
  event.currentTarget.muted = false
}

function formatTimestamp(value) {
  const totalSeconds = Math.max(0, Math.floor(Number(value) || 0))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  const minutePart = hours ? String(minutes).padStart(2, '0') : String(minutes)
  const secondPart = String(seconds).padStart(2, '0')

  return hours ? `${hours}:${minutePart}:${secondPart}` : `${minutePart}:${secondPart}`
}

function revokePreview(url) {
  if (url) {
    URL.revokeObjectURL(url)
  }
}

function handleVideoChange(event) {
  const file = event.target.files?.[0] || null
  errorMessage.value = ''

  if (!file) {
    return
  }
  if (!/\.mp4$/i.test(file.name)) {
    errorMessage.value = '视频只支持 MP4 格式'
    event.target.value = ''
    return
  }
  if (file.size <= 0 || file.size > MAX_VIDEO_SIZE) {
    errorMessage.value = '视频文件必须大于 0 且不能超过 200MB'
    event.target.value = ''
    return
  }

  revokePreview(videoPreviewUrl.value)
  videoFile.value = file
  videoPreviewUrl.value = URL.createObjectURL(file)
  frameCaptureReady.value = false
  capturedFrameTime.value = null
  uploadedVideoUrl.value = ''
  videoProgress.value = 0

  // 截帧封面属于原视频，切换视频后不能继续沿用；手动上传的封面则保留。
  if (coverSource.value === 'frame') {
    revokePreview(coverPreviewUrl.value)
    coverFile.value = null
    coverPreviewUrl.value = ''
    coverSource.value = ''
    uploadedCoverUrl.value = ''
    coverProgress.value = 0
  }
}

function handleCoverChange(event) {
  const file = event.target.files?.[0] || null
  errorMessage.value = ''

  if (!file) {
    return
  }
  if (!/\.(jpe?g|png)$/i.test(file.name)) {
    errorMessage.value = '封面只支持 JPG、JPEG 或 PNG 格式'
    event.target.value = ''
    return
  }
  if (file.size <= 0 || file.size > MAX_COVER_SIZE) {
    errorMessage.value = '封面文件必须大于 0 且不能超过 10MB'
    event.target.value = ''
    return
  }

  revokePreview(coverPreviewUrl.value)
  coverFile.value = file
  coverPreviewUrl.value = URL.createObjectURL(file)
  coverSource.value = 'upload'
  capturedFrameTime.value = null
  uploadedCoverUrl.value = ''
  coverProgress.value = 0
}

function handlePreviewFrameReady(event) {
  frameCaptureReady.value = event.currentTarget.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA
}

async function captureCurrentFrame() {
  const player = previewVideo.value
  errorMessage.value = ''

  if (!videoFile.value || !player) {
    errorMessage.value = '请先选择视频文件'
    return
  }
  if (!frameCaptureReady.value || player.readyState < HTMLMediaElement.HAVE_CURRENT_DATA) {
    errorMessage.value = '视频画面还在准备，请稍后再截取'
    return
  }

  capturingFrame.value = true

  try {
    player.pause()

    if (!player.videoWidth || !player.videoHeight) {
      throw new Error('暂时无法读取视频画面尺寸')
    }

    const canvas = document.createElement('canvas')
    canvas.width = CAPTURE_COVER_WIDTH
    canvas.height = CAPTURE_COVER_HEIGHT

    const context = canvas.getContext('2d', { alpha: false })

    if (!context) {
      throw new Error('当前浏览器无法创建封面画布')
    }

    const scale = Math.min(canvas.width / player.videoWidth, canvas.height / player.videoHeight)
    const frameWidth = Math.round(player.videoWidth * scale)
    const frameHeight = Math.round(player.videoHeight * scale)
    const frameX = Math.round((canvas.width - frameWidth) / 2)
    const frameY = Math.round((canvas.height - frameHeight) / 2)

    // 封面固定输出为 16:9；原始帧完整居中，比例不一致的区域使用深色补边，避免列表展示时裁掉内容。
    context.fillStyle = '#030407'
    context.fillRect(0, 0, canvas.width, canvas.height)
    context.drawImage(player, frameX, frameY, frameWidth, frameHeight)

    const blob = await new Promise((resolve) => canvas.toBlob(resolve, 'image/jpeg', 0.9))

    if (!blob) {
      throw new Error('当前视频帧转换失败')
    }
    if (blob.size > MAX_COVER_SIZE) {
      throw new Error('截取的封面超过 10MB，请换一个时间点重试')
    }

    const currentTime = player.currentTime
    const file = new File([blob], `video-cover-${Math.round(currentTime * 1000)}ms.jpg`, { type: 'image/jpeg' })

    revokePreview(coverPreviewUrl.value)
    coverFile.value = file
    coverPreviewUrl.value = URL.createObjectURL(file)
    coverSource.value = 'frame'
    capturedFrameTime.value = currentTime
    uploadedCoverUrl.value = ''
    coverProgress.value = 0
  } catch (error) {
    errorMessage.value = error.message || '截取视频封面失败'
  } finally {
    capturingFrame.value = false
  }
}

function updateProgress(target, event) {
  if (!event.total) {
    return
  }

  target.value = Math.min(100, Math.round((event.loaded / event.total) * 100))
}

function validateForm() {
  const title = form.title.trim()

  if (!title) {
    return '请输入视频标题'
  }
  if ([...title].length > 128) {
    return '视频标题不能超过 128 个字符'
  }
  if ([...form.description.trim()].length > 2000) {
    return '视频描述不能超过 2000 个字符'
  }
  if (!videoFile.value) {
    return '请选择 MP4 视频文件'
  }
  if (!coverFile.value) {
    return '请选择视频封面'
  }

  return ''
}

async function handleSubmit() {
  if (submitting.value) {
    return
  }

  errorMessage.value = validateForm()

  if (errorMessage.value) {
    return
  }

  try {
    if (!uploadedVideoUrl.value) {
      stage.value = 'upload-video'
      const videoData = await uploadVideoFile(videoFile.value, (event) => updateProgress(videoProgress, event))

      if (!videoData?.video_url) {
        throw new Error('视频上传响应中没有 video_url')
      }
      uploadedVideoUrl.value = videoData.video_url
      videoProgress.value = 100
    }

    if (!uploadedCoverUrl.value) {
      stage.value = 'upload-cover'
      const coverData = await uploadCoverFile(coverFile.value, (event) => updateProgress(coverProgress, event))

      if (!coverData?.cover_url) {
        throw new Error('封面上传响应中没有 cover_url')
      }
      uploadedCoverUrl.value = coverData.cover_url
      coverProgress.value = 100
    }

    stage.value = 'publishing'
    const data = await publishVideo({
      title: form.title.trim(),
      description: form.description.trim(),
      playUrl: uploadedVideoUrl.value,
      coverUrl: uploadedCoverUrl.value,
    })

    if (!data?.video?.id) {
      throw new Error('发布响应中没有视频 ID')
    }

    // 接口返回 ID 就代表后端已经发布成功；后续页面模块加载失败不能允许用户再次提交。
    publishedVideoId.value = String(data.video.id)
    stage.value = 'published'

    try {
      await router.replace({ name: 'video-detail', params: { videoId: publishedVideoId.value } })
    } catch {
      errorMessage.value = '视频已经发布成功，但详情页资源加载失败。请使用下方入口重新打开，不要再次发布。'
    }
  } catch (error) {
    if (error.response?.status === 401) {
      await router.replace({ name: 'login', query: { redirect: route.fullPath } })
      return
    }

    errorMessage.value = error.message
  } finally {
    if (!publishedVideoId.value) {
      stage.value = 'idle'
    }
  }
}

onBeforeUnmount(() => {
  revokePreview(videoPreviewUrl.value)
  revokePreview(coverPreviewUrl.value)
})
</script>

<template>
  <section class="publish-page">
    <header class="publish-heading">
      <div>
        <p class="eyebrow">CREATOR STUDIO</p>
        <h1>上传并发布</h1>
        <p>依次上传视频、封面，再将作品信息提交给后端。</p>
      </div>
      <ol class="publish-steps" aria-label="发布步骤">
        <li :class="{ active: stage === 'upload-video' || uploadedVideoUrl }"><span>1</span>视频</li>
        <li :class="{ active: stage === 'upload-cover' || uploadedCoverUrl }"><span>2</span>封面</li>
        <li :class="{ active: stage === 'publishing' || stage === 'published' }"><span>3</span>发布</li>
      </ol>
    </header>

    <div class="publish-layout">
      <form class="publish-form" @submit.prevent="handleSubmit">
        <label class="publish-field">
          <span>视频标题 <small>必填，最多 128 字</small></span>
          <input v-model="form.title" name="title" type="text" maxlength="128" placeholder="给作品起一个清晰的标题" :disabled="submitting" @input="errorMessage = ''">
        </label>

        <label class="publish-field">
          <span>视频描述 <small>选填，最多 2000 字</small></span>
          <textarea v-model="form.description" name="description" rows="5" maxlength="2000" placeholder="介绍一下这条视频…" :disabled="submitting" @input="errorMessage = ''"></textarea>
          <small class="publish-counter">{{ form.description.length }}/2000</small>
        </label>

        <div class="publish-upload-grid">
          <label class="publish-upload">
            <span class="publish-upload__type">MP4</span>
            <strong>选择视频</strong>
            <small>{{ videoFile ? videoFile.name : '最大 200MB' }}</small>
            <input type="file" accept="video/mp4,.mp4" :disabled="submitting" @change="handleVideoChange">
            <span v-if="videoProgress" class="publish-progress"><i :style="{ width: `${videoProgress}%` }"></i></span>
          </label>

          <label class="publish-upload">
            <span class="publish-upload__type">COVER</span>
            <strong>选择封面</strong>
            <small>{{ coverFile ? coverFile.name : 'JPG / JPEG / PNG，最大 10MB' }}</small>
            <input type="file" accept="image/jpeg,image/png,.jpg,.jpeg,.png" :disabled="submitting" @change="handleCoverChange">
            <span v-if="coverProgress" class="publish-progress"><i :style="{ width: `${coverProgress}%` }"></i></span>
          </label>
        </div>

        <p v-if="errorMessage" class="publish-error" role="alert">{{ errorMessage }}</p>

        <div v-if="publishedVideoId" class="publish-success" role="status">
          <strong>视频已经发布成功</strong>
          <span>视频 ID：{{ publishedVideoId }}</span>
          <div>
            <a :href="publishedDetailHref">重新打开详情页</a>
            <a href="/">返回首页</a>
          </div>
        </div>

        <button class="publish-submit" type="submit" :disabled="submitting">
          {{ stageLabel }}
        </button>
      </form>

      <aside class="publish-preview" aria-label="作品预览">
        <div class="publish-preview__heading">
          <span>PREVIEW</span>
          <small>本地预览，不会提前发布</small>
        </div>

        <div class="publish-preview__video">
          <video
            v-if="videoPreviewUrl"
            ref="previewVideo"
            :src="videoPreviewUrl"
            :poster="coverPreviewUrl"
            controls
            playsinline
            preload="auto"
            @loadedmetadata="preparePlayer"
            @loadeddata="handlePreviewFrameReady"
            @seeking="frameCaptureReady = false"
            @seeked="handlePreviewFrameReady"
          ></video>
          <div v-else>
            <strong>等待视频</strong>
            <span>选择 MP4 后可在这里预览</span>
          </div>
        </div>

        <div v-if="videoPreviewUrl" class="publish-frame-capture">
          <div>
            <strong>从视频截取封面</strong>
            <small>拖动上方进度条，停在想要的画面后截取</small>
          </div>
          <div>
            <span v-if="capturedFrameTime !== null">已截取 {{ formatTimestamp(capturedFrameTime) }}</span>
            <button type="button" :disabled="submitting || capturingFrame || !frameCaptureReady" @click="captureCurrentFrame">
              {{ capturingFrame ? '正在截取…' : frameCaptureReady ? '截取当前帧' : '画面准备中' }}
            </button>
          </div>
        </div>

        <div class="publish-preview__cover">
          <img v-if="coverPreviewUrl" :src="coverPreviewUrl" alt="待发布的视频封面预览">
          <div v-else>封面预览</div>
        </div>

        <div class="publish-preview__copy">
          <h2>{{ form.title.trim() || '未命名作品' }}</h2>
          <p>{{ form.description.trim() || '填写描述后会显示在这里。' }}</p>
        </div>
      </aside>
    </div>
  </section>
</template>

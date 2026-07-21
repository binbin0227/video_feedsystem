import { createRouter, createWebHistory } from 'vue-router'
import { isLoggedIn } from '../utils/auth'

const routes = [
  {
    path: '/',
    name: 'home',
    component: () => import('../views/PublicFeedView.vue'),
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/LoginView.vue'),
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('../views/RegisterView.vue'),
  },
  {
    path: '/following',
    name: 'following-feed',
    component: () => import('../views/FollowingFeedView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/videos/:videoId',
    name: 'video-detail',
    component: () => import('../views/VideoDetailView.vue'),
  },
  {
    path: '/publish',
    name: 'publish',
    component: () => import('../views/PublishVideoView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/users/:accountId',
    name: 'user-profile',
    component: () => import('../views/UserProfileView.vue'),
  },
  {
    path: '/search',
    name: 'search',
    component: () => import('../views/UserSearchView.vue'),
  },
  {
    path: '/me/liked',
    name: 'liked-videos',
    component: () => import('../views/LikedVideosView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/me/following',
    name: 'following-list',
    component: () => import('../views/FollowingListView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/me/followers',
    name: 'follower-list',
    component: () => import('../views/FollowerListView.vue'),
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !isLoggedIn()) {
    return {
      name: 'login',
      query: { redirect: to.fullPath },
    }
  }

  return true
})

export default router

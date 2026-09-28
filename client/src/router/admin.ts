import type { RouteComponent } from 'vue-router';

/**
 * 后台子路由（挂在 /admin 下）。name 为稳定契约：移动端后台按 name 映射自己的视图。
 * 由桌面后台负责维护；新增页面时同步在此登记。
 */
export interface AdminChild {
  path: string;
  name: string;
  component: () => Promise<RouteComponent>;
}

export const adminChildren: AdminChild[] = [
  { path: '', name: 'admin-today', component: () => import('../views/admin/AdminTodayView.vue') },
  { path: 'posts', name: 'admin-posts', component: () => import('../views/admin/AdminPostsView.vue') },
  { path: 'write/post', name: 'admin-write-post', component: () => import('../views/write/WritePostView.vue') },
  { path: 'write/note', name: 'admin-write-note', component: () => import('../views/write/WriteNoteView.vue') },
  { path: 'notes', name: 'admin-notes', component: () => import('../views/admin/AdminNotesView.vue') },
  { path: 'media', name: 'admin-media', component: () => import('../views/admin/AdminMediaView.vue') },
  { path: 'identity', name: 'admin-identity', component: () => import('../views/admin/AdminIdentityView.vue') },
  { path: 'about', name: 'admin-about', component: () => import('../views/admin/AdminAboutView.vue') },
  { path: 'appearance', name: 'admin-appearance', component: () => import('../views/admin/AdminAppearanceView.vue') },
  { path: 'settings', name: 'admin-settings', component: () => import('../views/admin/AdminSettingsView.vue') },
  { path: 'apikeys', name: 'admin-apikeys', component: () => import('../views/admin/AdminApiKeysView.vue') },
  { path: 'data', name: 'admin-data', component: () => import('../views/admin/AdminDataView.vue') },
  { path: 'comments', name: 'admin-comments', component: () => import('../views/admin/AdminCommentsView.vue') },
  { path: 'users', name: 'admin-users', component: () => import('../views/admin/AdminUsersView.vue') },
];

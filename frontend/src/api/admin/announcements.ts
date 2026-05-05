/**
 * Admin Announcements API endpoints
 */

import { apiClient } from '../client'
import type {
  Announcement,
  AnnouncementUserReadStatus,
  BasePaginationResponse,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest
} from '@/types'

export interface AnnouncementImageStorageConfig {
  endpoint: string
  region: string
  bucket: string
  access_key_id: string
  secret_access_key?: string
  has_secret_access_key?: boolean
  public_base_url: string
  object_prefix: string
}

export interface AnnouncementImageUploadResult {
  url: string
  key: string
  content_type: string
  size: number
}

export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: string
    search?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<BasePaginationResponse<Announcement>> {
  const { data } = await apiClient.get<BasePaginationResponse<Announcement>>('/admin/announcements', {
    params: { page, page_size: pageSize, ...filters },
    signal: options?.signal
  })
  return data
}

export async function getById(id: number): Promise<Announcement> {
  const { data } = await apiClient.get<Announcement>(`/admin/announcements/${id}`)
  return data
}

export async function create(request: CreateAnnouncementRequest): Promise<Announcement> {
  const { data } = await apiClient.post<Announcement>('/admin/announcements', request)
  return data
}

export async function update(id: number, request: UpdateAnnouncementRequest): Promise<Announcement> {
  const { data } = await apiClient.put<Announcement>(`/admin/announcements/${id}`, request)
  return data
}

export async function deleteAnnouncement(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/announcements/${id}`)
  return data
}

export async function getReadStatus(
  id: number,
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    search?: string
    sort_by?: string
    sort_order?: 'asc' | 'desc'
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<BasePaginationResponse<AnnouncementUserReadStatus>> {
  const { data } = await apiClient.get<BasePaginationResponse<AnnouncementUserReadStatus>>(
    `/admin/announcements/${id}/read-status`,
    {
      params: { page, page_size: pageSize, ...filters },
      signal: options?.signal
    }
  )
  return data
}

export async function getImageStorage(): Promise<AnnouncementImageStorageConfig> {
  const { data } = await apiClient.get<AnnouncementImageStorageConfig>('/admin/announcements/image-storage')
  return data
}

export async function updateImageStorage(request: AnnouncementImageStorageConfig): Promise<AnnouncementImageStorageConfig> {
  const { data } = await apiClient.put<AnnouncementImageStorageConfig>('/admin/announcements/image-storage', request)
  return data
}

export async function uploadImage(file: File): Promise<AnnouncementImageUploadResult> {
  const form = new FormData()
  form.append('image', file)
  const { data } = await apiClient.post<AnnouncementImageUploadResult>('/admin/announcements/images', form)
  return data
}

const announcementsAPI = {
  list,
  getById,
  create,
  update,
  delete: deleteAnnouncement,
  getReadStatus,
  getImageStorage,
  updateImageStorage,
  uploadImage
}

export default announcementsAPI

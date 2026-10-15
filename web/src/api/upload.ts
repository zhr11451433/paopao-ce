import { request } from '@/utils/request'
import type { UploadResult } from '@/types/api'

/**
 * 上传图片：POST /upload（multipart/form-data）
 *
 * 【两个必须记住的约定】
 * 1. 文件字段名必须叫 `image`（后台写死了 c.FormFile("image")），传多张就 append 多次。
 * 2. 千万不要手写 Content-Type: 'multipart/form-data'。
 *    multipart 的 header 里要带一个 boundary 分隔串，浏览器/axios 会自己生成；
 *    手写会把 boundary 弄丢，后端直接解析失败。
 */
export function uploadImages(files: File[]): Promise<UploadResult> {
  const formData = new FormData()
  files.forEach((file) => formData.append('image', file))
  return request<UploadResult>({ url: '/upload', method: 'post', data: formData })
}
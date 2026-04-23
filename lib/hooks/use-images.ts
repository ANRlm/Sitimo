'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { batchDeleteImages, deleteImage, editImage, getImage, hardDeleteImage, listImages, restoreImage, updateImage, uploadImage, type ImageEditInput, type ImageListQuery } from '@/lib/api/images';

export function useImages(query: ImageListQuery) {
  return useQuery({
    queryKey: ['images', query],
    queryFn: () => listImages(query),
  });
}

export function useImage(id?: string) {
  return useQuery({
    queryKey: ['images', id],
    queryFn: () => getImage(id!),
    enabled: Boolean(id),
  });
}

export function useUploadImage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: uploadImage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] });
      queryClient.invalidateQueries({ queryKey: ['meta'] });
      toast.success('图像已上传');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useUpdateImage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: { tagIds: string[]; description?: string } }) => updateImage(id, input),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['images'] });
      queryClient.invalidateQueries({ queryKey: ['images', variables.id] });
      toast.success('图像已更新');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useDeleteImage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteImage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] });
      toast.success('图像已移入回收站');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useRestoreImage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: restoreImage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] });
      toast.success('图像已恢复');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useHardDeleteImage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: hardDeleteImage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] });
      toast.success('图像已彻底删除');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useEditImage() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input, problemId }: { id: string; input: ImageEditInput; problemId?: string }) =>
      editImage(id, input, problemId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] });
      queryClient.invalidateQueries({ queryKey: ['problems'] });
      toast.success('图像处理已完成');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

export function useBatchDeleteImages() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: batchDeleteImages,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] });
      toast.success('图像已批量删除');
    },
    onError: (error: Error) => toast.error(error.message),
  });
}

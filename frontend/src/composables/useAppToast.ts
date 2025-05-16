import { useToast } from 'primevue/usetoast'

export function useAppToast() {
  const toast = useToast()

  return {
    showSuccess: (summary: string, detail?: string) =>
      toast.add({ severity: 'success', summary, detail, life: 3000 }),
    showError: (summary: string, detail?: string) =>
      toast.add({ severity: 'error', summary, detail, life: 5000 }),
  }
}
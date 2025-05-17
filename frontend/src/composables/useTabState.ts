import { ref } from 'vue'

const activeTab = ref('')

export function useTabState() {
    return {
        activeTab
    }
}
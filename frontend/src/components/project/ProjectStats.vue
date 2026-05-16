<template>
  <div>
    <h2>Project Statistics</h2>
    <!-- <Doughnut :data="{ -->
    <!--   labels: ['Time Spent', 'Remaining Time'], -->
    <!--   datasets: [ -->
    <!--     { -->
    <!--       data: [ -->
    <!--         (project?.totalTimeMs || 0) / (1000 * 60 * 60), -->
    <!--         Math.max( -->
    <!--           0, -->
    <!--           (project?.timeBudgetHours || 0) - -->
    <!--           (project?.totalTimeMs || 0) / (1000 * 60 * 60), -->
    <!--         ), -->
    <!--       ], -->
    <!--       backgroundColor: [nord.nord14, nord.nord3], -->
    <!--       borderColor: nord.nordc0, -->
    <!--       borderWidth: 2, -->
    <!--     }, -->
    <!--   ], -->
    <!-- }" :options="{ -->
    <!--   responsive: true, -->
    <!--   plugins: { -->
    <!--     tooltip: { -->
    <!--       callbacks: { -->
    <!--         label: function (context) { -->
    <!--           const label = context.label || ''; -->
    <!--           const value = context.parsed || 0; -->
    <!--           return `${label}: ${value}h`; -->
    <!--         }, -->
    <!--       }, -->
    <!--     }, -->
    <!--   }, -->
    <!-- }" /> -->
    {{ props.projectId }}
    {{ stats }}
  </div>
</template>

<script setup lang="ts">
// import { Doughnut } from "vue-chartjs";
import type { ProjectStats } from "@/model";
import { useProjectsStore } from "@/stores/projects";
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from "chart.js";
import { ref, watch } from "vue";
// import { nord } from "@/helpers/nord";

ChartJS.register(ArcElement, Tooltip, Legend);

const projectsStore = useProjectsStore();

const props = defineProps<{
  projectId: string;
}>();

const stats = ref<ProjectStats | undefined>();

async function updateProjectStats() {
  try {
    const result = await projectsStore.fetchProjectStats(props.projectId, "time_spent", "", "",
      Intl.DateTimeFormat().resolvedOptions().timeZone);
    if (result) {
      stats.value = result;
    }
  } catch {
  }
}

watch(
  () => props.projectId,
  async (newId, oldId) => {
    if (!newId || newId === oldId) {
      // No need to refetch if the ID hasn't changed
      return;
    }

    updateProjectStats();
  },
  { immediate: true },
);
</script>

<style scoped></style

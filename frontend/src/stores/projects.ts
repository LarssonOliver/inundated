import { projectsApi, type ProjectsApi, type PaginatedProjectsResponse } from "@/api/projects";
import type { Project, ProjectStats } from "@/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

function copyProject(project: Project): Project {
  return {
    ...project,
    tagIds: new Set(project.tagIds),
  };
}

export interface PaginationState {
  limit: number;
  offset: number;
  total: number;
}

function createProjectsStore(api: ProjectsApi, now: () => number = () => Date.now()) {
  return defineStore("projects", () => {
    const projects = ref<Map<string, Project>>(new Map<string, Project>());
    const _pending = ref<Promise<void> | null>(null);

    const lastFetched = ref<number | null>(null);
    const paginationState = ref<PaginationState | null>(null);
    const TTL = 60_000; // 1 minute

    const readOnlyProjects = computed<readonly Project[]>(() =>
      Array.from(projects.value.values()).map(copyProject),
    );

    /**
     * Fetches the first page of projects from the API and stores them locally.
     *
     * @returns A promise that resolves when the projects have been fetched.
     */
    async function fetchProjectsAlways(): Promise<void> {
      if (_pending.value) return _pending.value;

      _pending.value = (async () => {
        const result = await api.listProjectsPaginated(50, 0);
        projects.value = new Map(result.data.map((project) => [project.id, project]));
        paginationState.value = result.pagination;
      })();

      try {
        await _pending.value;
      } finally {
        _pending.value = null;
      }
    }

    /**
     * Fetches the first page of projects from the API if the cached projects are stale (older than TTL).
     *
     * @returns A promise that resolves when the projects have been fetched or if the cached projects are still valid.
     */
    async function fetchProjects(): Promise<void> {
      if (lastFetched.value && now() - lastFetched.value < TTL) {
        return;
      }

      await fetchProjectsAlways();
      lastFetched.value = now();
    }

    /**
     * Fetches a specific page of projects and accumulates them in the cache.
     * Used for infinite scrolling.
     *
     * @param limit - The number of items per page
     * @param offset - The offset to start from
     * @returns A promise that resolves when the page has been fetched
     */
    async function fetchProjectsPage(limit: number = 50, offset: number = 0): Promise<void> {
      if (_pending.value) return _pending.value;

      _pending.value = (async () => {
        const result = await api.listProjectsPaginated(limit, offset);
        // Accumulate items in the map instead of replacing
        for (const project of result.data) {
          projects.value.set(project.id, project);
        }
        paginationState.value = result.pagination;
      })();

      try {
        await _pending.value;
      } finally {
        _pending.value = null;
      }
    }

    /**
     * Gets pagination information for the currently loaded page.
     *
     * @returns Pagination state or null if no page has been fetched
     */
    function getPaginationState(): PaginationState | null {
      return paginationState.value;
    }

    /**
     * Checks if there are more items to fetch.
     *
     * @returns true if there are more items available to fetch
     */
    function hasMoreItems(): boolean {
      if (!paginationState.value) return false;
      const { limit, offset, total } = paginationState.value;
      return offset + limit < total;
    }

    /**
     * Creates a new project.
     *
     * @param project - The project to create.
     *
     * @returns A promise that resolves to the newly created project with
     *   correctly assigned ID.
     */
    async function createProject(project: Omit<Project, "id">): Promise<Project> {
      const newProject = await api.createProject(project);
      projects.value.set(newProject.id, newProject);
      return copyProject(newProject);
    }

    /**
     * Fetches a project by its ID.
     *
     * @param id - The ID of the project to fetch.
     *
     * @returns A promise that resolves to the project if found, or undefined
     *  if not found.
     */
    function getProjectById(id: string): Project | undefined {
      const project = projects.value.get(id);
      if (!project) return undefined;
      return copyProject(project);
    }

    /**
     * Fetches a project by its ID with detailed information (e.g., total time spent).
     *
     * @param id - The ID of the project to fetch.
     *
     * @returns A promise that resolves to the project with detailed information if found, or undefined
     */
    async function fetchDetailedProjectById(id: string): Promise<Project> {
      const project = await api.getProject(id, true);
      return copyProject(project);
    }

    /**
     * Updates an existing project.
     *
     * @param project - The project to update, identified by project.id.
     *
     * @returns A promise that resolves to the updated project, or undefined
     *  if not found.
     */
    async function updateProject(project: Project): Promise<Project> {
      const { id, ...fields } = project;
      const updated = await api.updateProject(id, fields);
      projects.value.set(updated.id, updated);
      return copyProject(updated);
    }

    /**
     * Deletes a project by its ID.
     *
     * @param id - The ID of the project to delete.
     *
     * @returns A promise that resolves when the project is deleted.
     */
    async function deleteProject(id: string): Promise<void> {
      await api.deleteProject(id);
      projects.value.delete(id);
    }

    async function fetchProjectStats(
      projectId: string,
      metric: string,
      interval: string,
      granularity: string,
      timezone: string,
    ): Promise<ProjectStats> {
      return await api.fetchProjectStats(projectId, metric, interval, granularity, timezone);
    }

    return {
      projects: readOnlyProjects,
      fetchProjects,
      fetchProjectsPage,
      getPaginationState,
      hasMoreItems,
      createProject,
      getProjectById,
      fetchDetailedProjectById,
      updateProject,
      deleteProject,
      fetchProjectStats,
    };
  });
}

export const useProjectsStore = createProjectsStore(projectsApi);
export const __test__ = { createProjectsStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useProjectsStore, import.meta.hot));
}

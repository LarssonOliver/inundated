import { projectsApi, type ProjectsApi } from "@/api/projects";
import type { Project } from "@/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

function copyProject(project: Project): Project {
  return {
    ...project,
    tagIds: new Set(project.tagIds),
  };
}

function createProjectsStore(api: ProjectsApi, now: () => number = () => Date.now()) {
  return defineStore("projects", () => {
    const projects = ref<Map<string, Project>>(new Map<string, Project>());
    const _pending = ref<Promise<void> | null>(null);

    const lastFetched = ref<number | null>(null);
    const TTL = 60_000; // 1 minute

    const readOnlyProjects = computed<readonly Project[]>(() =>
      Array.from(projects.value.values()).map(copyProject),
    );

    /**
     * Fetches all projects from the API and stores them locally.
     *
     * @returns A promise that resolves when the projects have been fetched.
     */
    async function fetchProjectsAlways(): Promise<void> {
      if (_pending.value) return _pending.value;

      _pending.value = (async () => {
        const fetched = await api.listProjects();
        projects.value = new Map(fetched.map((project) => [project.id, project]));
      })();

      try {
        await _pending.value;
      } finally {
        _pending.value = null;
      }
    }

    /**
     * Fetches all projects from the API if the cached projects are stale (older than TTL).
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

    return {
      projects: readOnlyProjects,
      fetchProjects,
      createProject,
      getProjectById,
      updateProject,
      deleteProject,
    };
  });
}

export const useProjectsStore = createProjectsStore(projectsApi);
export const __test__ = { createProjectsStore };

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useProjectsStore, import.meta.hot));
}

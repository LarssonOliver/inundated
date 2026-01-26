import type { Project } from "@/model/model";
import { acceptHMRUpdate } from "pinia";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

function copyProject(project: Project): Project {
  return {
    ...project,
    tagIds: [...project.tagIds],
  };
}

export const useProjectsStore = defineStore("projects", () => {
  const projects = ref<Project[]>([]);

  const readOnlyProjects = computed(() => projects.value.map(copyProject));

  /**
   * Creates a new project.
   *
   * @param project - The project to create.
   *
   * @returns A promise that resolves to the newly created project with
   *   correctly assigned ID.
   */
  async function createProject(project: Project): Promise<Project> {
    const newProject: Project = {
      id: projects.value.length + 1, // TODO: Use ID from server
      name: project.name,
      color: project.color,
      timeBudget: project.timeBudget,
      userId: project.userId, // TODO: Use the current user's id
      tagIds: [...project.tagIds],
    };

    projects.value.push(newProject);

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
  async function getProjectById(id: number): Promise<Project | undefined> {
    const project = projects.value.find((project) => project.id === id);

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
  async function updateProject(project: Project): Promise<Project | undefined> {
    const index = projects.value.findIndex((p) => p.id === project.id);

    if (index === -1) return undefined;

    const copy = {
      ...projects.value[index],
      ...copyProject(project),
    };

    projects.value.splice(index, 1, copy);

    return copyProject(copy);
  }

  /**
   * Deletes a project by its ID.
   *
   * @param id - The ID of the project to delete.
   *
   * @returns A promise that resolves when the project is deleted.
   */
  async function deleteProject(id: number): Promise<void> {
    const index = projects.value.findIndex((p) => p.id === id);
    if (index !== -1) projects.value.splice(index, 1);
  }

  return {
    projects: readOnlyProjects,
    createProject,
    getProjectById,
    updateProject,
    deleteProject,
  };
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useProjectsStore, import.meta.hot));
}

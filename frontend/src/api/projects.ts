import type { Project } from "@/model";
import { ProjectsApi as GeneratedProjectsApi, GetProjectIncludeEnum } from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { mapFromApiArray, projectMapper, toApiCreateProject, toApiUpdateProject } from "./mappers";

export interface ProjectsApi {
  listProjects(): Promise<Project[]>;
  getProject(id: string, detailed: boolean): Promise<Project>;
  createProject(project: Omit<Project, "id">): Promise<Project>;
  updateProject(id: string, project: Partial<Omit<Project, "id">>): Promise<Project>;
  deleteProject(id: string): Promise<void>;
}

const defaultGeneratedApi = new GeneratedProjectsApi(ApiConfig);

function createProjectsApi(api: GeneratedProjectsApi = defaultGeneratedApi): ProjectsApi {
  return {
    async listProjects(): Promise<Project[]> {
      const response = await api.listProjects();
      return mapFromApiArray(projectMapper, response);
    },

    async getProject(id: string, detailed: boolean): Promise<Project> {
      const include = new Set<GetProjectIncludeEnum>();
      if (detailed) {
        include.add(GetProjectIncludeEnum.TotalTimeMs);
      }

      const response = await api.getProject({ projectId: id, include: include });
      return projectMapper.fromApi(response);
    },

    async createProject(project: Omit<Project, "id">): Promise<Project> {
      const newProject = toApiCreateProject(project);
      const response = await api.createProject({ createProject: newProject });
      return projectMapper.fromApi(response);
    },

    async updateProject(id: string, project: Partial<Omit<Project, "id">>): Promise<Project> {
      const updateProject = toApiUpdateProject(project);
      const response = await api.updateProject({ projectId: id, updateProject: updateProject });
      return projectMapper.fromApi(response);
    },

    async deleteProject(id: string): Promise<void> {
      return await api.deleteProject({ projectId: id });
    },
  };
}

export const projectsApi = createProjectsApi();
export const __test__ = { createProjectsApi };

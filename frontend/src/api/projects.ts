import type { Project, ProjectStats } from "@/model";
import {
  ProjectsApi as GeneratedProjectsApi,
  GetProjectIncludeEnum,
  ProjectStatsMetricEnum,
  type PaginatedProjects,
} from "@/api/generated";
import { ApiConfig } from "@/api/config";
import { mapFromApiArray, projectMapper, toApiCreateProject, toApiUpdateProject } from "./mappers";
import { projectStatsMapper } from "./mappers/projectStatsMapper";

export interface PaginationMetadata {
  limit: number;
  offset: number;
  total: number;
}

export interface PaginatedProjectsResponse {
  data: Project[];
  pagination: PaginationMetadata;
}

export interface ProjectsApi {
  listProjects(): Promise<Project[]>;
  listProjectsPaginated(limit?: number, offset?: number): Promise<PaginatedProjectsResponse>;
  getProject(id: string, detailed: boolean): Promise<Project>;
  createProject(project: Omit<Project, "id">): Promise<Project>;
  updateProject(id: string, project: Partial<Omit<Project, "id">>): Promise<Project>;
  deleteProject(id: string): Promise<void>;
  fetchProjectStats(
    projectId: string,
    metric: string,
    interval: string,
    granularity: string,
    timezone: string,
  ): Promise<ProjectStats>;
}

const defaultGeneratedApi = new GeneratedProjectsApi(ApiConfig);

function createProjectsApi(api: GeneratedProjectsApi = defaultGeneratedApi): ProjectsApi {
  return {
    async listProjects(): Promise<Project[]> {
      const response = await api.listProjects({ limit: 50, offset: 0 });
      return mapFromApiArray(projectMapper, response.data);
    },

    async listProjectsPaginated(
      limit: number = 50,
      offset: number = 0,
    ): Promise<PaginatedProjectsResponse> {
      const response = await api.listProjects({ limit, offset });
      return {
        data: mapFromApiArray(projectMapper, response.data),
        pagination: {
          limit: response.pagination.limit,
          offset: response.pagination.offset,
          total: response.pagination.total,
        },
      };
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

    async fetchProjectStats(
      projectId: string,
      metric: string,
      interval: string,
      granularity: string,
      timezone: string,
    ): Promise<ProjectStats> {
      const response = await api.getProjectStats({
        projectId,
        metric: metric as ProjectStatsMetricEnum,
        interval,
        granularity,
        timezone,
      });
      return projectStatsMapper.fromApi(response);
    },
  };
}

export const projectsApi = createProjectsApi();
export const __test__ = { createProjectsApi };

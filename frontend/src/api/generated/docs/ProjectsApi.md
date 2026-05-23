# ProjectsApi

All URIs are relative to */api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createProject**](ProjectsApi.md#createproject) | **POST** /projects | Create project |
| [**deleteProject**](ProjectsApi.md#deleteproject) | **DELETE** /projects/{projectId} | Delete project |
| [**getProject**](ProjectsApi.md#getproject) | **GET** /projects/{projectId} | Get project |
| [**getProjectStats**](ProjectsApi.md#getprojectstats) | **GET** /projects/{projectId}/stats | Get timeseries stats for a project |
| [**listProjects**](ProjectsApi.md#listprojects) | **GET** /projects | List projects |
| [**updateProject**](ProjectsApi.md#updateproject) | **PATCH** /projects/{projectId} | Update project |



## createProject

> Project createProject(createProject)

Create project

### Example

```ts
import {
  Configuration,
  ProjectsApi,
} from '';
import type { CreateProjectRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ProjectsApi();

  const body = {
    // CreateProject
    createProject: ...,
  } satisfies CreateProjectRequest;

  try {
    const data = await api.createProject(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **createProject** | [CreateProject](CreateProject.md) |  | |

### Return type

[**Project**](Project.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Project created |  -  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## deleteProject

> deleteProject(projectId)

Delete project

### Example

```ts
import {
  Configuration,
  ProjectsApi,
} from '';
import type { DeleteProjectRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ProjectsApi();

  const body = {
    // string
    projectId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies DeleteProjectRequest;

  try {
    const data = await api.deleteProject(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **projectId** | `string` |  | [Defaults to `undefined`] |

### Return type

`void` (Empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Deleted |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getProject

> Project getProject(projectId, include)

Get project

### Example

```ts
import {
  Configuration,
  ProjectsApi,
} from '';
import type { GetProjectRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ProjectsApi();

  const body = {
    // string
    projectId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // Set<'totalTimeMs'> | Comma-separated list of optional computed fields to include. Supported values: totalTimeMs  (optional)
    include: ...,
  } satisfies GetProjectRequest;

  try {
    const data = await api.getProject(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **projectId** | `string` |  | [Defaults to `undefined`] |
| **include** | `totalTimeMs` | Comma-separated list of optional computed fields to include. Supported values: totalTimeMs  | [Optional] [Enum: totalTimeMs] |

### Return type

[**Project**](Project.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Project |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getProjectStats

> ProjectStats getProjectStats(projectId, metric, interval, granularity, timezone)

Get timeseries stats for a project

Returns aggregated timeseries data for a given metric on a project. Data is bucketed by the requested interval granularity within the specified time range. 

### Example

```ts
import {
  Configuration,
  ProjectsApi,
} from '';
import type { GetProjectStatsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ProjectsApi();

  const body = {
    // string
    projectId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // 'time_spent' | The metric to aggregate over time.
    metric: time_spent,
    // string | The time range to query as an ISO 8601 interval. Supports all three forms: - `{start}/{end}` — explicit start and end datetimes: `2024-01-01T00:00:00Z/2024-03-31T23:59:59Z` - `{start}/{duration}` — start datetime and a duration: `2024-01-01T00:00:00Z/P3M` - `{duration}/{end}` — a duration ending at a datetime: `P30D/2024-03-31T23:59:59Z` Datetime values must be full RFC 3339 timestamps including timezone (for example `...Z` or `...+01:00`). Duration/duration intervals are not supported. Defaults to `P30D/{now}` (the last 30 days) if omitted.  (optional)
    interval: 2024-01-01T00:00:00Z/2024-03-31T23:59:59Z,
    // string | The bucket size for each data point, expressed as an ISO 8601 duration. Common values: `PT1M` (minute), `PT1H` (hour), `P1D` (day), `P1W` (week), `P1M` (month). Defaults to `P1D`.  (optional)
    granularity: PT1M,
    // string | IANA timezone used for bucketing (e.g. `Europe/Stockholm`). Defaults to UTC. Affects how day/week/month boundaries are computed.  (optional)
    timezone: Europe/Stockholm,
  } satisfies GetProjectStatsRequest;

  try {
    const data = await api.getProjectStats(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **projectId** | `string` |  | [Defaults to `undefined`] |
| **metric** | `time_spent` | The metric to aggregate over time. | [Defaults to `undefined`] [Enum: time_spent] |
| **interval** | `string` | The time range to query as an ISO 8601 interval. Supports all three forms: - &#x60;{start}/{end}&#x60; — explicit start and end datetimes: &#x60;2024-01-01T00:00:00Z/2024-03-31T23:59:59Z&#x60; - &#x60;{start}/{duration}&#x60; — start datetime and a duration: &#x60;2024-01-01T00:00:00Z/P3M&#x60; - &#x60;{duration}/{end}&#x60; — a duration ending at a datetime: &#x60;P30D/2024-03-31T23:59:59Z&#x60; Datetime values must be full RFC 3339 timestamps including timezone (for example &#x60;...Z&#x60; or &#x60;...+01:00&#x60;). Duration/duration intervals are not supported. Defaults to &#x60;P30D/{now}&#x60; (the last 30 days) if omitted.  | [Optional] [Defaults to `undefined`] |
| **granularity** | `string` | The bucket size for each data point, expressed as an ISO 8601 duration. Common values: &#x60;PT1M&#x60; (minute), &#x60;PT1H&#x60; (hour), &#x60;P1D&#x60; (day), &#x60;P1W&#x60; (week), &#x60;P1M&#x60; (month). Defaults to &#x60;P1D&#x60;.  | [Optional] [Defaults to `&#39;P1D&#39;`] |
| **timezone** | `string` | IANA timezone used for bucketing (e.g. &#x60;Europe/Stockholm&#x60;). Defaults to UTC. Affects how day/week/month boundaries are computed.  | [Optional] [Defaults to `&#39;UTC&#39;`] |

### Return type

[**ProjectStats**](ProjectStats.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Timeseries data for the requested metric. |  -  |
| **400** | Invalid query parameters. |  -  |
| **404** | Project not found. |  -  |
| **422** | Unprocessable request — e.g. &#x60;from&#x60; is after &#x60;to&#x60;, or the requested range exceeds the maximum allowed window.  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listProjects

> PaginatedProjects listProjects(limit, offset)

List projects

### Example

```ts
import {
  Configuration,
  ProjectsApi,
} from '';
import type { ListProjectsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ProjectsApi();

  const body = {
    // number | Maximum number of items to return per page. Capped at 100 to prevent resource exhaustion.  (optional)
    limit: 50,
    // number | Number of items to skip from the beginning (zero-indexed). (optional)
    offset: 0,
  } satisfies ListProjectsRequest;

  try {
    const data = await api.listProjects(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **limit** | `number` | Maximum number of items to return per page. Capped at 100 to prevent resource exhaustion.  | [Optional] [Defaults to `25`] |
| **offset** | `number` | Number of items to skip from the beginning (zero-indexed). | [Optional] [Defaults to `0`] |

### Return type

[**PaginatedProjects**](PaginatedProjects.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of projects |  -  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## updateProject

> Project updateProject(projectId, updateProject)

Update project

### Example

```ts
import {
  Configuration,
  ProjectsApi,
} from '';
import type { UpdateProjectRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new ProjectsApi();

  const body = {
    // string
    projectId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // UpdateProject
    updateProject: ...,
  } satisfies UpdateProjectRequest;

  try {
    const data = await api.updateProject(body);
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters


| Name | Type | Description  | Notes |
|------------- | ------------- | ------------- | -------------|
| **projectId** | `string` |  | [Defaults to `undefined`] |
| **updateProject** | [UpdateProject](UpdateProject.md) |  | |

### Return type

[**Project**](Project.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Updated project |  -  |
| **400** | Bad request |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


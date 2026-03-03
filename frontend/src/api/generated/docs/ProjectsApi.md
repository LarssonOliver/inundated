# ProjectsApi

All URIs are relative to */api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createProject**](ProjectsApi.md#createproject) | **POST** /projects | Create project |
| [**deleteProject**](ProjectsApi.md#deleteproject) | **DELETE** /projects/{projectId} | Delete project |
| [**getProject**](ProjectsApi.md#getproject) | **GET** /projects/{projectId} | Get project |
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


## listProjects

> Array&lt;Project&gt; listProjects()

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

  try {
    const data = await api.listProjects();
    console.log(data);
  } catch (error) {
    console.error(error);
  }
}

// Run the test
example().catch(console.error);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**Array&lt;Project&gt;**](Project.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of projects |  -  |

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


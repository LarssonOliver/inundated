# TasksApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createTask**](TasksApi.md#createtask) | **POST** /api/tasks | Create task |
| [**deleteTask**](TasksApi.md#deletetask) | **DELETE** /api/tasks/{taskId} | Delete task |
| [**getTask**](TasksApi.md#gettask) | **GET** /api/tasks/{taskId} | Get task |
| [**listTasks**](TasksApi.md#listtasks) | **GET** /api/tasks | List tasks |
| [**moveTask**](TasksApi.md#movetask) | **POST** /api/tasks/{taskId}/move | Move task |
| [**updateTask**](TasksApi.md#updatetask) | **PATCH** /api/tasks/{taskId} | Update task |



## createTask

> Task createTask(createTask)

Create task

Creates a task along with its task tag. The task is placed last among its siblings. 

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '';
import type { CreateTaskRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TasksApi(config);

  const body = {
    // CreateTask
    createTask: ...,
  } satisfies CreateTaskRequest;

  try {
    const data = await api.createTask(body);
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
| **createTask** | [CreateTask](CreateTask.md) |  | |

### Return type

[**Task**](Task.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Task created |  -  |
| **400** | Bad request, e.g. an unknown parent or tag, or a task tag in tagIds.  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## deleteTask

> deleteTask(taskId)

Delete task

Deletes a task and its subtasks, along with their task tags. Tasks with logged time can\&#39;t be deleted; close them instead. 

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '';
import type { DeleteTaskRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TasksApi(config);

  const body = {
    // string
    taskId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies DeleteTaskRequest;

  try {
    const data = await api.deleteTask(body);
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
| **taskId** | `string` |  | [Defaults to `undefined`] |

### Return type

`void` (Empty response body)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Deleted |  -  |
| **404** | Not found |  -  |
| **409** | The task or one of its subtasks has logged time. |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getTask

> Task getTask(taskId)

Get task

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '';
import type { GetTaskRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TasksApi(config);

  const body = {
    // string
    taskId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies GetTaskRequest;

  try {
    const data = await api.getTask(body);
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
| **taskId** | `string` |  | [Defaults to `undefined`] |

### Return type

[**Task**](Task.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Task |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listTasks

> PaginatedTasks listTasks(limit, offset, includeClosed, parentId, tagId, dueFrom, dueTo)

List tasks

Lists tasks ordered by parent, then by rank among siblings. Closed tasks are hidden unless includeClosed is set. 

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '';
import type { ListTasksRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TasksApi(config);

  const body = {
    // number | Maximum number of items to return per page. Capped at 100 to prevent resource exhaustion.  (optional)
    limit: 50,
    // number | Number of items to skip from the beginning (zero-indexed). (optional)
    offset: 0,
    // boolean | Whether to include closed tasks in the results. Defaults to false.  (optional)
    includeClosed: true,
    // string | Only list direct subtasks of this task. (optional)
    parentId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // string | Only list tasks carrying this regular tag. (optional)
    tagId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // Date | Only list tasks due on or after this day. (optional)
    dueFrom: 2013-10-20,
    // Date | Only list tasks due on or before this day. (optional)
    dueTo: 2013-10-20,
  } satisfies ListTasksRequest;

  try {
    const data = await api.listTasks(body);
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
| **includeClosed** | `boolean` | Whether to include closed tasks in the results. Defaults to false.  | [Optional] [Defaults to `false`] |
| **parentId** | `string` | Only list direct subtasks of this task. | [Optional] [Defaults to `undefined`] |
| **tagId** | `string` | Only list tasks carrying this regular tag. | [Optional] [Defaults to `undefined`] |
| **dueFrom** | `Date` | Only list tasks due on or after this day. | [Optional] [Defaults to `undefined`] |
| **dueTo** | `Date` | Only list tasks due on or before this day. | [Optional] [Defaults to `undefined`] |

### Return type

[**PaginatedTasks**](PaginatedTasks.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of tasks |  -  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## moveTask

> Task moveTask(taskId, moveTask)

Move task

Moves a task to a new position, optionally under a new parent. The task keeps its subtasks. 

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '';
import type { MoveTaskRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TasksApi(config);

  const body = {
    // string
    taskId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // MoveTask
    moveTask: ...,
  } satisfies MoveTaskRequest;

  try {
    const data = await api.moveTask(body);
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
| **taskId** | `string` |  | [Defaults to `undefined`] |
| **moveTask** | [MoveTask](MoveTask.md) |  | |

### Return type

[**Task**](Task.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Moved task |  -  |
| **400** | Bad request, e.g. a parent that is the task itself or one of its subtasks, or an afterTaskId that isn\&#39;t a sibling under the new parent.  |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## updateTask

> Task updateTask(taskId, updateTask)

Update task

### Example

```ts
import {
  Configuration,
  TasksApi,
} from '';
import type { UpdateTaskRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TasksApi(config);

  const body = {
    // string
    taskId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // UpdateTask
    updateTask: ...,
  } satisfies UpdateTaskRequest;

  try {
    const data = await api.updateTask(body);
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
| **taskId** | `string` |  | [Defaults to `undefined`] |
| **updateTask** | [UpdateTask](UpdateTask.md) |  | |

### Return type

[**Task**](Task.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Updated task |  -  |
| **400** | Bad request |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


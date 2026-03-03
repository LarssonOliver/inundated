# TagsApi

All URIs are relative to */api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createTag**](TagsApi.md#createtag) | **POST** /tags | Create tag |
| [**deleteTag**](TagsApi.md#deletetag) | **DELETE** /tags/{tagId} | Delete tag |
| [**getTag**](TagsApi.md#gettag) | **GET** /tags/{tagId} | Get tag |
| [**listTags**](TagsApi.md#listtags) | **GET** /tags | List tags |
| [**updateTag**](TagsApi.md#updatetag) | **PATCH** /tags/{tagId} | Update tag |



## createTag

> Tag createTag(createTag)

Create tag

### Example

```ts
import {
  Configuration,
  TagsApi,
} from '';
import type { CreateTagRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TagsApi();

  const body = {
    // CreateTag
    createTag: ...,
  } satisfies CreateTagRequest;

  try {
    const data = await api.createTag(body);
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
| **createTag** | [CreateTag](CreateTag.md) |  | |

### Return type

[**Tag**](Tag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Tag created |  -  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## deleteTag

> deleteTag(tagId)

Delete tag

### Example

```ts
import {
  Configuration,
  TagsApi,
} from '';
import type { DeleteTagRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TagsApi();

  const body = {
    // string
    tagId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies DeleteTagRequest;

  try {
    const data = await api.deleteTag(body);
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
| **tagId** | `string` |  | [Defaults to `undefined`] |

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


## getTag

> Tag getTag(tagId, include)

Get tag

### Example

```ts
import {
  Configuration,
  TagsApi,
} from '';
import type { GetTagRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TagsApi();

  const body = {
    // string
    tagId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // Set<'totalTimeMs'> | Comma-separated list of optional computed fields to include. Supported values: totalTimeMs  (optional)
    include: ...,
  } satisfies GetTagRequest;

  try {
    const data = await api.getTag(body);
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
| **tagId** | `string` |  | [Defaults to `undefined`] |
| **include** | `totalTimeMs` | Comma-separated list of optional computed fields to include. Supported values: totalTimeMs  | [Optional] [Enum: totalTimeMs] |

### Return type

[**Tag**](Tag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Tag |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listTags

> Array&lt;Tag&gt; listTags()

List tags

### Example

```ts
import {
  Configuration,
  TagsApi,
} from '';
import type { ListTagsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TagsApi();

  try {
    const data = await api.listTags();
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

[**Array&lt;Tag&gt;**](Tag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of tags |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## updateTag

> Tag updateTag(tagId, updateTag)

Update tag

### Example

```ts
import {
  Configuration,
  TagsApi,
} from '';
import type { UpdateTagRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TagsApi();

  const body = {
    // string
    tagId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // UpdateTag
    updateTag: ...,
  } satisfies UpdateTagRequest;

  try {
    const data = await api.updateTag(body);
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
| **tagId** | `string` |  | [Defaults to `undefined`] |
| **updateTag** | [UpdateTag](UpdateTag.md) |  | |

### Return type

[**Tag**](Tag.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Updated tag |  -  |
| **400** | Bad request |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


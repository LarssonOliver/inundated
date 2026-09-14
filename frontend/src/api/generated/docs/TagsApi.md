# TagsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createTag**](TagsApi.md#createtag) | **POST** /api/tags | Create tag |
| [**deleteTag**](TagsApi.md#deletetag) | **DELETE** /api/tags/{tagId} | Delete tag |
| [**getTag**](TagsApi.md#gettag) | **GET** /api/tags/{tagId} | Get tag |
| [**getTagStats**](TagsApi.md#gettagstats) | **GET** /api/tags/{tagId}/stats | Get timeseries stats for a tag |
| [**listTags**](TagsApi.md#listtags) | **GET** /api/tags | List tags |
| [**updateTag**](TagsApi.md#updatetag) | **PATCH** /api/tags/{tagId} | Update tag |



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
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TagsApi(config);

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

[sessionCookie](../README.md#sessionCookie)

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
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TagsApi(config);

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

[sessionCookie](../README.md#sessionCookie)

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
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TagsApi(config);

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

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Tag |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getTagStats

> TagStats getTagStats(tagId, metric, interval, granularity, timezone)

Get timeseries stats for a tag

Returns aggregated timeseries data for a given metric on a tag. Data is bucketed by the requested interval granularity within the specified time range. 

### Example

```ts
import {
  Configuration,
  TagsApi,
} from '';
import type { GetTagStatsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TagsApi(config);

  const body = {
    // string
    tagId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // StatsMetric | The metric to aggregate over time.
    metric: time_spent,
    // string | The time range to query as an ISO 8601 interval. Supports all three forms: - `{start}/{end}` — explicit start and end datetimes: `2024-01-01T00:00:00Z/2024-03-31T23:59:59Z` - `{start}/{duration}` — start datetime and a duration: `2024-01-01T00:00:00Z/P3M` - `{duration}/{end}` — a duration ending at a datetime: `P30D/2024-03-31T23:59:59Z` Datetime values must be full RFC 3339 timestamps including timezone (for example `...Z` or `...+01:00`). Duration/duration intervals are not supported. Defaults to `P30D/{now}` (the last 30 days) if omitted.  (optional)
    interval: 2024-01-01T00:00:00Z/2024-03-31T23:59:59Z,
    // string | The bucket size for each data point, expressed as an ISO 8601 duration. Common values: `PT1M` (minute), `PT1H` (hour), `P1D` (day), `P1W` (week), `P1M` (month). Defaults to `P1D`.  (optional)
    granularity: PT1M,
    // string | IANA timezone used for bucketing (e.g. `Europe/Stockholm`). Defaults to UTC. Affects how day/week/month boundaries are computed.  (optional)
    timezone: Europe/Stockholm,
  } satisfies GetTagStatsRequest;

  try {
    const data = await api.getTagStats(body);
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
| **metric** | `StatsMetric` | The metric to aggregate over time. | [Defaults to `undefined`] [Enum: time_spent] |
| **interval** | `string` | The time range to query as an ISO 8601 interval. Supports all three forms: - &#x60;{start}/{end}&#x60; — explicit start and end datetimes: &#x60;2024-01-01T00:00:00Z/2024-03-31T23:59:59Z&#x60; - &#x60;{start}/{duration}&#x60; — start datetime and a duration: &#x60;2024-01-01T00:00:00Z/P3M&#x60; - &#x60;{duration}/{end}&#x60; — a duration ending at a datetime: &#x60;P30D/2024-03-31T23:59:59Z&#x60; Datetime values must be full RFC 3339 timestamps including timezone (for example &#x60;...Z&#x60; or &#x60;...+01:00&#x60;). Duration/duration intervals are not supported. Defaults to &#x60;P30D/{now}&#x60; (the last 30 days) if omitted.  | [Optional] [Defaults to `undefined`] |
| **granularity** | `string` | The bucket size for each data point, expressed as an ISO 8601 duration. Common values: &#x60;PT1M&#x60; (minute), &#x60;PT1H&#x60; (hour), &#x60;P1D&#x60; (day), &#x60;P1W&#x60; (week), &#x60;P1M&#x60; (month). Defaults to &#x60;P1D&#x60;.  | [Optional] [Defaults to `&#39;P1D&#39;`] |
| **timezone** | `string` | IANA timezone used for bucketing (e.g. &#x60;Europe/Stockholm&#x60;). Defaults to UTC. Affects how day/week/month boundaries are computed.  | [Optional] [Defaults to `&#39;UTC&#39;`] |

### Return type

[**TagStats**](TagStats.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Timeseries data for the requested metric. |  -  |
| **400** | Invalid query parameters. |  -  |
| **404** | Tag not found. |  -  |
| **422** | Unprocessable request — e.g. &#x60;from&#x60; is after &#x60;to&#x60;, or the requested range exceeds the maximum allowed window.  |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listTags

> PaginatedTags listTags(limit, offset, includeArchived)

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
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TagsApi(config);

  const body = {
    // number | Maximum number of items to return per page. Capped at 100 to prevent resource exhaustion.  (optional)
    limit: 50,
    // number | Number of items to skip from the beginning (zero-indexed). (optional)
    offset: 0,
    // boolean | Whether to include archived items in the results. Defaults to false, so archived items are hidden unless explicitly requested.  (optional)
    includeArchived: true,
  } satisfies ListTagsRequest;

  try {
    const data = await api.listTags(body);
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
| **includeArchived** | `boolean` | Whether to include archived items in the results. Defaults to false, so archived items are hidden unless explicitly requested.  | [Optional] [Defaults to `false`] |

### Return type

[**PaginatedTags**](PaginatedTags.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of tags |  -  |
| **400** | Bad request |  -  |

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
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TagsApi(config);

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

[sessionCookie](../README.md#sessionCookie)

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


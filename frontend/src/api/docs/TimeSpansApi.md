# TimeSpansApi

All URIs are relative to */api*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createTimeSpan**](TimeSpansApi.md#createtimespan) | **POST** /time-spans | Create time span |
| [**deleteTimeSpan**](TimeSpansApi.md#deletetimespan) | **DELETE** /time-spans/{timeSpanId} | Delete time span |
| [**getTimeSpan**](TimeSpansApi.md#gettimespan) | **GET** /time-spans/{timeSpanId} | Get time span |
| [**listTimeSpans**](TimeSpansApi.md#listtimespans) | **GET** /time-spans | List time spans |
| [**updateTimeSpan**](TimeSpansApi.md#updatetimespan) | **PATCH** /time-spans/{timeSpanId} | Update time span |



## createTimeSpan

> TimeSpan createTimeSpan(createTimeSpan)

Create time span

### Example

```ts
import {
  Configuration,
  TimeSpansApi,
} from '';
import type { CreateTimeSpanRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TimeSpansApi();

  const body = {
    // CreateTimeSpan
    createTimeSpan: ...,
  } satisfies CreateTimeSpanRequest;

  try {
    const data = await api.createTimeSpan(body);
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
| **createTimeSpan** | [CreateTimeSpan](CreateTimeSpan.md) |  | |

### Return type

[**TimeSpan**](TimeSpan.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Time span created |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## deleteTimeSpan

> deleteTimeSpan(timeSpanId)

Delete time span

### Example

```ts
import {
  Configuration,
  TimeSpansApi,
} from '';
import type { DeleteTimeSpanRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TimeSpansApi();

  const body = {
    // string
    timeSpanId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies DeleteTimeSpanRequest;

  try {
    const data = await api.deleteTimeSpan(body);
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
| **timeSpanId** | `string` |  | [Defaults to `undefined`] |

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

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getTimeSpan

> TimeSpan getTimeSpan(timeSpanId)

Get time span

### Example

```ts
import {
  Configuration,
  TimeSpansApi,
} from '';
import type { GetTimeSpanRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TimeSpansApi();

  const body = {
    // string
    timeSpanId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies GetTimeSpanRequest;

  try {
    const data = await api.getTimeSpan(body);
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
| **timeSpanId** | `string` |  | [Defaults to `undefined`] |

### Return type

[**TimeSpan**](TimeSpan.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Time span |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listTimeSpans

> Array&lt;TimeSpan&gt; listTimeSpans()

List time spans

### Example

```ts
import {
  Configuration,
  TimeSpansApi,
} from '';
import type { ListTimeSpansRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TimeSpansApi();

  try {
    const data = await api.listTimeSpans();
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

[**Array&lt;TimeSpan&gt;**](TimeSpan.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | List of time spans |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## updateTimeSpan

> TimeSpan updateTimeSpan(timeSpanId, updateTimeSpan)

Update time span

### Example

```ts
import {
  Configuration,
  TimeSpansApi,
} from '';
import type { UpdateTimeSpanRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new TimeSpansApi();

  const body = {
    // string
    timeSpanId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // UpdateTimeSpan
    updateTimeSpan: ...,
  } satisfies UpdateTimeSpanRequest;

  try {
    const data = await api.updateTimeSpan(body);
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
| **timeSpanId** | `string` |  | [Defaults to `undefined`] |
| **updateTimeSpan** | [UpdateTimeSpan](UpdateTimeSpan.md) |  | |

### Return type

[**TimeSpan**](TimeSpan.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Updated time span |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


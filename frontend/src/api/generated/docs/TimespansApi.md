# TimespansApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createTimespan**](TimespansApi.md#createtimespan) | **POST** /api/timespans | Create time span |
| [**deleteTimespan**](TimespansApi.md#deletetimespan) | **DELETE** /api/timespans/{timespanId} | Delete time span |
| [**getTimespan**](TimespansApi.md#gettimespan) | **GET** /api/timespans/{timespanId} | Get time span |
| [**listTimespans**](TimespansApi.md#listtimespans) | **GET** /api/timespans | List time spans |
| [**updateTimespan**](TimespansApi.md#updatetimespan) | **PATCH** /api/timespans/{timespanId} | Update time span |



## createTimespan

> Timespan createTimespan(createTimespan)

Create time span

### Example

```ts
import {
  Configuration,
  TimespansApi,
} from '';
import type { CreateTimespanRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
    // To configure API key authorization: xsrfToken
    apiKey: "YOUR API KEY",
  });
  const api = new TimespansApi(config);

  const body = {
    // CreateTimespan
    createTimespan: ...,
  } satisfies CreateTimespanRequest;

  try {
    const data = await api.createTimespan(body);
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
| **createTimespan** | [CreateTimespan](CreateTimespan.md) |  | |

### Return type

[**Timespan**](Timespan.md)

### Authorization

[sessionCookie](../README.md#sessionCookie), [xsrfToken](../README.md#xsrfToken)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | Time span created |  -  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## deleteTimespan

> deleteTimespan(timespanId)

Delete time span

### Example

```ts
import {
  Configuration,
  TimespansApi,
} from '';
import type { DeleteTimespanRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
    // To configure API key authorization: xsrfToken
    apiKey: "YOUR API KEY",
  });
  const api = new TimespansApi(config);

  const body = {
    // string
    timespanId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies DeleteTimespanRequest;

  try {
    const data = await api.deleteTimespan(body);
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
| **timespanId** | `string` |  | [Defaults to `undefined`] |

### Return type

`void` (Empty response body)

### Authorization

[sessionCookie](../README.md#sessionCookie), [xsrfToken](../README.md#xsrfToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Deleted |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## getTimespan

> Timespan getTimespan(timespanId)

Get time span

### Example

```ts
import {
  Configuration,
  TimespansApi,
} from '';
import type { GetTimespanRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TimespansApi(config);

  const body = {
    // string
    timespanId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies GetTimespanRequest;

  try {
    const data = await api.getTimespan(body);
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
| **timespanId** | `string` |  | [Defaults to `undefined`] |

### Return type

[**Timespan**](Timespan.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Time span |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listTimespans

> PaginatedTimespans listTimespans(limit, offset)

List time spans

### Example

```ts
import {
  Configuration,
  TimespansApi,
} from '';
import type { ListTimespansRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new TimespansApi(config);

  const body = {
    // number | Maximum number of items to return per page. Capped at 100 to prevent resource exhaustion.  (optional)
    limit: 50,
    // number | Number of items to skip from the beginning (zero-indexed). (optional)
    offset: 0,
  } satisfies ListTimespansRequest;

  try {
    const data = await api.listTimespans(body);
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

[**PaginatedTimespans**](PaginatedTimespans.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Paginated list of time spans |  -  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## updateTimespan

> Timespan updateTimespan(timespanId, updateTimespan)

Update time span

### Example

```ts
import {
  Configuration,
  TimespansApi,
} from '';
import type { UpdateTimespanRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
    // To configure API key authorization: xsrfToken
    apiKey: "YOUR API KEY",
  });
  const api = new TimespansApi(config);

  const body = {
    // string
    timespanId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
    // UpdateTimespan
    updateTimespan: ...,
  } satisfies UpdateTimespanRequest;

  try {
    const data = await api.updateTimespan(body);
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
| **timespanId** | `string` |  | [Defaults to `undefined`] |
| **updateTimespan** | [UpdateTimespan](UpdateTimespan.md) |  | |

### Return type

[**Timespan**](Timespan.md)

### Authorization

[sessionCookie](../README.md#sessionCookie), [xsrfToken](../README.md#xsrfToken)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Updated time span |  -  |
| **400** | Bad request |  -  |
| **404** | Not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


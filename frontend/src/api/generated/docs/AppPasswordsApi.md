# AppPasswordsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**createAppPassword**](AppPasswordsApi.md#createapppassword) | **POST** /api/app-passwords | Create app password |
| [**deleteAppPassword**](AppPasswordsApi.md#deleteapppassword) | **DELETE** /api/app-passwords/{appPasswordId} | Delete app password |
| [**listAppPasswords**](AppPasswordsApi.md#listapppasswords) | **GET** /api/app-passwords | List app passwords |



## createAppPassword

> CreatedAppPassword createAppPassword(createAppPassword)

Create app password

### Example

```ts
import {
  Configuration,
  AppPasswordsApi,
} from '';
import type { CreateAppPasswordRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new AppPasswordsApi(config);

  const body = {
    // CreateAppPassword
    createAppPassword: ...,
  } satisfies CreateAppPasswordRequest;

  try {
    const data = await api.createAppPassword(body);
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
| **createAppPassword** | [CreateAppPassword](CreateAppPassword.md) |  | |

### Return type

[**CreatedAppPassword**](CreatedAppPassword.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **201** | App password created |  -  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## deleteAppPassword

> deleteAppPassword(appPasswordId)

Delete app password

### Example

```ts
import {
  Configuration,
  AppPasswordsApi,
} from '';
import type { DeleteAppPasswordRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new AppPasswordsApi(config);

  const body = {
    // string
    appPasswordId: 38400000-8cf0-11bd-b23e-10b96e4ef00d,
  } satisfies DeleteAppPasswordRequest;

  try {
    const data = await api.deleteAppPassword(body);
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
| **appPasswordId** | `string` |  | [Defaults to `undefined`] |

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
| **204** | App password deleted |  -  |
| **404** | App password not found |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## listAppPasswords

> Array&lt;AppPassword&gt; listAppPasswords()

List app passwords

### Example

```ts
import {
  Configuration,
  AppPasswordsApi,
} from '';
import type { ListAppPasswordsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new AppPasswordsApi(config);

  try {
    const data = await api.listAppPasswords();
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

[**Array&lt;AppPassword&gt;**](AppPassword.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | App passwords of the current scope, without their secrets |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


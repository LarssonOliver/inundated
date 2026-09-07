# UsersApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getCurrentUser**](UsersApi.md#getcurrentuser) | **GET** /api/me | Get current user |
| [**updateCurrentUser**](UsersApi.md#updatecurrentuser) | **PUT** /api/me | Update current user |



## getCurrentUser

> User getCurrentUser()

Get current user

### Example

```ts
import {
  Configuration,
  UsersApi,
} from '';
import type { GetCurrentUserRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new UsersApi(config);

  try {
    const data = await api.getCurrentUser();
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

[**User**](User.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Current authenticated user |  -  |
| **401** | Unauthorized - missing or invalid token |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## updateCurrentUser

> User updateCurrentUser(xXSRFTOKEN, updateUser)

Update current user

### Example

```ts
import {
  Configuration,
  UsersApi,
} from '';
import type { UpdateCurrentUserRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new UsersApi(config);

  const body = {
    // string | Anti-CSRF token extracted from the XSRF-TOKEN cookie.
    xXSRFTOKEN: xXSRFTOKEN_example,
    // UpdateUser
    updateUser: ...,
  } satisfies UpdateCurrentUserRequest;

  try {
    const data = await api.updateCurrentUser(body);
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
| **xXSRFTOKEN** | `string` | Anti-CSRF token extracted from the XSRF-TOKEN cookie. | [Defaults to `undefined`] |
| **updateUser** | [UpdateUser](UpdateUser.md) |  | |

### Return type

[**User**](User.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Updated user |  -  |
| **400** | Bad request |  -  |
| **401** | Unauthorized - missing or invalid token |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


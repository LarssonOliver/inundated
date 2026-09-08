# AuthApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**authCallback**](AuthApi.md#authcallback) | **GET** /api/auth/callback | OIDC callback |
| [**authLogin**](AuthApi.md#authlogin) | **GET** /api/auth/login | Initiate OIDC login |
| [**authLogout**](AuthApi.md#authlogout) | **POST** /api/auth/logout | Log out |



## authCallback

> authCallback(code, state, inundatedLogin)

OIDC callback

Callback endpoint invoked by the identity provider after the user has authenticated. Exchanges the authorization code for tokens, establishes a session, and redirects the user back to the application. 

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { AuthCallbackRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // string
    code: code_example,
    // string
    state: state_example,
    // string | Browser-binding token planted as an HttpOnly cookie by the login redirect. The callback only establishes a session when this matches the `state` query parameter, so an attacker cannot complete their own authorization in a victim\'s browser (login CSRF / session fixation).  (optional)
    inundatedLogin: inundatedLogin_example,
  } satisfies AuthCallbackRequest;

  try {
    const data = await api.authCallback(body);
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
| **code** | `string` |  | [Defaults to `undefined`] |
| **state** | `string` |  | [Defaults to `undefined`] |
| **inundatedLogin** | `string` | Browser-binding token planted as an HttpOnly cookie by the login redirect. The callback only establishes a session when this matches the &#x60;state&#x60; query parameter, so an attacker cannot complete their own authorization in a victim\&#39;s browser (login CSRF / session fixation).  | [Optional] [Defaults to `undefined`] |

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
| **200** | This endpoint is part of the login flow and typically redirects the user to the identity provider. A 200 response is not expected in normal browser usage.  |  -  |
| **302** | Session established and user redirected. |  * Set-Cookie - Session cookie. <br>  * Location - Redirect target. <br>  |
| **400** | Bad request |  -  |
| **401** | Unaurhotrized |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## authLogin

> authLogin(redirect)

Initiate OIDC login

Starts the OpenID Connect authorization flow by redirecting the user to the configured identity provider. 

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { AuthLoginRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const api = new AuthApi();

  const body = {
    // string | Optional application-relative path to return to after successful authentication.  (optional)
    redirect: redirect_example,
  } satisfies AuthLoginRequest;

  try {
    const data = await api.authLogin(body);
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
| **redirect** | `string` | Optional application-relative path to return to after successful authentication.  | [Optional] [Defaults to `undefined`] |

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
| **200** | This endpoint initiates the login flow and typically redirects the user to the identity provider. A 200 response is not expected in normal browser usage.  |  -  |
| **302** | Redirect to the identity provider. |  * Location - Authorization endpoint. <br>  * Set-Cookie - HttpOnly browser-binding cookie echoed back by the callback to defeat login CSRF / session fixation.  <br>  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## authLogout

> authLogout()

Log out

### Example

```ts
import {
  Configuration,
  AuthApi,
} from '';
import type { AuthLogoutRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
    // To configure API key authorization: xsrfToken
    apiKey: "YOUR API KEY",
  });
  const api = new AuthApi(config);

  try {
    const data = await api.authLogout();
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

`void` (Empty response body)

### Authorization

[sessionCookie](../README.md#sessionCookie), [xsrfToken](../README.md#xsrfToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: Not defined


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **204** | Session terminated |  * Set-Cookie - Clears the session cookie. <br>  |
| **302** | Session terminated and browser redirected |  -  |
| **401** | Unaurhotrized |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


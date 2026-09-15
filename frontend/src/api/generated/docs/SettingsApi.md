# SettingsApi

All URIs are relative to *http://localhost*

| Method | HTTP request | Description |
|------------- | ------------- | -------------|
| [**getSettings**](SettingsApi.md#getsettings) | **GET** /api/settings | Get current settings |
| [**updateSettings**](SettingsApi.md#updatesettings) | **PATCH** /api/settings | Update settings |



## getSettings

> Settings getSettings()

Get current settings

Returns the settings for the current scope (the authenticated user, or the shared unowned settings in userless mode). If none exist yet, a default row is created on first access. 

### Example

```ts
import {
  Configuration,
  SettingsApi,
} from '';
import type { GetSettingsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new SettingsApi(config);

  try {
    const data = await api.getSettings();
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

[**Settings**](Settings.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Current settings |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


## updateSettings

> Settings updateSettings(updateSettings)

Update settings

### Example

```ts
import {
  Configuration,
  SettingsApi,
} from '';
import type { UpdateSettingsRequest } from '';

async function example() {
  console.log("🚀 Testing  SDK...");
  const config = new Configuration({ 
    // To configure API key authorization: sessionCookie
    apiKey: "YOUR API KEY",
  });
  const api = new SettingsApi(config);

  const body = {
    // UpdateSettings
    updateSettings: ...,
  } satisfies UpdateSettingsRequest;

  try {
    const data = await api.updateSettings(body);
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
| **updateSettings** | [UpdateSettings](UpdateSettings.md) |  | |

### Return type

[**Settings**](Settings.md)

### Authorization

[sessionCookie](../README.md#sessionCookie)

### HTTP request headers

- **Content-Type**: `application/json`
- **Accept**: `application/json`


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
| **200** | Updated settings |  -  |
| **400** | Bad request |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)


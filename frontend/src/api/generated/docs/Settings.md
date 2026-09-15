
# Settings


## Properties

Name | Type
------------ | -------------
`weekStartDay` | [WeekStartDay](WeekStartDay.md)
`timezone` | string
`durationFormat` | [DurationFormat](DurationFormat.md)
`timeFormat` | [TimeFormat](TimeFormat.md)

## Example

```typescript
import type { Settings } from ''

// TODO: Update the object below with actual values
const example = {
  "weekStartDay": null,
  "timezone": null,
  "durationFormat": null,
  "timeFormat": null,
} satisfies Settings

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as Settings
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



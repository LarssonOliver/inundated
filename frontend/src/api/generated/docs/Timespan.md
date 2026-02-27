
# Timespan


## Properties

Name | Type
------------ | -------------
`id` | string
`name` | string
`startTime` | Date
`endTime` | Date
`tagIds` | Set&lt;string&gt;

## Example

```typescript
import type { Timespan } from ''

// TODO: Update the object below with actual values
const example = {
  "id": null,
  "name": null,
  "startTime": null,
  "endTime": null,
  "tagIds": null,
} satisfies Timespan

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as Timespan
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



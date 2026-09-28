
# UpdateTask


## Properties

Name | Type
------------ | -------------
`name` | string
`tagIds` | Set&lt;string&gt;
`dueDate` | Date
`clearDueDate` | boolean
`estimateHours` | number
`clearEstimate` | boolean
`closed` | boolean
`closeReason` | [CloseReason](CloseReason.md)

## Example

```typescript
import type { UpdateTask } from ''

// TODO: Update the object below with actual values
const example = {
  "name": null,
  "tagIds": null,
  "dueDate": null,
  "clearDueDate": null,
  "estimateHours": null,
  "clearEstimate": null,
  "closed": null,
  "closeReason": null,
} satisfies UpdateTask

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as UpdateTask
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



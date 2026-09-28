
# Task


## Properties

Name | Type
------------ | -------------
`id` | string
`name` | string
`tagId` | string
`parentId` | string
`tagIds` | Set&lt;string&gt;
`dueDate` | Date
`estimateHours` | number
`rank` | string
`closed` | boolean
`closeReason` | [CloseReason](CloseReason.md)
`closedAt` | Date

## Example

```typescript
import type { Task } from ''

// TODO: Update the object below with actual values
const example = {
  "id": null,
  "name": null,
  "tagId": null,
  "parentId": null,
  "tagIds": null,
  "dueDate": null,
  "estimateHours": null,
  "rank": null,
  "closed": null,
  "closeReason": null,
  "closedAt": null,
} satisfies Task

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as Task
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



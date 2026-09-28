
# CreateTask


## Properties

Name | Type
------------ | -------------
`name` | string
`parentId` | string
`tagIds` | Set&lt;string&gt;
`dueDate` | Date
`estimateHours` | number

## Example

```typescript
import type { CreateTask } from ''

// TODO: Update the object below with actual values
const example = {
  "name": null,
  "parentId": null,
  "tagIds": null,
  "dueDate": null,
  "estimateHours": null,
} satisfies CreateTask

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as CreateTask
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



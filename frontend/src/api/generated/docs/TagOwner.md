
# TagOwner

Set on owned tags, naming the item that owns the tag, such as the task that owns a task tag. An owned tag\'s name, color and archived state follow its owner, so it can\'t be changed through the tag API.

## Properties

Name | Type
------------ | -------------
`kind` | [TagOwnerKind](TagOwnerKind.md)
`id` | string

## Example

```typescript
import type { TagOwner } from ''

// TODO: Update the object below with actual values
const example = {
  "kind": null,
  "id": null,
} satisfies TagOwner

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as TagOwner
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



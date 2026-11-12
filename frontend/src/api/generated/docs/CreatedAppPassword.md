
# CreatedAppPassword

A newly created app password. The password is only returned here and cannot be retrieved again.

## Properties

Name | Type
------------ | -------------
`id` | string
`name` | string
`createdAt` | Date
`lastUsedAt` | Date
`password` | string

## Example

```typescript
import type { CreatedAppPassword } from ''

// TODO: Update the object below with actual values
const example = {
  "id": null,
  "name": null,
  "createdAt": null,
  "lastUsedAt": null,
  "password": null,
} satisfies CreatedAppPassword

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as CreatedAppPassword
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



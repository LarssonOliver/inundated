
# SeriesPoint


## Properties

Name | Type
------------ | -------------
`interval` | string
`value` | number

## Example

```typescript
import type { SeriesPoint } from ''

// TODO: Update the object below with actual values
const example = {
  "interval": 2024-01-01T00:00:00Z/2024-01-08T00:00:00Z,
  "value": null,
} satisfies SeriesPoint

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as SeriesPoint
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



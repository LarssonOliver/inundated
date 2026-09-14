
# TagStats


## Properties

Name | Type
------------ | -------------
`tagId` | string
`metric` | [StatsMetric](StatsMetric.md)
`interval` | string
`granularity` | string
`unit` | string
`series` | [Array&lt;SeriesPoint&gt;](SeriesPoint.md)

## Example

```typescript
import type { TagStats } from ''

// TODO: Update the object below with actual values
const example = {
  "tagId": null,
  "metric": null,
  "interval": 2024-01-01T00:00:00Z/2024-03-31T23:59:59Z,
  "granularity": P1W,
  "unit": seconds,
  "series": null,
} satisfies TagStats

console.log(example)

// Convert the instance to a JSON string
const exampleJSON: string = JSON.stringify(example)
console.log(exampleJSON)

// Parse the JSON string back to an object
const exampleParsed = JSON.parse(exampleJSON) as TagStats
console.log(exampleParsed)
```

[[Back to top]](#) [[Back to API list]](../README.md#api-endpoints) [[Back to Model list]](../README.md#models) [[Back to README]](../README.md)



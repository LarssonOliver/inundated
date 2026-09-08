package utils

import "github.com/google/uuid"

// DedupeUUIDs returns ids with duplicates removed, preserving first-seen order.
// A tag-id list on a project or timespan is a set: a caller may repeat an id,
// and the repositories must normalise it the same way regardless of backend
// (Postgres counts distinct rows, the in-memory store loops per id).
func DedupeUUIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) < 2 {
		return ids
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

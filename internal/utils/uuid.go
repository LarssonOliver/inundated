package utils

import "github.com/google/uuid"

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

// SameUUIDSet reports whether a and b hold the same ids, ignoring order and
// duplicates.
func SameUUIDSet(a, b []uuid.UUID) bool {
	inA := make(map[uuid.UUID]bool, len(a))
	for _, id := range a {
		inA[id] = true
	}
	inB := make(map[uuid.UUID]bool, len(b))
	for _, id := range b {
		if !inA[id] {
			return false
		}
		inB[id] = true
	}
	return len(inA) == len(inB)
}

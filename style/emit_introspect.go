//go:build !wasm

package style

import (
	"sort"

	"webtyp.com/fmt"
	"webtyp.com/widget"
)

func (s *Sheet) Parts() []widget.Part {
	var parts []widget.Part
	for p := range s.partRules {
		parts = append(parts, p)
	}
	sort.Slice(parts, func(i, j int) bool {
		return parts[i] < parts[j]
	})
	return parts
}

func (s *Sheet) StateAttrs() []fmt.KeyValue {
	seen := make(map[string]bool)
	var result []fmt.KeyValue

	collect := func(st widget.State) {
		kv := st.Attr()
		key := kv.Key() + "=" + kv.Value()
		if !seen[key] {
			seen[key] = true
			result = append(result, fmt.KeyValue{Key: kv.Key(), Value: kv.Value()})
		}
	}

	if s.rootRule.hasRevealed {
		collect(s.rootRule.revealedBy)
	}
	for _, pr := range s.partRules {
		if pr.hasRevealed {
			collect(pr.revealedBy)
		}
	}
	for k := range s.stateRules {
		collect(k.state)
	}
	// A state reached through an ancestor (WhenWithin) still has to be written
	// by the markup — the attribute simply lands on the container instead of
	// on the part being repainted, which does not make it any less of a state
	// this sheet depends on.
	for k := range s.stateWithin {
		collect(k.state)
	}
	for _, dr := range s.deviceRules {
		if dr.hasRevealed {
			collect(dr.revealedBy)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		return result[i].Value < result[j].Value
	})
	return result
}

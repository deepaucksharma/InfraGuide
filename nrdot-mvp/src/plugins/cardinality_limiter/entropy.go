package cardinalitylimiter

import (
	"math"
	"sort"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

// EntropyCalculator calculates the entropy score for attribute sets.
type EntropyCalculator struct {
	// Historical data for calculating entropy
	labelValues map[string]map[string]int // Maps label name -> value -> count
	totalCount  int
}

// NewEntropyCalculator creates a new entropy calculator.
func NewEntropyCalculator() *EntropyCalculator {
	return &EntropyCalculator{
		labelValues: make(map[string]map[string]int),
		totalCount:  0,
	}
}

// AddLabelSet adds a set of labels to the historical data.
func (e *EntropyCalculator) AddLabelSet(labelSet map[string]string) {
	e.totalCount++
	
	for name, value := range labelSet {
		if _, exists := e.labelValues[name]; !exists {
			e.labelValues[name] = make(map[string]int)
		}
		
		e.labelValues[name][value]++
	}
}

// AddAttributes adds a set of attributes to the historical data.
func (e *EntropyCalculator) AddAttributes(attrs pcommon.Map) {
	labelSet := attributesToMap(attrs)
	e.AddLabelSet(labelSet)
}

// CalculateEntropyScore calculates an entropy-based score for a label set.
// Higher scores mean more important (should be kept).
func (e *EntropyCalculator) CalculateEntropyScore(labelSet map[string]string) float64 {
	if e.totalCount == 0 {
		return 0
	}
	
	// Calculate information content of each label based on historical data
	labelScores := make(map[string]float64)
	for name, value := range labelSet {
		valueMap, exists := e.labelValues[name]
		if !exists {
			// New label name, high entropy
			labelScores[name] = 1.0
			continue
		}
		
		count, exists := valueMap[value]
		if !exists {
			// New value for this label, high entropy
			labelScores[name] = 1.0
			continue
		}
		
		// Calculate probability of this value occurring
		probability := float64(count) / float64(e.totalCount)
		
		// Calculate entropy (information content) of this label
		// Rare values have higher entropy (more information)
		entropy := -math.Log2(probability)
		
		// Normalize to 0-1 range
		normalizedEntropy := math.Min(1.0, entropy/16.0) // Cap at 16 bits of entropy
		
		labelScores[name] = normalizedEntropy
	}
	
	// Calculate the average entropy score across all labels
	var totalScore float64
	for _, score := range labelScores {
		totalScore += score
	}
	
	// Also consider the number of labels as a factor
	// More labels might indicate more specificity
	labelCount := float64(len(labelSet))
	labelCountFactor := math.Min(1.0, labelCount/10.0) // Normalize to 0-1 range, cap at 10 labels
	
	// Combine both factors
	if len(labelScores) > 0 {
		averageScore := totalScore / float64(len(labelScores))
		return averageScore * (0.8 + 0.2*labelCountFactor) // 80% entropy, 20% label count
	}
	
	return 0
}

// attributesToMap converts attributes to a string map.
func attributesToMap(attrs pcommon.Map) map[string]string {
	result := make(map[string]string, attrs.Len())
	
	attrs.Range(func(k string, v pcommon.Value) bool {
		result[k] = valueToString(v)
		return true
	})
	
	return result
}

// valueToString converts a pcommon.Value to a string.
func valueToString(v pcommon.Value) string {
	switch v.Type() {
	case pcommon.ValueTypeStr:
		return v.Str()
	case pcommon.ValueTypeInt:
		return string(v.Int())
	case pcommon.ValueTypeDouble:
		return string(v.Double())
	case pcommon.ValueTypeBool:
		return string(v.Bool())
	case pcommon.ValueTypeMap:
		// Simplified handling of maps for entropy calculation
		var parts []string
		v.Map().Range(func(k string, v pcommon.Value) bool {
			parts = append(parts, k+"="+valueToString(v))
			return true
		})
		return strings.Join(parts, ",")
	case pcommon.ValueTypeSlice:
		// Simplified handling of slices for entropy calculation
		var parts []string
		for i := 0; i < v.Slice().Len(); i++ {
			parts = append(parts, valueToString(v.Slice().At(i)))
		}
		return strings.Join(parts, ",")
	default:
		return ""
	}
}

// EntropyBasedCardinalityControl applies entropy-based cardinality control.
func EntropyBasedCardinalityControl(
	keySetTable map[string]keySetInfo,
	maxKeySets int,
) ([]string, []string) {
	// If we're under the limit, no need to drop/aggregate anything
	if len(keySetTable) <= maxKeySets {
		return nil, nil
	}
	
	// Calculate how many to drop
	toDrop := len(keySetTable) - maxKeySets
	
	// Convert map to slice for sorting
	keySets := make([]keySetEntry, 0, len(keySetTable))
	for key, info := range keySetTable {
		keySets = append(keySets, keySetEntry{
			key:         key,
			entropyScore: info.entropyScore,
			lastSeen:    info.lastSeen,
			accessCount: info.accessCount,
		})
	}
	
	// Sort by entropy score (lowest first - these will be dropped)
	sort.Slice(keySets, func(i, j int) bool {
		// Primary sort by entropy score
		if keySets[i].entropyScore != keySets[j].entropyScore {
			return keySets[i].entropyScore < keySets[j].entropyScore
		}
		
		// Secondary sort by access count
		if keySets[i].accessCount != keySets[j].accessCount {
			return keySets[i].accessCount < keySets[j].accessCount
		}
		
		// Tertiary sort by last seen (older first)
		return keySets[i].lastSeen < keySets[j].lastSeen
	})
	
	// Select the keys to drop and aggregate
	toDropKeys := make([]string, toDrop)
	toAggregateKeys := make([]string, 0, toDrop)
	
	// Take the first 'toDrop' entries for dropping or aggregation
	for i := 0; i < toDrop; i++ {
		toDropKeys[i] = keySets[i].key
		
		// If the entropy score is above a threshold, consider it for aggregation
		// instead of dropping completely
		if keySets[i].entropyScore > 0.3 { // Threshold for aggregation
			toAggregateKeys = append(toAggregateKeys, keySets[i].key)
		}
	}
	
	return toDropKeys, toAggregateKeys
}

// keySetEntry is used for sorting key-sets by entropy score.
type keySetEntry struct {
	key          string
	entropyScore float64
	lastSeen     int64
	accessCount  int64
}

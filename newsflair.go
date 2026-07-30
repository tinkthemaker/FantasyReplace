package main

import (
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
)

var (
	reFlairQuantity  = regexp.MustCompile(`(?i)(?:[$€£¥]\s*)?(?:\d[\d,.]*\s+)?(?:millions|billions|trillions|million|billion|trillion)\b(?:(?:\s+of)?(?:\s+(?:monthly|daily|annual|active|registered|paying)){0,2}\s+(?:dollars?|euros?|pounds?|yen|users?|people|customers?|subscribers?|viewers?|readers?|downloads?|installs?|views?|visits?|miles?|kilometers?|kilometres?))?`)
	reFlairSensitive = regexp.MustCompile(`(?i)\b(?:death|deaths|dead|died|killed|casualt(?:y|ies)|injur(?:y|ies|ed)|hospitali[sz]ed|victims?|fatalit(?:y|ies)|murder|suicide|assault|abuse|rape|crime|criminal|attack|shooting|war|battle|bomb|missile|invasion|genocide|disease|cancer|pandemic|epidemic|outbreak|diagnos(?:is|ed)|patients?|disaster|earthquake|flood|wildfire|hurricane|famine)\b`)
	reFlairSource    = regexp.MustCompile(`(?i)^\s*(?:sources?|source links?|attribution|original article|reported by)\s*:`)
	reFlairProtected = regexp.MustCompile("`+[^`\\n]*`+|\\[[^]\\n]+\\]\\([^\\n)]+\\)|https?://[^\\s)]+|\"[^\"\\n]*\"|“[^”\\n]*”")
)

var flairPhrases = map[string][][]string{
	"money": {
		{"a princely sum"},
		{"a hoard worthy of guarded vaults", "enough coin to trouble a royal treasurer"},
		{"a dragon-worthy hoard", "enough gold to make a kingdom take notice"},
	},
	"audience": {
		nil,
		{"a crowd fit to fill kingdoms", "a truly formidable multitude"},
		{"a host stretching beyond the horizon", "a multitude to make census-takers despair"},
	},
	"downloads": {
		nil,
		{"a digital host of remarkable size", "a torrent of small arrivals"},
		{"a digital host fit to overrun the castle gates", "a torrent vast enough to tax the kingdom's ledgers"},
	},
	"distance": {
		nil,
		{"a journey beyond many mapped realms", "a formidable span"},
		{"a voyage beyond the edges of a royal cartographer's map", "a span fit for the grandest quest"},
	},
	"generic": {
		nil,
		nil,
		{"a veritable multitude", "a quantity of distinctly legendary proportions"},
	},
}

// applyNewsFlair preserves the factual quantity and appends category-aware
// imagery. It intentionally skips structural, attributed, quoted, linked, and
// sensitive copy. Level 1 handles only explicit money; level 2 adds confident
// semantic categories; level 3 permits restrained generic quantity flair.
func applyNewsFlair(text string, seed int64, level int) string {
	if level < 1 || level > 3 {
		return text
	}
	lines := strings.Split(text, "\n")
	occurrence := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ">") ||
			reFlairSource.MatchString(line) || reFlairSensitive.MatchString(line) || hasUnbalancedQuotes(line) {
			continue
		}
		lines[i] = replaceOutsideProtected(line, func(segment string) string {
			return reFlairQuantity.ReplaceAllStringFunc(segment, func(quantity string) string {
				category := flairCategory(quantity, segment)
				phrases := flairPhrases[category][level-1]
				if len(phrases) == 0 {
					return quantity
				}
				phrase := phrases[stableFlairIndex(seed, category, quantity, occurrence, len(phrases))]
				occurrence++
				return quantity + "—" + phrase
			})
		})
	}
	return strings.Join(lines, "\n")
}

func replaceOutsideProtected(line string, transform func(string) string) string {
	locs := reFlairProtected.FindAllStringIndex(line, -1)
	if len(locs) == 0 {
		return transform(line)
	}
	var b strings.Builder
	cur := 0
	for _, loc := range locs {
		b.WriteString(transform(line[cur:loc[0]]))
		b.WriteString(line[loc[0]:loc[1]])
		cur = loc[1]
	}
	b.WriteString(transform(line[cur:]))
	return b.String()
}

func hasUnbalancedQuotes(line string) bool {
	return strings.Count(line, `"`)%2 != 0 || strings.Count(line, "“") != strings.Count(line, "”")
}

func flairCategory(quantity, context string) string {
	q := strings.ToLower(quantity)
	window := strings.ToLower(context)
	switch {
	case strings.ContainsAny(q, "$€£¥") || containsAny(q, "dollar", "euro", "pound", "yen") || containsAny(window, "revenue", "funding", "valuation", "cost", "sales"):
		return "money"
	case containsAny(q, "user", "people", "customer", "subscriber", "viewer", "reader") || containsAny(window, "monthly active users", "daily active users"):
		return "audience"
	case containsAny(q, "download", "install", "view", "visit"):
		return "downloads"
	case containsAny(q, "mile", "kilometer", "kilometre"):
		return "distance"
	default:
		return "generic"
	}
}

func containsAny(value string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func stableFlairIndex(seed int64, category, quantity string, occurrence, size int) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strconv.FormatInt(seed, 10)))
	_, _ = h.Write([]byte("|" + category + "|" + strings.ToLower(quantity) + "|" + strconv.Itoa(occurrence)))
	return int(h.Sum64() % uint64(size))
}

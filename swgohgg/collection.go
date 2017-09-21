package swgohgg

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (c *Client) Collection() (collection Collection, err error) {
	url := fmt.Sprintf("https://swgoh.gg/u/%s/collection/", c.profile)
	doc, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	doc.Find(".collection-char-list .collection-char").Each(func(i int, s *goquery.Selection) {
		char := parseChar(s)
		if !collection.Contains(char.Name) {
			collection = append(collection, char)
		}
	})
	sort.Sort(ByStars(collection, false))
	return collection, nil
}

type Collection []*Char

func (r Collection) Contains(char string) bool {
	for i := range r {
		if r[i].Name == char {
			return true
		}
	}
	return false
}

func (r Collection) ContainsAll(chars ...string) bool {
	for _, char := range chars {
		if !r.Contains(char) {
			return false
		}
	}
	return true
}

type Char struct {
	Name  string
	Stars int
	Level int
	Gear  int
}

func (c *Char) String() string {
	if c == nil {
		return "nil"
	}
	return fmt.Sprintf("%s %d* G%d Lvl%d", c.Name, c.Stars, c.Gear, c.Level)
}

func parseChar(s *goquery.Selection) *Char {
	var char Char
	char.Name = s.Find(".collection-char-name-link").Text()
	char.Level, _ = strconv.Atoi(s.Find(".char-portrait-full-level").Text())
	char.Gear = gearLevel(s)
	char.Stars = stars(s)
	return &char
}

func stars(s *goquery.Selection) int {
	level := 0
	s.Find(".star").Each(func(i int, star *goquery.Selection) {
		if star.HasClass("star-inactive") {
			return
		}
		level++
	})
	return level
}

func gearLevel(s *goquery.Selection) int {
	switch s.Find(".char-portrait-full-gear-level").Text() {
	case "XII":
		return 12
	case "XI":
		return 11
	case "X":
		return 10
	case "IX":
		return 9
	case "VIII":
		return 8
	case "VII":
		return 7
	case "VI":
		return 6
	case "V":
		return 5
	case "IV":
		return 4
	case "III":
		return 3
	case "II":
		return 2
	case "I":
		return 1
	default:
		return 0
	}
}

type CharacterStats struct {
	Name  string
	Level int64
	Stars int64

	// Current character gallactic power
	GalacticPower int64

	// List of skils of this character
	Skills []Skill

	// Basic Stats
	STR                int64
	AGI                int64
	INT                int64
	StrenghGrowth      float64
	AgilityGrowth      float64
	IntelligenceGrowth float64

	// General
	Health         int64
	Protection     int64
	Speed          int64
	CriticalDamage int64
	Potency        float64
	Tenacity       float64
	HealthSteal    int64
}

type Skill struct {
	Name  string
	Level int64
}

func (c *Client) CharacterStats(char string) (*CharacterStats, error) {
	charSlug := CharSlug(CharName(char))
	doc, err := c.Get(fmt.Sprintf("https://swgoh.gg/u/%s/collection/%s/", c.profile, charSlug))
	if err != nil {
		return nil, err
	}

	charStats := &CharacterStats{}
	charStats.Name = doc.Find(".pc-char-overview-name").Text()
	charStats.Level = atoi(doc.Find(".char-portrait-full-level").Text())
	charStats.Stars = int64(stars(doc.Find(".player-char-portrait")))
	charStats.GalacticPower = atoi(doc.Find(".unit-gp-stat-amount-current").First().Text())
	// Skills
	doc.Find(".pc-skills-list").First().Find(".pc-skill").Each(func(i int, s *goquery.Selection) {
		skill := Skill{}
		skill.Name = s.Find(".pc-skill-name").First().Text()
		skill.Level = skillLevel(s)
		charStats.Skills = append(charStats.Skills, skill)
	})
	//Stats
	doc.Find(".media-body .pc-stat").Each(func(i int, s *goquery.Selection) {
		name, value := s.Find(".pc-stat-label").Text(), s.Find(".pc-stat-value").Text()
		switch strings.TrimSpace(name) {
		case "Strength (STR)":
			charStats.STR = atoi(value)
		case "Agility (AGI)":
			charStats.AGI = atoi(value)
		case "Intelligence (INT)":
			charStats.INT = atoi(value)
		case "Strength Growth":
			charStats.StrenghGrowth = float64(atoi(value)) / 10
		case "Agility Growth":
			charStats.AgilityGrowth = float64(atoi(value)) / 10
		case "Intelligence Growth":
			charStats.IntelligenceGrowth = float64(atoi(value)) / 10
		case "Health":
			charStats.Health = atoi(value)
		case "Protection":
			charStats.Protection = atoi(value)
		case "Speed":
			charStats.Speed = atoi(value)
		case "Critical Damage":
			charStats.CriticalDamage = atoi(value)
		case "Potency":
			charStats.Potency = float64(atoi(value)) / 100.0
		case "Tenacity":
			charStats.Tenacity = float64(atoi(value)) / 100.0
		case "Health Steal":
			charStats.HealthSteal = atoi(value)
		}
	})
	return charStats, nil
}

func skillLevel(s *goquery.Selection) int64 {
	title := s.Find(".pc-skill-levels").First().AttrOr("data-title", "Level -1")
	// Title is in the form 'Level X of Y'
	fields := strings.Fields(title)
	if len(fields) >= 2 {
		return atoi(fields[1])
	}
	return -1
}

// atoi best-effort convertion to int, return 0 if unparseable
func atoi(src string) int64 {
	src = strings.Replace(src, ",", "", -1)
	src = strings.Replace(src, ".", "", -1)
	src = strings.Replace(src, "%", "", -1)
	v, _ := strconv.ParseInt(src, 10, 64)
	return v
}

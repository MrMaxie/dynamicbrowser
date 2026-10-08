package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	Default   string
	Autostart bool
	Browsers  Browsers
	Rules     []Rule
}

type Browser struct {
	Name string `yaml:"-"`
	Exe  string
	Args Arguments
}

type Arguments []string

func (a *Arguments) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return errors.New("args must be an array of strings")
	}
	*a = nil
	for _, item := range node.Content {
		if item.Kind == yaml.AliasNode {
			item = item.Alias
		}
		if item.Tag != "!!str" {
			return errors.New("args must contain only strings")
		}
		*a = append(*a, item.Value)
	}
	return nil
}

type Browsers []Browser

func (b *Browsers) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.New("browsers must be a mapping of names to browser configurations")
	}
	var values map[string]Browser
	if err := node.Decode(&values); err != nil {
		return err
	}
	*b = nil
	seen := make(map[string]bool)
	var appendBrowsers func(*yaml.Node) error
	appendBrowsers = func(node *yaml.Node) error {
		if node.Kind == yaml.AliasNode {
			return appendBrowsers(node.Alias)
		}
		if node.Kind == yaml.SequenceNode {
			for _, item := range node.Content {
				if err := appendBrowsers(item); err != nil {
					return err
				}
			}
			return nil
		}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.ShortTag() == "!!merge" {
				if err := appendBrowsers(node.Content[i+1]); err != nil {
					return err
				}
				continue
			}
			if key.ShortTag() != "!!str" {
				return errors.New("browser names must be strings")
			}
			var name string
			if err := key.Decode(&name); err != nil {
				return err
			}
			if !seen[name] {
				browser := values[name]
				browser.Name = name
				*b = append(*b, browser)
				seen[name] = true
			}
		}
		return nil
	}
	return appendBrowsers(node)
}

type Rule struct {
	Browser string
	Source  struct {
		Process Patterns
		Window  Patterns
	}
	Target struct {
		URL    Patterns
		Domain Patterns
		Path   Patterns
	}
}

type Patterns struct {
	present     bool
	expressions []*regexp.Regexp
}

func (p *Patterns) UnmarshalYAML(node *yaml.Node) error {
	*p = Patterns{present: true}
	var nodes []*yaml.Node
	switch node.Kind {
	case yaml.ScalarNode:
		nodes = []*yaml.Node{node}
	case yaml.SequenceNode:
		nodes = node.Content
	default:
		return errors.New("patterns must be a string or an array of strings")
	}
	for _, item := range nodes {
		if item.Kind == yaml.AliasNode {
			item = item.Alias
		}
		if item.Tag != "!!str" {
			return errors.New("patterns must contain only strings")
		}
		expression, err := compilePattern(item.Value)
		if err != nil {
			return fmt.Errorf("pattern %q: %w", item.Value, err)
		}
		p.expressions = append(p.expressions, expression)
	}
	return nil
}

func (p Patterns) Match(value string) bool {
	if !p.present {
		return true
	}
	for _, expression := range p.expressions {
		if expression.MatchString(value) {
			return true
		}
	}
	return false
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if expression, ok := strings.CutPrefix(pattern, "regex:"); ok {
		return regexp.Compile(expression)
	}
	var expression strings.Builder
	expression.WriteString("(?s)^")
	characters := []rune(pattern)
	for i := 0; i < len(characters); i++ {
		switch characters[i] {
		case '*':
			expression.WriteString(".*")
		case '?':
			expression.WriteByte('.')
		case '\\':
			i++
			if i == len(characters) {
				return nil, errors.New("incomplete glob escape")
			}
			expression.WriteString(regexp.QuoteMeta(string(characters[i])))
		default:
			expression.WriteString(regexp.QuoteMeta(string(characters[i])))
		}
	}
	expression.WriteByte('$')
	return regexp.Compile(expression.String())
}

func decode(data []byte) (*Config, error) {
	var configuration Config
	if err := yaml.Unmarshal(data, &configuration); err != nil {
		return nil, err
	}
	return &configuration, nil
}

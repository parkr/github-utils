package langclassify

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/parkr/github-utils/gh"
	"github.com/parkr/github-utils/repos"
	yaml "gopkg.in/yaml.v2"
)

var dockerfileFromRegexp = regexp.MustCompile(`FROM (\S+):(\S+)`)

type Language interface {
	GetName() string
	GetVersion(*gh.Client) string
}

// ClassifyRepository takes a repository and returns language and language version information.
func ClassifyRepository(client *gh.Client, repo *repos.Repository) (Language, error) {
	lang, version, err := LanguageAndVersionFromDockerfile(client, repo)
	if err != nil {
		lang, version, err = LanguageAndVersionFromTravisConfiguration(client, repo)
	}

	switch lang {
	case "Go", "go", "golang":
		return &Go{version: version}, nil
	case "Ruby", "ruby":
		return &Ruby{version: version}, nil
	case "node_js", "Node", "node", "JavaScript", "js", "javascript":
		return &JavaScript{version: version}, nil
	case "python", "Python":
		return &Python{version: version}, nil
	}

	return nil, fmt.Errorf("language %q is not supported", lang)
}

// LanguageAndVersionFromDockerfile fetches the Dockerfile from a repo and parses the language and version.
func LanguageAndVersionFromDockerfile(client *gh.Client, repo *repos.Repository) (string, string, error) {
	content, err := client.GetFileContents(repo.GetOwner().GetLogin(), repo.GetName(), "Dockerfile")
	if err != nil {
		return repo.GetLanguage(), "", err
	}
	if len(content) == 0 {
		return repo.GetLanguage(), "", errors.New("Dockerfile is empty")
	}

	matches := dockerfileFromRegexp.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 && len(matches[0]) >= 3 {
		return matches[0][1], matches[0][2], nil
	}

	return repo.GetLanguage(), "", errors.New("Dockerfile doesn't have a FROM declaration with a language")
}

type travisLanguageConfig struct {
	Language string `yaml:"language"`

	RubyVersions   []string `yaml:"rvm"`
	PythonVersions []string `yaml:"python"`
	GoVersions     []string `yaml:"go"`
	NodeJSVersions []string `yaml:"node_js"`
}

// LanguageAndVersionFromTravisConfiguration fetches the .travis.yml file from a repo and parses the language and version.
func LanguageAndVersionFromTravisConfiguration(client *gh.Client, repo *repos.Repository) (string, string, error) {
	content, err := client.GetFileContents(repo.GetOwner().GetLogin(), repo.GetName(), ".travis.yml")
	if err != nil {
		return repo.GetLanguage(), "", err
	}
	if len(content) == 0 {
		return repo.GetLanguage(), "", errors.New(".travis.yml is empty")
	}

	cfg := &travisLanguageConfig{}
	err = yaml.NewDecoder(strings.NewReader(content)).Decode(cfg)
	if err == nil && len(cfg.Language) != 0 {
		switch cfg.Language {
		case "go":
			if len(cfg.GoVersions) > 0 {
				return "Go", cfg.GoVersions[len(cfg.GoVersions)-1], nil
			}
		case "", "ruby":
			if len(cfg.RubyVersions) > 0 {
				return "Ruby", cfg.RubyVersions[len(cfg.RubyVersions)-1], nil
			}
		case "python":
			if len(cfg.PythonVersions) > 0 {
				return "Python", cfg.PythonVersions[len(cfg.PythonVersions)-1], nil
			}
		case "node_js":
			if len(cfg.NodeJSVersions) > 0 {
				return "JavaScript", cfg.NodeJSVersions[len(cfg.NodeJSVersions)-1], nil
			}
		default:
			log.Printf("repo(%s): .travis.yml:\n%s\n", repo.GetFullName(), content)
		}
		return cfg.Language, "", nil
	}

	return repo.GetLanguage(), "", nil
}

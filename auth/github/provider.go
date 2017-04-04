package github

import (
	"errors"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"fmt"
	"strings"

	"encoding/json"

	"github.com/concourse/atc"
	"github.com/concourse/atc/auth"
	"github.com/concourse/atc/auth/provider"
	"github.com/concourse/atc/auth/verifier"
	"github.com/hashicorp/go-multierror"
	flags "github.com/jessevdk/go-flags"
	"github.com/tedsuo/rata"
)

const ProviderName = "github"
const DisplayName = "GitHub"

var Scopes = []string{"read:org"}

type GitHubAuthConfig struct {
	ClientID      string             `json:"client_id"     long:"client-id"     description:"Application client ID for enabling GitHub OAuth."`
	ClientSecret  string             `json:"client_secret" long:"client-secret" description:"Application client secret for enabling GitHub OAuth."`
	Organizations []string           `json:"organizations" long:"organization"  description:"GitHub organization whose members will have access." value-name:"ORG"`
	Teams         []GitHubTeamConfig `json:"teams"         long:"team"          description:"GitHub team whose members will have access." value-name:"ORG/TEAM"`
	Users         []string           `json:"users"         long:"user"          description:"GitHub user to permit access." value-name:"LOGIN"`
	AuthURL       string             `json:"auth_url"      long:"auth-url"      description:"Override default endpoint AuthURL for Github Enterprise."`
	TokenURL      string             `json:"token_url"     long:"token-url"     description:"Override default endpoint TokenURL for Github Enterprise."`
	APIURL        string             `json:"api_url"       long:"api-url"       description:"Override default API endpoint URL for Github Enterprise."`
}

func (*GitHubAuthConfig) AuthMethod(oauthBaseURL string) atc.AuthMethod {
	path, err := auth.OAuthRoutes.CreatePathForRoute(
		auth.OAuthBegin,
		rata.Params{"provider": ProviderName},
	)
	if err != nil {
		panic("failed to construct oauth begin handler route: " + err.Error())
	}

	return atc.AuthMethod{
		Type:        atc.AuthTypeOAuth,
		DisplayName: DisplayName,
		AuthURL:     oauthBaseURL + path,
	}
}

func (auth *GitHubAuthConfig) IsConfigured() bool {
	return auth.ClientID != "" ||
		auth.ClientSecret != "" ||
		len(auth.Organizations) > 0 ||
		len(auth.Teams) > 0 ||
		len(auth.Users) > 0
}

func (auth *GitHubAuthConfig) Validate() error {
	var errs *multierror.Error
	if auth.ClientID == "" || auth.ClientSecret == "" {
		errs = multierror.Append(
			errs,
			errors.New("must specify --github-auth-client-id and --github-auth-client-secret to use GitHub OAuth."),
		)
	}
	if len(auth.Organizations) == 0 && len(auth.Teams) == 0 && len(auth.Users) == 0 {
		errs = multierror.Append(
			errs,
			errors.New("at least one of the following is required for github-auth: organizations, teams, users."),
		)
	}
	return errs.ErrorOrNil()
}

type GitHubTeamConfig struct {
	OrganizationName string
	TeamName         string
}

func (flag *GitHubTeamConfig) UnmarshalFlag(value string) error {
	s := strings.SplitN(value, "/", 2)
	if len(s) != 2 {
		return fmt.Errorf("malformed GitHub team specification: '%s'", value)
	}

	flag.OrganizationName = s[0]
	flag.TeamName = s[1]

	return nil
}

type GitHubProvider struct {
	*oauth2.Config
	verifier.Verifier
}

func init() {
	provider.Register(ProviderName, GitHubTeamProvider{})
}

type GitHubTeamProvider struct {
}

func (GitHubTeamProvider) AddAuthGroup(parser *flags.Parser) provider.AuthConfig {
	flags := &GitHubAuthConfig{}

	group, err := parser.Group.AddGroup("Github Auth", "Github Authentication", flags)
	if err != nil {
		panic(err)
	}

	group.Namespace = "github-auth"

	return flags
}

func (GitHubTeamProvider) UnmarshalConfig(config *json.RawMessage) (provider.AuthConfig, error) {
	flags := &GitHubAuthConfig{}
	if config != nil {
		err := json.Unmarshal(*config, &flags)
		if err != nil {
			return nil, err
		}
	}
	return flags, nil
}

func (GitHubTeamProvider) ProviderConstructor(
	config provider.AuthConfig,
	redirectURL string,
) (provider.Provider, bool) {
	githubAuth := config.(*GitHubAuthConfig)

	client := NewClient(githubAuth.APIURL)

	endpoint := github.Endpoint
	if githubAuth.AuthURL != "" && githubAuth.TokenURL != "" {
		endpoint.AuthURL = githubAuth.AuthURL
		endpoint.TokenURL = githubAuth.TokenURL
	}

	return GitHubProvider{
		Verifier: verifier.NewVerifierBasket(
			NewTeamVerifier(teamConfigsToTeam(githubAuth.Teams), client),
			NewOrganizationVerifier(githubAuth.Organizations, client),
			NewUserVerifier(githubAuth.Users, client),
		),
		Config: &oauth2.Config{
			ClientID:     githubAuth.ClientID,
			ClientSecret: githubAuth.ClientSecret,
			Endpoint:     endpoint,
			Scopes:       Scopes,
			RedirectURL:  redirectURL,
		},
	}, true
}

func (GitHubProvider) PreTokenClient() (*http.Client, error) {
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}, nil
}

func teamConfigsToTeam(dbteams []GitHubTeamConfig) []Team {
	teams := []Team{}
	for _, team := range dbteams {
		teams = append(teams, Team{
			Name:         team.TeamName,
			Organization: team.OrganizationName,
		})
	}
	return teams
}

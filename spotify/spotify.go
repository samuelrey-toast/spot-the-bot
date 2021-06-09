package spotify

import (
	"log"

	"github.com/pkg/errors"
	"github.com/samuelrey/spot-discord/framework"
	"github.com/zmb3/spotify"
)

var (
	authURL              string
	config               *Config
	spotifyAuthenticator spotify.Authenticator
	tknHandler           *TokenHandler
)

const (
	StrAuthMessageFmt = "Click this link so we can create your playlist on Spotify:\n%s"
)

func init() {
	config = LoadConfig("secrets_spotify.json")
	spotifyAuthenticator = spotify.NewAuthenticator(
		config.RedirectURL, spotify.ScopePlaylistModifyPublic)
	spotifyAuthenticator.SetAuthInfo(config.ClientID, config.Secret)
	authURL = spotifyAuthenticator.AuthURL(config.State)
	tknHandler = NewTokenHandler()
}

type SpotifyBuilder struct {
	authenticator *spotify.Authenticator
	authURL       string
	client        *spotify.Client
	user          *spotify.User
}

func CreateSpotifyBuilder(config *Config) *SpotifyBuilder {
	spotifyAuthenticator := spotify.NewAuthenticator(
		config.RedirectURL,
		spotify.ScopePlaylistModifyPrivate,
	)
	spotifyAuthenticator.SetAuthInfo(config.ClientID, config.Secret)

	authURL = spotifyAuthenticator.AuthURL(config.State)

	return &SpotifyBuilder{
		authenticator: &spotifyAuthenticator,
		authURL:       authURL,
	}
}

// AuthorizeUser takes a user through the Spotify authorization flow. The
// token we receive from Spotify is used to create a reusable client. The
// client is tied to a single Spotify user. We recommend creating an account
// on Spotify specific for this bot.
func (sb *SpotifyBuilder) AuthorizeUser() error {
	log.Printf("Navigate here to authorize Spotify user: %s\n", authURL)

	token, err := getToken()
	if err != nil {
		return errors.Wrap(err, "Authorize Spotify user")
	}

	client := spotifyAuthenticator.NewClient(token)
	spotifyUser, err := client.CurrentUser()
	if err != nil {
		return errors.Wrap(err, "Authorize Spotify user")
	}

	sb.client = &client
	sb.user = &spotifyUser.User
	return nil
}

// CreatePlaylist creates a playlist with the given name for the authorized
// user.
func (sb *SpotifyBuilder) CreatePlaylist(playlistName string) (*framework.Playlist, error) {
	// TODO make the playlist collaborative
	playlist, err := sb.client.CreatePlaylistForUser(sb.user.ID, playlistName, "", false)
	if err != nil {
		return nil, errors.Wrap(err, "Create Spotify playlist")
	}

	return &framework.Playlist{
		ID:  playlist.ID.String(),
		URL: playlist.ExternalURLs["spotify"],
	}, nil
}

// NewClient provides an interface to perform actions in Spotify on behalf
// of an authenticated user. Authenticate the user if they haven't already.
func NewClient(
	msgUserID string,
	sendAuthURL func(string),
) (*SpotifyPlaylister, error) {
	token, found := tknHandler.Get(msgUserID)
	if !found {
		sendAuthURL(authURL)

		var err error
		token, err = getToken()
		if err != nil {
			return nil, errors.Wrap(err, "Authenticate music service")
		}

		tknHandler.Register(msgUserID, token)
	}

	client := spotifyAuthenticator.NewClient(token)
	return &SpotifyPlaylister{client: &client}, nil
}

type SpotifyPlaylister struct {
	client *spotify.Client
}

func (s SpotifyPlaylister) CreatePlaylist(
	userID string,
	playlistName string,
) (*framework.Playlist, error) {
	playlist, err := s.client.CreatePlaylistForUser(
		userID, playlistName, "", true)
	if err != nil {
		return nil, err
	}

	return &framework.Playlist{
		ID:  playlist.ID.String(),
		URL: playlist.ExternalURLs["spotify"],
	}, nil
}

func (s SpotifyPlaylister) CurrentUserID() (*string, error) {
	user, err := s.client.CurrentUser()
	if err != nil {
		return nil, err
	}

	return &user.ID, nil
}

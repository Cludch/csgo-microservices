package demoparser

import (
	"errors"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/Cludch/csgo-microservices/demoparser/internal/config"
	"github.com/Cludch/csgo-microservices/demoparser/pkg/files"
	shared_config "github.com/Cludch/csgo-microservices/shared/pkg/config"
	"github.com/Cludch/csgo-microservices/shared/pkg/entity"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
	log "github.com/sirupsen/logrus"
)

// DemoParser holds the instance of one demo consisting of the file handle and the parsed data.
type ParserService struct {
	config        *shared_config.GlobalConfig
	parser        demoinfocs.Parser
	Match         *MatchData
	CurrentRound  byte
	RoundStart    time.Duration
	RoundOngoing  bool
	SidesSwitched bool
	GameOver      bool
}

func NewService(c *config.ConfigService) *ParserService {
	return &ParserService{
		config: c.GetConfig().Global,
	}
}

// MatchData holds information about the match itself.
type MatchData struct {
	ID       entity.ID
	Map      string
	Header   *common.DemoHeader
	Players  []*Player
	Teams    [2]*Team
	Duration time.Duration
	Time     time.Time
	Rounds   []*Round
}

// Team represents a team and links to it's players.
type Team struct {
	StartedAs common.Team
	State     *common.TeamState
	Players   []*Player
}

// Player represents one player either as T or CT.
type Player struct {
	SteamID  uint64
	Name     string
	Team     *Team
	WinCount int
	RankOld  int
	RankNew  int
}

// Round contains information about one round.
type Round struct {
	Duration  time.Duration
	Kills     []*Kill
	Damage    []*Damage
	Winner    *Team
	WinReason events.RoundEndReason
	MVP       *Player
}

// Kill holds information about a kill that happenend during the match.
type Kill struct {
	Tick            time.Duration
	Victim          *Player
	Killer          *Player
	Assister        *Player
	Weapon          common.EquipmentType
	IsDuringRound   bool
	IsHeadshot      bool
	IsFlashAssist   bool
	IsAttackerBlind bool
	IsNoScope       bool
	IsThroughSmoke  bool
	IsThroughWall   bool
}

type Damage struct {
	Attacker          *Player
	HealthDamageTaken int
}

// Parse takes a demo file and starts parsing by registering all required event handlers.
func (s *ParserService) Parse(dir string, demoFile *files.Demo) error {
	s.Match = &MatchData{ID: demoFile.ID, Time: demoFile.MatchTime}

	f, err := os.Open(path.Join(dir, demoFile.Filename))

	if err != nil {
		return err
	}

	const msg = "Starting demo parsing of match %s (file %s)"
	log.Infof(msg, s.Match.ID, demoFile.Filename)

	s.parser = demoinfocs.NewParser(f)
	defer s.parser.Close()
	defer f.Close()

	// Parsing the header within an event handler crashes.
	header, _ := s.parser.ParseHeader()
	s.Match.Header = &header

	// Register all handler.
	s.parser.RegisterEventHandler(s.handleMatchStart)
	s.parser.RegisterEventHandler(s.handleGamePhaseChanged)
	s.parser.RegisterEventHandler(s.handleKill)
	s.parser.RegisterEventHandler(s.handlePlayerHurt)
	s.parser.RegisterEventHandler(s.handleMVP)
	s.parser.RegisterEventHandler(s.handleRoundStart)
	s.parser.RegisterEventHandler(s.handleRoundEnd)
	s.parser.RegisterEventHandler(s.handleRankUpdate)
	s.parser.RegisterEventHandler(s.handleParserWarn)

	return s.parser.ParseToEnd()
}

// AddPlayer adds a player to the game and returns the pointer.
func (s *ParserService) AddPlayer(player *common.Player) *Player {
	teamID := GetTeamIndex(player.Team, s.SidesSwitched)
	teams := s.Match.Teams
	teamPlayers := teams[teamID].Players

	customPlayer := &Player{SteamID: player.SteamID64, Name: player.Name, Team: teams[teamID], RankOld: -1, RankNew: -1}

	teams[teamID].Players = append(teamPlayers, customPlayer)
	s.Match.Players = append(s.Match.Players, customPlayer)

	return customPlayer
}

func (s *ParserService) getPlayer(player *common.Player) (*Player, error) {
	if player.IsBot {
		return nil, errors.New("Player is a bot")
	}

	for _, localPlayer := range s.Match.Players {
		if player.SteamID64 == localPlayer.SteamID {
			return localPlayer, nil
		}
	}

	for _, gamePlayer := range s.parser.GameState().Participants().Playing() {
		if player.SteamID64 == gamePlayer.SteamID64 {
			return s.AddPlayer(player), nil
		}
	}

	return nil, errors.New("Player not found in local match struct " + strconv.FormatUint(player.SteamID64, 10))
}

func (s *ParserService) debug(message string) {
	if s.config.IsTrace() {
		log.WithFields(log.Fields{
			"Match": s.Match.ID,
			"Round": s.CurrentRound,
		}).Trace(message)
	} else {
		log.Debug(message)
	}
}
